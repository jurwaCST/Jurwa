package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

// Small helpers so each test reads like a list of events.
var evN int

func submitted(report string, c Category, at time.Time) Event {
	evN++
	return Event{ID: fmt.Sprintf("ev-%d", evN), Type: ReportSubmitted, ReportID: report, Category: c, OccurredAt: at}
}

func changed(report string, to Status, at time.Time) Event {
	evN++
	return Event{ID: fmt.Sprintf("ev-%d", evN), Type: StatusChanged, ReportID: report, ToStatus: to, OccurredAt: at}
}

func apply(p *Projection, events ...Event) {
	for _, e := range events {
		p.Apply(e)
	}
}

func TestNoEventsMeansAllZero(t *testing.T) {
	p := NewProjection()
	s := Summarize(p.Reports(), Period{})
	assert.Equal(t, Counts{}, s.Counts)
	assert.Equal(t, 0.0, s.AverageResolutionHours)

	b := Breakdown(p.Reports(), Period{})
	require.Len(t, b, 3, "all three categories are always listed")
	assert.Equal(t, []Category{Hostel, AcademicBlock, Others}, []Category{b[0].Category, b[1].Category, b[2].Category})
	for _, c := range b {
		assert.Equal(t, Counts{}, c.Counts)
	}
}

func TestSummaryAfterASequenceOfEvents(t *testing.T) {
	p := NewProjection()
	apply(p,
		submitted("r1", Hostel, t0),
		submitted("r2", Hostel, t0),
		submitted("r3", AcademicBlock, t0),
		submitted("r4", Others, t0),
		submitted("r5", Others, t0),
		changed("r2", Assigned, t0.Add(time.Hour)),
		changed("r3", Assigned, t0.Add(time.Hour)),
		changed("r3", InProgress, t0.Add(2*time.Hour)),
		changed("r4", Assigned, t0.Add(time.Hour)),
		changed("r4", InProgress, t0.Add(2*time.Hour)),
		changed("r4", Resolved, t0.Add(10*time.Hour)),
		changed("r5", Escalated, t0.Add(50*time.Hour)),
	)

	s := Summarize(p.Reports(), Period{})
	assert.Equal(t, Counts{Total: 5, Open: 1, Assigned: 1, InProgress: 1, Resolved: 1, Escalated: 1}, s.Counts)
	assert.Equal(t, 10.0, s.AverageResolutionHours)
}

func TestEscalatedCountsReportsEscalatedAtLeastOnce(t *testing.T) {
	p := NewProjection()
	apply(p,
		submitted("r1", Hostel, t0),
		changed("r1", Escalated, t0.Add(50*time.Hour)),
		changed("r1", InProgress, t0.Add(51*time.Hour)), // picked up again after escalation
		changed("r1", Resolved, t0.Add(60*time.Hour)),
	)
	s := Summarize(p.Reports(), Period{})
	assert.Equal(t, Counts{Total: 1, Resolved: 1, Escalated: 1}, s.Counts,
		"a report that was escalated and later resolved counts as resolved now and as escalated once")
}

func TestAverageResolutionHours(t *testing.T) {
	p := NewProjection()
	apply(p,
		submitted("r1", Hostel, t0),
		submitted("r2", Hostel, t0),
		submitted("r3", Hostel, t0),
		changed("r1", Resolved, t0.Add(4*time.Hour)),
		changed("r2", Resolved, t0.Add(8*time.Hour)),
		changed("r3", InProgress, t0.Add(100*time.Hour)), // not resolved, not counted
	)
	assert.Equal(t, 6.0, Summarize(p.Reports(), Period{}).AverageResolutionHours)
}

func TestCategoryBreakdown(t *testing.T) {
	p := NewProjection()
	apply(p,
		submitted("h1", Hostel, t0),
		submitted("h2", Hostel, t0),
		submitted("a1", AcademicBlock, t0),
		changed("h2", Assigned, t0.Add(time.Hour)),
		changed("a1", Resolved, t0.Add(time.Hour)),
	)
	b := Breakdown(p.Reports(), Period{})
	assert.Equal(t, []CategoryCounts{
		{Category: Hostel, Counts: Counts{Total: 2, Open: 1, Assigned: 1}},
		{Category: AcademicBlock, Counts: Counts{Total: 1, Resolved: 1}},
		{Category: Others, Counts: Counts{}},
	}, b)
}

