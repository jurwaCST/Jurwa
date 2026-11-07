// Package events reads the escalation Saga's events from the shared event bus
// and calls reassignment's use cases. The consumer group is "reassignment".
package events

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/proto"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/reassignment/application"
)

// Group is the consumer group, named after the service (pkg/eventbus rule).
const Group = "reassignment"

// Saga is the part of the application the consumers call.
type Saga interface {
	EscalateOnLapse(ctx context.Context, ev application.SagaEvent) error
	CompensateEscalation(ctx context.Context, ev application.SagaEvent) error
}

// Start subscribes to sla.lapsed and escalation.compensated in goroutines.
// They stop when ctx is cancelled. consumer names this instance (the hostname).
func Start(ctx context.Context, bus eventbus.Subscriber, consumer string, app Saga) {
	go subscribe(ctx, bus, eventbus.StreamSlaLapsed, consumer, OnSlaLapsed(app))
	go subscribe(ctx, bus, eventbus.StreamEscalationCompensated, consumer, OnEscalationCompensated(app))
}

func subscribe(ctx context.Context, bus eventbus.Subscriber, stream, consumer string, h eventbus.Handler) {
	if err := bus.Subscribe(ctx, stream, Group, consumer, h); err != nil {
		slog.ErrorContext(ctx, "subscription stopped", "stream", stream, "error", err)
	}
}

// OnSlaLapsed escalates the report to its Category Head.
func OnSlaLapsed(app Saga) eventbus.Handler {
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

// OnEscalationCompensated undoes the escalation when the step it undoes is ours.
func OnEscalationCompensated(app Saga) eventbus.Handler {
	return func(ctx context.Context, payload []byte) error {
		var e eventsv1.EscalationCompensated
		if err := proto.Unmarshal(payload, &e); err != nil {
			return err
		}
		if e.GetUndoneStep() != "reassignment.escalate" {
			return nil
		}
		return app.CompensateEscalation(ctx, application.SagaEvent{
			EventID: e.GetMeta().GetEventId(), CorrelationID: e.GetMeta().GetCorrelationId(),
			ReportID: e.GetReportId(), Reason: e.GetReason(),
		})
	}
}
