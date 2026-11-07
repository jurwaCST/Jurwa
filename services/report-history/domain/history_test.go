package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	reportA   = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	reportB   = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	studentA  = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
	studentB  = "2c3d4e5f-6071-4b8c-9d0e-1f2a3b4c5d6e"
	entryID   = "3d4e5f60-7182-4c9d-8e1f-2a3b4c5d6e7f"
	headActor = "4e5f6071-8293-4d0e-9f2a-3b4c5d6e7f80"
)

var t0 = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

var hostelLoc = Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}

func TestNewSummaryStartsOpen(t *testing.T) {
	s, err := NewSummary(reportA, studentA, CategoryHostel, hostelLoc, t0)

	require.NoError(t, err)
	assert.Equal(t, reportA, s.ReportID)
	assert.Equal(t, studentA, s.ReporterID)
	assert.Equal(t, CategoryHostel, s.Category)
	assert.Equal(t, hostelLoc, s.Location)
	assert.Equal(t, StatusOpen, s.Status)
	assert.Empty(t, s.AssigneeID)
	assert.Equal(t, t0, s.CreatedAt)
	assert.Equal(t, t0, s.UpdatedAt)
}

func TestNewSummaryRejectsBadInput(t *testing.T) {
	tests := []struct {
		name     string
		reportID string
		reporter string
		category Category
		loc      Location
		at       time.Time
		want     error
	}{
		{name: "bad report id", reportID: "r-1", reporter: studentA, category: CategoryHostel, loc: hostelLoc, at: t0, want: ErrInvalidReportID},
		{name: "empty reporter", reportID: reportA, reporter: "", category: CategoryHostel, loc: hostelLoc, at: t0, want: ErrInvalidReporterID},
		{name: "missing category", reportID: reportA, reporter: studentA, category: "", loc: hostelLoc, at: t0, want: ErrInvalidCategory},
		{name: "unknown category", reportID: reportA, reporter: studentA, category: "KITCHEN", loc: hostelLoc, at: t0, want: ErrInvalidCategory},
		{name: "empty building", reportID: reportA, reporter: studentA, category: CategoryHostel, loc: Location{Building: "  "}, at: t0, want: ErrMissingBuilding},
		{name: "zero time", reportID: reportA, reporter: studentA, category: CategoryHostel, loc: hostelLoc, want: ErrMissingTime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSummary(tt.reportID, tt.reporter, tt.category, tt.loc, tt.at)

			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func validEntryInput() EntryInput {
	return EntryInput{
		ReportID:   reportA,
		EventType:  EventStatusChanged,
		FromStatus: StatusOpen,
		ToStatus:   StatusAssigned,
		ActorID:    headActor,
		Summary:    "Assigned to DSA",
		OccurredAt: t0.Add(time.Hour),
	}
}

func TestNewEntry(t *testing.T) {
	e, err := NewEntry(entryID, validEntryInput())

	require.NoError(t, err)
	assert.Equal(t, entryID, e.ID)
	assert.Equal(t, reportA, e.ReportID)
	assert.Equal(t, EventStatusChanged, e.EventType)
	assert.Equal(t, StatusOpen, e.FromStatus)
	assert.Equal(t, StatusAssigned, e.ToStatus)
	assert.Equal(t, headActor, e.ActorID)
	assert.Equal(t, "Assigned to DSA", e.Summary)
}

func TestNewEntrySystemActorAndNoStatusChange(t *testing.T) {
	in := validEntryInput()
	in.EventType = "report.reassigned"
	in.FromStatus, in.ToStatus, in.ActorID = "", "", ""

	e, err := NewEntry(entryID, in)

	require.NoError(t, err)
	assert.Empty(t, e.ActorID, "the system acted")
	assert.Empty(t, e.ToStatus)
}

func TestNewEntryRejectsBadInput(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		change func(*EntryInput)
		want   error
	}{
		{name: "bad entry id", id: "e-1", change: func(*EntryInput) {}, want: ErrInvalidEntryID},
		{name: "bad report id", id: entryID, change: func(in *EntryInput) { in.ReportID = "" }, want: ErrInvalidReportID},
		{name: "missing event type", id: entryID, change: func(in *EntryInput) { in.EventType = " " }, want: ErrMissingEventType},
		{name: "missing summary", id: entryID, change: func(in *EntryInput) { in.Summary = "" }, want: ErrMissingSummary},
		{name: "zero time", id: entryID, change: func(in *EntryInput) { in.OccurredAt = time.Time{} }, want: ErrMissingTime},
		{name: "actor not a uuid", id: entryID, change: func(in *EntryInput) { in.ActorID = "dsa" }, want: ErrInvalidActorID},
		{name: "unknown to status", id: entryID, change: func(in *EntryInput) { in.ToStatus = "CLOSED" }, want: ErrInvalidStatus},
		{name: "unknown from status", id: entryID, change: func(in *EntryInput) { in.FromStatus = "NEW" }, want: ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validEntryInput()
			tt.change(&in)

			_, err := NewEntry(tt.id, in)

			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestApplyStatusChangeUpdatesSummary(t *testing.T) {
	s, err := NewSummary(reportA, studentA, CategoryHostel, hostelLoc, t0)
	require.NoError(t, err)
	e, err := NewEntry(entryID, validEntryInput())
	require.NoError(t, err)

	got := s.Apply(e)

	assert.Equal(t, StatusAssigned, got.Status)
	assert.Equal(t, e.OccurredAt, got.UpdatedAt)
	assert.Equal(t, StatusOpen, s.Status, "Apply returns a new value and leaves s unchanged")
}

func TestApplyKeepsStatusWhenNotAStatusChange(t *testing.T) {
	s, err := NewSummary(reportA, studentA, CategoryHostel, hostelLoc, t0)
	require.NoError(t, err)
	in := validEntryInput()
	in.FromStatus, in.ToStatus = "", ""
	e, err := NewEntry(entryID, in)
	require.NoError(t, err)

	assert.Equal(t, StatusOpen, s.Apply(e).Status)
}

func TestApplyOlderEntryDoesNotMoveSummaryBack(t *testing.T) {
	s, err := NewSummary(reportA, studentA, CategoryHostel, hostelLoc, t0)
	require.NoError(t, err)
	newer := validEntryInput()
	newer.FromStatus, newer.ToStatus, newer.OccurredAt = StatusAssigned, StatusInProgress, t0.Add(2*time.Hour)
	older := validEntryInput() // OPEN -> ASSIGNED, one hour after t0, arrives late
	e1, err := NewEntry(entryID, newer)
	require.NoError(t, err)
	e2, err := NewEntry(entryID, older)
	require.NoError(t, err)

	got := s.Apply(e1).Apply(e2)

	assert.Equal(t, StatusInProgress, got.Status)
	assert.Equal(t, t0.Add(2*time.Hour), got.UpdatedAt)
}

func TestSortTimelineOldestFirst(t *testing.T) {
	at := func(h int) Entry {
		return Entry{Summary: string(rune('a' + h)), OccurredAt: t0.Add(time.Duration(h) * time.Hour)}
	}
	entries := []Entry{at(2), at(0), at(3), at(1)}

	SortTimeline(entries)

	assert.Equal(t, []Entry{at(0), at(1), at(2), at(3)}, entries)
}

func TestSortTimelineKeepsRecordOrderForSameTime(t *testing.T) {
	first := Entry{Summary: "first", OccurredAt: t0}
	second := Entry{Summary: "second", OccurredAt: t0}
	entries := []Entry{first, second}

	SortTimeline(entries)

	assert.Equal(t, []Entry{first, second}, entries)
}

func TestSortNewestFirst(t *testing.T) {
	older := Summary{ReportID: reportA, CreatedAt: t0}
	newer := Summary{ReportID: reportB, CreatedAt: t0.Add(time.Hour)}
	summaries := []Summary{older, newer}

	SortNewestFirst(summaries)

	assert.Equal(t, []Summary{newer, older}, summaries)
}

func TestFilterByReporter(t *testing.T) {
	a := Summary{ReportID: reportA, ReporterID: studentA}
	b := Summary{ReportID: reportB, ReporterID: studentB}

	assert.Equal(t, []Summary{a}, FilterByReporter([]Summary{a, b}, studentA))
	assert.Equal(t, []Summary{b}, FilterByReporter([]Summary{a, b}, studentB))
	assert.Empty(t, FilterByReporter([]Summary{a, b}, "5f607182-93a4-4e1f-8a3b-4c5d6e7f8091"))
}

func TestPaginate(t *testing.T) {
	items := make([]Summary, 45)
	for i := range items {
		items[i] = Summary{Location: Location{Area: string(rune('A' + i))}}
	}
	tests := []struct {
		name      string
		size      int32
		token     string
		wantFrom  int
		wantTo    int
		wantToken string
	}{
		{name: "default size is 20", size: 0, token: "", wantFrom: 0, wantTo: 20, wantToken: "20"},
		{name: "second page", size: 0, token: "20", wantFrom: 20, wantTo: 40, wantToken: "40"},
		{name: "last page has no next token", size: 0, token: "40", wantFrom: 40, wantTo: 45, wantToken: ""},
		{name: "size 1", size: 1, token: "", wantFrom: 0, wantTo: 1, wantToken: "1"},
		{name: "size 100 returns all", size: 100, token: "", wantFrom: 0, wantTo: 45, wantToken: ""},
		{name: "exact end has no next token", size: 5, token: "40", wantFrom: 40, wantTo: 45, wantToken: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, next, err := Paginate(items, tt.size, tt.token)

			require.NoError(t, err)
			assert.Equal(t, items[tt.wantFrom:tt.wantTo], page)
			assert.Equal(t, tt.wantToken, next)
		})
	}
}

func TestPaginateEmptyList(t *testing.T) {
	page, next, err := Paginate(nil, 0, "")

	require.NoError(t, err)
	assert.Empty(t, page)
	assert.Empty(t, next)
}

func TestPaginateRejectsBadInput(t *testing.T) {
	items := make([]Summary, 3)
	tests := []struct {
		name  string
		size  int32
		token string
		want  error
	}{
		{name: "size below 1", size: -1, want: ErrInvalidPageSize},
		{name: "size above 100", size: 101, want: ErrInvalidPageSize},
		{name: "token not a number", token: "abc", want: ErrInvalidPageToken},
		{name: "negative token", token: "-1", want: ErrInvalidPageToken},
		{name: "token past the end", token: "4", want: ErrInvalidPageToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Paginate(items, tt.size, tt.token)

			require.ErrorIs(t, err, tt.want)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID(reportA))
	assert.False(t, IsUUID("report-1"))
}

func TestNewEntryAssignee(t *testing.T) {
	in := validEntryInput()
	in.AssigneeID = headActor
	e, err := NewEntry(entryID, in)
	require.NoError(t, err)
	assert.Equal(t, headActor, e.AssigneeID)

	in.AssigneeID = "dsa"
	_, err = NewEntry(entryID, in)
	assert.ErrorIs(t, err, ErrInvalidAssigneeID)
}

func TestApplySetsAssigneeOnlyWhenTheEntryHasOne(t *testing.T) {
	s, err := NewSummary(reportA, studentA, CategoryHostel, hostelLoc, t0)
	require.NoError(t, err)

	routed := Entry{ReportID: reportA, EventType: EventReportRouted, AssigneeID: headActor, Summary: "Routed to DSA", OccurredAt: t0.Add(time.Minute)}
	s = s.Apply(routed)
	assert.Equal(t, headActor, s.AssigneeID)

	change := Entry{ReportID: reportA, EventType: EventStatusChanged, FromStatus: StatusOpen, ToStatus: StatusAssigned, Summary: "Assigned", OccurredAt: t0.Add(2 * time.Minute)}
	s = s.Apply(change)
	assert.Equal(t, headActor, s.AssigneeID, "a status change keeps the assignee")
	assert.Equal(t, StatusAssigned, s.Status)
}

func TestFilterByAssignee(t *testing.T) {
	a := Summary{ReportID: reportA, AssigneeID: headActor}
	b := Summary{ReportID: reportB, AssigneeID: ""}
	assert.Equal(t, []Summary{a}, FilterByAssignee([]Summary{a, b}, headActor))
	assert.Empty(t, FilterByAssignee([]Summary{b}, headActor))
}

// Events on different streams can be handled in any order: a status change
// recorded after a newer reassignment must still set the status.
func TestApplyStatusAndAssigneeHaveTheirOwnTime(t *testing.T) {
	s, err := NewSummary(reportA, studentA, CategoryHostel, hostelLoc, t0)
	require.NoError(t, err)

	reassigned := Entry{ReportID: reportA, EventType: EventReportReassigned, AssigneeID: headActor, Summary: "Reassigned", OccurredAt: t0.Add(time.Hour)}
	assigned := Entry{ReportID: reportA, EventType: EventStatusChanged, FromStatus: StatusOpen, ToStatus: StatusAssigned, Summary: "Assigned", OccurredAt: t0.Add(2 * time.Minute)}
	olderRouting := Entry{ReportID: reportA, EventType: EventReportRouted, AssigneeID: studentB, Summary: "Routed", OccurredAt: t0.Add(time.Minute)}

	s = s.Apply(reassigned)
	s = s.Apply(assigned)
	s = s.Apply(olderRouting)

	assert.Equal(t, StatusAssigned, s.Status, "the older status change is still applied")
	assert.Equal(t, headActor, s.AssigneeID, "the older routing does not undo the reassignment")
	assert.Equal(t, t0.Add(time.Hour), s.UpdatedAt, "updated_at is the newest event")
}
