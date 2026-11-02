package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jurwaCST/Jurwa/services/audit-log/domain"
	"github.com/jurwaCST/Jurwa/services/audit-log/ports"
)

// ErrInvalidArgument: a required field is missing or badly formed.
var ErrInvalidArgument = errors.New("invalid argument")

// Service holds the audit-log use cases.
type Service struct {
	store ports.EventStore
	now   func() time.Time
	newID func() string
}

// NewService wires the use cases. now is time.Now and newID is uuid.NewString outside tests.
func NewService(store ports.EventStore, now func() time.Time, newID func() string) *Service {
	return &Service{store: store, now: now, newID: newID}
}

// Append checks and stores one entry, and returns its new id.
func (s *Service) Append(ctx context.Context, in domain.Input) (string, error) {
	e, err := domain.NewEvent(s.newID(), in, s.now().UTC())
	if err != nil {
		return "", err
	}
	if err := s.store.Append(ctx, e); err != nil {
		return "", err
	}
	return e.ID, nil
}

// ListByReport returns a report's entries, oldest first. Only the System
// Administrator may read them. An unknown report gives an empty list.
func (s *Service) ListByReport(ctx context.Context, role domain.Role, reportID string) ([]domain.Event, error) {
	if err := domain.CheckCanList(role); err != nil {
		return nil, err
	}
	if !domain.IsUUID(reportID) {
		return nil, fmt.Errorf("%w: report_id must be a UUID", ErrInvalidArgument)
	}
	return s.store.ListByReport(ctx, reportID)
}
