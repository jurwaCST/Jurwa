// Package memory is an in-memory TimerRepository. It is replaced by the Redis
// adapter (keys sla:deadlines and sla:timer:<report_id>) later.
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

// Repository keeps timers in a map, safe for concurrent use.
type Repository struct {
	mu     sync.RWMutex
	timers map[string]domain.Timer
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{timers: make(map[string]domain.Timer)}
}

func (r *Repository) Create(_ context.Context, t domain.Timer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.timers[t.ReportID]; ok {
		return domain.ErrAlreadyExists
	}
	r.timers[t.ReportID] = t
	return nil
}

func (r *Repository) Get(_ context.Context, reportID string) (domain.Timer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.timers[reportID]
	if !ok {
		return domain.Timer{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *Repository) Update(_ context.Context, t domain.Timer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.timers[t.ReportID]; !ok {
		return domain.ErrNotFound
	}
	r.timers[t.ReportID] = t
	return nil
}

// ListDue works like ZRANGEBYSCORE sla:deadlines 0 <now> LIMIT 0 <limit>.
func (r *Repository) ListDue(_ context.Context, now time.Time, limit int) ([]domain.Timer, error) {
	r.mu.RLock()
	due := make([]domain.Timer, 0)
	for _, t := range r.timers {
		if t.IsDue(now) {
			due = append(due, t)
		}
	}
	r.mu.RUnlock()

	sort.Slice(due, func(i, j int) bool {
		if !due[i].Deadline.Equal(due[j].Deadline) {
			return due[i].Deadline.Before(due[j].Deadline)
		}
		return due[i].ReportID < due[j].ReportID
	})
	if len(due) > limit {
		due = due[:limit]
	}
	return due, nil
}
