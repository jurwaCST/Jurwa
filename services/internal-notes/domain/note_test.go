package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	noteID   = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	reportA  = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	reportB  = "2c3d4e5f-6071-4b8c-9d0e-1f2a3b4c5d6e"
	authorID = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
)

var t0 = time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

func validInput() NewNoteInput {
	return NewNoteInput{ReportID: reportA, AuthorID: authorID, Role: RoleStaff, Body: "Plumber booked for Monday."}
}

func TestNewNote(t *testing.T) {
	n, err := NewNote(noteID, validInput(), t0)

	require.NoError(t, err)
	assert.Equal(t, noteID, n.ID)
	assert.Equal(t, reportA, n.ReportID)
	assert.Equal(t, authorID, n.AuthorID)
	assert.Equal(t, "Plumber booked for Monday.", n.Body)
	assert.Equal(t, t0, n.CreatedAt)
}

func TestNewNoteTrimsBody(t *testing.T) {
	in := validInput()
	in.Body = "  Parts ordered.\n"

	n, err := NewNote(noteID, in, t0)

	require.NoError(t, err)
	assert.Equal(t, "Parts ordered.", n.Body)
}

func TestNewNoteRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		change func(*NewNoteInput)
		want   error
	}{
		{name: "bad note id", id: "n-1", change: func(*NewNoteInput) {}, want: ErrInvalidID},
		{name: "missing report id", id: noteID, change: func(in *NewNoteInput) { in.ReportID = "" }, want: ErrInvalidReportID},
		{name: "report id not a uuid", id: noteID, change: func(in *NewNoteInput) { in.ReportID = "report-1" }, want: ErrInvalidReportID},
		{name: "missing author id", id: noteID, change: func(in *NewNoteInput) { in.AuthorID = "" }, want: ErrInvalidAuthorID},
		{name: "empty body", id: noteID, change: func(in *NewNoteInput) { in.Body = "" }, want: ErrBlankBody},
		{name: "blank body", id: noteID, change: func(in *NewNoteInput) { in.Body = " \t\n " }, want: ErrBlankBody},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.change(&in)

			_, err := NewNote(tt.id, in, t0)

			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestNewNoteBodyLengthBoundaries(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error // nil means accepted
	}{
		{name: "1 character", body: "x", want: nil},
		{name: "1000 characters", body: strings.Repeat("x", MaxBodyChars), want: nil},
		{name: "1001 characters", body: strings.Repeat("x", MaxBodyChars+1), want: ErrBodyTooLong},
		{name: "1000 Dzongkha characters count as 1000, not as bytes", body: strings.Repeat("ཀ", MaxBodyChars), want: nil},
		{name: "spaces around 1000 characters are not counted", body: "  " + strings.Repeat("x", MaxBodyChars) + "  ", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			in.Body = tt.body

			_, err := NewNote(noteID, in, t0)

			if tt.want == nil {
				assert.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

// TestWhoMayUseNotes follows docs/design/api-routes.md: Category Head and Staff
// add notes; Category Head, Staff and System Admin read them.
func TestWhoMayUseNotes(t *testing.T) {
	tests := []struct {
		role    Role
		mayAdd  bool
		mayList bool
	}{
		{role: RoleStaff, mayAdd: true, mayList: true},
		{role: RoleCategoryHead, mayAdd: true, mayList: true},
		{role: RoleSystemAdmin, mayAdd: false, mayList: true},
		{role: RoleStudent, mayAdd: false, mayList: false},
		{role: RolePresident, mayAdd: false, mayList: false},
		{role: "", mayAdd: false, mayList: false},
		{role: "JANITOR", mayAdd: false, mayList: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.role)+" add", func(t *testing.T) {
			in := validInput()
			in.Role = tt.role

			_, err := NewNote(noteID, in, t0)

			if tt.mayAdd {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrPermissionDenied)
			}
		})
		t.Run(string(tt.role)+" list", func(t *testing.T) {
			err := CheckCanList(tt.role)

			if tt.mayList {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrPermissionDenied)
			}
		})
	}
}

func TestStudentIsDeniedBeforeInputIsChecked(t *testing.T) {
	in := NewNoteInput{Role: RoleStudent} // everything else is missing too

	_, err := NewNote(noteID, in, t0)

	assert.ErrorIs(t, err, ErrPermissionDenied, "a student learns nothing about the note rules")
}

func TestSortOldestFirst(t *testing.T) {
	at := func(id string, h int) Note { return Note{ID: id, CreatedAt: t0.Add(time.Duration(h) * time.Hour)} }
	notes := []Note{at("c", 2), at("a", 0), at("b", 1), at("a2", 0)}

	SortOldestFirst(notes)

	assert.Equal(t, []Note{at("a", 0), at("a2", 0), at("b", 1), at("c", 2)}, notes)
}

func TestFilterByReport(t *testing.T) {
	a := Note{ID: "1", ReportID: reportA}
	b := Note{ID: "2", ReportID: reportB}

	assert.Equal(t, []Note{a}, FilterByReport([]Note{a, b}, reportA))
	assert.Empty(t, FilterByReport([]Note{b}, reportA))
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID(noteID))
	assert.False(t, IsUUID("note-1"))
}
