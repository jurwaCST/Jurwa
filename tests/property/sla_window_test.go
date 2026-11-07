// Package property checks the SLA rules against random inputs with gopter
// (SCRUM-107). Each property runs on at least 100 generated cases.
// Results and reasons: docs/testing/property-tests.md.
package property

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	policy "github.com/jurwaCST/Jurwa/services/sla-policy/domain"
	tracking "github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
)

const reportID = "11111111-1111-1111-1111-111111111111"

// base is the start of the two-year range the start times are drawn from.
var base = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

const twoYears = int64(2 * 365 * 24 * time.Hour / time.Second)

// startGen gives a start time anywhere in 2026 to 2027, to the second.
func startGen() gopter.Gen {
	return gen.Int64Range(0, twoYears).Map(func(s int64) time.Time { return base.Add(time.Duration(s) * time.Second) })
}

// categoryGen gives one of the three report categories (HOSTEL, ACADEMIC_BLOCK, OTHERS).
func categoryGen() gopter.Gen {
	return gen.Int32Range(1, 3)
}

// hoursGen gives an SLA window from 1 hour to 30 days, as sla-policy accepts.
func hoursGen() gopter.Gen {
	return gen.Int32Range(1, 30*24)
}

func parameters() *gopter.TestParameters {
	p := gopter.DefaultTestParameters()
	p.MinSuccessfulTests = 100
	return p
}

// timerFor starts a timer with the window sla-policy gives for the category.
func timerFor(t *testing.T, category, hours int32, start time.Time) (tracking.Timer, time.Duration) {
	t.Helper()
	pol, err := policy.NewPolicy(policy.Category(category), hours, "22222222-2222-2222-2222-222222222222", start)
	if err != nil {
		t.Fatalf("policy %d %dh: %v", category, hours, err)
	}
	tm, err := tracking.NewTimer(reportID, tracking.Category(category), pol.Window(), start)
	if err != nil {
		t.Fatalf("timer: %v", err)
	}
	return tm, pol.Window()
}

func TestSLAWindowProperties(t *testing.T) {
	props := gopter.NewProperties(parameters())

	props.Property("1. the deadline is never before the start time", prop.ForAll(
		func(start time.Time, category, hours int32) bool {
			tm, _ := timerFor(t, category, hours, start)
			return !tm.Deadline.Before(tm.StartedAt)
		},
		startGen(), categoryGen(), hoursGen(),
	))

	props.Property("2. deadline minus start equals the policy window for the category", prop.ForAll(
		func(start time.Time, category, hours int32) bool {
			tm, window := timerFor(t, category, hours, start)
			return tm.Deadline.Sub(tm.StartedAt) == window && window == time.Duration(hours)*time.Hour
		},
		startGen(), categoryGen(), hoursGen(),
	))

	// now is the deadline moved by -3 to +3 days, to the second, so cases fall
	// before, at and after the deadline.
	props.Property("3. a timer is due exactly when now is at or after the deadline", prop.ForAll(
		func(start time.Time, category, hours int32, offset int64) bool {
			tm, _ := timerFor(t, category, hours, start)
			now := tm.Deadline.Add(time.Duration(offset) * time.Second)
			return tm.IsDue(now) == !now.Before(tm.Deadline)
		},
		startGen(), categoryGen(), hoursGen(), gen.Int64Range(-3*24*3600, 3*24*3600),
	))

	props.Property("3b. Lapse succeeds exactly when the timer is due, and only once", prop.ForAll(
		func(start time.Time, category, hours int32, offset int64) bool {
			tm, _ := timerFor(t, category, hours, start)
			now := tm.Deadline.Add(time.Duration(offset) * time.Second)
			due := tm.IsDue(now)
			err := tm.Lapse(now)
			if due != (err == nil) {
				return false
			}
			if due {
				return tm.State == tracking.Lapsed && tm.Lapse(now) != nil // not twice
			}
			return tm.State == tracking.Running
		},
		startGen(), categoryGen(), hoursGen(), gen.Int64Range(-3*24*3600, 3*24*3600),
	))

	props.Property("4. a longer window never gives an earlier deadline", prop.ForAll(
		func(start time.Time, category, a, b int32) bool {
			short, long := min(a, b), max(a, b)
			t1, _ := timerFor(t, category, short, start)
			t2, _ := timerFor(t, category, long, start)
			return !t2.Deadline.Before(t1.Deadline)
		},
		startGen(), categoryGen(), hoursGen(), hoursGen(),
	))

	props.Property("5. starting later never gives an earlier deadline", prop.ForAll(
		func(start time.Time, category, hours int32, later int64) bool {
			t1, _ := timerFor(t, category, hours, start)
			t2, _ := timerFor(t, category, hours, start.Add(time.Duration(later)*time.Second))
			return !t2.Deadline.Before(t1.Deadline)
		},
		startGen(), categoryGen(), hoursGen(), gen.Int64Range(0, 30*24*3600),
	))

	props.Property("6. a stopped timer is never due, however late it is", prop.ForAll(
		func(start time.Time, category, hours int32, offset int64) bool {
			tm, _ := timerFor(t, category, hours, start)
			tm.Stop()
			return !tm.IsDue(tm.Deadline.Add(time.Duration(offset) * time.Second))
		},
		startGen(), categoryGen(), hoursGen(), gen.Int64Range(-3*24*3600, 365*24*3600),
	))

	props.TestingRun(t, gopter.ConsoleReporter(true))
}

// The window rules of sla-policy: every hour count of 1 or more is accepted,
// 0 and below are refused, for every category.
func TestPolicyWindowProperties(t *testing.T) {
	props := gopter.NewProperties(parameters())

	props.Property("7. sla-policy accepts every window of 1 hour or more", prop.ForAll(
		func(category, hours int32) bool {
			p, err := policy.NewPolicy(policy.Category(category), hours, "u", base)
			return err == nil && p.Window() == time.Duration(hours)*time.Hour
		},
		categoryGen(), gen.Int32Range(1, 1<<30),
	))

	props.Property("8. sla-policy refuses a window of 0 hours or less", prop.ForAll(
		func(category, hours int32) bool {
			_, err := policy.NewPolicy(policy.Category(category), hours, "u", base)
			return err != nil
		},
		categoryGen(), gen.Int32Range(-1<<30, 0),
	))

	props.TestingRun(t, gopter.ConsoleReporter(true))
}
