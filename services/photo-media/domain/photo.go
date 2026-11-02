package domain

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// MaxPhotoBytes is the largest photo accepted: 5 MB (5 * 1024 * 1024 bytes).
const MaxPhotoBytes = 5 * 1024 * 1024

// The only two content types the report form accepts.
const (
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
)

// signatures are the first bytes ("magic numbers") of each accepted file type.
// They are checked so a file cannot claim to be a JPEG while holding something else.
var signatures = map[string][]byte{
	ContentTypeJPEG: {0xFF, 0xD8, 0xFF},
	ContentTypePNG:  {0x89, 0x50, 0x4E, 0x47},
}

// Photo is one stored image. ReportID is "" until the report is created.
type Photo struct {
	ID          string
	UploaderID  string
	ContentType string
	FileName    string
	Data        []byte
	ReportID    string
	UploadedAt  time.Time
}

// Size returns the photo size in bytes.
func (p Photo) Size() int64 {
	return int64(len(p.Data))
}

// NewPhotoInput is what the student uploads.
type NewPhotoInput struct {
	UploaderID  string
	ContentType string
	Data        []byte
	FileName    string // optional, for display only
}

// ErrInvalidPhoto is wrapped by every upload rule error, so callers can map all
// of them to one response (INVALID_ARGUMENT) with errors.Is.
var ErrInvalidPhoto = errors.New("invalid photo")

var (
	ErrInvalidID       = fmt.Errorf("%w: id must be a UUID", ErrInvalidPhoto)
	ErrMissingUploader = fmt.Errorf("%w: uploader id must be a UUID", ErrInvalidPhoto)
	ErrUnsupportedType = fmt.Errorf("%w: content type must be image/jpeg or image/png", ErrInvalidPhoto)
	ErrContentMismatch = fmt.Errorf("%w: file bytes do not match the content type", ErrInvalidPhoto)
	ErrEmptyPhoto      = fmt.Errorf("%w: photo is empty", ErrInvalidPhoto)
	ErrPhotoTooLarge   = fmt.Errorf("%w: photo is larger than 5 MB", ErrInvalidPhoto)
)

// ErrPhotoNotFound is returned when no photo has the requested id.
var ErrPhotoNotFound = errors.New("photo not found")

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s has the canonical UUID form.
func IsUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// NewPhoto checks the upload rules and returns the photo. The photo keeps its
// own copy of the bytes, so the caller may reuse its buffer.
func NewPhoto(id string, in NewPhotoInput, now time.Time) (Photo, error) {
	if !IsUUID(id) {
		return Photo{}, ErrInvalidID
	}
	if !IsUUID(in.UploaderID) {
		return Photo{}, ErrMissingUploader
	}
	contentType := strings.ToLower(strings.TrimSpace(in.ContentType))
	sig, ok := signatures[contentType]
	if !ok {
		return Photo{}, ErrUnsupportedType
	}
	switch {
	case len(in.Data) == 0:
		return Photo{}, ErrEmptyPhoto
	case len(in.Data) > MaxPhotoBytes:
		return Photo{}, ErrPhotoTooLarge
	case !bytes.HasPrefix(in.Data, sig):
		return Photo{}, ErrContentMismatch
	}
	return Photo{
		ID:          id,
		UploaderID:  in.UploaderID,
		ContentType: contentType,
		FileName:    strings.TrimSpace(in.FileName),
		Data:        bytes.Clone(in.Data),
		UploadedAt:  now,
	}, nil
}
