package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-history/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-history/application"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
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

func TestSeedSample(t *testing.T) {
	ctx := context.Background()
	app := application.NewService(memory.NewRepository(), uuid.NewString)

	require.NoError(t, seedSample(ctx, app))

	reports, _, err := app.ListReportsByReporter(ctx, sampleStudentID, 0, "")
	require.NoError(t, err)
	require.Len(t, reports, 2)
	assert.Equal(t, sampleReport2, reports[0].ReportID, "newest first")
	assert.Equal(t, domain.StatusInProgress, reports[1].Status)

	timeline, err := app.GetReportTimeline(ctx, sampleReport1)
	require.NoError(t, err)
	assert.Len(t, timeline, 4, "submitted, routed, assigned, in progress")

	assigned, _, err := app.ListReportsByAssignee(ctx, sampleDSAID, 0, "")
	require.NoError(t, err)
	require.Len(t, assigned, 1)
	assert.Equal(t, sampleReport1, assigned[0].ReportID, "the sample DSA has one report")

	assert.Error(t, seedSample(ctx, app), "seeding twice is rejected")
}
