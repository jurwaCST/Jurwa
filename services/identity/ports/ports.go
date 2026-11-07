// Package ports holds the interfaces the identity core depends on. Adapters
// implement them, and tests replace them with fakes.
package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

// UserRepository stores user accounts: adapters/out/memory, or
// adapters/out/postgres when DATABASE_URL is set.
type UserRepository interface {
	// Create saves a new user. Returns domain.ErrEmailTaken if the email exists.
	Create(ctx context.Context, u domain.User) error
	// FindByEmail returns domain.ErrUserNotFound if there is no such user.
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	// FindByID returns domain.ErrUserNotFound if there is no such user.
	FindByID(ctx context.Context, id string) (domain.User, error)
}
