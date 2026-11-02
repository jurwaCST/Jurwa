package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
)

const (
	reportA = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	reportB = "2c3d4e5f-6071-4b8c-9d0e-1f2a3b4c5d6e"
)

func TestSaveAndListByReport(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	first := domain.Note{ID: "1", ReportID: reportA, Body: "first"}
	other := domain.Note{ID: "2", ReportID: reportB, Body: "other report"}
	second := domain.Note{ID: "3", ReportID: reportA, Body: "second"}
	for _, n := range []domain.Note{first, other, second} {
		require.NoError(t, repo.Save(ctx, n))
	}

	got, err := repo.ListByReport(ctx, reportA)

	require.NoError(t, err)
	assert.Equal(t, []domain.Note{first, second}, got)
}

func TestListUnknownReportIsEmpty(t *testing.T) {
	got, err := NewRepository().ListByReport(context.Background(), reportA)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestListReturnsACopy(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	require.NoError(t, repo.Save(ctx, domain.Note{ID: "1", ReportID: reportA, Body: "first"}))

	got, err := repo.ListByReport(ctx, reportA)
	require.NoError(t, err)
	got[0].Body = "changed by caller"

	again, err := repo.ListByReport(ctx, reportA)
	require.NoError(t, err)
	assert.Equal(t, "first", again[0].Body)
}
