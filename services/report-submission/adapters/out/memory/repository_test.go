package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

func TestSaveAndFind(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	report := domain.Report{ID: "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", Status: domain.StatusOpen}

	require.NoError(t, repo.Save(ctx, report))

	got, err := repo.FindByID(ctx, report.ID)
	require.NoError(t, err)
	assert.Equal(t, report, got)

	_, err = repo.FindByID(ctx, "unknown")
	assert.ErrorIs(t, err, domain.ErrReportNotFound)
}
