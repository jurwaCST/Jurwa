// Package events reads the event bus and tells the student about their report
// (SCRUM-155): report.submitted gives "We received your report", and
// status.changed to In progress, Resolved or Escalated gives a message too.
// status.changed does not name the student, so the report is looked up in
// report-submission. Each message's id is made from the event id, so an event
// delivered twice gives one message.
package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

// Group is the consumer group of notification on every stream.
const Group = "notification"

// maxBody is the notification body limit (domain), counted in letters.
const maxBody = 200

// Streams returns the streams notification reads.
func Streams() []string {
	return []string{eventbus.StreamReportSubmitted, eventbus.StreamStatusChanged}
}

// UseCases is what the consumer needs from the application layer.
type UseCases interface {
	SendForEvent(ctx context.Context, eventID string, in domain.Input) (domain.Notification, bool, error)
}

// Consumer turns report events into messages for the student.
type Consumer struct {
	app     UseCases
	reports ports.Reports
}

func NewConsumer(app UseCases, reports ports.Reports) *Consumer {
	return &Consumer{app: app, reports: reports}
}

// Run reads the streams until ctx is cancelled. It returns the first error of
// a subscription, or nil after ctx is cancelled.
func (c *Consumer) Run(ctx context.Context, sub eventbus.Subscriber, consumer string) error {
	handlers := map[string]eventbus.Handler{
		eventbus.StreamReportSubmitted: c.ReportSubmitted,
		eventbus.StreamStatusChanged:   c.StatusChanged,
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, len(handlers))
	for stream, h := range handlers {
		go func() { errs <- sub.Subscribe(ctx, stream, Group, consumer, h) }()
	}
	var first error
	for range handlers {
		if err := <-errs; err != nil && first == nil {
			first = err
			cancel() // one subscription failed: stop the others
		}
	}
	return first
}

// ReportSubmitted tells the student their report was received. The event
// carries the report, so no lookup is needed.
func (c *Consumer) ReportSubmitted(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportSubmitted
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.submitted: %w", err)
	}
	r := ev.GetReport()
	l := r.GetLocation()
	return c.send(ctx, ev.GetMeta().GetEventId(), domain.Input{
		RecipientUserID: r.GetReporterId(),
		Channel:         domain.InApp,
		EventType:       domain.EventReportSubmitted,
		ReportID:        r.GetId(),
		Subject:         "We received your report",
		Body:            body(placeOf(l.GetBuilding(), l.GetWing(), l.GetArea()), ". You can follow it in My reports."),
	})
}

// What the student is told for each status. Open is the start, and Assigned
// comes a moment after "received", so neither sends a message.
var statusMessages = map[commonv1.Status]struct{ subject, rest string }{
	commonv1.Status_IN_PROGRESS: {"Your report is in progress", ": someone is working on it now."},
	commonv1.Status_RESOLVED:    {"Your report is resolved", ": it is marked as fixed. You can rate the fix in My reports."},
	commonv1.Status_ESCALATED:   {"Your report was escalated", ": it passed its time limit, so it went to the Category Head."},
}

// StatusChanged tells the student about a new status. An error from
// report-submission goes back to the bus, which tries the event again and
// then keeps it in the dead stream, so it is never dropped.
func (c *Consumer) StatusChanged(ctx context.Context, payload []byte) error {
	var ev eventsv1.StatusChanged
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("status.changed: %w", err)
	}
	msg, ok := statusMessages[ev.GetToStatus()]
	if !ok {
		return nil
	}
	r, err := c.reports.GetReport(ctx, ev.GetReportId())
	if errors.Is(err, ports.ErrReportNotFound) {
		// No report, so no student to tell; trying again would not help.
		slog.Warn("status.changed for an unknown report, no message sent", "report_id", ev.GetReportId(), "error", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("status.changed: %w", err)
	}
	return c.send(ctx, ev.GetMeta().GetEventId(), domain.Input{
		RecipientUserID: r.ReporterID,
		Channel:         domain.InApp,
		EventType:       domain.EventStatusChanged,
		ReportID:        ev.GetReportId(),
		Subject:         msg.subject,
		Body:            body(placeOf(r.Building, r.Wing, r.Area), msg.rest),
	})
}

func (c *Consumer) send(ctx context.Context, eventID string, in domain.Input) error {
	_, _, err := c.app.SendForEvent(ctx, eventID, in)
	return err
}

// placeOf joins the parts of a location that are given: "HC, Floor 1, Left wing bathroom".
func placeOf(parts ...string) string {
	var out []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return "Your report"
	}
	return strings.Join(out, ", ")
}

// body is the place followed by rest, the place shortened when the whole
// would pass the 200-letter limit.
func body(where, rest string) string {
	room := maxBody - len([]rune(rest))
	if r := []rune(where); len(r) > room {
		where = string(r[:room-1]) + "…"
	}
	return where + rest
}
