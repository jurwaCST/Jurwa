package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryHead(t *testing.T) {
	routed := routed()
	moved, err := Reassign("a2", routed, staffID, "Plumbing", headID, now)
	require.NoError(t, err)

	head, err := CategoryHead([]Assignment{routed, moved})
	require.NoError(t, err)
	assert.Equal(t, headID, head.AssigneeUserID)

	_, err = CategoryHead([]Assignment{moved})
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = CategoryHead(nil)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestNewEscalated(t *testing.T) {
	e := NewEscalated("a3", routed(), now)
	assert.Equal(t, Assignment{
		ID: "a3", ReportID: reportID, AssigneeUserID: headID, AssigneePositionCode: "DSA", Kind: Escalated,
		Reason: "SLA lapsed: escalated to the Category Head", Current: true, AssignedAt: now,
	}, e)
	assert.Empty(t, e.AssignedBy, "the system acted")
}

func TestNewCompensated(t *testing.T) {
	staff, err := Reassign("a2", routed(), staffID, "Plumbing", headID, now)
	require.NoError(t, err)
	esc := NewEscalated("a3", routed(), now)

	c, err := NewCompensated("a4", esc, staff, "report already RESOLVED", now)
	require.NoError(t, err)
	assert.Equal(t, staffID, c.AssigneeUserID, "back to the person before the escalation")
	assert.Equal(t, Compensated, c.Kind)
	assert.Equal(t, "Escalation undone: report already RESOLVED", c.Reason)
	assert.Empty(t, c.AssignedBy)
	assert.True(t, c.Current)

	_, err = NewCompensated("a5", c, staff, "again", now)
	assert.ErrorIs(t, err, ErrNotEscalated, "an escalation is undone only once")
}
