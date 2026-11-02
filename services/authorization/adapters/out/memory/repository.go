// Package memory is an in-memory PolicyRepository that holds one fixed
// permission table. A database replaces it in Sprint 3.
package memory

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

// Repository holds the policy. The policy never changes after start, so no
// lock is needed.
type Repository struct {
	policy domain.Policy
}

// New returns a repository with the default role table.
func New() *Repository {
	return WithPolicy(domain.DefaultPolicy())
}

// WithPolicy returns a repository with the given policy. Used by tests.
func WithPolicy(p domain.Policy) *Repository {
	return &Repository{policy: p}
}

func (r *Repository) Load(_ context.Context) (domain.Policy, error) {
	return r.policy, nil
}
