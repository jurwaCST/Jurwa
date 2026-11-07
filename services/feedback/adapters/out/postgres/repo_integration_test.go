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

	"github.com/jurwaCST/Jurwa/services/feedback/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/feedback/domain"
	"github.com/jurwaCST/Jurwa/services/feedback/ports"
)

// startDB starts a throwaway PostgreSQL 16 in Docker and returns its URL.
func startDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("feedback"),
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
	repotest.Run(t, func(t *testing.T) ports.FeedbackRepository {
		repo := connect(t, url)
		_, err := repo.pool.Exec(context.Background(), "TRUNCATE feedback")
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
	f := domain.Feedback{
		ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", ReportID: "33333333-3333-4333-8333-333333333333",
		StudentUserID: "22222222-2222-4222-8222-222222222222", Rating: 5, CreatedAt: time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, first.Create(ctx, f))
	first.pool.Close()

	second := connect(t, url) // runs Migrate again
	got, err := second.GetByReport(ctx, f.ReportID)
	require.NoError(t, err)
	assert.Equal(t, f.ID, got.ID)
	assert.Equal(t, int32(5), got.Rating)
}
