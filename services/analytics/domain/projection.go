package domain

import (
	"errors"
	"sort"
	"time"
)

// ErrInvalidPeriod: the period ends before it starts.
var ErrInvalidPeriod = errors.New("period.from must not be after period.to")

// Category matches jurwa.common.v1.Category.
type Category int32

const (
	CategoryUnspecified Category = 0
	Hostel              Category = 1
	AcademicBlock       Category = 2
	Others              Category = 3
)

// Categories in the order the dashboard shows them.
var Categories = []Category{Hostel, AcademicBlock, Others}

func (c Category) String() string {
	switch c {
	case Hostel:
		return "HOSTEL"
	case AcademicBlock:
		return "ACADEMIC_BLOCK"
	case Others:
		return "OTHERS"
	default:
		return "CATEGORY_UNSPECIFIED"
	}
}

func (c Category) valid() bool { return c >= Hostel && c <= Others }

// Status matches jurwa.common.v1.Status.
type Status int32

const (
	StatusUnspecified Status = 0
	Open              Status = 1
	Assigned          Status = 2
	InProgress        Status = 3
	Resolved          Status = 4
	Escalated         Status = 5
)

func (s Status) String() string {
	switch s {
	case Open:
		return "OPEN"
	case Assigned:
		return "ASSIGNED"
	case InProgress:
		return "IN_PROGRESS"
	case Resolved:
		return "RESOLVED"
	case Escalated:
		return "ESCALATED"
	default:
		return "STATUS_UNSPECIFIED"
	}
}

func (s Status) valid() bool { return s >= Open && s <= Escalated }

// EventType says which event this is.
type EventType int32

const (
	ReportSubmitted EventType = 1 // events:report.submitted
	StatusChanged   EventType = 2 // events:status.changed
)

// Event is the part of an event analytics needs. The Redis Streams consumer
// (Sprint 3) turns proto/events/v1 messages into this.
type Event struct {
	ID         string // EventMeta.event_id, used to ignore duplicates
	Type       EventType
	ReportID   string
	Category   Category // ReportSubmitted only
	ToStatus   Status   // StatusChanged only
	OccurredAt time.Time
}

// Report is analytics' own copy of one report: only what the counts need.
type Report struct {
	ID           string
	Category     Category
	Status       Status
	CreatedAt    time.Time
	ResolvedAt   time.Time // zero unless Status is Resolved
	WasEscalated bool
}

// Projection is the state built from events.
type Projection struct {
	reports map[string]Report
	seen    map[string]bool
}

// NewProjection returns an empty projection.
func NewProjection() *Projection {
	return &Projection{reports: make(map[string]Report), seen: make(map[string]bool)}
}

// Apply updates the projection with one event. It returns false, and changes
// nothing, when the event is a repeat or cannot be used (for example an
// unknown status). It never panics on bad events.
func (p *Projection) Apply(e Event) bool {
	if p.seen[e.ID] {
		return false
	}
	current, exists := p.reports[e.ReportID]
	next, ok := Next(current, exists, e)
	if !ok {
		return false
	}
	p.reports[e.ReportID] = next
	p.seen[e.ID] = true
	return true
}

// Next returns the report after event e. current is the stored report and
// exists says whether there is one. ok is false, and the report must not
// change, when the event cannot be used. Repeats are not checked here: the
// store keeps the seen event ids. Both the in-memory and the PostgreSQL store
// use this, so they follow the same rules.
func Next(current Report, exists bool, e Event) (Report, bool) {
	if e.ID == "" || e.ReportID == "" {
		return Report{}, false
	}
	switch e.Type {
	case ReportSubmitted:
		if exists || !e.Category.valid() {
			return Report{}, false
		}
		return Report{ID: e.ReportID, Category: e.Category, Status: Open, CreatedAt: e.OccurredAt}, true
	case StatusChanged:
		if !exists || !e.ToStatus.valid() {
			return Report{}, false
		}
		r := current
		r.Status = e.ToStatus
		r.ResolvedAt = time.Time{}
		if e.ToStatus == Resolved {
			r.ResolvedAt = e.OccurredAt
		}
		if e.ToStatus == Escalated {
			r.WasEscalated = true
		}
		return r, true
	}
	return Report{}, false
}

// Reports returns a copy of every report, sorted by id.
func (p *Projection) Reports() []Report {
	out := make([]Report, 0, len(p.reports))
	for _, r := range p.reports {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Period limits the counts to reports created in it. From is included, To is
// not. nil means no limit on that side.
type Period struct {
	From *time.Time
	To   *time.Time
}

// Validate rejects a period that ends before it starts.
func (p Period) Validate() error {
	if p.From != nil && p.To != nil && p.From.After(*p.To) {
		return ErrInvalidPeriod
	}
	return nil
}

func (p Period) contains(t time.Time) bool {
	if p.From != nil && t.Before(*p.From) {
		return false
	}
	if p.To != nil && !t.Before(*p.To) {
		return false
	}
	return true
}

// Counts are the numbers on the dashboard. Open to Resolved count reports by
// their status now. Escalated counts reports escalated at least once, even if
// they were later resolved, so Total is not always the sum of the others.
type Counts struct {
	Total      int64
	Open       int64
	Assigned   int64
	InProgress int64
	Resolved   int64
	Escalated  int64
}

func (c *Counts) add(r Report) {
	c.Total++
	switch r.Status {
	case Open:
		c.Open++
	case Assigned:
		c.Assigned++
	case InProgress:
		c.InProgress++
	case Resolved:
		c.Resolved++
	}
	if r.WasEscalated {
		c.Escalated++
	}
}

// Summary is GetSummary's answer.
type Summary struct {
	Counts                 Counts
	AverageResolutionHours float64 // 0 when nothing is resolved
}

// Summarize counts the reports created in the period.
func Summarize(reports []Report, period Period) Summary {
	var s Summary
	var resolvedHours float64
	for _, r := range reports {
		if !period.contains(r.CreatedAt) {
			continue
		}
		s.Counts.add(r)
		if r.Status == Resolved {
			resolvedHours += r.ResolvedAt.Sub(r.CreatedAt).Hours()
		}
	}
	if s.Counts.Resolved > 0 {
		s.AverageResolutionHours = resolvedHours / float64(s.Counts.Resolved)
	}
	return s
}

// CategoryCounts is one row of GetCategoryBreakdown.
type CategoryCounts struct {
	Category Category
	Counts   Counts
}

// Breakdown counts the reports created in the period, per category. All three
// categories are always returned, in the order of Categories.
func Breakdown(reports []Report, period Period) []CategoryCounts {
	out := make([]CategoryCounts, len(Categories))
	index := make(map[Category]int, len(Categories))
	for i, c := range Categories {
		out[i].Category = c
		index[c] = i
	}
	for _, r := range reports {
		if !period.contains(r.CreatedAt) {
			continue
		}
		if i, ok := index[r.Category]; ok {
			out[i].Counts.add(r)
		}
	}
	return out
}
