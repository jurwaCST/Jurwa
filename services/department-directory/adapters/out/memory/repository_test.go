package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	seedDepts := []domain.Department{{Code: "DEPT_B", Name: "B"}, {Code: "DEPT_A", Name: "A"}}
	repo := NewRepository(seedDepts, domain.DefaultPositions())

	depts, err := repo.ListDepartments(ctx)
	require.NoError(t, err)
	require.Len(t, depts, 2)
	assert.Equal(t, "DEPT_A", depts[0].Code, "sorted by code")

	// Changing the returned slice or the seed does not change the repository.
	depts[0].Name = "changed"
	seedDepts[0].Name = "changed"
	again, err := repo.ListDepartments(ctx)
	require.NoError(t, err)
	assert.Equal(t, "A", again[0].Name)
	assert.Equal(t, "B", again[1].Name)

	positions, err := repo.ListPositions(ctx)
	require.NoError(t, err)
	assert.Len(t, positions, len(domain.DefaultPositions()))
	for i := 1; i < len(positions); i++ {
		assert.Less(t, positions[i-1].Code, positions[i].Code)
	}
}
