// Package memory is an in-memory PhotoRepository. Data is lost on restart.
// It is used in Sprint 2 and in tests; MongoDB GridFS replaces it in Sprint 3.
package memory

import (
	"bytes"
	"context"
	"sync"

	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

// Repository keeps photos in a map. It is safe for concurrent use.
type Repository struct {
	mu     sync.RWMutex
	photos map[string]domain.Photo
}

// NewRepository returns an empty repository.
func NewRepository() *Repository {
	return &Repository{photos: make(map[string]domain.Photo)}
}

// Save stores a copy of the photo, replacing any photo with the same id.
func (r *Repository) Save(_ context.Context, photo domain.Photo) error {
	photo.Data = bytes.Clone(photo.Data)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.photos[photo.ID] = photo
	return nil
}

// FindByID returns a copy of the photo or domain.ErrPhotoNotFound.
func (r *Repository) FindByID(_ context.Context, id string) (domain.Photo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	photo, ok := r.photos[id]
	if !ok {
		return domain.Photo{}, domain.ErrPhotoNotFound
	}
	photo.Data = bytes.Clone(photo.Data)
	return photo, nil
}
