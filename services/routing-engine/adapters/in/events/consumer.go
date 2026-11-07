// Package events is routing-engine's event input: it reads report.submitted
// and routes each new report, which publishes report.routed.
package events

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	eventsv1 "github.com/jurwaCST/Jurwa/proto/gen/events/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// Group is the consumer group, named after the service (pkg/eventbus rule).
const Group = "routing-engine"

// Router is the use case the consumer calls (application.Service).
type Router interface {
	RouteSubmitted(ctx context.Context, sourceEventID string, req domain.Request) (domain.Decision, error)
}

// Consumer handles events:report.submitted.
type Consumer struct{ router Router }

// New returns a Consumer that routes with router.
func New(router Router) *Consumer { return &Consumer{router: router} }

// Handle is an eventbus.Handler. Only the report id, category and building are
// passed on: the description never affects routing (decision 1, FR-11).
// An error makes the bus try again and then move the message to
// events:report.submitted.dead. A repeated event is routed again, which is safe:
// routing is stateless and report.routed gets the same event id.
func (c *Consumer) Handle(ctx context.Context, payload []byte) error {
	ev := &eventsv1.ReportSubmitted{}
	if err := proto.Unmarshal(payload, ev); err != nil {
		return fmt.Errorf("decode report.submitted: %w", err)
	}
	r := ev.GetReport()
	_, err := c.router.RouteSubmitted(ctx, ev.GetMeta().GetEventId(), domain.Request{
		ReportID: r.GetId(),
		Category: category(r.GetCategory()),
		Building: r.GetLocation().GetBuilding(),
	})
	if err != nil {
		return fmt.Errorf("route report %s: %w", r.GetId(), err)
	}
	return nil
}

func category(c commonv1.Category) domain.Category {
	switch c {
	case commonv1.Category_HOSTEL:
		return domain.Hostel
	case commonv1.Category_ACADEMIC_BLOCK:
		return domain.AcademicBlock
	case commonv1.Category_OTHERS:
		return domain.Others
	default:
		return ""
	}
}
