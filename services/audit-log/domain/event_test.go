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
	actorID  = "22222222-2222-2222-2222-222222222222"
)

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func validInput() Input {
	return Input{
		OccurredAt: now.Add(-time.Minute),
		Service:    "status-lifecycle",
		Action:     "STATUS_CHANGED",
		ReportID:   reportID,
		ActorID:    actorID,
		Details:    map[string]string{"from": "OPEN", "to": "ASSIGNED"},
	}
}

func TestNewEvent(t *testing.T) {
	in := validInput()
	in.Service = "  status-lifecycle "
	e, err := NewEvent("e1", in, now)
	require.NoError(t, err)
	assert.Equal(t, Event{
		ID: "e1", OccurredAt: now.Add(-time.Minute), Service: "status-lifecycle", Action: StatusChanged,
		ReportID: reportID, ActorID: actorID, Details: map[string]string{"from": "OPEN", "to": "ASSIGNED"},
	}, e)
}

func TestNewEvent_Defaults(t *testing.T) {
	in := Input{Service: "sla-tracking", Action: "ESCALATED"}
	e, err := NewEvent("e1", in, now)
	require.NoError(t, err)
	assert.Equal(t, now, e.OccurredAt, "missing time uses server time")
	assert.Empty(t, e.ReportID)
	assert.Empty(t, e.ActorID, "the system acted")
	assert.Empty(t, e.Details)
}

func TestNewEvent_CopiesDetails(t *testing.T) {
	in := validInput()
	e, err := NewEvent("e1", in, now)
	require.NoError(t, err)
	in.Details["to"] = "changed"
	assert.Equal(t, "ASSIGNED", e.Details["to"], "the caller cannot change a stored entry")
}

func TestNewEvent_Errors(t *testing.T) {
	tests := []struct {
		name   string
		change func(in *Input)
	}{
		{"service missing", func(in *Input) { in.Service = " " }},
		{"service too long", func(in *Input) { in.Service = strings.Repeat("s", MaxServiceLen+1) }},
		{"action missing", func(in *Input) { in.Action = "" }},
		{"action unknown", func(in *Input) { in.Action = "DELETED" }},
		{"action lower case", func(in *Input) { in.Action = "routed" }},
		{"report id not a UUID", func(in *Input) { in.ReportID = "abc" }},
		{"actor id not a UUID", func(in *Input) { in.ActorID = "abc" }},
		{"on behalf of too long", func(in *Input) { in.OnBehalfOf = strings.Repeat("a", MaxValueLen+1) }},
		{"too many details", func(in *Input) {
			in.Details = map[string]string{}
			for i := 0; i <= MaxDetails; i++ {
				in.Details[strings.Repeat("k", i+1)] = "v"
			}
		}},
		{"empty detail key", func(in *Input) { in.Details = map[string]string{"": "v"} }},
		{"detail key too long", func(in *Input) { in.Details = map[string]string{strings.Repeat("k", MaxKeyLen+1): "v"} }},
		{"detail value too long", func(in *Input) { in.Details = map[string]string{"k": strings.Repeat("v", MaxValueLen+1)} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.change(&in)
			_, err := NewEvent("e1", in, now)
			assert.ErrorIs(t, err, ErrInvalidEvent)
		})
	}
}

func TestNewEvent_Boundaries(t *testing.T) {
	in := validInput()
	in.Service = strings.Repeat("s", MaxServiceLen)
	in.OnBehalfOf = strings.Repeat("ཀ", MaxValueLen) // counted in characters
	in.Details = map[string]string{strings.Repeat("k", MaxKeyLen): strings.Repeat("v", MaxValueLen)}
	_, err := NewEvent("e1", in, now)
	assert.NoError(t, err)

	in = validInput()
	in.Details = map[string]string{}
	for i := 0; i < MaxDetails; i++ {
		in.Details[strings.Repeat("k", i+1)] = "v"
	}
	_, err = NewEvent("e1", in, now)
	assert.NoError(t, err, "exactly MaxDetails entries")
}

func TestAllActionsAccepted(t *testing.T) {
	for _, a := range []string{"REPORT_SUBMITTED", "ROUTED", "STATUS_CHANGED", "REASSIGNED", "ESCALATED", "COMPENSATED"} {
		in := validInput()
		in.Action = a
		_, err := NewEvent("e1", in, now)
		assert.NoError(t, err, a)
	}
}

func TestCheckCanList(t *testing.T) {
	assert.NoError(t, CheckCanList("SYSTEM_ADMIN"))
	for _, r := range []Role{"CATEGORY_HEAD", "STAFF", "STUDENT", "PRESIDENT", ""} {
		assert.ErrorIs(t, CheckCanList(r), ErrPermissionDenied, r)
	}
}

func TestIsUUID(t *testing.T) {
	assert.True(t, IsUUID(reportID))
	assert.False(t, IsUUID("abc"))
}
