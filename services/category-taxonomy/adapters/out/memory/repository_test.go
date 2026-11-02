package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(domain.DefaultCategories())

	all, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 3)
	assert.Equal(t, []domain.Category{domain.Hostel, domain.AcademicBlock, domain.Others},
		[]domain.Category{all[0].Category, all[1].Category, all[2].Category}, "sorted by SortOrder")

	renamed := all[0]
	renamed.Name = "Hostels"
	require.NoError(t, repo.Save(ctx, renamed))
	all, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hostels", all[0].Name)

	err = repo.Save(ctx, domain.Info{Category: "LIBRARY", Name: "Library"})
	assert.ErrorIs(t, err, domain.ErrNotFound, "categories are never added")
	all, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)
}
