package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/photo-media/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

const (
	photoID    = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	uploaderID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
)

var now = time.Date(2026, 10, 30, 9, 0, 0, 0, time.FixedZone("BTT", 6*60*60))

func newTestService(repo *memory.Repository) *Service {
	return NewService(repo, func() string { return photoID }, func() time.Time { return now })
}

func pngInput() domain.NewPhotoInput {
	return domain.NewPhotoInput{UploaderID: uploaderID, ContentType: domain.ContentTypePNG, Data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D}, FileName: "tap.png"}
}

func TestUploadThenGetReturnsSamePhoto(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(memory.NewRepository())

	uploaded, err := svc.UploadPhoto(ctx, pngInput())
	require.NoError(t, err)
	got, err := svc.GetPhoto(ctx, uploaded.ID)

	require.NoError(t, err)
	assert.Equal(t, photoID, got.ID)
	assert.Equal(t, domain.ContentTypePNG, got.ContentType)
	assert.Equal(t, pngInput().Data, got.Data)
	assert.Equal(t, now.UTC(), got.UploadedAt, "times are stored in UTC")
}

func TestUploadInvalidPhotoStoresNothing(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	in := pngInput()
	in.ContentType = "image/gif"

	_, err := newTestService(repo).UploadPhoto(ctx, in)

	require.ErrorIs(t, err, domain.ErrUnsupportedType)
	_, err = repo.FindByID(ctx, photoID)
	assert.ErrorIs(t, err, domain.ErrPhotoNotFound)
}

func TestGetPhotoErrors(t *testing.T) {
	svc := newTestService(memory.NewRepository())

	_, err := svc.GetPhoto(context.Background(), "photo-1")
	assert.ErrorIs(t, err, domain.ErrInvalidID)

	_, err = svc.GetPhoto(context.Background(), photoID)
	assert.ErrorIs(t, err, domain.ErrPhotoNotFound)
}

// failingRepo fails every Save, to check the error is passed on.
type failingRepo struct{}

func (failingRepo) Save(context.Context, domain.Photo) error { return errors.New("disk full") }

func (failingRepo) FindByID(context.Context, string) (domain.Photo, error) {
	return domain.Photo{}, domain.ErrPhotoNotFound
}

func TestUploadSaveError(t *testing.T) {
	svc := NewService(failingRepo{}, func() string { return photoID }, func() time.Time { return now })

	_, err := svc.UploadPhoto(context.Background(), pngInput())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "save photo")
	assert.NotErrorIs(t, err, domain.ErrInvalidPhoto)
}
