package application

import (
	"context"
	"time"

	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
	"github.com/jurwaCST/Jurwa/services/routing-rules/ports"
)

// Service holds the routing-rules use cases.
type Service struct {
	repo ports.RuleRepository
	now  func() time.Time
}

func NewService(repo ports.RuleRepository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// GetRule finds the rule for a category and building. The description is never an input.
func (s *Service) GetRule(ctx context.Context, q domain.Query) (domain.Rule, error) {
	if !q.Category.Valid() {
		return domain.Match(nil, q) // returns ErrInvalidCategory without a repository call
	}
	rules, err := s.repo.FindByCategory(ctx, q.Category)
	if err != nil {
		return domain.Rule{}, err
	}
	return domain.Match(rules, q)
}

// ListRules returns all rules, for the System Administrator.
func (s *Service) ListRules(ctx context.Context, includeInactive bool) ([]domain.Rule, error) {
	return s.repo.List(ctx, includeInactive)
}

// UpsertRule validates and saves a rule. A priority of 0 means the default (100).
func (s *Service) UpsertRule(ctx context.Context, rule domain.Rule) (domain.Rule, error) {
	if rule.Priority == 0 {
		rule.Priority = domain.DefaultPriority
	}
	if err := rule.Validate(); err != nil {
		return domain.Rule{}, err
	}
	rule.UpdatedAt = s.now().UTC()
	return s.repo.Upsert(ctx, rule)
}
