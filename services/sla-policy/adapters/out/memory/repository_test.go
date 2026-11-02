package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := New()

	_, err := repo.Get(ctx, domain.Hostel)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	p, _ := domain.NewPolicy(domain.Hostel, 48, "admin", time.Now())
	require.NoError(t, repo.Save(ctx, p))
	got, err := repo.Get(ctx, domain.Hostel)
	require.NoError(t, err)
	assert.Equal(t, int32(48), got.WindowHours)

	// Saving again replaces the window (upsert).
	p.WindowHours = 24
	require.NoError(t, repo.Save(ctx, p))
	got, _ = repo.Get(ctx, domain.Hostel)
	assert.Equal(t, int32(24), got.WindowHours)

	_, err = repo.Get(ctx, domain.Others)
	assert.ErrorIs(t, err, domain.ErrNotFound, "other categories are unaffected")
}
