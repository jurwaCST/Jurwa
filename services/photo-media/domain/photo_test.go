package domain

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	photoID    = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	uploaderID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
)

var (
	now     = time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	jpegSig = []byte{0xFF, 0xD8, 0xFF}
	pngSig  = []byte{0x89, 0x50, 0x4E, 0x47}
	gifSig  = []byte("GIF89a")
)

// image returns a file of exactly size bytes that starts with sig.
func image(sig []byte, size int) []byte {
	data := make([]byte, size)
	copy(data, sig)
	return data
}

func validInput() NewPhotoInput {
	return NewPhotoInput{UploaderID: uploaderID, ContentType: ContentTypeJPEG, Data: image(jpegSig, 64), FileName: "tap.jpg"}
}

func TestNewPhotoAcceptsJPEGAndPNG(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		data        []byte
	}{
		{name: "jpeg", contentType: ContentTypeJPEG, data: image(jpegSig, 64)},
		{name: "png", contentType: ContentTypePNG, data: image(pngSig, 64)},
		{name: "content type is case insensitive", contentType: "IMAGE/PNG", data: image(pngSig, 64)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			in.ContentType = tt.contentType
			in.Data = tt.data

			p, err := NewPhoto(photoID, in, now)

			require.NoError(t, err)
			assert.Equal(t, photoID, p.ID)
			assert.Equal(t, uploaderID, p.UploaderID)
			assert.Equal(t, tt.data, p.Data)
			assert.Equal(t, int64(len(tt.data)), p.Size())
			assert.Equal(t, now, p.UploadedAt)
			assert.Empty(t, p.ReportID)
		})
	}
}

func TestNewPhotoRejectsWrongType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		data        []byte
		want        error
	}{
		{name: "gif", contentType: "image/gif", data: image(gifSig, 64), want: ErrUnsupportedType},
		{name: "pdf", contentType: "application/pdf", data: image([]byte("%PDF"), 64), want: ErrUnsupportedType},
		{name: "missing content type", contentType: "", data: image(jpegSig, 64), want: ErrUnsupportedType},
		{name: "says jpeg but bytes are png", contentType: ContentTypeJPEG, data: image(pngSig, 64), want: ErrContentMismatch},
		{name: "says png but bytes are gif", contentType: ContentTypePNG, data: image(gifSig, 64), want: ErrContentMismatch},
		{name: "shorter than the signature", contentType: ContentTypePNG, data: []byte{0x89, 0x50}, want: ErrContentMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			in.ContentType = tt.contentType
			in.Data = tt.data

			_, err := NewPhoto(photoID, in, now)

			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidPhoto)
		})
	}
}

func TestNewPhotoSizeBoundaries(t *testing.T) {
	tests := []struct {
		name string
		size int
		want error // nil means accepted
	}{
		{name: "0 bytes", size: 0, want: ErrEmptyPhoto},
		{name: "tiny file", size: 4, want: nil},
		{name: "exactly 5 MB", size: MaxPhotoBytes, want: nil},
		{name: "5 MB plus 1 byte", size: MaxPhotoBytes + 1, want: ErrPhotoTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			in.Data = image(jpegSig, tt.size)

			_, err := NewPhoto(photoID, in, now)

			if tt.want == nil {
				assert.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidPhoto)
		})
	}
}

func TestNewPhotoChecksIDs(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		uploader string
		want     error
	}{
		{name: "bad photo id", id: "photo-1", uploader: uploaderID, want: ErrInvalidID},
		{name: "missing uploader", id: photoID, uploader: "", want: ErrMissingUploader},
		{name: "uploader not a uuid", id: photoID, uploader: "student-7", want: ErrMissingUploader},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			in.UploaderID = tt.uploader

			_, err := NewPhoto(tt.id, in, now)

			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidPhoto)
		})
	}
}

func TestNewPhotoKeepsItsOwnCopyOfTheBytes(t *testing.T) {
	in := validInput()
	original := bytes.Clone(in.Data)

	p, err := NewPhoto(photoID, in, now)
	require.NoError(t, err)
	in.Data[10] = 0x42 // the caller reuses its buffer

	assert.Equal(t, original, p.Data)
}

func TestNewPhotoStoresTheCanonicalContentType(t *testing.T) {
	in := validInput()
	in.ContentType = " Image/JPEG "

	p, err := NewPhoto(photoID, in, now)

	require.NoError(t, err)
	assert.Equal(t, ContentTypeJPEG, p.ContentType)
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID(photoID))
	assert.False(t, IsUUID(""))
	assert.False(t, IsUUID("not-a-uuid"))
}
