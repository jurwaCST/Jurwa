// Package repotest holds the behaviour every ports.RuleRepository must have.
// The in-memory adapter runs it in its unit tests and the PostgreSQL adapter in
// its integration test, so both store rules the same way.
package repotest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
	"github.com/jurwaCST/Jurwa/services/routing-rules/ports"
)

// Run checks repo behaviour. newRepo must return a repository holding only the
// three default rules (domain.DefaultRules).
func Run(t *testing.T, newRepo func(t *testing.T) ports.RuleRepository) {
	ctx := context.Background()

	t.Run("seeded with the three default rules", func(t *testing.T) {
		repo := newRepo(t)
		all, err := repo.List(ctx, false)
		require.NoError(t, err)
		require.Len(t, all, 3)
		want := map[domain.Category]string{domain.Hostel: "DSA", domain.AcademicBlock: "DAA", domain.Others: "SSO"}
		for _, r := range all {
			assert.NotEmpty(t, r.ID, "every rule has an id")
			assert.Equal(t, want[r.Category], r.TargetPosition)
			assert.Empty(t, r.Building)
			assert.Equal(t, domain.DefaultPriority, r.Priority)
			assert.True(t, r.Active)
		}
	})

	t.Run("list is ordered by category then building", func(t *testing.T) {
		repo := newRepo(t)
		_, err := repo.Upsert(ctx, domain.Rule{Category: domain.Hostel, Building: "Hostel B", TargetPosition: "SSO", Priority: 100, Active: true})
		require.NoError(t, err)
		all, err := repo.List(ctx, true)
		require.NoError(t, err)
		var got []string
		for _, r := range all {
			got = append(got, string(r.Category)+"/"+r.Building)
		}
		assert.Equal(t, []string{"ACADEMIC_BLOCK/", "HOSTEL/", "HOSTEL/Hostel B", "OTHERS/"}, got)
	})

	t.Run("upsert of the same category and building keeps the id", func(t *testing.T) {
		repo := newRepo(t)
		before, err := repo.FindByCategory(ctx, domain.Hostel)
		require.NoError(t, err)
		require.Len(t, before, 1)

		saved, err := repo.Upsert(ctx, domain.Rule{Category: domain.Hostel, TargetPosition: "SSO", Priority: 50, Active: true})
		require.NoError(t, err)
		assert.Equal(t, before[0].ID, saved.ID)

		after, err := repo.FindByCategory(ctx, domain.Hostel)
		require.NoError(t, err)
		require.Len(t, after, 1, "replaced, not added")
		assert.Equal(t, "SSO", after[0].TargetPosition)
		assert.Equal(t, 50, after[0].Priority)
	})

	t.Run("a new building rule gets a new id and inactive rules are hidden", func(t *testing.T) {
		repo := newRepo(t)
		hostel, err := repo.FindByCategory(ctx, domain.Hostel)
		require.NoError(t, err)

		added, err := repo.Upsert(ctx, domain.Rule{Category: domain.Hostel, Building: "Hostel C", TargetPosition: "DSA", Priority: 100, Active: false})
		require.NoError(t, err)
		assert.NotEmpty(t, added.ID)
		assert.NotEqual(t, hostel[0].ID, added.ID)

		active, err := repo.List(ctx, false)
		require.NoError(t, err)
		assert.Len(t, active, 3, "the inactive rule is hidden")
		everything, err := repo.List(ctx, true)
		require.NoError(t, err)
		assert.Len(t, everything, 4)

		byCategory, err := repo.FindByCategory(ctx, domain.Hostel)
		require.NoError(t, err)
		assert.Len(t, byCategory, 2, "FindByCategory returns inactive rules too")
	})

	t.Run("other categories are not returned", func(t *testing.T) {
		repo := newRepo(t)
		others, err := repo.FindByCategory(ctx, domain.Others)
		require.NoError(t, err)
		require.Len(t, others, 1)
		assert.Equal(t, domain.Others, others[0].Category)
	})
}
