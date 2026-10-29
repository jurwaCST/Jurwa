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
		{ID: "a1", Category: domain.CategoryAcademicBlock, Location: domain.Location{Building: "Academic Block 1"}, Active: true},
	})

	got, err := repo.ListByCategory(context.Background(), domain.CategoryHostel)

	require.NoError(t, err)
	require.Len(t, got, 2, "inactive entries are kept; the use case filters them")
	assert.Equal(t, "h1", got[0].ID, "order is kept")
	assert.Equal(t, "h2", got[1].ID)
}

func TestListByCategoryUnknownCategoryIsEmpty(t *testing.T) {
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

func TestSeedEntriesAreValid(t *testing.T) {
	seen := map[string]bool{}
	categories := map[domain.Category]int{}
	for _, e := range SeedEntries() {
		assert.True(t, domain.IsUUID(e.ID), "entry id %q is a UUID", e.ID)
		assert.False(t, seen[e.ID], "entry id %q is unique", e.ID)
		seen[e.ID] = true
		assert.NoError(t, domain.ValidateCategory(e.Category))
		assert.NotEmpty(t, e.Location.Building)
		categories[e.Category]++
	}
	assert.Positive(t, categories[domain.CategoryHostel])
	assert.Positive(t, categories[domain.CategoryAcademicBlock])
	assert.Positive(t, categories[domain.CategoryOthers])
}
