package spec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	photo "github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

const (
	photoID  = "3f9d6c1e-0a2b-4c3d-8e9f-1a2b3c4d5e6f"
	uploader = reporter
)

var (
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47}
)

// file returns n bytes that start with the given signature (or n plain bytes
// when n is shorter than the signature).
func file(magic []byte, n int) []byte {
	b := bytes.Repeat([]byte{0x00}, n)
	copy(b, magic)
	return b
}

func checkPhoto(t *testing.T, contentType string, data []byte, wantErr error) {
	t.Helper()
	got, err := photo.NewPhoto(photoID, photo.NewPhotoInput{UploaderID: uploader, ContentType: contentType, Data: data}, now)
	if wantErr != nil {
		assert.ErrorIs(t, err, wantErr)
		assert.ErrorIs(t, err, photo.ErrInvalidPhoto, "every upload error maps to INVALID_ARGUMENT")
		return
	}
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), got.Size())
}

// Size: at most 5 MB (5 * 1024 * 1024 bytes). Boundaries 0, 1, smallest real
// file, exactly 5 MB and 5 MB + 1, for both JPG and PNG.
func TestPhotoSizeBoundaries(t *testing.T) {
	const fiveMB = photo.MaxPhotoBytes
	require.Equal(t, 5*1024*1024, fiveMB)

	for _, f := range []struct {
		name, contentType string
		magic             []byte
	}{
		{"JPG", photo.ContentTypeJPEG, jpegMagic},
		{"PNG", photo.ContentTypePNG, pngMagic},
	} {
		tests := []struct {
			name    string
			size    int
			wantErr error
		}{
			{name: "0 bytes (empty)", size: 0, wantErr: photo.ErrEmptyPhoto},
			{name: "1 byte (too short to be a real image)", size: 1, wantErr: photo.ErrContentMismatch},
			{name: "smallest real file (signature only)", size: len(f.magic)},
			{name: "5 MB - 1 byte", size: fiveMB - 1},
			{name: "exactly 5 MB (upper bound)", size: fiveMB},
			{name: "5 MB + 1 byte (just over)", size: fiveMB + 1, wantErr: photo.ErrPhotoTooLarge},
		}
		for _, tt := range tests {
			t.Run(f.name+" "+tt.name, func(t *testing.T) {
				checkPhoto(t, f.contentType, file(f.magic, tt.size), tt.wantErr)
			})
		}
	}
}

// Type: JPG and PNG only, and the bytes must really be that type.
func TestPhotoTypeClasses(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		data        []byte
		wantErr     error
	}{
		{name: "JPG", contentType: "image/jpeg", data: file(jpegMagic, 100)},
		{name: "PNG", contentType: "image/png", data: file(pngMagic, 100)},
		{name: "upper case and spaces in the type", contentType: " IMAGE/PNG ", data: file(pngMagic, 100)},
		{name: "GIF", contentType: "image/gif", data: file([]byte("GIF89a"), 100), wantErr: photo.ErrUnsupportedType},
		{name: "PDF", contentType: "application/pdf", data: file([]byte("%PDF"), 100), wantErr: photo.ErrUnsupportedType},
		{name: "no type", contentType: "", data: file(jpegMagic, 100), wantErr: photo.ErrUnsupportedType},
		{name: "says JPG but holds PNG bytes", contentType: "image/jpeg", data: file(pngMagic, 100), wantErr: photo.ErrContentMismatch},
		{name: "says PNG but holds a PDF", contentType: "image/png", data: file([]byte("%PDF"), 100), wantErr: photo.ErrContentMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkPhoto(t, tt.contentType, tt.data, tt.wantErr)
		})
	}
}
