// Package dispatcher builds the daily digest: one in-app message per Category
// Head with the open and escalated reports in their category. It talks to
// other services only through the three interfaces below, so tests use fakes.
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// MaxBodyChars is notification's body limit.
const MaxBodyChars = 200

// Category matches jurwa.common.v1.Category.
type Category int32

const (
	Hostel        Category = 1
	AcademicBlock Category = 2
	Others        Category = 3
)

// Categories in the order digests are sent.
var Categories = []Category{Hostel, AcademicBlock, Others}

func (c Category) String() string {
	switch c {
	case Hostel:
		return "Hostel"
	case AcademicBlock:
		return "Academic Block"
	case Others:
		return "Others"
	default:
		return "Unknown"
	}
}

// HeadPosition is the Category Head position for each category, the same
// routing as in the proposal: Hostel to DSA, Academic Block to DAA, Others to SSO.
var HeadPosition = map[Category]string{
	Hostel:        "DSA",
	AcademicBlock: "DAA",
	Others:        "SSO",
}

// Counts are analytics' StatusCounts for one category.
type Counts struct {
	Total      int64
	Open       int64
	Assigned   int64
	InProgress int64
	Resolved   int64
	Escalated  int64 // escalated at least once
}

// Unresolved is every report not resolved yet, including escalated ones.
func (c Counts) Unresolved() int64 { return c.Total - c.Resolved }

// CountsSource is analytics GetCategoryBreakdown.
type CountsSource interface {
	CategoryCounts(ctx context.Context) (map[Category]Counts, error)
}

// HeadDirectory is department-directory GetStaff: who holds a position.
// It returns "" when nobody holds it yet.
type HeadDirectory interface {
	HolderOf(ctx context.Context, position string) (userID string, err error)
}

// Sender is notification SendNotification.
type Sender interface {
	Send(ctx context.Context, userID, subject, body string) error
}

// Result is what one run returns as JSON.
type Result struct {
	Digests int `json:"digests"` // messages sent
	Errors  int `json:"errors"`  // heads that should have had one but did not
}

// Dispatcher runs the digest.
type Dispatcher struct {
	counts CountsSource
	heads  HeadDirectory
	sender Sender
}

// New wires the dispatcher to its three dependencies.
func New(counts CountsSource, heads HeadDirectory, sender Sender) *Dispatcher {
	return &Dispatcher{counts: counts, heads: heads, sender: sender}
}

var errNoHolder = errors.New("nobody holds this position")

// Run sends one digest to each Category Head who has unresolved reports.
// If analytics cannot be read, nothing is sent and the error is returned.
// A failure for one head is logged and counted, and the others still get
// their digest.
func (d *Dispatcher) Run(ctx context.Context) (Result, error) {
	all, err := d.counts.CategoryCounts(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("read analytics: %w", err)
	}
	var res Result
	for _, c := range Categories {
		counts := all[c]
		if counts.Unresolved() == 0 {
			continue
		}
		if err := d.sendOne(ctx, c, counts); err != nil {
			slog.WarnContext(ctx, "digest not sent", "category", c.String(), "position", HeadPosition[c], "error", err)
			res.Errors++
			continue
		}
		res.Digests++
	}
	return res, nil
}

func (d *Dispatcher) sendOne(ctx context.Context, c Category, counts Counts) error {
	userID, err := d.heads.HolderOf(ctx, HeadPosition[c])
	if err != nil {
		return err
	}
	if userID == "" {
		return errNoHolder
	}
	return d.sender.Send(ctx, userID, "Daily digest: "+c.String(), Body(c, counts))
}

// Body is the message text. Even with the largest int64 numbers it is 189
// characters at most, under MaxBodyChars (tested).
func Body(c Category, n Counts) string {
	return fmt.Sprintf("%s: %d unresolved (%d open, %d assigned, %d in progress), %d escalated. Please check Jurwa.",
		c.String(), n.Unresolved(), n.Open, n.Assigned, n.InProgress, n.Escalated)
}
