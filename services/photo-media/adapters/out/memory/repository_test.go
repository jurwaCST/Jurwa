package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

const photoID = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"

func TestSaveAndFind(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	photo := domain.Photo{ID: photoID, ContentType: domain.ContentTypePNG, Data: []byte{0x89, 0x50, 0x4E, 0x47}}

	require.NoError(t, repo.Save(ctx, photo))
	got, err := repo.FindByID(ctx, photoID)

	require.NoError(t, err)
	assert.Equal(t, photo, got)
}

func TestFindUnknownID(t *testing.T) {
	_, err := NewRepository().FindByID(context.Background(), photoID)

	assert.ErrorIs(t, err, domain.ErrPhotoNotFound)
}

func TestStoredBytesCannotBeChangedFromOutside(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	data := []byte{0xFF, 0xD8, 0xFF, 0x00}
	require.NoError(t, repo.Save(ctx, domain.Photo{ID: photoID, Data: data}))

	data[3] = 0x42 // caller changes its buffer after Save
	got, err := repo.FindByID(ctx, photoID)
	require.NoError(t, err)
	got.Data[2] = 0x42 // caller changes the returned copy

	again, err := repo.FindByID(ctx, photoID)
	require.NoError(t, err)
	assert.Equal(t, []byte{0xFF, 0xD8, 0xFF, 0x00}, again.Data)
}
