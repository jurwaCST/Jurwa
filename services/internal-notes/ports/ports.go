package ports

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
)

// NoteRepository stores notes. The in-memory adapter is used in Sprint 2;
// a MongoDB adapter (collection "notes") replaces it in Sprint 3.
type NoteRepository interface {
	Save(ctx context.Context, note domain.Note) error
	// ListByReport returns the notes of one report, in any order.
	ListByReport(ctx context.Context, reportID string) ([]domain.Note, error)
}
