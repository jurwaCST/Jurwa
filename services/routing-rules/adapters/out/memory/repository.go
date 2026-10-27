// Package memory is an in-memory RuleRepository, seeded with the default rules.
// Data is lost when the service stops.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

type key struct {
	category domain.Category
	building string
}

// Repository implements ports.RuleRepository. It is safe for concurrent use.
type Repository struct {
	mu    sync.RWMutex
	rules map[key]domain.Rule
}

// NewRepository returns a repository holding the given rules. Rules without an
// id get a new UUID.
func NewRepository(seed []domain.Rule) *Repository {
	r := &Repository{rules: make(map[key]domain.Rule)}
	for _, rule := range seed {
		if rule.ID == "" {
			rule.ID = uuid.NewString()
		}
		r.rules[key{rule.Category, rule.Building}] = rule
	}
	return r
}

func (r *Repository) FindByCategory(_ context.Context, category domain.Category) ([]domain.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Rule
	for k, rule := range r.rules {
		if k.category == category {
			out = append(out, rule)
		}
	}
	sortRules(out)
	return out, nil
}

func (r *Repository) List(_ context.Context, includeInactive bool) ([]domain.Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		if rule.Active || includeInactive {
			out = append(out, rule)
		}
	}
	sortRules(out)
	return out, nil
}

func (r *Repository) Upsert(_ context.Context, rule domain.Rule) (domain.Rule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := key{rule.Category, rule.Building}
	if existing, ok := r.rules[k]; ok {
		rule.ID = existing.ID
	} else {
		rule.ID = uuid.NewString()
	}
	r.rules[k] = rule
	return rule, nil
}

func sortRules(rules []domain.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Category != rules[j].Category {
			return rules[i].Category < rules[j].Category
		}
		return rules[i].Building < rules[j].Building
	})
}
