package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/report-history/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-history/application"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

const (
	reportA  = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	studentA = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
	dsa      = "4e5f6071-8293-4d0e-9f2a-3b4c5d6e7f80"
	hod      = "6071a2b3-c4d5-4e6f-9a0b-1c2d3e4f5a6b"
)

var t0 = time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)

func eventID(n int) string { return fmt.Sprintf("0f000000-0000-4000-8000-%012d", n) }

func meta(n int, at time.Time) *eventsv1.EventMeta {
	return &eventsv1.EventMeta{EventId: eventID(n), OccurredAt: timestamppb.New(at), Producer: "test", CorrelationId: reportA}
}

func submittedEvent(n int) *eventsv1.ReportSubmitted {
	return &eventsv1.ReportSubmitted{Meta: meta(n, t0), Report: &commonv1.Report{
		Id: reportA, ReporterId: studentA, Category: commonv1.Category_HOSTEL,
		Location: &commonv1.Location{Building: "Hostel B", Wing: "East"}, Status: commonv1.Status_OPEN, CreatedAt: timestamppb.New(t0),
	}}
}

func routedEvent(n int) *eventsv1.ReportRouted {
	return &eventsv1.ReportRouted{Meta: meta(n, t0.Add(time.Minute)), ReportId: reportA, Category: commonv1.Category_HOSTEL,
		CategoryHeadId: dsa, TargetPosition: "DSA", RuleId: "rule-hostel"}
}

func statusEvent(n int, from, to commonv1.Status, at time.Time) *eventsv1.StatusChanged {
	actor := dsa
	return &eventsv1.StatusChanged{Meta: meta(n, at), ReportId: reportA, FromStatus: from, ToStatus: to, ActorId: &actor, Version: 2}
}

func reassignedEvent(n int) *eventsv1.ReportReassigned {
	actor := dsa
	return &eventsv1.ReportReassigned{Meta: meta(n, t0.Add(time.Hour)), ReportId: reportA, FromUserId: dsa, ToUserId: hod,
		Reason: "Lab equipment", ActorId: &actor}
}

func payload(t *testing.T, m proto.Message) []byte {
	t.Helper()
	b, err := proto.Marshal(m)
	require.NoError(t, err)
	return b
}

func newApp() *application.Service {
	n := 0
	return application.NewService(memory.NewRepository(), func() string { n++; return fmt.Sprintf("0a000000-0000-4000-8000-%012d", n) })
}

// run starts the consumer on an in-memory bus and stops it when the test ends.
func run(t *testing.T, c *Consumer, bus *eventbus.Memory) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Run(ctx, bus, "test-consumer") }()
	t.Cleanup(func() {
		cancel()
		assert.NoError(t, <-done)
	})
}

func TestTheFourStreamsBuildTheReadModel(t *testing.T) {
	ctx := context.Background()
	app := newApp()
	bus := eventbus.NewMemory()
	run(t, NewConsumer(app), bus)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportSubmitted, submittedEvent(1)))
	require.Eventually(t, func() bool {
		r, _, _ := app.ListReportsByReporter(ctx, studentA, 0, "")
		return len(r) == 1
	}, 5*time.Second, 20*time.Millisecond)

	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportRouted, routedEvent(2)))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamStatusChanged, statusEvent(3, commonv1.Status_OPEN, commonv1.Status_ASSIGNED, t0.Add(2*time.Minute))))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportReassigned, reassignedEvent(4)))

	require.Eventually(t, func() bool {
		tl, _ := app.GetReportTimeline(ctx, reportA)
		return len(tl) == 4
	}, 5*time.Second, 20*time.Millisecond)

	tl, err := app.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	assert.Equal(t, []string{domain.EventReportSubmitted, domain.EventReportRouted, domain.EventStatusChanged, domain.EventReportReassigned},
		[]string{tl[0].EventType, tl[1].EventType, tl[2].EventType, tl[3].EventType})
	assert.Equal(t, "Routed to DSA", tl[1].Summary)
	assert.Equal(t, "Status changed to ASSIGNED", tl[2].Summary)
	assert.Equal(t, "Reassigned: Lab equipment", tl[3].Summary)
	assert.Equal(t, eventID(2), tl[1].ID, "the event id is the entry id")

	mine, _, err := app.ListReportsByReporter(ctx, studentA, 0, "")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusAssigned, mine[0].Status)
	assert.Equal(t, hod, mine[0].AssigneeID, "the reassignment wins")

	forDSA, _, err := app.ListReportsByAssignee(ctx, dsa, 0, "")
	require.NoError(t, err)
	assert.Empty(t, forDSA)
	forHoD, _, err := app.ListReportsByAssignee(ctx, hod, 0, "")
	require.NoError(t, err)
	assert.Len(t, forHoD, 1)
}

