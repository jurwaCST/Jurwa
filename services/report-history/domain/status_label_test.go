package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The timeline uses the same words as the status badges in the web app
// (web/src/components/labels.ts), not the system names.
func TestStatusLabelMatchesTheBadges(t *testing.T) {
	tests := map[Status]string{
		StatusOpen:       "Open",
		StatusAssigned:   "Assigned",
		StatusInProgress: "In progress",
		StatusResolved:   "Resolved",
		StatusEscalated:  "Escalated",
	}
	for s, want := range tests {
		assert.Equal(t, want, StatusLabel(s), "label of %s", s)
	}
}

func TestStatusLabelOfAnUnknownStatusIsTheStatusItself(t *testing.T) {
	assert.Equal(t, "CLOSED", StatusLabel("CLOSED"))
}
