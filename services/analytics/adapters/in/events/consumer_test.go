package events

// SCRUM-160: analytics reads report.submitted and status.changed, so every new
// report and every status change reaches the admin and President dashboards.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/analytics/application"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

const reportID = "33333333-3333-4333-8333-333333333333"

var sentAt = time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)

func setup(wait time.Duration) (*Consumer, *application.Service) {
	app := application.NewService(memory.New())
	return NewConsumer(app, wait), app
}

func submittedEvent(eventID string, category commonv1.Category) *eventsv1.ReportSubmitted {
	return &eventsv1.ReportSubmitted{
		Meta:   &eventsv1.EventMeta{EventId: eventID, OccurredAt: timestamppb.New(sentAt), Producer: "report-submission"},
		Report: &commonv1.Report{Id: reportID, Category: category, CreatedAt: timestamppb.New(sentAt)},
	}
}

func changedEvent(eventID string, to commonv1.Status) *eventsv1.StatusChanged {
	return &eventsv1.StatusChanged{
		Meta:     &eventsv1.EventMeta{EventId: eventID, OccurredAt: timestamppb.New(sentAt.Add(time.Hour)), Producer: "status-lifecycle"},
		ReportId: reportID,
		ToStatus: to,
	}
}

func submitted(t *testing.T, eventID string, category commonv1.Category) []byte {
	t.Helper()
	b, err := proto.Marshal(submittedEvent(eventID, category))
	require.NoError(t, err)
	return b
}

func changed(t *testing.T, eventID string, to commonv1.Status) []byte {
	t.Helper()
	b, err := proto.Marshal(changedEvent(eventID, to))
	require.NoError(t, err)
	return b
}

func summary(t *testing.T, app *application.Service) domain.Summary {
	t.Helper()
	s, err := app.Summary(context.Background(), domain.Period{})
	require.NoError(t, err)
	return s
}

func TestStreamsAndGroup(t *testing.T) {
	assert.Equal(t, []string{eventbus.StreamReportSubmitted, eventbus.StreamStatusChanged}, Streams())
	assert.Equal(t, "analytics", Group)
}

func TestANewReportIsCounted(t *testing.T) {
	c, app := setup(time.Second)
	require.NoError(t, c.ReportSubmitted(context.Background(), submitted(t, "e1", commonv1.Category_HOSTEL)))

	s := summary(t, app)
	assert.EqualValues(t, 1, s.Counts.Total)
	assert.EqualValues(t, 1, s.Counts.Open)

	rows, err := app.Breakdown(context.Background(), domain.Period{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, rows[0].Counts.Total, "counted under Hostel")
}

func TestAStatusChangeMovesTheCount(t *testing.T) {
	c, app := setup(time.Second)
	ctx := context.Background()
	require.NoError(t, c.ReportSubmitted(ctx, submitted(t, "e1", commonv1.Category_HOSTEL)))
	require.NoError(t, c.StatusChanged(ctx, changed(t, "e2", commonv1.Status_IN_PROGRESS)))

	s := summary(t, app)
	assert.EqualValues(t, 0, s.Counts.Open)
	assert.EqualValues(t, 1, s.Counts.InProgress)

	require.NoError(t, c.StatusChanged(ctx, changed(t, "e3", commonv1.Status_RESOLVED)))
	assert.EqualValues(t, 1, summary(t, app).Counts.Resolved)
}

func TestAnEventTwiceIsCountedOnce(t *testing.T) {
	c, app := setup(time.Second)
	ctx := context.Background()
	require.NoError(t, c.ReportSubmitted(ctx, submitted(t, "e1", commonv1.Category_HOSTEL)))
	require.NoError(t, c.ReportSubmitted(ctx, submitted(t, "e1", commonv1.Category_HOSTEL)))
	assert.EqualValues(t, 1, summary(t, app).Counts.Total)
}

// The two streams are read at the same time, so after a restart a status
// change can come before its report. It waits for the report instead of being
// lost.
func TestAStatusChangeBeforeItsReportWaitsForIt(t *testing.T) {
	c, app := setup(2 * time.Second)
	ctx := context.Background()

	done := make(chan error, 1)
	go func() { done <- c.StatusChanged(ctx, changed(t, "e2", commonv1.Status_ASSIGNED)) }()
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, c.ReportSubmitted(ctx, submitted(t, "e1", commonv1.Category_ACADEMIC_BLOCK)))

	require.NoError(t, <-done)
	s := summary(t, app)
	assert.EqualValues(t, 1, s.Counts.Assigned)
	assert.EqualValues(t, 0, s.Counts.Open)
}

// A report that never arrives: the change is returned as an error, so the bus
// tries again and finally keeps it in the dead-letter stream, never lost quietly.
func TestAStatusChangeForAReportThatNeverArrivesFails(t *testing.T) {
	c, app := setup(50 * time.Millisecond)
	err := c.StatusChanged(context.Background(), changed(t, "e2", commonv1.Status_ASSIGNED))
	assert.ErrorIs(t, err, ErrUnknownReport)
	assert.EqualValues(t, 0, summary(t, app).Counts.Total)
}

func TestABrokenEventIsAnError(t *testing.T) {
	c, _ := setup(time.Second)
	assert.Error(t, c.ReportSubmitted(context.Background(), []byte("not a protobuf")))
	assert.Error(t, c.StatusChanged(context.Background(), []byte("not a protobuf")))
}

// Run reads both streams from a bus until it is stopped.
func TestRunReadsBothStreams(t *testing.T) {
	c, app := setup(2 * time.Second)
	bus := eventbus.NewMemory()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Run(ctx, bus, "test") }()

	require.NoError(t, bus.Publish(ctx, eventbus.StreamReportSubmitted, submittedEvent("e1", commonv1.Category_OTHERS)))
	require.NoError(t, bus.Publish(ctx, eventbus.StreamStatusChanged, changedEvent("e2", commonv1.Status_ASSIGNED)))
	assert.Eventually(t, func() bool { return summary(t, app).Counts.Assigned == 1 }, 3*time.Second, 20*time.Millisecond)

	cancel()
	assert.NoError(t, <-done)
}