func TestARepeatedEventIsIgnored(t *testing.T) {
	ctx := context.Background()
	app := newApp()
	c := NewConsumer(app)

	require.NoError(t, c.ReportSubmitted(ctx, payload(t, submittedEvent(1))))
	require.NoError(t, c.ReportSubmitted(ctx, payload(t, submittedEvent(1))), "same report.submitted again")
	require.NoError(t, c.ReportRouted(ctx, payload(t, routedEvent(2))))
	require.NoError(t, c.ReportRouted(ctx, payload(t, routedEvent(2))), "same report.routed again")

	tl, err := app.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	assert.Len(t, tl, 2)
}

func TestAnEventBeforeItsReportWaitsForTheReport(t *testing.T) {
	ctx := context.Background()
	app := newApp()
	c := NewConsumer(app)
	c.WaitForReport = 2 * time.Second
	c.PollEvery = 10 * time.Millisecond

	result := make(chan error, 1)
	go func() { result <- c.ReportRouted(ctx, payload(t, routedEvent(2))) }()
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, c.ReportSubmitted(ctx, payload(t, submittedEvent(1))))

	require.NoError(t, <-result)
	forDSA, _, err := app.ListReportsByAssignee(ctx, dsa, 0, "")
	require.NoError(t, err)
	assert.Len(t, forDSA, 1)
}

func TestAnEventForAReportThatNeverArrivesFails(t *testing.T) {
	c := NewConsumer(newApp())
	c.WaitForReport = 50 * time.Millisecond
	c.PollEvery = 10 * time.Millisecond

	err := c.StatusChanged(context.Background(), payload(t, statusEvent(3, commonv1.Status_OPEN, commonv1.Status_ASSIGNED, t0)))
	assert.ErrorIs(t, err, domain.ErrReportNotFound)
}

func TestBadMessagesAreErrors(t *testing.T) {
	ctx := context.Background()
	c := NewConsumer(newApp())

	assert.Error(t, c.ReportSubmitted(ctx, []byte("not protobuf")))
	assert.Error(t, c.ReportRouted(ctx, []byte("not protobuf")))
	assert.Error(t, c.StatusChanged(ctx, []byte("not protobuf")))
	assert.Error(t, c.ReportReassigned(ctx, []byte("not protobuf")))

	noCategory := submittedEvent(1)
	noCategory.Report.Category = commonv1.Category_CATEGORY_UNSPECIFIED
	assert.ErrorIs(t, c.ReportSubmitted(ctx, payload(t, noCategory)), domain.ErrInvalidCategory)

	noEventID := routedEvent(2)
	noEventID.Meta.EventId = ""
	assert.ErrorIs(t, c.ReportRouted(ctx, payload(t, noEventID)), domain.ErrInvalidEntryID)
}

func TestReassignedWithoutReasonAndSystemStatusChange(t *testing.T) {
	ctx := context.Background()
	app := newApp()
	c := NewConsumer(app)
	require.NoError(t, c.ReportSubmitted(ctx, payload(t, submittedEvent(1))))

	r := reassignedEvent(4)
	r.Reason = ""
	r.ActorId = nil
	require.NoError(t, c.ReportReassigned(ctx, payload(t, r)))

	s := statusEvent(5, commonv1.Status_ASSIGNED, commonv1.Status_ESCALATED, t0.Add(2*time.Hour))
	s.ActorId = nil
	require.NoError(t, c.StatusChanged(ctx, payload(t, s)))

	tl, err := app.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	require.Len(t, tl, 3)
	assert.Equal(t, "Reassigned", tl[1].Summary)
	assert.Empty(t, tl[2].ActorID, "the system escalated it")
	assert.Equal(t, domain.StatusEscalated, tl[2].ToStatus)
}

func TestStreamsAndGroup(t *testing.T) {
	assert.Equal(t, "report-history", Group)
	assert.ElementsMatch(t, []string{
		eventbus.StreamReportSubmitted, eventbus.StreamReportRouted, eventbus.StreamStatusChanged, eventbus.StreamReportReassigned,
	}, Streams())
}
