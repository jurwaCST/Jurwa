package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	reportID = "11111111-1111-1111-1111-111111111111"
	headID   = "22222222-2222-2222-2222-222222222222"
	staffID  = "33333333-3333-3333-3333-333333333333"
	otherID  = "44444444-4444-4444-4444-444444444444"
)

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func routed() Assignment {
	return NewRouted("a1", reportID, headID, "DSA", now)
}

func TestNewRouted(t *testing.T) {
	a := routed()
	assert.Equal(t, Assignment{
		ID: "a1", ReportID: reportID, AssigneeUserID: headID, AssigneePositionCode: "DSA",
		Kind: Routed, Current: true, AssignedAt: now,
	}, a)
	assert.Empty(t, a.AssignedBy, "the system routed it")
}

func TestReassign(t *testing.T) {
	cur := routed()
	later := now.Add(time.Hour)

	next, err := Reassign("a2", cur, staffID, "  Plumbing, needs estate staff  ", headID, later)
	require.NoError(t, err)
	assert.Equal(t, Assignment{
		ID: "a2", ReportID: reportID, AssigneeUserID: staffID, Kind: Reassigned,
		AssignedBy: headID, Reason: "Plumbing, needs estate staff", Current: true, AssignedAt: later,
	}, next)
}

func TestReassign_Errors(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		reason  string
		wantErr error
	}{
		{"blank reason", staffID, "   ", ErrInvalidReason},
		{"reason 501 characters", staffID, strings.Repeat("a", 501), ErrInvalidReason},
		{"same assignee", headID, "no change", ErrSameAssignee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Reassign("a2", routed(), tt.to, tt.reason, headID, now)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}

	// Boundary: exactly 500 characters, counted as characters not bytes.
	_, err := Reassign("a2", routed(), staffID, strings.Repeat("ཀ", 500), headID, now)
	assert.NoError(t, err, "500 Dzongkha characters")
}

func TestCheckCanReassign(t *testing.T) {
	cur := routed()
	cur.AssigneeUserID = staffID // a staff member holds the report

	tests := []struct {
		name    string
		role    Role
		actor   string
		allowed bool
	}{
		{"category head, not assignee", RoleCategoryHead, headID, true},
		{"staff who is the assignee", RoleStaff, staffID, true},
		{"staff who is not the assignee", RoleStaff, otherID, false},
		{"student", RoleStudent, staffID, false},
		{"system admin", RoleSystemAdmin, otherID, false},
		{"president", RolePresident, otherID, false},
		{"no role", "", staffID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCanReassign(tt.role, tt.actor, cur)
			if tt.allowed {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrPermissionDenied)
			}
		})
	}
}

func TestCheckCanList(t *testing.T) {
	for _, r := range []Role{RoleCategoryHead, RoleStaff, RoleSystemAdmin} {
		assert.NoError(t, CheckCanList(r), r)
	}
	for _, r := range []Role{RoleStudent, RolePresident, ""} {
		assert.ErrorIs(t, CheckCanList(r), ErrPermissionDenied, r)
	}
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID(reportID))
	assert.False(t, IsUUID(""))
	assert.False(t, IsUUID("abc"))
}

func TestKindString(t *testing.T) {
	assert.Equal(t, "ROUTED", Routed.String())
	assert.Equal(t, "REASSIGNED", Reassigned.String())
	assert.Equal(t, "ESCALATED", Escalated.String())
	assert.Equal(t, "COMPENSATED", Compensated.String())
	assert.Equal(t, "ASSIGNMENT_KIND_UNSPECIFIED", KindUnspecified.String())
}
