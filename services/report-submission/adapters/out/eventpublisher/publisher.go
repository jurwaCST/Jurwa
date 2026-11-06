// Package eventpublisher sends report.submitted on the event bus (Redis Streams
// in the cluster). It is used when REDIS_URL is set.
package eventpublisher

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

const producer = "report-submission"

// Publisher implements ports.EventPublisher with an eventbus.Publisher.
type Publisher struct {
	bus     eventbus.Publisher
	eventID func() string
	now     func() time.Time
}

// New returns a Publisher. eventID makes a new UUID per event; now is usually time.Now.
func New(bus eventbus.Publisher, eventID func() string, now func() time.Time) *Publisher {
	return &Publisher{bus: bus, eventID: eventID, now: now}
}

// ReportSubmitted publishes the report on events:report.submitted. The report id
// is the correlation id, so every later event about this report can be linked to it.
func (p *Publisher) ReportSubmitted(ctx context.Context, r domain.Report) error {
	return p.bus.Publish(ctx, eventbus.StreamReportSubmitted, &eventsv1.ReportSubmitted{
		Meta: &eventsv1.EventMeta{
			EventId:       p.eventID(),
			OccurredAt:    timestamppb.New(p.now()),
			Producer:      producer,
			CorrelationId: r.ID,
		},
		Report: toProto(r),
	})
}

func toProto(r domain.Report) *commonv1.Report {
	return &commonv1.Report{
		Id:          r.ID,
		ReporterId:  r.ReporterID,
		Category:    category(r.Category),
		Location:    &commonv1.Location{Building: r.Location.Building, Wing: r.Location.Wing, Area: r.Location.Area},
		Description: r.Description,
		PhotoId:     r.PhotoID,
		Status:      commonv1.Status_OPEN,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
}

func category(c domain.Category) commonv1.Category {
	switch c {
	case domain.CategoryHostel:
		return commonv1.Category_HOSTEL
	case domain.CategoryAcademicBlock:
		return commonv1.Category_ACADEMIC_BLOCK
	case domain.CategoryOthers:
		return commonv1.Category_OTHERS
	default:
		return commonv1.Category_CATEGORY_UNSPECIFIED
	}
}
