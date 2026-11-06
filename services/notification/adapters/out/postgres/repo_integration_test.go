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

	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/notification/domain"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

// startDB starts a throwaway PostgreSQL 16 in Docker and returns its URL.
func startDB(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("notification"),
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
	repotest.Run(t, func(t *testing.T) ports.NotificationRepository {
		repo := connect(t, url)
		_, err := repo.pool.Exec(context.Background(), "TRUNCATE notifications")
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
	n := domain.Notification{
		ID: "00000000-0000-4000-8000-000000000001", RecipientUserID: "22222222-2222-4222-8222-222222222222",
		Channel: domain.InApp, EventType: domain.EventStatusChanged, Subject: "Report updated", Body: "Now In Progress.",
		Status: domain.Sent, CreatedAt: time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, first.Save(ctx, n))
	first.pool.Close()

	second := connect(t, url) // runs Migrate again
	list, _, err := second.ListByRecipient(ctx, n.RecipientUserID, false, 0, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, n.ID, list[0].ID)
}
