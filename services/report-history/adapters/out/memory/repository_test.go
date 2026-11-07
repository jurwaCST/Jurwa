package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

const (
	reportA  = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	reportB  = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	studentA = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
	studentB = "2c3d4e5f-6071-4b8c-9d0e-1f2a3b4c5d6e"
)

var t0 = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

func TestCreateFindUpdateSummary(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	s := domain.Summary{ReportID: reportA, ReporterID: studentA, Status: domain.StatusOpen, CreatedAt: t0}

	require.NoError(t, repo.CreateSummary(ctx, s))
	assert.ErrorIs(t, repo.CreateSummary(ctx, s), domain.ErrReportExists)

	s.Status = domain.StatusAssigned
	require.NoError(t, repo.UpdateSummary(ctx, s))
	got, err := repo.FindSummary(ctx, reportA)
	require.NoError(t, err)
	assert.Equal(t, s, got)
}

func TestUnknownSummary(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()

	_, err := repo.FindSummary(ctx, reportA)
	assert.ErrorIs(t, err, domain.ErrReportNotFound)
	assert.ErrorIs(t, repo.UpdateSummary(ctx, domain.Summary{ReportID: reportA}), domain.ErrReportNotFound)
}

func TestListSummariesByReporter(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	a := domain.Summary{ReportID: reportA, ReporterID: studentA}
	b := domain.Summary{ReportID: reportB, ReporterID: studentB}
	require.NoError(t, repo.CreateSummary(ctx, a))
	require.NoError(t, repo.CreateSummary(ctx, b))

	got, err := repo.ListSummariesByReporter(ctx, studentA)
	require.NoError(t, err)
	assert.Equal(t, []domain.Summary{a}, got)

	got, err = repo.ListSummariesByReporter(ctx, "3d4e5f60-7182-4c9d-8e1f-2a3b4c5d6e7f")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestEntriesKeepAppendOrderPerReport(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	first := domain.Entry{ReportID: reportA, Summary: "first"}
	second := domain.Entry{ReportID: reportA, Summary: "second"}
	other := domain.Entry{ReportID: reportB, Summary: "other report"}
	for _, e := range []domain.Entry{first, other, second} {
		require.NoError(t, repo.AppendEntry(ctx, e))
	}

	got, err := repo.ListEntries(ctx, reportA)
	require.NoError(t, err)
	assert.Equal(t, []domain.Entry{first, second}, got)

	got[0].Summary = "changed by caller"
	again, err := repo.ListEntries(ctx, reportA)
	require.NoError(t, err)
	assert.Equal(t, "first", again[0].Summary, "the caller gets a copy")
}

func TestListSummariesByAssignee(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	dsa := "4e5f6071-8293-4d0e-9f2a-3b4c5d6e7f80"
	a := domain.Summary{ReportID: reportA, ReporterID: studentA, AssigneeID: dsa}
	b := domain.Summary{ReportID: reportB, ReporterID: studentB}
	require.NoError(t, repo.CreateSummary(ctx, a))
	require.NoError(t, repo.CreateSummary(ctx, b))

	got, err := repo.ListSummariesByAssignee(ctx, dsa)
	require.NoError(t, err)
	assert.Equal(t, []domain.Summary{a}, got)

	got, err = repo.ListSummariesByAssignee(ctx, studentA)
	require.NoError(t, err)
	assert.Empty(t, got)
}
