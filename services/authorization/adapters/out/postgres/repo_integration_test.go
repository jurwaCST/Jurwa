//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/jurwaCST/Jurwa/services/authorization/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/authorization/domain"
	"github.com/jurwaCST/Jurwa/services/authorization/ports"
)

// startDB starts a throwaway PostgreSQL 16 in Docker and returns its URL.
func startDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("authorization_db"),
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

// The seeded SQL table is the same as domain.DefaultPolicy, so the database
// and the in-memory adapter give the same answers.
func TestRepositoryContract(t *testing.T) {
	url := startDB(t)
	repotest.Run(t, func(t *testing.T) ports.PolicyRepository { return connect(t, url) })
}

// A permission added in the database is loaded on the next Load, and running
// the schema again (as every start does) keeps it.
func TestChangesInTheDatabaseAreLoaded(t *testing.T) {
	ctx := context.Background()
	url := startDB(t)
	repo := connect(t, url)
	_, err := repo.pool.Exec(ctx, `INSERT INTO role_permissions (action, role_code) VALUES ('dashboard.read', 'CATEGORY_HEAD')`)
	require.NoError(t, err)

	again := connect(t, url) // runs Migrate again
	p, err := again.Load(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.Role{domain.SystemAdmin, domain.President, domain.CategoryHead}, p.Roles(domain.ActionDashboardRead))
}
