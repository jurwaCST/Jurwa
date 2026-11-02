package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
	"github.com/jurwaCST/Jurwa/services/sla-policy/ports"
)

// ErrInvalidArgument: a required field is missing or badly formed.
var ErrInvalidArgument = errors.New("invalid argument")

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service holds the sla-policy use cases.
type Service struct {
	repo ports.PolicyRepository
	now  func() time.Time
}

// NewService wires the use cases to the repository. now is time.Now outside tests.
func NewService(repo ports.PolicyRepository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

// GetPolicy returns the SLA window for a category.
func (s *Service) GetPolicy(ctx context.Context, category domain.Category) (domain.Policy, error) {
	if !category.Valid() {
		return domain.Policy{}, fmt.Errorf("%w: %d", domain.ErrInvalidCategory, int32(category))
	}
	return s.repo.Get(ctx, category)
}

// UpsertPolicy creates or replaces the SLA window for a category.
// Only the System Administrator may call it; roles are checked by the
// api-gateway and authorization service, so here updatedBy must be a UUID.
func (s *Service) UpsertPolicy(ctx context.Context, category domain.Category, windowHours int32, updatedBy string) (domain.Policy, error) {
	if !uuidPattern.MatchString(updatedBy) {
		return domain.Policy{}, fmt.Errorf("%w: updated_by must be a UUID", ErrInvalidArgument)
	}
	p, err := domain.NewPolicy(category, windowHours, updatedBy, s.now().UTC())
	if err != nil {
		return domain.Policy{}, err
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return domain.Policy{}, err
	}
	return p, nil
}
