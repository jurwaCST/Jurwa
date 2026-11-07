// Package eventpublisher sends status-lifecycle's events on the shared event
// bus (pkg/eventbus, Redis Streams).
package eventpublisher

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

// Producer is the name written in every event's meta.producer.
const Producer = "status-lifecycle"

// Publisher implements ports.EventPublisher with eventbus.Publisher.
type Publisher struct {
	bus   eventbus.Publisher
	newID func() string
	now   func() time.Time
}

// New returns a Publisher. newID is uuid.NewString and now is time.Now outside tests.
func New(bus eventbus.Publisher, newID func() string, now func() time.Time) *Publisher {
	return &Publisher{bus: bus, newID: newID, now: now}
}

// meta builds the event meta. Outside a Saga run the event starts its own correlation.
func (p *Publisher) meta(correlationID string) *eventsv1.EventMeta {
	if correlationID == "" {
		correlationID = p.newID()
	}
	return &eventsv1.EventMeta{
		EventId: p.newID(), OccurredAt: timestamppb.New(p.now()), Producer: Producer, CorrelationId: correlationID,
	}
}

// PublishStatusChanged sends events:status.changed. actor_id is left out when the system acted.
func (p *Publisher) PublishStatusChanged(ctx context.Context, ev ports.StatusChanged) error {
	msg := &eventsv1.StatusChanged{
		Meta:       p.meta(ev.CorrelationID),
		ReportId:   ev.ReportID,
		FromStatus: commonv1.Status(ev.From),
		ToStatus:   commonv1.Status(ev.To),
		Version:    ev.Version,
	}
	if ev.ActorID != "" {
		msg.ActorId = proto.String(ev.ActorID)
	}
	return p.bus.Publish(ctx, eventbus.StreamStatusChanged, msg)
}

// PublishEscalationCompensated sends events:escalation.compensated.
func (p *Publisher) PublishEscalationCompensated(ctx context.Context, ev ports.EscalationCompensated) error {
	return p.bus.Publish(ctx, eventbus.StreamEscalationCompensated, &eventsv1.EscalationCompensated{
		Meta:       p.meta(ev.CorrelationID),
		ReportId:   ev.ReportID,
		FailedStep: ev.FailedStep,
		UndoneStep: ev.UndoneStep,
		Reason:     ev.Reason,
	})
}
