package main

import (
	"context"
	"time"

	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/analytics/application"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestDemoEventsGiveTheDocumentedCounts(t *testing.T) {
	app := application.NewService(memory.New())
	events := demoEvents(time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC))
	assert.Equal(t, len(events), seed(app, events), "every demo event is used")
	assert.Equal(t, 0, seed(app, events), "seeding twice changes nothing")

	s, err := app.Summary(context.Background(), domain.Period{})
	assert.NoError(t, err)
	assert.Equal(t, domain.Counts{Total: 6, Open: 1, Assigned: 1, InProgress: 1, Resolved: 2, Escalated: 1}, s.Counts)
	assert.Equal(t, 9.0, s.AverageResolutionHours, "(6 + 12) / 2")
}
