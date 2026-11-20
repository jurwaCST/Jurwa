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

// Streams returns the streams notification reads. SCRUM-162 added
// report.routed and report.reassigned: the student hears who has the report,
// and the person it reaches hears that it arrived.
func Streams() []string {
	return []string{
		eventbus.StreamReportSubmitted, eventbus.StreamStatusChanged, eventbus.StreamReportRouted, eventbus.StreamReportReassigned,
		eventbus.StreamReportWithdrawn, // SCRUM-163
	}
}

// UseCases is what the consumer needs from the application layer.
type UseCases interface {
	SendForEvent(ctx context.Context, eventID string, in domain.Input) (domain.Notification, bool, error)
}

// Consumer turns report events into messages for the student and for the
// person a report reaches.
type Consumer struct {
	app       UseCases
	reports   ports.Reports
	positions ports.Positions
}

func NewConsumer(app UseCases, reports ports.Reports, positions ports.Positions) *Consumer {
	return &Consumer{app: app, reports: reports, positions: positions}
}

// Run reads the streams until ctx is cancelled. It returns the first error of
// a subscription, or nil after ctx is cancelled.
func (c *Consumer) Run(ctx context.Context, sub eventbus.Subscriber, consumer string) error {
	handlers := map[string]eventbus.Handler{
		eventbus.StreamReportSubmitted:  c.ReportSubmitted,
		eventbus.StreamStatusChanged:    c.StatusChanged,
		eventbus.StreamReportRouted:     c.ReportRouted,
		eventbus.StreamReportReassigned: c.ReportReassigned,
		eventbus.StreamReportWithdrawn:  c.ReportWithdrawn,
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
	commonv1.Status_ESCALATED:   {"Your report was escalated", ": it passed its time limit, so it went up to the DSA or the DAA."},
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

// ReportRouted tells the student who has their report, and that person that a
// report arrived (SCRUM-162).
func (c *Consumer) ReportRouted(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportRouted
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.routed: %w", err)
	}
	return c.handOver(ctx, ev.GetMeta().GetEventId(), domain.EventReportRouted, ev.GetReportId(), ev.GetCategoryHeadId(), "Your report was sent to ")
}

// ReportReassigned tells the student and the new holder when a person passes a
// report on. When the system acted (no actor), an escalation is told to the
// person it went up to; the student hears it from status.changed (ESCALATED).
// An undone escalation sends nothing.
func (c *Consumer) ReportReassigned(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportReassigned
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.reassigned: %w", err)
	}
	if ev.ActorId == nil {
		if !strings.HasPrefix(ev.GetReason(), "SLA lapsed") {
			return nil
		}
		return c.toHolder(ctx, ev.GetMeta().GetEventId(), domain.EventReportReassigned, ev.GetReportId(), ev.GetToUserId(),
			"A late report was escalated to you", ": it passed its time limit. Open it on your board.")
	}
	return c.handOver(ctx, ev.GetMeta().GetEventId(), domain.EventReportReassigned, ev.GetReportId(), ev.GetToUserId(), "Your report was passed to ")
}

// ReportWithdrawn tells whoever held a report that its student deleted it
// (SCRUM-163). The report is gone, so the message cannot name its place.
func (c *Consumer) ReportWithdrawn(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportWithdrawn
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.withdrawn: %w", err)
	}
	holder := ev.GetCurrentAssigneeId()
	if holder == "" {
		return nil
	}
	return c.send(ctx, ev.GetMeta().GetEventId()+"/holder", domain.Input{
		RecipientUserID: holder,
		Channel:         domain.InApp,
		EventType:       domain.EventReportWithdrawn,
		ReportID:        ev.GetReportId(),
		Subject:         "Report removed by the student",
		Body:            "The student deleted a report you held before work started, so it has left your board.",
	})
}

// handOver sends two messages for one event: to the student, naming the post of
// the new holder, and to the holder. The holder's message id is made from the
// event id plus "/holder", so the two never share an id.
func (c *Consumer) handOver(ctx context.Context, eventID, eventType, reportID, holder, studentSubject string) error {
	r, err := c.reports.GetReport(ctx, reportID)
	if errors.Is(err, ports.ErrReportNotFound) {
		slog.Warn(eventType+" for an unknown report, no message sent", "report_id", reportID, "error", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", eventType, err)
	}
	where := placeOf(r.Building, r.Wing, r.Area)
	if r.ReporterID != "" {
		if err := c.send(ctx, eventID, domain.Input{
			RecipientUserID: r.ReporterID,
			Channel:         domain.InApp,
			EventType:       eventType,
			ReportID:        reportID,
			Subject:         studentSubject + c.titleOf(ctx, holder),
			Body:            body(where, ": they will fix it or pass it to the right person."),
		}); err != nil {
			return err
		}
	}
	if holder == "" {
		return nil
	}
	return c.send(ctx, eventID+"/holder", domain.Input{
		RecipientUserID: holder,
		Channel:         domain.InApp,
		EventType:       eventType,
		ReportID:        reportID,
		Subject:         "New report assigned to you",
		Body:            body(where, ": open it on your board."),
	})
}

// toHolder sends one message, to the person a report reached.
func (c *Consumer) toHolder(ctx context.Context, eventID, eventType, reportID, holder, subject, rest string) error {
	if holder == "" {
		return nil
	}
	r, err := c.reports.GetReport(ctx, reportID)
	if errors.Is(err, ports.ErrReportNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", eventType, err)
	}
	return c.send(ctx, eventID+"/holder", domain.Input{
		RecipientUserID: holder,
		Channel:         domain.InApp,
		EventType:       eventType,
		ReportID:        reportID,
		Subject:         subject,
		Body:            body(placeOf(r.Building, r.Wing, r.Area), rest),
	})
}

// titleOf is "the <post>" of a user, or "the person who will fix it" when the
// directory cannot say: the message still goes out.
func (c *Consumer) titleOf(ctx context.Context, userID string) string {
	title, err := c.positions.TitleOf(ctx, userID)
	if err != nil {
		slog.Warn("no post name for a message", "user_id", userID, "error", err)
	}
	if err != nil || title == "" {
		return "the person who will fix it"
	}
	return "the " + title
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
