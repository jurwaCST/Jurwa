package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
	"github.com/jurwaCST/Jurwa/services/photo-media/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/photo-media/application"
	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

const (
	photoID    = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	uploaderID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
)

var jpeg = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}

func newTestHandler() *Handler {
	now := func() time.Time { return time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC) }
	return NewHandler(application.NewService(memory.NewRepository(), func() string { return photoID }, now))
}

// brokenApp fails every call, to check internal errors are hidden.
type brokenApp struct{}

func (brokenApp) UploadPhoto(context.Context, domain.NewPhotoInput) (domain.Photo, error) {
	return domain.Photo{}, errors.New("save photo: mongodb password is wrong")
}

func (brokenApp) GetPhoto(context.Context, string) (domain.Photo, error) {
	return domain.Photo{}, errors.New("find photo: mongodb password is wrong")
}

func TestUploadThenGet(t *testing.T) {
	ctx := context.Background()
	h := newTestHandler()

	up, err := h.UploadPhoto(ctx, &photomediav1.UploadPhotoRequest{UploaderId: uploaderID, ContentType: "image/jpeg", Data: jpeg, FileName: "tap.jpg"})
	require.NoError(t, err)
	assert.Equal(t, photoID, up.GetPhotoId())

	got, err := h.GetPhoto(ctx, &photomediav1.GetPhotoRequest{PhotoId: up.GetPhotoId()})
	require.NoError(t, err)
	assert.Equal(t, photoID, got.GetPhotoId())
	assert.Equal(t, "image/jpeg", got.GetContentType())
	assert.Equal(t, jpeg, got.GetData())
	assert.Equal(t, int64(len(jpeg)), got.GetSizeBytes())
	assert.Equal(t, uploaderID, got.GetUploaderId())
	assert.Nil(t, got.ReportId, "report_id is absent until the report is created")
}

func TestUploadInvalidPhoto(t *testing.T) {
	tests := []struct {
		name string
		req  *photomediav1.UploadPhotoRequest
	}{
		{name: "gif", req: &photomediav1.UploadPhotoRequest{UploaderId: uploaderID, ContentType: "image/gif", Data: []byte("GIF89a")}},
		{name: "empty", req: &photomediav1.UploadPhotoRequest{UploaderId: uploaderID, ContentType: "image/jpeg"}},
		{name: "bytes do not match type", req: &photomediav1.UploadPhotoRequest{UploaderId: uploaderID, ContentType: "image/png", Data: jpeg}},
		{name: "too large", req: &photomediav1.UploadPhotoRequest{UploaderId: uploaderID, ContentType: "image/jpeg", Data: append(jpeg, make([]byte, domain.MaxPhotoBytes)...)}},
		{name: "no uploader", req: &photomediav1.UploadPhotoRequest{ContentType: "image/jpeg", Data: jpeg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTestHandler().UploadPhoto(context.Background(), tt.req)

			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestGetPhotoErrors(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want codes.Code
	}{
		{name: "unknown id", id: photoID, want: codes.NotFound},
		{name: "malformed id", id: "photo-1", want: codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTestHandler().GetPhoto(context.Background(), &photomediav1.GetPhotoRequest{PhotoId: tt.id})

			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}

func TestGetPhotoWithReportID(t *testing.T) {
	reportID := "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	h := NewHandler(fixedApp{photo: domain.Photo{ID: photoID, ReportID: reportID, Data: jpeg}})

	got, err := h.GetPhoto(context.Background(), &photomediav1.GetPhotoRequest{PhotoId: photoID})

	require.NoError(t, err)
	require.NotNil(t, got.ReportId)
	assert.Equal(t, reportID, got.GetReportId())
}

func TestInternalErrorsAreHidden(t *testing.T) {
	h := NewHandler(brokenApp{})

	_, err := h.UploadPhoto(context.Background(), &photomediav1.UploadPhotoRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")

	_, err = h.GetPhoto(context.Background(), &photomediav1.GetPhotoRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")
}

// fixedApp returns one stored photo, for fields the application cannot set yet.
type fixedApp struct{ photo domain.Photo }

func (a fixedApp) UploadPhoto(context.Context, domain.NewPhotoInput) (domain.Photo, error) {
	return a.photo, nil
}

func (a fixedApp) GetPhoto(context.Context, string) (domain.Photo, error) {
	return a.photo, nil
}
