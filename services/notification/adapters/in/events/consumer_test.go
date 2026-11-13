package events

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/notification/application"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

const (
	student  = "22222222-2222-4222-8222-222222222222"
	reportID = "33333333-3333-4333-8333-333333333333"
	event1   = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	event2   = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
)

var place = &commonv1.Location{Building: "HC", Wing: "Floor 1", Area: "Left wing bathroom"}

// okSender delivers everything.
type okSender struct{ calls int }

func (s *okSender) Send(context.Context, domain.Notification) error { s.calls++; return nil }

// fakeReports answers GetReport like report-submission, or fails.
type fakeReports struct {
	calls int
	err   error
}

func (f *fakeReports) GetReport(_ context.Context, id string) (ports.Report, error) {
	f.calls++
	if f.err != nil {
		return ports.Report{}, f.err
	}
	return ports.Report{ID: id, ReporterID: student, Building: "HC", Wing: "Floor 1", Area: "Left wing bathroom"}, nil
}

func setup() (*Consumer, *application.Service, *fakeReports, *okSender) {
	sender := &okSender{}
	app := application.NewService(memory.New(), sender)
	reports := &fakeReports{}
	return NewConsumer(app, reports), app, reports, sender
}

func submitted(t *testing.T, eventID string) []byte {
	t.Helper()
	b, err := proto.Marshal(&eventsv1.ReportSubmitted{
		Meta:   &eventsv1.EventMeta{EventId: eventID, Producer: "report-submission"},
		Report: &commonv1.Report{Id: reportID, ReporterId: student, Category: commonv1.Category_HOSTEL, Location: place},
	})
	require.NoError(t, err)
	return b
}

func changed(t *testing.T, eventID string, from, to commonv1.Status) []byte {
	t.Helper()
	b, err := proto.Marshal(&eventsv1.StatusChanged{
		Meta:     &eventsv1.EventMeta{EventId: eventID, Producer: "status-lifecycle"},
		ReportId: reportID, FromStatus: from, ToStatus: to, Version: 3,
	})
	require.NoError(t, err)
	return b
}

func inbox(t *testing.T, app *application.Service) []domain.Notification {
	t.Helper()
	list, _, err := app.List(context.Background(), student, false, 0, "")
	require.NoError(t, err)
	return list
}

func TestReadsTheTwoStreams(t *testing.T) {
	assert.ElementsMatch(t, []string{"events:report.submitted", "events:status.changed"}, Streams())
	assert.Equal(t, "notification", Group)
}

func TestReportSubmittedTellsTheStudentItWasReceived(t *testing.T) {
	c, app, reports, _ := setup()

	require.NoError(t, c.ReportSubmitted(context.Background(), submitted(t, event1)))

	list := inbox(t, app)
	require.Len(t, list, 1)
	n := list[0]
	assert.Equal(t, student, n.RecipientUserID)
	assert.Equal(t, domain.InApp, n.Channel)
	assert.Equal(t, domain.EventReportSubmitted, n.EventType)
	assert.Equal(t, reportID, n.ReportID)
	assert.Equal(t, "We received your report", n.Subject)
	assert.Contains(t, n.Body, "HC, Floor 1, Left wing bathroom")
	assert.Equal(t, 0, reports.calls, "the event already names the student")
}

func TestStatusChangesTheStudentIsToldAbout(t *testing.T) {
	tests := []struct {
		to      commonv1.Status
		subject string
	}{
		{commonv1.Status_IN_PROGRESS, "Your report is in progress"},
		{commonv1.Status_RESOLVED, "Your report is resolved"},
		{commonv1.Status_ESCALATED, "Your report was escalated"},
	}
	for _, tt := range tests {
		t.Run(tt.to.String(), func(t *testing.T) {
			c, app, reports, _ := setup()

			require.NoError(t, c.StatusChanged(context.Background(), changed(t, event1, commonv1.Status_ASSIGNED, tt.to)))

			list := inbox(t, app)
			require.Len(t, list, 1)
			n := list[0]
			assert.Equal(t, student, n.RecipientUserID, "the student comes from report-submission")
			assert.Equal(t, domain.EventStatusChanged, n.EventType)
			assert.Equal(t, reportID, n.ReportID)
			assert.Equal(t, tt.subject, n.Subject)
			assert.Contains(t, n.Body, "HC, Floor 1, Left wing bathroom")
			assert.LessOrEqual(t, len([]rune(n.Body)), 200)
			assert.Equal(t, 1, reports.calls)
		})
	}
}

// Assigned comes a moment after "received", so it would be a second message
// at once; Open is the start. Neither needs the report looked up.
func TestStatusChangesThatSendNothing(t *testing.T) {
	for _, to := range []commonv1.Status{commonv1.Status_OPEN, commonv1.Status_ASSIGNED} {
		t.Run(to.String(), func(t *testing.T) {
			c, app, reports, _ := setup()

			require.NoError(t, c.StatusChanged(context.Background(), changed(t, event1, commonv1.Status_OPEN, to)))

			assert.Empty(t, inbox(t, app))
			assert.Equal(t, 0, reports.calls)
		})
	}
}

func TestTheSameEventTwiceGivesOneMessage(t *testing.T) {
	c, app, _, sender := setup()
	ctx := context.Background()

	require.NoError(t, c.ReportSubmitted(ctx, submitted(t, event1)))
	require.NoError(t, c.ReportSubmitted(ctx, submitted(t, event1)))
	require.NoError(t, c.StatusChanged(ctx, changed(t, event2, commonv1.Status_ASSIGNED, commonv1.Status_RESOLVED)))
	require.NoError(t, c.StatusChanged(ctx, changed(t, event2, commonv1.Status_ASSIGNED, commonv1.Status_RESOLVED)))

	assert.Len(t, inbox(t, app), 2, "one for received, one for resolved")
	assert.Equal(t, 2, sender.calls)
}

// report-submission down: the error goes back to the bus, which tries again
// and then keeps the event in the dead stream. The event is never dropped.
func TestReportSubmissionDownIsAnErrorAndARetryWorks(t *testing.T) {
	c, app, reports, _ := setup()
	ctx := context.Background()
	reports.err = errors.Join(ports.ErrReportsUnavailable, errors.New("connection refused"))

	err := c.StatusChanged(ctx, changed(t, event1, commonv1.Status_ASSIGNED, commonv1.Status_IN_PROGRESS))
	assert.ErrorIs(t, err, ports.ErrReportsUnavailable)
	assert.Empty(t, inbox(t, app))

	reports.err = nil
	require.NoError(t, c.StatusChanged(ctx, changed(t, event1, commonv1.Status_ASSIGNED, commonv1.Status_IN_PROGRESS)))
	assert.Len(t, inbox(t, app), 1)
}

// A report report-submission does not know (for example, its memory was
// reset) has no student to tell. Trying again would not help, so it is skipped.
func TestUnknownReportIsSkipped(t *testing.T) {
	c, app, reports, _ := setup()
	reports.err = ports.ErrReportNotFound

	require.NoError(t, c.StatusChanged(context.Background(), changed(t, event1, commonv1.Status_ASSIGNED, commonv1.Status_RESOLVED)))
	assert.Empty(t, inbox(t, app))
}

func TestBadPayloadIsAnError(t *testing.T) {
	c, _, _, _ := setup()
	ctx := context.Background()
	assert.Error(t, c.ReportSubmitted(ctx, []byte("not protobuf")))
	assert.Error(t, c.StatusChanged(ctx, []byte("not protobuf")))
}
