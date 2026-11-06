//go:build integration

package mongo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tcmongo "github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/report-submission/ports"
)

// TestMongoRepositoryContract runs the same checks as the in-memory adapter on a
// real MongoDB. It starts MongoDB in a container (Docker must be running), or uses
// MONGO_TEST_URL when it is set (for example the MongoDB of dev/docker-compose.yml).
func TestMongoRepositoryContract(t *testing.T) {
	ctx := context.Background()
	url := os.Getenv("MONGO_TEST_URL")
	if url == "" {
		container, err := tcmongo.Run(ctx, "mongo:7")
		require.NoError(t, err)
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		url, err = container.ConnectionString(ctx)
		require.NoError(t, err)
	}

	run := 0
	repotest.RunRepoContract(t, func(t *testing.T) ports.ReportRepository {
		run++
		// A new database per check, so the checks do not see each other's data.
		db := fmt.Sprintf("report_submission_test_%d_%d", time.Now().UnixNano(), run)
		repo, err := Connect(ctx, url, db)
		require.NoError(t, err)
		require.NoError(t, repo.EnsureIndexes(ctx))
		t.Cleanup(func() {
			_ = repo.DropDatabase(context.Background())
			_ = repo.Close(context.Background())
		})
		return repo
	})
}

func TestConnectFailsOnBadURL(t *testing.T) {
	_, err := Connect(context.Background(), "not-a-mongo-url", "x")
	require.Error(t, err)
}
