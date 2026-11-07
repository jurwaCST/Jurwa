// Package eventpublisher sends sla-tracking's events on the shared event bus
// (pkg/eventbus, Redis Streams). Each sla.lapsed starts one escalation Saga run,
// so it gets a new correlation_id that every later event of the run repeats.
package eventpublisher

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

// Producer is the name written in every event's meta.producer.
const Producer = "sla-tracking"

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

// PublishSLALapsed sends events:sla.lapsed.
func (p *Publisher) PublishSLALapsed(ctx context.Context, ev ports.SLALapsed) error {
	return p.bus.Publish(ctx, eventbus.StreamSlaLapsed, &eventsv1.SlaLapsed{
		Meta: &eventsv1.EventMeta{
			EventId:       p.newID(),
			OccurredAt:    timestamppb.New(p.now()),
			Producer:      Producer,
			CorrelationId: p.newID(), // a new Saga run
		},
		ReportId: ev.ReportID,
		Category: commonv1.Category(ev.Category),
		Deadline: timestamppb.New(ev.Deadline),
	})
}
