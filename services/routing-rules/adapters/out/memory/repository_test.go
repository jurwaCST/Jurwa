package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(domain.DefaultRules())

	all, err := repo.List(ctx, false)
	require.NoError(t, err)
	require.Len(t, all, 3)
	for _, r := range all {
		assert.NotEmpty(t, r.ID, "seeded rules get an id")
	}

	hostel, err := repo.FindByCategory(ctx, domain.Hostel)
	require.NoError(t, err)
	require.Len(t, hostel, 1)
	assert.Equal(t, "DSA", hostel[0].TargetPosition)

	// Upserting the same (category, building) keeps the id and replaces the rule.
	updated, err := repo.Upsert(ctx, domain.Rule{Category: domain.Hostel, TargetPosition: "SSO", Priority: 100, Active: true})
	require.NoError(t, err)
	assert.Equal(t, hostel[0].ID, updated.ID)

	// A new building rule gets a new id.
	added, err := repo.Upsert(ctx, domain.Rule{Category: domain.Hostel, Building: "Hostel B", TargetPosition: "DSA", Priority: 100, Active: false})
	require.NoError(t, err)
	assert.NotEqual(t, hostel[0].ID, added.ID)

	active, err := repo.List(ctx, false)
	require.NoError(t, err)
	assert.Len(t, active, 3, "inactive rule is hidden")

	everything, err := repo.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, everything, 4)
}
