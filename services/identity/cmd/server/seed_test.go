package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/identity/application"
	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

func newApp(t *testing.T) (*application.Service, *memory.Users) {
	t.Helper()
	tokens, err := domain.NewTokenIssuer([]byte("test-secret-at-least-32-bytes-long!!"), 15*time.Minute)
	require.NoError(t, err)
	users := memory.NewUsers()
	return application.NewService(users, tokens), users
}

func TestSeedFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantCreated int
		wantErr     string
	}{
		{name: "flag not set: nothing is created", env: map[string]string{}, wantCreated: 0},
		{name: "flag false: nothing is created", env: map[string]string{"SEED_DEMO_USERS": "false", "SEED_DEMO_PASSWORD": "demo-pass-123"}, wantCreated: 0},
		{name: "flag true: six users", env: map[string]string{"SEED_DEMO_USERS": "true", "SEED_DEMO_PASSWORD": "demo-pass-123"}, wantCreated: 6},
		{name: "flag true without a password fails", env: map[string]string{"SEED_DEMO_USERS": "true"}, wantErr: "SEED_DEMO_PASSWORD"},
		{name: "flag true with a short password fails", env: map[string]string{"SEED_DEMO_USERS": "true", "SEED_DEMO_PASSWORD": "short"}, wantErr: "SEED_DEMO_PASSWORD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, users := newApp(t)
			created, err := seedFromEnv(context.Background(), app, func(k string) string { return tt.env[k] })
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				if pw := tt.env["SEED_DEMO_PASSWORD"]; pw != "" {
					assert.NotContains(t, err.Error(), pw, "the password is never in the message")
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCreated, created)
			_, findErr := users.FindByEmail(context.Background(), "dsa.demo@jurwa.example")
			if tt.wantCreated == 0 {
				assert.ErrorIs(t, findErr, domain.ErrUserNotFound)
			} else {
				assert.NoError(t, findErr)
			}
		})
	}
}
