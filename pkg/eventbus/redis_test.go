package eventbus_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
)

// TestRedisBusOnMiniredis runs the contract against the Redis bus with miniredis,
// an in-process Redis, so make test needs no Docker. The integration test runs
// the same contract on a real Redis.
func TestRedisBusOnMiniredis(t *testing.T) {
	server := miniredis.RunT(t)
	runBusContract(t, func(t *testing.T) Bus {
		bus, err := eventbus.NewRedis("redis://" + server.Addr() + "/0")
		require.NoError(t, err)
		bus.Block = 50 * time.Millisecond
		bus.RetryDelay = time.Millisecond
		require.NoError(t, bus.Ping(context.Background()))
		t.Cleanup(func() { _ = bus.Close() })
		return bus
	})
}

func TestNewRedisRejectsBadURL(t *testing.T) {
	_, err := eventbus.NewRedis("not-a-url")
	assert.Error(t, err)
}

func TestRedisSubscribeFailsWhenRedisIsDown(t *testing.T) {
	server := miniredis.RunT(t)
	bus, err := eventbus.NewRedis("redis://" + server.Addr() + "/0")
	require.NoError(t, err)
	server.Close()

	err = bus.Subscribe(context.Background(), "s", "g", "c", func(context.Context, []byte) error { return nil })
	assert.Error(t, err)
}

func TestRedisEntryWithoutPayloadGoesToDeadStream(t *testing.T) {
	server := miniredis.RunT(t)
	bus, err := eventbus.NewRedis("redis://" + server.Addr() + "/0")
	require.NoError(t, err)
	bus.Block = 50 * time.Millisecond

	_, err = server.XAdd("broken", "*", []string{"other", "x"})
	require.NoError(t, err)

	called := false
	subscribe(t, bus, "broken", "g1", "c1", func(context.Context, []byte) error {
		called = true
		return nil
	})

	require.Eventually(t, func() bool {
		entries, _ := server.Stream("broken" + eventbus.DeadSuffix)
		return len(entries) == 1
	}, 5*time.Second, 20*time.Millisecond)
	assert.False(t, called, "handler must not get an entry without payload")
}
