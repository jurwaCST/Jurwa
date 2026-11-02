package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// RuleLookup finds the routing rule (routing-rules GetRule). Implementations
// return domain.ErrRuleNotFound, domain.ErrInvalidCategory or domain.ErrUnavailable.
type RuleLookup interface {
	GetRule(ctx context.Context, category domain.Category, building string) (domain.Rule, error)
}

// HolderLookup finds who holds a position (department-directory GetStaff).
// It returns "" when nobody holds it, and domain.ErrUnavailable when the
// directory cannot be reached.
type HolderLookup interface {
	GetHolder(ctx context.Context, positionCode string) (string, error)
}

// ReportRouted is published on events:report.routed after a successful route.
type ReportRouted struct {
	ReportID       string
	Category       domain.Category
	CategoryHeadID string
	TargetPosition string
	RuleID         string
}

// EventPublisher sends events to other services (Redis Streams in Sprint 3).
type EventPublisher interface {
	PublishReportRouted(ctx context.Context, ev ReportRouted) error
}
