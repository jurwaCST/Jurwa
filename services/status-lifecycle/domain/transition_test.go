package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every (from, to) pair, checked for a user and for the system.
// Valid for a user: OPEN>ASSIGNED, ASSIGNED>IN_PROGRESS, IN_PROGRESS>RESOLVED, ESCALATED>IN_PROGRESS.
// Valid only for the system: OPEN/ASSIGNED/IN_PROGRESS>ESCALATED, and ESCALATED back to its
// previous status (Saga compensation). In this table previous is ASSIGNED for ESCALATED rows.
func TestCheckTransition_AllPairs(t *testing.T) {
	tests := []struct {
		from, to Status
		userOK   bool
		systemOK bool
	}{
		// from OPEN
		{Open, Open, false, false},
		{Open, Assigned, true, true},
		{Open, InProgress, false, false},
		{Open, Resolved, false, false},
		{Open, Escalated, false, true},
		// from ASSIGNED
		{Assigned, Open, false, false},
		{Assigned, Assigned, false, false},
		{Assigned, InProgress, true, true},
		{Assigned, Resolved, false, false},
		{Assigned, Escalated, false, true},
		// from IN_PROGRESS
		{InProgress, Open, false, false},
		{InProgress, Assigned, false, false},
		{InProgress, InProgress, false, false},
		{InProgress, Resolved, true, true},
		{InProgress, Escalated, false, true},
		// from RESOLVED: final, nothing is allowed
		{Resolved, Open, false, false},
		{Resolved, Assigned, false, false},
		{Resolved, InProgress, false, false},
		{Resolved, Resolved, false, false},
		{Resolved, Escalated, false, false},
		// from ESCALATED (previous = ASSIGNED)
		{Escalated, Open, false, false},
		{Escalated, Assigned, false, true}, // compensation back to previous
		{Escalated, InProgress, true, true},
		{Escalated, Resolved, false, false},
		{Escalated, Escalated, false, false},
	}

	for _, tt := range tests {
		previous := StatusUnspecified
		if tt.from == Escalated {
			previous = Assigned
		}
		for _, actor := range []struct {
			name     string
			bySystem bool
			wantOK   bool
		}{
			{"user", false, tt.userOK},
			{"system", true, tt.systemOK},
		} {
			t.Run(tt.from.String()+"_to_"+tt.to.String()+"_by_"+actor.name, func(t *testing.T) {
				err := CheckTransition(tt.from, tt.to, previous, actor.bySystem)
				if actor.wantOK {
					assert.NoError(t, err)
				} else {
					assert.ErrorIs(t, err, ErrInvalidTransition)
				}
			})
		}
	}
}

func TestCheckTransition_CompensationGoesOnlyToPrevious(t *testing.T) {
	tests := []struct {
		name     string
		previous Status
		to       Status
		wantOK   bool
	}{
		{"back to OPEN", Open, Open, true},
		{"back to ASSIGNED", Assigned, Assigned, true},
		{"back to IN_PROGRESS", InProgress, InProgress, true},
		{"not to a different status", Open, Assigned, false},
		{"not when previous is unknown", StatusUnspecified, Open, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckTransition(Escalated, tt.to, tt.previous, true)
			if tt.wantOK {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrInvalidTransition)
			}
		})
	}
}

func TestCheckTransition_RejectsUnknownStatus(t *testing.T) {
	for _, to := range []Status{StatusUnspecified, Status(99)} {
		err := CheckTransition(Open, to, StatusUnspecified, true)
		assert.ErrorIs(t, err, ErrInvalidStatus)
	}
}

func TestReportStatus_ChangeTo(t *testing.T) {
	created := time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC)
	later := created.Add(time.Hour)

	rs := NewReportStatus("r1", created)
	assert.Equal(t, Open, rs.Status)
	assert.Equal(t, StatusUnspecified, rs.Previous)
	assert.Equal(t, int32(1), rs.Version)

	require.NoError(t, rs.ChangeTo(Assigned, "", later))
	assert.Equal(t, Assigned, rs.Status)
	assert.Equal(t, Open, rs.Previous)
	assert.Equal(t, int32(2), rs.Version)
	assert.Equal(t, "", rs.UpdatedBy)
	assert.Equal(t, later, rs.UpdatedAt)

	// Escalate, then undo it: the report goes back to ASSIGNED.
	require.NoError(t, rs.ChangeTo(Escalated, "", later))
	assert.Equal(t, Assigned, rs.Previous)
	require.NoError(t, rs.ChangeTo(Assigned, "", later))
	assert.Equal(t, Assigned, rs.Status)
	assert.Equal(t, int32(4), rs.Version)

	require.NoError(t, rs.ChangeTo(InProgress, "staff-1", later))
	assert.Equal(t, "staff-1", rs.UpdatedBy)
}

func TestReportStatus_ChangeToInvalidLeavesItUnchanged(t *testing.T) {
	rs := NewReportStatus("r1", time.Now())
	before := rs

	err := rs.ChangeTo(Resolved, "user-1", time.Now())

	assert.True(t, errors.Is(err, ErrInvalidTransition))
	assert.Equal(t, before, rs)
}

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "OPEN", Open.String())
	assert.Equal(t, "IN_PROGRESS", InProgress.String())
	assert.Equal(t, "ESCALATED", Escalated.String())
	assert.Equal(t, "STATUS_UNSPECIFIED", StatusUnspecified.String())
	assert.Equal(t, "STATUS_UNSPECIFIED", Status(42).String())
}
