package main

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvOr(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{name: "uses value when set", value: "9100", fallback: "9000", want: "9100"},
		{name: "uses fallback when empty", value: "", fallback: "9000", want: "9000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JURWA_TEST_VAR", tt.value)
			assert.Equal(t, tt.want, envOr("JURWA_TEST_VAR", tt.fallback))
		})
	}
}

func TestDialIsLazy(t *testing.T) {
	// Nothing listens on this port: dial must still succeed, calls fail later with Unavailable.
	conn, err := dial("localhost:1")
	assert.NoError(t, err)
	assert.NoError(t, conn.Close())
}

func TestOpenBus(t *testing.T) {
	ctx := context.Background()

	bus, err := openBus(ctx, "")
	require.NoError(t, err)
	assert.Nil(t, bus, "no REDIS_URL: events go to the log and nothing is consumed")

	_, err = openBus(ctx, "not a url")
	assert.Error(t, err)

	_, err = openBus(ctx, "redis://localhost:1/0")
	assert.Error(t, err, "a Redis that does not answer stops the start")

	mr := miniredis.RunT(t)
	bus, err = openBus(ctx, "redis://"+mr.Addr()+"/0")
	require.NoError(t, err)
	require.NotNil(t, bus)
	assert.NoError(t, bus.Close())
}
