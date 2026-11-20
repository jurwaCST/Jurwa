// Package events reads the event bus into analytics (SCRUM-160): report.submitted
// adds a report to the counts and status.changed moves it, so the admin and
// President dashboards show every report as it happens.
package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// Group is the consumer group of analytics on every stream.
const Group = "analytics"

// ErrUnknownReport: a status change for a report that did not arrive in time.
var ErrUnknownReport = errors.New("status change for a report analytics does not know")

// Streams returns the streams analytics reads.
func Streams() []string {
	return []string{eventbus.StreamReportSubmitted, eventbus.StreamStatusChanged, eventbus.StreamReportWithdrawn}
}

// UseCases is what the consumer needs from the application layer.
type UseCases interface {
	Record(ctx context.Context, e domain.Event) (bool, error)
	Knows(ctx context.Context, reportID string) (bool, error)
}

// Consumer turns report events into counts.
type Consumer struct {
	app UseCases
	// wait is how long a status change waits for its report: the two streams
	// are read at the same time, so after a restart the change can come first.
	wait time.Duration
}

func NewConsumer(app UseCases, wait time.Duration) *Consumer {
	return &Consumer{app: app, wait: wait}
}

// Run reads the streams until ctx is cancelled. It returns the first error of
// a subscription, or nil after ctx is cancelled.
func (c *Consumer) Run(ctx context.Context, sub eventbus.Subscriber, consumer string) error {
	handlers := map[string]eventbus.Handler{
		eventbus.StreamReportSubmitted: c.ReportSubmitted,
		eventbus.StreamStatusChanged:   c.StatusChanged,
		eventbus.StreamReportWithdrawn: c.ReportWithdrawn,
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
			cancel() // one subscription failed: stop the other
		}
	}
	return first
}

// ReportSubmitted counts a new report as Open, under its category.
func (c *Consumer) ReportSubmitted(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportSubmitted
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.submitted: %w", err)
	}
	r := ev.GetReport()
	at := r.GetCreatedAt().AsTime()
	if r.GetCreatedAt() == nil {
		at = ev.GetMeta().GetOccurredAt().AsTime()
	}
	_, err := c.app.Record(ctx, domain.Event{
		ID:         ev.GetMeta().GetEventId(),
		Type:       domain.ReportSubmitted,
		ReportID:   r.GetId(),
		Category:   domain.Category(r.GetCategory()),
		OccurredAt: at,
	})
	return err
}

// StatusChanged moves a report to its new status. When the report is not known
// yet it waits for it; if it never comes, the error makes the bus try again and
// finally keep the event in the dead-letter stream.
func (c *Consumer) StatusChanged(ctx context.Context, payload []byte) error {
	var ev eventsv1.StatusChanged
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("status.changed: %w", err)
	}
	if err := c.waitFor(ctx, ev.GetReportId()); err != nil {
		return err
	}
	_, err := c.app.Record(ctx, domain.Event{
		ID:         ev.GetMeta().GetEventId(),
		Type:       domain.StatusChanged,
		ReportID:   ev.GetReportId(),
		ToStatus:   domain.Status(ev.GetToStatus()),
		OccurredAt: ev.GetMeta().GetOccurredAt().AsTime(),
	})
	return err
}

func (c *Consumer) waitFor(ctx context.Context, reportID string) error {
	deadline := time.Now().Add(c.wait)
	for {
		known, err := c.app.Knows(ctx, reportID)
		if err != nil {
			return err
		}
		if known {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %s", ErrUnknownReport, reportID)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// ReportWithdrawn takes a report its student deleted out of the counts
// (SCRUM-163). It does not wait for the report: the mark is kept either way.
func (c *Consumer) ReportWithdrawn(ctx context.Context, payload []byte) error {
	var ev eventsv1.ReportWithdrawn
	if err := proto.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("report.withdrawn: %w", err)
	}
	_, err := c.app.Record(ctx, domain.Event{
		ID:         ev.GetMeta().GetEventId(),
		Type:       domain.ReportWithdrawn,
		ReportID:   ev.GetReportId(),
		OccurredAt: ev.GetMeta().GetOccurredAt().AsTime(),
	})
	return err
}
