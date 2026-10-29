package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

func TestListByCategoryReturnsOnlyThatCategory(t *testing.T) {
	repo := NewRepository([]domain.Entry{
		{ID: "h1", Category: domain.CategoryHostel, Location: domain.Location{Building: "Hostel A"}, Active: true},
		{ID: "h2", Category: domain.CategoryHostel, Location: domain.Location{Building: "Hostel B"}, Active: false},
		{ID: "b1", Category: domain.CategoryAcademicBlock, Location: domain.Location{Building: "Academic Block 1"}, Active: true},
	})

	got, err := repo.ListByCategory(context.Background(), domain.CategoryHostel)

	require.NoError(t, err)
	require.Len(t, got, 2, "inactive rows are kept; the use case filters them")
	assert.Equal(t, "h1", got[0].ID, "order is kept")
	assert.Equal(t, "h2", got[1].ID)
}

func TestListByCategoryWithNoRowsIsEmpty(t *testing.T) {
	repo := NewRepository(nil)

	got, err := repo.ListByCategory(context.Background(), domain.CategoryOthers)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestListByCategoryReturnsACopy(t *testing.T) {
	repo := NewRepository([]domain.Entry{
		{ID: "h1", Category: domain.CategoryHostel, Location: domain.Location{Building: "Hostel A"}, Active: true},
	})

	first, err := repo.ListByCategory(context.Background(), domain.CategoryHostel)
	require.NoError(t, err)
	first[0].Location.Building = "changed by caller"

	second, err := repo.ListByCategory(context.Background(), domain.CategoryHostel)
	require.NoError(t, err)
	assert.Equal(t, "Hostel A", second[0].Location.Building, "callers cannot change the stored list")
}

func TestSampleEntriesFollowTheRules(t *testing.T) {
	seen := map[string]bool{}
	perCategory := map[domain.Category]int{}
	for _, e := range SampleEntries() {
		_, err := domain.NewEntry(e.ID, e.Category, e.Location)
		assert.NoError(t, err, "entry %s follows the hierarchy rules", e.ID)
		assert.False(t, seen[e.ID], "entry id %s is unique", e.ID)
		seen[e.ID] = true
		perCategory[e.Category]++
	}
	assert.Positive(t, perCategory[domain.CategoryHostel])
	assert.Positive(t, perCategory[domain.CategoryAcademicBlock])
	assert.Positive(t, perCategory[domain.CategoryOthers])
}
