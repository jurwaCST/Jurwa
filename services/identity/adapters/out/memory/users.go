// Package memory is an in-memory UserRepository for local runs and tests.
// Data is lost when the server stops. Used when DATABASE_URL is not set.
package memory

import (
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

// Users is safe to use from many gRPC calls at once.
type Users struct {
	mu      sync.RWMutex
	byID    map[string]domain.User
	byEmail map[string]string // email -> id
}

func NewUsers() *Users {
	return &Users{byID: map[string]domain.User{}, byEmail: map[string]string{}}
}

func (r *Users) Create(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, taken := r.byEmail[u.Email]; taken {
		return domain.ErrEmailTaken
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u.ID
	return nil
}

func (r *Users) FindByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return r.byID[id], nil
}

func (r *Users) FindByID(_ context.Context, id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}
