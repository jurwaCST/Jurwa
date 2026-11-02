package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reportID = "11111111-1111-1111-1111-111111111111"

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func TestNewTimer(t *testing.T) {
	tm, err := NewTimer(reportID, Hostel, 48*time.Hour, now)
	require.NoError(t, err)
	assert.Equal(t, reportID, tm.ReportID)
	assert.Equal(t, Hostel, tm.Category)
	assert.Equal(t, now, tm.StartedAt)
	assert.Equal(t, now.Add(48*time.Hour), tm.Deadline)
	assert.Equal(t, Running, tm.State)
}

func TestNewTimer_Errors(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		window   time.Duration
		wantErr  error
	}{
		{"category unspecified", CategoryUnspecified, time.Hour, ErrInvalidCategory},
		{"category out of range", Category(4), time.Hour, ErrInvalidCategory},
		{"zero window", Hostel, 0, ErrInvalidWindow},
		{"negative window", Hostel, -time.Hour, ErrInvalidWindow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTimer(reportID, tt.category, tt.window, now)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIsDue(t *testing.T) {
	tm, _ := NewTimer(reportID, Others, time.Hour, now)
	deadline := now.Add(time.Hour)

	assert.False(t, tm.IsDue(deadline.Add(-time.Second)), "one second before the deadline")
	assert.True(t, tm.IsDue(deadline), "exactly at the deadline")
	assert.True(t, tm.IsDue(deadline.Add(time.Second)), "after the deadline")

	tm.State = Stopped
	assert.False(t, tm.IsDue(deadline.Add(time.Hour)), "a stopped timer is never due")
	tm.State = Lapsed
	assert.False(t, tm.IsDue(deadline.Add(time.Hour)), "a lapsed timer is not due again")
}

func TestLapse(t *testing.T) {
	tm, _ := NewTimer(reportID, AcademicBlock, time.Hour, now)

	assert.ErrorIs(t, tm.Lapse(now), ErrNotDue, "before the deadline")
	assert.Equal(t, Running, tm.State)

	require.NoError(t, tm.Lapse(now.Add(time.Hour)))
	assert.Equal(t, Lapsed, tm.State)

	assert.ErrorIs(t, tm.Lapse(now.Add(2*time.Hour)), ErrNotDue, "already lapsed")
}

func TestStop(t *testing.T) {
	for _, from := range []State{Running, Lapsed, Stopped} {
		t.Run(from.String(), func(t *testing.T) {
			tm, _ := NewTimer(reportID, Hostel, time.Hour, now)
			tm.State = from
			tm.Stop()
			assert.Equal(t, Stopped, tm.State)
		})
	}
}

func TestCategory(t *testing.T) {
	tests := []struct {
		c     Category
		valid bool
		name  string
	}{
		{CategoryUnspecified, false, "CATEGORY_UNSPECIFIED"},
		{Hostel, true, "HOSTEL"},
		{AcademicBlock, true, "ACADEMIC_BLOCK"},
		{Others, true, "OTHERS"},
		{Category(-1), false, "CATEGORY_UNSPECIFIED"},
		{Category(4), false, "CATEGORY_UNSPECIFIED"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.valid, tt.c.Valid(), int32(tt.c))
		assert.Equal(t, tt.name, tt.c.String(), int32(tt.c))
	}
}

func TestStateString(t *testing.T) {
	assert.Equal(t, "TIMER_STATE_UNSPECIFIED", StateUnspecified.String())
	assert.Equal(t, "RUNNING", Running.String())
	assert.Equal(t, "STOPPED", Stopped.String())
	assert.Equal(t, "LAPSED", Lapsed.String())
	assert.Equal(t, "TIMER_STATE_UNSPECIFIED", State(9).String())
}
