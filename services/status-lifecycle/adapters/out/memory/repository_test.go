package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := New()
	rs := domain.NewReportStatus("r1", time.Now())

	_, err := repo.Get(ctx, "r1")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	require.NoError(t, repo.Create(ctx, rs))
	assert.ErrorIs(t, repo.Create(ctx, rs), domain.ErrAlreadyExists)

	got, err := repo.Get(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, rs, got)

	require.NoError(t, rs.ChangeTo(domain.Assigned, "", time.Now()))
	assert.ErrorIs(t, repo.Update(ctx, rs, 5), domain.ErrVersionConflict, "wrong expected version")
	require.NoError(t, repo.Update(ctx, rs, 1))

	got, _ = repo.Get(ctx, "r1")
	assert.Equal(t, domain.Assigned, got.Status)

	assert.ErrorIs(t, repo.Update(ctx, domain.NewReportStatus("nope", time.Now()), 1), domain.ErrNotFound)
}
