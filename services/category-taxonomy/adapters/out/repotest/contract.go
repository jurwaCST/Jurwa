// Package repotest holds the behaviour every ports.CategoryRepository must have.
// The in-memory adapter runs it in its unit tests and the PostgreSQL adapter in
// its integration test, so both store categories the same way.
package repotest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/ports"
)

// Run checks repo behaviour. newRepo must return a repository holding only the
// three default categories (domain.DefaultCategories).
func Run(t *testing.T, newRepo func(t *testing.T) ports.CategoryRepository) {
	ctx := context.Background()

	t.Run("seeded with the three default categories in order", func(t *testing.T) {
		all, err := newRepo(t).List(ctx)
		require.NoError(t, err)
		assert.Equal(t, domain.DefaultCategories(), all, "same names, help text, order and Category Head as the code")
	})

	t.Run("save replaces a category", func(t *testing.T) {
		repo := newRepo(t)
		hostel := domain.DefaultCategories()[0]
		hostel.Name, hostel.Description = "Hostels", "Rooms and bathrooms"
		require.NoError(t, repo.Save(ctx, hostel))

		all, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, all, 3)
		assert.Equal(t, hostel, all[0])
		assert.Equal(t, "DSA", all[0].HeadPosition, "renaming never changes the Category Head")
	})

	t.Run("save of an unknown category is not found and adds nothing", func(t *testing.T) {
		repo := newRepo(t)
		err := repo.Save(ctx, domain.Info{Category: "LIBRARY", Name: "Library", SortOrder: 4, Active: true})
		assert.ErrorIs(t, err, domain.ErrNotFound)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, all, 3)
	})
}
