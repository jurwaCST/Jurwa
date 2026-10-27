package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
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

func TestSeedReports(t *testing.T) {
	app := application.NewService(memory.New(), logpublisher.New(slog.Default()), time.Now)
	seedReports(app, " 11111111-1111-1111-1111-111111111111, not-a-uuid,,")

	rs, err := app.GetStatus(context.Background(), "11111111-1111-1111-1111-111111111111")
	assert.NoError(t, err)
	assert.Equal(t, domain.Open, rs.Status)
}
