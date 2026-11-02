package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/authorization/domain"
)

func TestNewLoadsTheDefaultPolicy(t *testing.T) {
	p, err := New().Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultPolicy().Actions(), p.Actions())
}

func TestWithPolicyLoadsTheGivenPolicy(t *testing.T) {
	custom := domain.NewPolicy(map[string][]domain.Role{domain.ActionMeRead: {domain.Student}})
	p, err := WithPolicy(custom).Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{domain.ActionMeRead}, p.Actions())
}
