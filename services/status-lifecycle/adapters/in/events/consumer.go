// Package events reads status-lifecycle's events from the shared event bus:
// report.submitted gives a report OPEN, report.routed sets ASSIGNED, and
// sla.lapsed sets ESCALATED (escalation Saga). The consumer group is
// "status-lifecycle".
package events

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
)

// Group is the consumer group, named after the service (pkg/eventbus rule).
const Group = "status-lifecycle"

// App is the part of the application the consumers call.
type App interface {
	OnReportSubmitted(ctx context.Context, reportID string) error
	OnReportRouted(ctx context.Context, reportID, targetPosition, correlationID string) error
	EscalateOnLapse(ctx context.Context, ev application.SagaEvent) error
}

// Start subscribes to the three streams in goroutines. They stop when ctx is cancelled.
func Start(ctx context.Context, bus eventbus.Subscriber, consumer string, app App) {
	go subscribe(ctx, bus, eventbus.StreamReportSubmitted, consumer, OnReportSubmitted(app))
	go subscribe(ctx, bus, eventbus.StreamReportRouted, consumer, OnReportRouted(app))
	go subscribe(ctx, bus, eventbus.StreamSlaLapsed, consumer, OnSlaLapsed(app))
}

func subscribe(ctx context.Context, bus eventbus.Subscriber, stream, consumer string, h eventbus.Handler) {
	if err := bus.Subscribe(ctx, stream, Group, consumer, h); err != nil {
		slog.ErrorContext(ctx, "subscription stopped", "stream", stream, "error", err)
	}
}

// OnReportSubmitted gives the new report the status OPEN.
func OnReportSubmitted(app App) eventbus.Handler {
	return func(ctx context.Context, payload []byte) error {
		var e eventsv1.ReportSubmitted
		if err := proto.Unmarshal(payload, &e); err != nil {
			return err
		}
		return app.OnReportSubmitted(ctx, e.GetReport().GetId())
	}
}

// OnReportRouted sets the report to ASSIGNED.
func OnReportRouted(app App) eventbus.Handler {
	return func(ctx context.Context, payload []byte) error {
		var e eventsv1.ReportRouted
		if err := proto.Unmarshal(payload, &e); err != nil {
			return err
		}
		return app.OnReportRouted(ctx, e.GetReportId(), e.GetTargetPosition(), e.GetMeta().GetCorrelationId())
	}
}

// OnSlaLapsed sets the report to ESCALATED, or asks for compensation.
func OnSlaLapsed(app App) eventbus.Handler {
	return func(ctx context.Context, payload []byte) error {
		var e eventsv1.SlaLapsed
		if err := proto.Unmarshal(payload, &e); err != nil {
			return err
		}
		return app.EscalateOnLapse(ctx, application.SagaEvent{
			EventID: e.GetMeta().GetEventId(), CorrelationID: e.GetMeta().GetCorrelationId(), ReportID: e.GetReportId(),
		})
	}
}
