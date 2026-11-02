package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

// PhotoRepository stores photos. The in-memory adapter is used in Sprint 2;
// a MongoDB GridFS adapter (bucket "photos") replaces it in Sprint 3.
type PhotoRepository interface {
	Save(ctx context.Context, photo domain.Photo) error
	// FindByID returns domain.ErrPhotoNotFound when no photo has the id.
	FindByID(ctx context.Context, id string) (domain.Photo, error)
}
