//go:build integration

package eventbus_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
)

// TestRedisBus runs the same contract as the in-memory bus against a real Redis.
// By default it starts Redis in a container (Docker must be running). Set
// REDIS_TEST_URL to use a Redis that is already running, for example the one in
// dev/docker-compose.yml.
func TestRedisBus(t *testing.T) {
	url := os.Getenv("REDIS_TEST_URL")
	if url == "" {
		ctx := context.Background()
		container, err := tcredis.Run(ctx, "redis:7-alpine")
		require.NoError(t, err)
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		url, err = container.ConnectionString(ctx)
		require.NoError(t, err)
	}
	t.Logf("redis: %s", hidePassword(url))

	runBusContract(t, func(t *testing.T) Bus {
		bus, err := eventbus.NewRedis(url)
		require.NoError(t, err)
		bus.Block = 200 * time.Millisecond
		bus.RetryDelay = 10 * time.Millisecond
		require.NoError(t, bus.Ping(context.Background()))
		t.Cleanup(func() { _ = bus.Close() })
		return bus
	})
}

func hidePassword(url string) string {
	for i := 0; i < len(url); i++ {
		if url[i] == '@' {
			return "redis://***" + url[i:]
		}
	}
	return url
}
