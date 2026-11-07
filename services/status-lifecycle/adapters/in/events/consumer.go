// Package events reads sla.lapsed from the shared event bus and sets the report
// to ESCALATED (escalation Saga). The consumer group is "status-lifecycle".
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

// Saga is the part of the application the consumer calls.
type Saga interface {
	EscalateOnLapse(ctx context.Context, ev application.SagaEvent) error
}

// Start subscribes to sla.lapsed in a goroutine. It stops when ctx is cancelled.
func Start(ctx context.Context, bus eventbus.Subscriber, consumer string, app Saga) {
	go func() {
		if err := bus.Subscribe(ctx, eventbus.StreamSlaLapsed, Group, consumer, OnSlaLapsed(app)); err != nil {
			slog.ErrorContext(ctx, "subscription stopped", "stream", eventbus.StreamSlaLapsed, "error", err)
		}
	}()
}

// OnSlaLapsed sets the report to ESCALATED, or asks for compensation.
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
