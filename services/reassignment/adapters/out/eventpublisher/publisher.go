// Package eventpublisher sends reassignment's events on the shared event bus
// (pkg/eventbus, Redis Streams).
package eventpublisher

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

// Producer is the name written in every event's meta.producer.
const Producer = "reassignment"

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

// PublishReportReassigned sends events:report.reassigned. actor_id is left out
// when the system acted. Outside a Saga run the event starts its own correlation.
func (p *Publisher) PublishReportReassigned(ctx context.Context, ev ports.ReportReassigned) error {
	correlation := ev.CorrelationID
	if correlation == "" {
		correlation = p.newID()
	}
	msg := &eventsv1.ReportReassigned{
		Meta: &eventsv1.EventMeta{
			EventId:       p.newID(),
			OccurredAt:    timestamppb.New(p.now()),
			Producer:      Producer,
			CorrelationId: correlation,
		},
		ReportId:   ev.ReportID,
		FromUserId: ev.FromUserID,
		ToUserId:   ev.ToUserID,
		Reason:     ev.Reason,
	}
	if ev.ActorID != "" {
		actor := ev.ActorID
		msg.ActorId = &actor
	}
	return p.bus.Publish(ctx, eventbus.StreamReportReassigned, msg)
}
