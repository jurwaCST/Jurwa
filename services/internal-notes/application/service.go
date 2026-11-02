package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
	"github.com/jurwaCST/Jurwa/services/internal-notes/ports"
)

// Service holds the internal-notes use cases.
type Service struct {
	repo  ports.NoteRepository
	newID func() string
	now   func() time.Time
}

// NewService wires the use cases to their port. newID is usually uuid.NewString
// and now is usually time.Now; tests pass fixed values.
func NewService(repo ports.NoteRepository, newID func() string, now func() time.Time) *Service {
	return &Service{repo: repo, newID: newID, now: now}
}

// AddNote checks the role and the note, stores it and returns it.
// Returns domain.ErrPermissionDenied for a role that may not add notes, or an
// error wrapping domain.ErrInvalidInput; nothing is stored in either case.
func (s *Service) AddNote(ctx context.Context, in domain.NewNoteInput) (domain.Note, error) {
	note, err := domain.NewNote(s.newID(), in, s.now().UTC())
	if err != nil {
		return domain.Note{}, err
	}
	if err := s.repo.Save(ctx, note); err != nil {
		return domain.Note{}, fmt.Errorf("save note: %w", err)
	}
	return note, nil
}

// ListNotes returns the notes of one report, oldest first. An unknown report
// has no notes, so the list is empty, not an error.
func (s *Service) ListNotes(ctx context.Context, role domain.Role, reportID string) ([]domain.Note, error) {
	if err := domain.CheckCanList(role); err != nil {
		return nil, err
	}
	if !domain.IsUUID(reportID) {
		return nil, domain.ErrInvalidReportID
	}
	notes, err := s.repo.ListByReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	notes = domain.FilterByReport(notes, reportID)
	domain.SortOldestFirst(notes)
	return notes, nil
}
