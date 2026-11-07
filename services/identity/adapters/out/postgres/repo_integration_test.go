//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/jurwaCST/Jurwa/services/identity/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/identity/domain"
	"github.com/jurwaCST/Jurwa/services/identity/ports"
)

// startDB starts a throwaway PostgreSQL 16 in Docker and returns its URL.
func startDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("identity"),
		tcpostgres.WithUsername("jurwa"),
		tcpostgres.WithPassword("test-only"),
		tcpostgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)
	url, err := c.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	return url
}

func connect(t *testing.T, url string) *Repository {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	repo := New(pool)
	require.NoError(t, repo.Migrate(context.Background()))
	return repo
}

func TestRepositoryContract(t *testing.T) {
	url := startDB(t)
	repotest.Run(t, func(t *testing.T) ports.UserRepository {
		repo := connect(t, url)
		_, err := repo.pool.Exec(context.Background(), "TRUNCATE users")
		require.NoError(t, err)
		return repo
	})
}

// Emails are unique whatever their case, as the design file says.
func TestEmailIsUniqueIgnoringCase(t *testing.T) {
	ctx := context.Background()
	repo := connect(t, startDB(t))
	u := domain.User{
		ID: "5a6b7c8d-9e0f-4a1b-8c2d-3e4f5a6b7c8d", Email: "pema@rub.edu.bt", FullName: "Pema",
		PasswordHash: "x", Roles: []domain.Role{domain.RoleStudent}, Active: true, CreatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, u))

	u.ID = "6b7c8d9e-0f1a-4b2c-9d3e-4f5a6b7c8d9e"
	u.Email = "Pema@RUB.edu.bt"
	assert.ErrorIs(t, repo.Create(ctx, u), domain.ErrEmailTaken)
}

// Data survives a restart: a new connection pool sees what the first one saved,
// and running the schema again (as every start does) changes nothing.
func TestDataSurvivesARestart(t *testing.T) {
	ctx := context.Background()
	url := startDB(t)
	first := connect(t, url)
	u := domain.User{
		ID: "00000000-0000-4000-8000-000000000030", Email: "president.demo@jurwa.example", FullName: "President (demo)",
		PasswordHash: "x", Roles: []domain.Role{domain.RolePresident}, Active: true,
		CreatedAt: time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, first.Create(ctx, u))
	first.pool.Close()

	second := connect(t, url) // runs Migrate again
	got, err := second.FindByEmail(ctx, u.Email)
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)
	assert.Equal(t, []domain.Role{domain.RolePresident}, got.Roles)
}
