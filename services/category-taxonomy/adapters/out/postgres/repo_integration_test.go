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

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/ports"
)

// startDB starts a throwaway PostgreSQL 16 in Docker and returns its URL.
func startDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("category_taxonomy"),
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
	repotest.Run(t, func(t *testing.T) ports.CategoryRepository {
		repo := connect(t, url)
		// Start every check from the seeded rows only.
		_, err := repo.pool.Exec(context.Background(), "TRUNCATE categories")
		require.NoError(t, err)
		require.NoError(t, repo.Migrate(context.Background()))
		return repo
	})
}

// Data survives a restart: a new connection pool sees what the first one saved,
// and running the schema again (as every start does) changes nothing.
func TestDataSurvivesARestart(t *testing.T) {
	ctx := context.Background()
	url := startDB(t)
	first := connect(t, url)
	before, err := first.List(ctx)
	require.NoError(t, err)
	require.Len(t, before, len(domain.DefaultCategories()))
	first.pool.Close()

	second := connect(t, url) // runs Migrate again
	after, err := second.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}
