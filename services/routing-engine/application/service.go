package application

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

// Service holds the routing-engine use case. It is stateless.
type Service struct {
	rules   ports.RuleLookup
	holders ports.HolderLookup
	pub     ports.EventPublisher
}

func NewService(rules ports.RuleLookup, holders ports.HolderLookup, pub ports.EventPublisher) *Service {
	return &Service{rules: rules, holders: holders, pub: pub}
}

// RouteReport decides who receives a report, using only category and building.
// When the request has a report id it publishes report.routed.
func (s *Service) RouteReport(ctx context.Context, req domain.Request) (domain.Decision, error) {
	return s.route(ctx, req, "")
}

// RouteSubmitted routes a report because of the report.submitted event
// sourceEventID. It always publishes report.routed, so it needs a report id.
func (s *Service) RouteSubmitted(ctx context.Context, sourceEventID string, req domain.Request) (domain.Decision, error) {
	if req.ReportID == "" {
		return domain.Decision{}, domain.ErrMissingReportID
	}
	return s.route(ctx, req, sourceEventID)
}

func (s *Service) route(ctx context.Context, req domain.Request, sourceEventID string) (domain.Decision, error) {
	if err := req.Validate(); err != nil {
		return domain.Decision{}, err
	}
	rule, err := s.rules.GetRule(ctx, req.Category, req.Building)
	if err != nil {
		return domain.Decision{}, err
	}
	holder, err := s.holders.GetHolder(ctx, rule.TargetPosition)
	if err != nil {
		return domain.Decision{}, err
	}
	d, err := domain.Decide(req, rule, holder)
	if err != nil {
		return domain.Decision{}, err
	}
	if d.ReportID == "" {
		return d, nil
	}
	ev := ports.ReportRouted{
		ReportID: d.ReportID, Category: d.Category, CategoryHeadID: d.CategoryHeadID,
		TargetPosition: d.TargetPosition, RuleID: d.RuleID, SourceEventID: sourceEventID,
	}
	if err := s.pub.PublishReportRouted(ctx, ev); err != nil {
		return domain.Decision{}, fmt.Errorf("publish report.routed: %w", err)
	}
	return d, nil
}
