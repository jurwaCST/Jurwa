package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

const testSecret = "test-secret-at-least-32-bytes-long!!"

// newTestService returns a service with an in-memory store, predictable ids
// and a clock the test can move.
func newTestService(t *testing.T) (*Service, *time.Time) {
	t.Helper()
	now := time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC)
	tokens, err := domain.NewTokenIssuer([]byte(testSecret), 15*time.Minute)
	require.NoError(t, err)
	svc := NewService(memory.NewUsers(), tokens)
	n := 0
	svc.newID = func() string { n++; return fmt.Sprintf("user-%d", n) }
	svc.now = func() time.Time { return now }
	tokens.SetClock(func() time.Time { return now })
	return svc, &now
}

func TestRegister(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, err := svc.Register(ctx, "gayley@rub.edu.bt", "Gayley Choden", "secret123")
	require.NoError(t, err)

	tests := []struct {
		name     string
		email    string
		fullName string
		password string
		wantErr  error
	}{
		{name: "new email is accepted", email: "choki@rub.edu.bt", fullName: "Sonam Choki", password: "secret123"},
		{name: "duplicate email is rejected", email: "gayley@rub.edu.bt", fullName: "Someone Else", password: "secret123", wantErr: domain.ErrEmailTaken},
		{name: "duplicate email in other case is rejected", email: "GAYLEY@rub.edu.bt", fullName: "Someone Else", password: "secret123", wantErr: domain.ErrEmailTaken},
		{name: "invalid email is rejected", email: "gayley", fullName: "Gayley Choden", password: "secret123", wantErr: domain.ErrInvalidEmail},
		{name: "short password is rejected", email: "zangmo@rub.edu.bt", fullName: "Sonam Zangmo", password: "1234567", wantErr: domain.ErrPasswordTooShort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := svc.Register(ctx, tt.email, tt.fullName, tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, u.ID)
			assert.Equal(t, []domain.Role{domain.RoleStudent}, u.Roles)
		})
	}
}

func TestLogin(t *testing.T) {
	svc, now := newTestService(t)
	ctx := context.Background()
	registered, err := svc.Register(ctx, "gayley@rub.edu.bt", "Gayley Choden", "secret123")
	require.NoError(t, err)

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{name: "correct login returns a token", email: "gayley@rub.edu.bt", password: "secret123"},
		{name: "email is matched without case", email: " Gayley@RUB.edu.bt", password: "secret123"},
		{name: "wrong password is rejected", email: "gayley@rub.edu.bt", password: "wrong-password", wantErr: domain.ErrInvalidCredentials},
		{name: "unknown email gets the same error", email: "nobody@rub.edu.bt", password: "secret123", wantErr: domain.ErrInvalidCredentials},
		{name: "invalid email gets the same error", email: "not-an-email", password: "secret123", wantErr: domain.ErrInvalidCredentials},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.Login(ctx, tt.email, tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, res.AccessToken)
			assert.Equal(t, now.Add(15*time.Minute), res.ExpiresAt)
			assert.Equal(t, registered.ID, res.User.ID)
		})
	}
}

func TestValidateToken(t *testing.T) {
	svc, now := newTestService(t)
	ctx := context.Background()
	registered, err := svc.Register(ctx, "gayley@rub.edu.bt", "Gayley Choden", "secret123")
	require.NoError(t, err)
	login, err := svc.Login(ctx, "gayley@rub.edu.bt", "secret123")
	require.NoError(t, err)

	t.Run("valid token returns the user and role", func(t *testing.T) {
		u, err := svc.ValidateToken(ctx, login.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, registered.ID, u.ID)
		assert.Equal(t, "gayley@rub.edu.bt", u.Email)
		assert.Equal(t, []domain.Role{domain.RoleStudent}, u.Roles)
	})

	t.Run("tampered token is rejected", func(t *testing.T) {
		_, err := svc.ValidateToken(ctx, login.AccessToken+"x")
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("token for a deleted user is rejected", func(t *testing.T) {
		other, _ := newTestService(t) // same secret, empty store
		_, err := other.ValidateToken(ctx, login.AccessToken)
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		*now = now.Add(16 * time.Minute)
		_, err := svc.ValidateToken(ctx, login.AccessToken)
		assert.ErrorIs(t, err, domain.ErrInvalidToken)
	})
}
