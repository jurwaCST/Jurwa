// Package repotest holds the behaviour every ports.PolicyRepository must have.
// The in-memory adapter runs it in its unit tests and the PostgreSQL adapter in
// its integration test, so both load the same permission table.
package repotest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
	"github.com/jurwaCST/Jurwa/services/authorization/ports"
)

// Run checks repo behaviour. newRepo must return a repository holding the
// default permission table (domain.DefaultPolicy).
func Run(t *testing.T, newRepo func(t *testing.T) ports.PolicyRepository) {
	ctx := context.Background()

	t.Run("loads the default permission table", func(t *testing.T) {
		got, err := newRepo(t).Load(ctx)
		require.NoError(t, err)
		AssertSamePolicy(t, domain.DefaultPolicy(), got)
	})

	t.Run("loading twice gives the same table", func(t *testing.T) {
		repo := newRepo(t)
		first, err := repo.Load(ctx)
		require.NoError(t, err)
		second, err := repo.Load(ctx)
		require.NoError(t, err)
		AssertSamePolicy(t, first, second)
	})
}

// AssertSamePolicy checks that both policies know the same actions and allow
// the same roles for each one.
func AssertSamePolicy(t *testing.T, want, got domain.Policy) {
	t.Helper()
	require.Equal(t, want.Actions(), got.Actions())
	for _, action := range want.Actions() {
		assert.ElementsMatch(t, want.Roles(action), got.Roles(action), action)
	}
}
