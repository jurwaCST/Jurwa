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

	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
	"github.com/jurwaCST/Jurwa/services/analytics/ports"
)

// startDB starts a throwaway PostgreSQL 16 in Docker and returns its URL.
func startDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("analytics"),
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
	repotest.Run(t, func(t *testing.T) ports.ProjectionStore {
		repo := connect(t, url)
		_, err := repo.pool.Exec(context.Background(), "TRUNCATE report_facts, processed_events")
		require.NoError(t, err)
		return repo
	})
}

// Data survives a restart: a new connection pool sees what the first one saved,
// and running the schema again (as every start does) changes nothing.
func TestDataSurvivesARestart(t *testing.T) {
	ctx := context.Background()
	url := startDB(t)
	first := connect(t, url)
	e := domain.Event{ID: "demo-01", Type: domain.ReportSubmitted, ReportID: "hostel-1", Category: domain.Hostel, OccurredAt: time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)}
	applied, err := first.Apply(ctx, e)
	require.NoError(t, err)
	require.True(t, applied)
	first.pool.Close()

	second := connect(t, url) // runs Migrate again
	reports, err := second.Reports(ctx)
	require.NoError(t, err)
	require.Len(t, reports, 1)
	assert.Equal(t, "hostel-1", reports[0].ID)

	// The seen event ids are kept too, so replaying the stream after a
	// restart does not count anything twice.
	applied, err = second.Apply(ctx, e)
	require.NoError(t, err)
	assert.False(t, applied)
}
