// Package repotest holds the behaviour every ReportRepository must have. The
// in-memory adapter runs it in its unit test and the MongoDB adapter in its
// integration test, so both are checked against the same rules.
package repotest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
	"github.com/jurwaCST/Jurwa/services/report-submission/ports"
)

// RunRepoContract runs the checks against a new, empty repository from newRepo.
func RunRepoContract(t *testing.T, newRepo func(t *testing.T) ports.ReportRepository) {
	// MongoDB keeps times to the millisecond, so the sample time has no smaller part.
	created := time.Date(2026, 11, 6, 4, 30, 0, 123_000_000, time.UTC)
	photo := "0b9c8d7e-6f5a-4b3c-9d2e-1f0a9b8c7d6e"

	full := domain.Report{
		ID:          "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f",
		ReporterID:  "11111111-2222-4333-8444-555555555555",
		Category:    domain.CategoryHostel,
		Location:    domain.Location{Building: "Hostel B", Wing: "East", Area: "Bathroom 2"},
		Description: "Tap is leaking since Monday",
		PhotoID:     &photo,
		Status:      domain.StatusOpen,
		CreatedAt:   created,
	}

	t.Run("save then find returns the same report", func(t *testing.T) {
		repo := newRepo(t)
		require.NoError(t, repo.Save(context.Background(), full))

		got, err := repo.FindByID(context.Background(), full.ID)
		require.NoError(t, err)
		assert.Equal(t, full, got)
	})

	t.Run("report without photo and optional location parts keeps them empty", func(t *testing.T) {
		repo := newRepo(t)
		r := full
		r.ID = "7a7a7a7a-1b1b-4c2c-8d3d-4e4e4e4e4e4e"
		r.PhotoID = nil
		r.Location = domain.Location{Building: "Academic Block A"}
		r.Category = domain.CategoryAcademicBlock
		require.NoError(t, repo.Save(context.Background(), r))

		got, err := repo.FindByID(context.Background(), r.ID)
		require.NoError(t, err)
		assert.Equal(t, r, got)
		assert.Nil(t, got.PhotoID)
	})

	t.Run("unknown id is ErrReportNotFound", func(t *testing.T) {
		repo := newRepo(t)
		_, err := repo.FindByID(context.Background(), "99999999-9999-4999-8999-999999999999")
		assert.ErrorIs(t, err, domain.ErrReportNotFound)
	})

	t.Run("saving the same id again replaces the report", func(t *testing.T) {
		repo := newRepo(t)
		require.NoError(t, repo.Save(context.Background(), full))
		changed := full
		changed.Description = "Tap is now flooding the floor"
		require.NoError(t, repo.Save(context.Background(), changed))

		got, err := repo.FindByID(context.Background(), full.ID)
		require.NoError(t, err)
		assert.Equal(t, "Tap is now flooding the floor", got.Description)
	})

	t.Run("reports are kept apart", func(t *testing.T) {
		repo := newRepo(t)
		other := full
		other.ID = "8b8b8b8b-2c2c-4d3d-8e4e-5f5f5f5f5f5f"
		other.Description = "Light broken in corridor"
		require.NoError(t, repo.Save(context.Background(), full))
		require.NoError(t, repo.Save(context.Background(), other))

		got, err := repo.FindByID(context.Background(), full.ID)
		require.NoError(t, err)
		assert.Equal(t, full.Description, got.Description)
	})
}