func TestRepeatedEventIsIgnored(t *testing.T) {
	p := NewProjection()
	sub := submitted("r1", Hostel, t0)
	assert.True(t, p.Apply(sub))
	assert.False(t, p.Apply(sub), "same event id again")

	ch := changed("r1", Assigned, t0.Add(time.Hour))
	assert.True(t, p.Apply(ch))
	assert.False(t, p.Apply(ch))

	assert.Equal(t, Counts{Total: 1, Assigned: 1}, Summarize(p.Reports(), Period{}).Counts)
}

func TestEventsThatAreIgnored(t *testing.T) {
	tests := []struct {
		name string
		e    Event
	}{
		{"unknown status", Event{ID: "x1", Type: StatusChanged, ReportID: "r1", ToStatus: Status(9), OccurredAt: t0}},
		{"status unspecified", Event{ID: "x2", Type: StatusChanged, ReportID: "r1", ToStatus: StatusUnspecified, OccurredAt: t0}},
		{"status change for an unknown report", Event{ID: "x3", Type: StatusChanged, ReportID: "nobody", ToStatus: Resolved, OccurredAt: t0}},
		{"unknown category", Event{ID: "x4", Type: ReportSubmitted, ReportID: "r2", Category: Category(9), OccurredAt: t0}},
		{"report submitted twice with a new event id", Event{ID: "x5", Type: ReportSubmitted, ReportID: "r1", Category: Others, OccurredAt: t0}},
		{"no event id", Event{Type: ReportSubmitted, ReportID: "r3", Category: Hostel, OccurredAt: t0}},
		{"no report id", Event{ID: "x6", Type: ReportSubmitted, Category: Hostel, OccurredAt: t0}},
		{"unknown event type", Event{ID: "x7", Type: EventType(9), ReportID: "r1", OccurredAt: t0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProjection()
			p.Apply(submitted("r1", Hostel, t0))

			assert.NotPanics(t, func() { assert.False(t, p.Apply(tt.e)) })
			assert.Equal(t, Counts{Total: 1, Open: 1}, Summarize(p.Reports(), Period{}).Counts, "nothing changed")
		})
	}
}

func TestPeriodFilter(t *testing.T) {
	p := NewProjection()
	apply(p,
		submitted("oct31", Hostel, time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)),
		submitted("nov1", Hostel, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)),
		submitted("nov2", Hostel, time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)),
	)
	nov1 := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	nov2 := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		period Period
		want   int64
	}{
		{"no limits", Period{}, 3},
		{"from is included", Period{From: &nov1}, 2},
		{"to is not included", Period{To: &nov1}, 1},
		{"one day", Period{From: &nov1, To: &nov2}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Summarize(p.Reports(), tt.period).Counts.Total)
		})
	}
}

func TestBreakdownUsesThePeriod(t *testing.T) {
	p := NewProjection()
	apply(p,
		submitted("old", Hostel, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)),
		submitted("new", Hostel, time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)),
	)
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	b := Breakdown(p.Reports(), Period{From: &from})
	assert.Equal(t, int64(1), b[0].Counts.Total)
}

func TestPeriodValidate(t *testing.T) {
	a := t0
	b := t0.Add(time.Hour)
	assert.NoError(t, Period{}.Validate())
	assert.NoError(t, Period{From: &a, To: &b}.Validate())
	assert.NoError(t, Period{From: &a, To: &a}.Validate(), "an empty period is allowed")
	assert.ErrorIs(t, Period{From: &b, To: &a}.Validate(), ErrInvalidPeriod)
}

func TestReportsReturnsCopies(t *testing.T) {
	p := NewProjection()
	p.Apply(submitted("r1", Hostel, t0))
	list := p.Reports()
	list[0].Status = Resolved
	assert.Equal(t, Open, p.Reports()[0].Status, "changing the returned list does not change the projection")
}

func TestStringNames(t *testing.T) {
	assert.Equal(t, "HOSTEL", Hostel.String())
	assert.Equal(t, "ACADEMIC_BLOCK", AcademicBlock.String())
	assert.Equal(t, "OTHERS", Others.String())
	assert.Equal(t, "CATEGORY_UNSPECIFIED", Category(9).String())
	assert.Equal(t, "OPEN", Open.String())
	assert.Equal(t, "ASSIGNED", Assigned.String())
	assert.Equal(t, "IN_PROGRESS", InProgress.String())
	assert.Equal(t, "RESOLVED", Resolved.String())
	assert.Equal(t, "ESCALATED", Escalated.String())
	assert.Equal(t, "STATUS_UNSPECIFIED", Status(9).String())
}
