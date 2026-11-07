// Package repotest holds the behaviour every ports.UserRepository must have.
// The in-memory adapter runs it in its unit tests and the PostgreSQL adapter in
// its integration test, so both store users the same way.
package repotest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/identity/domain"
	"github.com/jurwaCST/Jurwa/services/identity/ports"
)

// Run checks repo behaviour. newRepo must return an empty repository.
func Run(t *testing.T, newRepo func(t *testing.T) ports.UserRepository) {
	ctx := context.Background()

	t.Run("a saved user is found by email and by id with every field", func(t *testing.T) {
		repo := newRepo(t)
		want := user("00000000-0000-4000-8000-000000000001", "dsa.demo@jurwa.example", domain.RoleCategoryHead)
		require.NoError(t, repo.Create(ctx, want))

		byEmail, err := repo.FindByEmail(ctx, want.Email)
		require.NoError(t, err)
		assertSameUser(t, want, byEmail)

		byID, err := repo.FindByID(ctx, want.ID)
		require.NoError(t, err)
		assertSameUser(t, want, byID)
	})

	t.Run("several roles are kept in order", func(t *testing.T) {
		repo := newRepo(t)
		want := user("7b1d2c3e-4f50-4a61-8b72-9c8d7e6f5a40", "two.roles@jurwa.example", domain.RoleStaff, domain.RoleSystemAdmin)
		require.NoError(t, repo.Create(ctx, want))

		got, err := repo.FindByID(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, []domain.Role{domain.RoleStaff, domain.RoleSystemAdmin}, got.Roles)
	})

	t.Run("an inactive user stays inactive", func(t *testing.T) {
		repo := newRepo(t)
		want := user("1c2d3e4f-5a6b-4c7d-8e9f-0a1b2c3d4e5f", "left@jurwa.example", domain.RoleStudent)
		want.Active = false
		require.NoError(t, repo.Create(ctx, want))

		got, err := repo.FindByEmail(ctx, want.Email)
		require.NoError(t, err)
		assert.False(t, got.Active)
	})

	t.Run("a second account with the same email is refused", func(t *testing.T) {
		repo := newRepo(t)
		require.NoError(t, repo.Create(ctx, user("2d3e4f5a-6b7c-4d8e-9f0a-1b2c3d4e5f60", "pema@rub.edu.bt", domain.RoleStudent)))

		err := repo.Create(ctx, user("3e4f5a6b-7c8d-4e9f-8a1b-2c3d4e5f6071", "pema@rub.edu.bt", domain.RoleStudent))
		assert.ErrorIs(t, err, domain.ErrEmailTaken)
	})

	t.Run("an unknown email or id is not found", func(t *testing.T) {
		repo := newRepo(t)
		_, err := repo.FindByEmail(ctx, "nobody@rub.edu.bt")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)

		_, err = repo.FindByID(ctx, "4f5a6b7c-8d9e-4f0a-9b1c-2d3e4f5a6b7c")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("an id that is not a UUID is not found", func(t *testing.T) {
		repo := newRepo(t)
		_, err := repo.FindByID(ctx, "not-a-uuid")
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

// user returns a valid account. CreatedAt is in UTC and whole microseconds,
// the precision PostgreSQL keeps.
func user(id, email string, roles ...domain.Role) domain.User {
	return domain.User{
		ID:           id,
		Email:        email,
		FullName:     "Test " + email,
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuuQGr4WzHgJ0f7yGq8m3m2bq2Yy5u7Z1i", // a bcrypt-shaped string, not a real password
		Roles:        roles,
		Active:       true,
		CreatedAt:    time.Date(2026, 11, 6, 9, 30, 15, 123456000, time.UTC),
	}
}

func assertSameUser(t *testing.T, want, got domain.User) {
	t.Helper()
	assert.True(t, want.CreatedAt.Equal(got.CreatedAt), "created_at: want %v, got %v", want.CreatedAt, got.CreatedAt)
	got.CreatedAt = want.CreatedAt
	assert.Equal(t, want, got)
}
