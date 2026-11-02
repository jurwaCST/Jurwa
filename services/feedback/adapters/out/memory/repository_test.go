package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

const reportID = "33333333-3333-3333-3333-333333333333"

func TestCreateThenGet(t *testing.T) {
	r := New()
	ctx := context.Background()
	f := domain.Feedback{ID: "f1", ReportID: reportID, Rating: 5}
	require.NoError(t, r.Create(ctx, f))

	got, err := r.GetByReport(ctx, reportID)
	require.NoError(t, err)
	assert.Equal(t, f, got)
}

func TestCreateTwiceIsAlreadyExists(t *testing.T) {
	r := New()
	ctx := context.Background()
	require.NoError(t, r.Create(ctx, domain.Feedback{ID: "f1", ReportID: reportID, Rating: 5}))

	err := r.Create(ctx, domain.Feedback{ID: "f2", ReportID: reportID, Rating: 1})
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)

	got, err := r.GetByReport(ctx, reportID)
	require.NoError(t, err)
	assert.Equal(t, "f1", got.ID, "the first feedback is kept")
}

func TestGetUnknownIsNotFound(t *testing.T) {
	_, err := New().GetByReport(context.Background(), reportID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
