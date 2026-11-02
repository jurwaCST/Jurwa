package application

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
	"github.com/jurwaCST/Jurwa/services/authorization/ports"
)

// Service holds the authorization use cases.
type Service struct {
	repo ports.PolicyRepository
}

// NewService wires the use cases to the policy store.
func NewService(repo ports.PolicyRepository) *Service {
	return &Service{repo: repo}
}

// CheckPermission loads the policy and asks it whether role may do action.
func (s *Service) CheckPermission(ctx context.Context, role domain.Role, action, resource string) (domain.Decision, error) {
	p, err := s.repo.Load(ctx)
	if err != nil {
		return domain.Decision{}, fmt.Errorf("load policy: %w", err)
	}
	return p.Check(role, action, resource)
}
