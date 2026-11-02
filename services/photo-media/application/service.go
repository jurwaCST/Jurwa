package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
	"github.com/jurwaCST/Jurwa/services/photo-media/ports"
)

// Service holds the photo-media use cases.
type Service struct {
	repo  ports.PhotoRepository
	newID func() string
	now   func() time.Time
}

// NewService wires the use cases to their port. newID is usually uuid.NewString
// and now is usually time.Now; tests pass fixed values.
func NewService(repo ports.PhotoRepository, newID func() string, now func() time.Time) *Service {
	return &Service{repo: repo, newID: newID, now: now}
}

// UploadPhoto checks the upload rules, stores the photo and returns it.
// A rule error wraps domain.ErrInvalidPhoto and nothing is stored.
func (s *Service) UploadPhoto(ctx context.Context, in domain.NewPhotoInput) (domain.Photo, error) {
	photo, err := domain.NewPhoto(s.newID(), in, s.now().UTC())
	if err != nil {
		return domain.Photo{}, err
	}
	if err := s.repo.Save(ctx, photo); err != nil {
		return domain.Photo{}, fmt.Errorf("save photo: %w", err)
	}
	return photo, nil
}

// GetPhoto returns the photo with the given id. A malformed id returns
// domain.ErrInvalidID; an unknown id returns domain.ErrPhotoNotFound.
func (s *Service) GetPhoto(ctx context.Context, id string) (domain.Photo, error) {
	if !domain.IsUUID(id) {
		return domain.Photo{}, domain.ErrInvalidID
	}
	return s.repo.FindByID(ctx, id)
}
