package spec

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lifecycle "github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

var statuses = []lifecycle.Status{lifecycle.Open, lifecycle.Assigned, lifecycle.InProgress, lifecycle.Resolved, lifecycle.Escalated}

// The expected tables come from docs/pm/requirements.md section 3, FR-15, FR-16
// and FR-20, written out by hand. Rows are "from", columns are "to", in the
// order Open, Assigned, In Progress, Resolved, Escalated. true = valid.

// A person (Category Head or staff) acting. Escalated is never set by a person.
var userTable = map[lifecycle.Status][5]bool{
	//                     OPEN   ASSIGNED IN_PROG RESOLVED ESCALATED
	lifecycle.Open:       {false, true, false, false, false},
	lifecycle.Assigned:   {false, false, true, false, false},
	lifecycle.InProgress: {false, false, false, true, false},
	lifecycle.Resolved:   {false, false, false, false, false},
	lifecycle.Escalated:  {false, false, true, false, false},
}

// The system acting (SLA escalation and Saga compensation). The Escalated row
// assumes the report was escalated from Assigned, so going back to Assigned is
// the compensation (FR-16).
var systemTable = map[lifecycle.Status][5]bool{
	//                     OPEN   ASSIGNED IN_PROG RESOLVED ESCALATED
	lifecycle.Open:       {false, true, false, false, true},
	lifecycle.Assigned:   {false, false, true, false, true},
	lifecycle.InProgress: {false, false, false, true, true},
	lifecycle.Resolved:   {false, false, false, false, false},
	lifecycle.Escalated:  {false, true, true, false, false},
}

func runTable(t *testing.T, table map[lifecycle.Status][5]bool, bySystem bool) {
	for _, from := range statuses {
		for j, to := range statuses {
			want := table[from][j]
			t.Run(fmt.Sprintf("%s to %s", from, to), func(t *testing.T) {
				previous := lifecycle.StatusUnspecified
				if from == lifecycle.Escalated {
					previous = lifecycle.Assigned
				}
				err := lifecycle.CheckTransition(from, to, previous, bySystem)
				if want {
					assert.NoError(t, err)
				} else {
					assert.ErrorIs(t, err, lifecycle.ErrInvalidTransition)
				}
			})
		}
	}
}

func TestTransitionTableByUser(t *testing.T)   { runTable(t, userTable, false) }
func TestTransitionTableBySystem(t *testing.T) { runTable(t, systemTable, true) }

// Compensation goes back only to the status before the escalation.
func TestCompensationOnlyToPrevious(t *testing.T) {
	assert.NoError(t, lifecycle.CheckTransition(lifecycle.Escalated, lifecycle.Open, lifecycle.Open, true))
	assert.ErrorIs(t, lifecycle.CheckTransition(lifecycle.Escalated, lifecycle.Assigned, lifecycle.Open, true), lifecycle.ErrInvalidTransition)
	assert.ErrorIs(t, lifecycle.CheckTransition(lifecycle.Escalated, lifecycle.Open, lifecycle.Open, false), lifecycle.ErrInvalidTransition, "only the system compensates")
}

func TestInvalidStatusValues(t *testing.T) {
	for _, to := range []lifecycle.Status{lifecycle.StatusUnspecified, 6, -1} {
		err := lifecycle.CheckTransition(lifecycle.Open, to, lifecycle.StatusUnspecified, true)
		assert.ErrorIs(t, err, lifecycle.ErrInvalidStatus, "to %d", to)
	}
}

// One full path through the lifecycle, including an escalation, with versions.
func TestFullLifecyclePath(t *testing.T) {
	t0 := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	rs := lifecycle.NewReportStatus("r1", t0)
	require.Equal(t, lifecycle.Open, rs.Status)
	require.Equal(t, int32(1), rs.Version)

	steps := []struct {
		to    lifecycle.Status
		actor string // "" = system
	}{
		{lifecycle.Assigned, ""},           // routed
		{lifecycle.Escalated, ""},          // SLA passed
		{lifecycle.InProgress, "user-dsa"}, // new assignee starts
		{lifecycle.Resolved, "user-dsa"},
	}
	for i, s := range steps {
		require.NoError(t, rs.ChangeTo(s.to, s.actor, t0.Add(time.Duration(i+1)*time.Hour)), "step %d to %s", i+1, s.to)
		assert.Equal(t, int32(i+2), rs.Version)
	}

	err := rs.ChangeTo(lifecycle.Escalated, "", t0.Add(10*time.Hour))
	assert.ErrorIs(t, err, lifecycle.ErrInvalidTransition, "a resolved report is never escalated")
	assert.Equal(t, lifecycle.Resolved, rs.Status, "a failed change leaves the report unchanged")
}
