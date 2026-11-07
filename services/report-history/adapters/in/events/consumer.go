// Package events reads the event bus and records each event in report-history:
// report.submitted creates the summary, report.routed and report.reassigned set
// the assignee, status.changed sets the status. Every event becomes one
// timeline entry whose id is the event id, so an event delivered twice is
// recorded once.
package events

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/report-history/application"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

// Group is the consumer group of report-history on every stream.
const Group = "report-history"

// Streams returns the streams report-history reads.
func Streams() []string {
	return []string{
		eventbus.StreamReportSubmitted,
		eventbus.StreamReportRouted,
		eventbus.StreamStatusChanged,
		eventbus.StreamReportReassigned,
	}
}

// UseCases is what the consumer needs from the application layer.
type UseCases interface {
	RecordSubmitted(ctx context.Context, in application.SubmittedInput) (domain.Summary, error)
	RecordEvent(ctx context.Context, eventID string, in domain.EntryInput) error
}

// Consumer turns events into report-history records.
type Consumer struct {
	app UseCases
	// WaitForReport is how long an event waits for its report.submitted, which
	// comes on another stream and can arrive later.
	WaitForReport time.Duration
	// PollEvery is how often the report is looked for while waiting.
	PollEvery time.Duration
}

// NewConsumer returns a Consumer that waits up to 5 seconds for a report.
func NewConsumer(app UseCases) *Consumer {
	return &Consumer{app: app, WaitForReport: 5 * time.Second, PollEvery: 200 * time.Millisecond}
}

// Run reads the four streams until ctx is cancelled. It returns the first
// error of a subscription, or nil after ctx is cancelled.
func (c *Consumer) Run(ctx context.Context, sub eventbus.Subscriber, consumer string) error {
	handlers := map[string]eventbus.Handler{
		eventbus.StreamReportSubmitted:  c.ReportSubmitted,
		eventbus.StreamReportRouted:     c.ReportRouted,
		eventbus.StreamStatusChanged:    c.StatusChanged,
		eventbus.StreamReportReassigned: c.ReportReassigned,
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

// ReportSubmitted records a new report. A report recorded before is ignored.
func (c *Consumer) ReportSubmitted(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportSubmitted
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.submitted: %w", err)
	}
	r := ev.GetReport()
	_, err := c.app.RecordSubmitted(ctx, application.SubmittedInput{
		EventID:    ev.GetMeta().GetEventId(),
		ReportID:   r.GetId(),
		ReporterID: r.GetReporterId(),
		Category:   category(r.GetCategory()),
		Location:   domain.Location{Building: r.GetLocation().GetBuilding(), Wing: r.GetLocation().GetWing(), Area: r.GetLocation().GetArea()},
		CreatedAt:  r.GetCreatedAt().AsTime(),
	})
	if errors.Is(err, domain.ErrReportExists) {
		return nil
	}
	return err
}

// ReportRouted sets the Category Head as the assignee.
func (c *Consumer) ReportRouted(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportRouted
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.routed: %w", err)
	}
	summary := "Routed"
	if p := strings.TrimSpace(ev.GetTargetPosition()); p != "" {
		summary = "Routed to " + p
	}
	return c.record(ctx, ev.GetMeta(), domain.EntryInput{
		ReportID:   ev.GetReportId(),
		EventType:  domain.EventReportRouted,
		AssigneeID: ev.GetCategoryHeadId(),
		Summary:    summary,
	})
}

// StatusChanged records the new status.
func (c *Consumer) StatusChanged(ctx context.Context, payload []byte) error {
	var ev eventsv1.StatusChanged
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("status.changed: %w", err)
	}
	to := status(ev.GetToStatus())
	return c.record(ctx, ev.GetMeta(), domain.EntryInput{
		ReportID:   ev.GetReportId(),
		EventType:  domain.EventStatusChanged,
		FromStatus: status(ev.GetFromStatus()),
		ToStatus:   to,
		ActorID:    ev.GetActorId(),
		Summary:    "Status changed to " + string(to),
	})
}

// ReportReassigned sets the new assignee.
func (c *Consumer) ReportReassigned(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportReassigned
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.reassigned: %w", err)
	}
	summary := "Reassigned"
	if r := strings.TrimSpace(ev.GetReason()); r != "" {
		summary = "Reassigned: " + r
	}
	return c.record(ctx, ev.GetMeta(), domain.EntryInput{
		ReportID:   ev.GetReportId(),
		EventType:  domain.EventReportReassigned,
		ActorID:    ev.GetActorId(),
		AssigneeID: ev.GetToUserId(),
		Summary:    summary,
	})
}

// record stores the entry. When the report is not known yet (its
// report.submitted is still on the way) it waits up to WaitForReport.
func (c *Consumer) record(ctx context.Context, meta *eventsv1.EventMeta, in domain.EntryInput) error {
	in.OccurredAt = meta.GetOccurredAt().AsTime()
	deadline := time.Now().Add(c.WaitForReport)
	for {
		err := c.app.RecordEvent(ctx, meta.GetEventId(), in)
		if !errors.Is(err, domain.ErrReportNotFound) || time.Now().After(deadline) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.PollEvery):
		}
	}
}

func category(c commonv1.Category) domain.Category {
	if c == commonv1.Category_CATEGORY_UNSPECIFIED {
		return ""
	}
	return domain.Category(c.String())
}

func status(s commonv1.Status) domain.Status {
	if s == commonv1.Status_STATUS_UNSPECIFIED {
		return ""
	}
	return domain.Status(s.String())
}
