package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/internal-notes/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/internal-notes/domain"
)

const (
	reportA  = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	reportB  = "2c3d4e5f-6071-4b8c-9d0e-1f2a3b4c5d6e"
	authorID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
)

var t0 = time.Date(2026, 11, 1, 9, 0, 0, 0, time.FixedZone("BTT", 6*60*60))

// testService returns a service whose clock moves one minute per note.
func testService(repo *memory.Repository) *Service {
	n := 0
	ids := func() string {
		n++
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", n)
	}
	clock := func() time.Time { return t0.Add(time.Duration(n) * time.Minute) }
	return NewService(repo, ids, clock)
}

func note(reportID string, role domain.Role, body string) domain.NewNoteInput {
	return domain.NewNoteInput{ReportID: reportID, AuthorID: authorID, Role: role, Body: body}
}

func TestAddThenListOldestFirstAndOnlyThatReport(t *testing.T) {
	ctx := context.Background()
	svc := testService(memory.NewRepository())
	for _, in := range []domain.NewNoteInput{
		note(reportA, domain.RoleCategoryHead, "Forwarded to the maintenance councillor."),
		note(reportB, domain.RoleStaff, "Other report."),
		note(reportA, domain.RoleStaff, "Plumber booked for Monday."),
	} {
		_, err := svc.AddNote(ctx, in)
		require.NoError(t, err)
	}

	notes, err := svc.ListNotes(ctx, domain.RoleSystemAdmin, reportA)

	require.NoError(t, err)
	require.Len(t, notes, 2)
	assert.Equal(t, "Forwarded to the maintenance councillor.", notes[0].Body)
	assert.Equal(t, "Plumber booked for Monday.", notes[1].Body)
	assert.True(t, notes[0].CreatedAt.Before(notes[1].CreatedAt))
	assert.Equal(t, time.UTC, notes[0].CreatedAt.Location(), "times are stored in UTC")
}

func TestStudentCannotAddOrList(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	svc := testService(repo)

	_, err := svc.AddNote(ctx, note(reportA, domain.RoleStudent, "Any update?"))
	assert.ErrorIs(t, err, domain.ErrPermissionDenied)
	stored, _ := repo.ListByReport(ctx, reportA)
	assert.Empty(t, stored, "nothing is stored")

	_, err = svc.ListNotes(ctx, domain.RoleStudent, reportA)
	assert.ErrorIs(t, err, domain.ErrPermissionDenied)
}

func TestAddNoteInvalidInput(t *testing.T) {
	_, err := testService(memory.NewRepository()).AddNote(context.Background(), note(reportA, domain.RoleStaff, "   "))

	assert.ErrorIs(t, err, domain.ErrBlankBody)
}

func TestListNotesErrors(t *testing.T) {
	svc := testService(memory.NewRepository())

	_, err := svc.ListNotes(context.Background(), domain.RoleStaff, "report-1")
	assert.ErrorIs(t, err, domain.ErrInvalidReportID)

	notes, err := svc.ListNotes(context.Background(), domain.RoleStaff, reportA)
	require.NoError(t, err)
	assert.Empty(t, notes, "a report with no notes gives an empty list")
}

// brokenRepo fails every call, to check storage errors are passed on.
type brokenRepo struct{}

var errDB = errors.New("database is down")

func (brokenRepo) Save(context.Context, domain.Note) error { return errDB }

func (brokenRepo) ListByReport(context.Context, string) ([]domain.Note, error) { return nil, errDB }

func TestStorageErrorsArePassedOn(t *testing.T) {
	svc := NewService(brokenRepo{}, func() string { return reportA }, func() time.Time { return t0 })

	_, err := svc.AddNote(context.Background(), note(reportA, domain.RoleStaff, "Hello"))
	assert.ErrorIs(t, err, errDB)

	_, err = svc.ListNotes(context.Background(), domain.RoleStaff, reportA)
	assert.ErrorIs(t, err, errDB)
}
