// Package eventpublisher sends report.routed on the event bus (Redis Streams in
// the cluster). It is used when REDIS_URL is set; otherwise logpublisher is.
package eventpublisher

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

const producer = "routing-engine"

// routedNamespace makes report.routed ids from report.submitted ids (UUID v5).
var routedNamespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://github.com/jurwaCST/Jurwa/events/report.routed"))

// Publisher implements ports.EventPublisher with an eventbus.Publisher.
type Publisher struct {
	bus   eventbus.Publisher
	newID func() string
	now   func() time.Time
}

// New returns a Publisher. newID makes a new UUID for a routing with no source
// event (the RPC); now is usually time.Now.
func New(bus eventbus.Publisher, newID func() string, now func() time.Time) *Publisher {
	return &Publisher{bus: bus, newID: newID, now: now}
}

// PublishReportRouted publishes ev on events:report.routed. The report id is
// the correlation id, as in report.submitted.
func (p *Publisher) PublishReportRouted(ctx context.Context, ev ports.ReportRouted) error {
	return p.bus.Publish(ctx, eventbus.StreamReportRouted, &eventsv1.ReportRouted{
		Meta: &eventsv1.EventMeta{
			EventId:       p.eventID(ev.SourceEventID),
			OccurredAt:    timestamppb.New(p.now()),
			Producer:      producer,
			CorrelationId: ev.ReportID,
		},
		ReportId:       ev.ReportID,
		Category:       category(ev.Category),
		CategoryHeadId: ev.CategoryHeadID,
		TargetPosition: ev.TargetPosition,
		RuleId:         ev.RuleID,
	})
}

// eventID is the same for the same report.submitted, so readers can drop a
// repeat (pkg/eventbus: handlers ignore a repeated meta.event_id).
func (p *Publisher) eventID(sourceEventID string) string {
	if sourceEventID == "" {
		return p.newID()
	}
	return uuid.NewSHA1(routedNamespace, []byte(sourceEventID)).String()
}

func category(c domain.Category) commonv1.Category {
	switch c {
	case domain.Hostel:
		return commonv1.Category_HOSTEL
	case domain.AcademicBlock:
		return commonv1.Category_ACADEMIC_BLOCK
	case domain.Others:
		return commonv1.Category_OTHERS
	default:
		return commonv1.Category_CATEGORY_UNSPECIFIED
	}
}
