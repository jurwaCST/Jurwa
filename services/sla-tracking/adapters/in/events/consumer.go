// Package events reads the report flow from the shared event bus: report.routed
// starts a report's SLA timer and status.changed to RESOLVED stops it. The
// consumer group is "sla-tracking".
package events

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

// Group is the consumer group, named after the service (pkg/eventbus rule).
const Group = "sla-tracking"

// Timers is the part of the application the consumers call.
type Timers interface {
	StartTimer(ctx context.Context, reportID string, category domain.Category) (domain.Timer, error)
	StopTimer(ctx context.Context, reportID string) (domain.Timer, error)
}

// Start subscribes to report.routed and status.changed in goroutines. They stop
// when ctx is cancelled. consumer names this instance (the hostname).
func Start(ctx context.Context, bus eventbus.Subscriber, consumer string, app Timers) {
	go subscribe(ctx, bus, eventbus.StreamReportRouted, consumer, OnReportRouted(app))
	go subscribe(ctx, bus, eventbus.StreamStatusChanged, consumer, OnStatusChanged(app))
}

func subscribe(ctx context.Context, bus eventbus.Subscriber, stream, consumer string, h eventbus.Handler) {
	if err := bus.Subscribe(ctx, stream, Group, consumer, h); err != nil {
		slog.ErrorContext(ctx, "subscription stopped", "stream", stream, "error", err)
	}
}

// OnReportRouted starts the SLA timer. A timer that already exists (a repeated
// event) counts as done. Any other error (for example no SLA window set in
// sla-policy) is returned, so the event ends in events:report.routed.dead
// where it can be seen.
func OnReportRouted(app Timers) eventbus.Handler {
	return func(ctx context.Context, payload []byte) error {
		var e eventsv1.ReportRouted
		if err := proto.Unmarshal(payload, &e); err != nil {
			return err
		}
		_, err := app.StartTimer(ctx, e.GetReportId(), domain.Category(e.GetCategory()))
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil
		}
		if err != nil {
			slog.WarnContext(ctx, "cannot start SLA timer", "report_id", e.GetReportId(), "error", err)
		}
		return err
	}
}

// OnStatusChanged stops the timer when the report is RESOLVED. Other changes,
// and a report without a timer, are ignored.
func OnStatusChanged(app Timers) eventbus.Handler {
	return func(ctx context.Context, payload []byte) error {
		var e eventsv1.StatusChanged
		if err := proto.Unmarshal(payload, &e); err != nil {
			return err
		}
		if e.GetToStatus() != commonv1.Status_RESOLVED {
			return nil
		}
		_, err := app.StopTimer(ctx, e.GetReportId())
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
}
