package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/audit-log/domain"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func TestAppendAndList(t *testing.T) {
	s := New()
	ctx := context.Background()

	list, err := s.ListByReport(ctx, "r1")
	require.NoError(t, err)
	assert.Empty(t, list)

	// Appended out of time order; same time keeps append order.
	require.NoError(t, s.Append(ctx, domain.Event{ID: "late", ReportID: "r1", OccurredAt: t0.Add(time.Hour)}))
	require.NoError(t, s.Append(ctx, domain.Event{ID: "early", ReportID: "r1", OccurredAt: t0}))
	require.NoError(t, s.Append(ctx, domain.Event{ID: "early-2", ReportID: "r1", OccurredAt: t0}))
	require.NoError(t, s.Append(ctx, domain.Event{ID: "other", ReportID: "r2", OccurredAt: t0}))
	require.NoError(t, s.Append(ctx, domain.Event{ID: "no-report", OccurredAt: t0}))

	list, err = s.ListByReport(ctx, "r1")
	require.NoError(t, err)
	ids := []string{}
	for _, e := range list {
		ids = append(ids, e.ID)
	}
	assert.Equal(t, []string{"early", "early-2", "late"}, ids)
}

func TestListReturnsCopies(t *testing.T) {
	s := New()
	ctx := context.Background()
	require.NoError(t, s.Append(ctx, domain.Event{ID: "e1", ReportID: "r1", Details: map[string]string{"to": "ASSIGNED"}}))

	list, _ := s.ListByReport(ctx, "r1")
	list[0].Details["to"] = "changed"
	list[0].Service = "changed"

	again, _ := s.ListByReport(ctx, "r1")
	assert.Equal(t, "ASSIGNED", again[0].Details["to"], "stored entries cannot be changed")
	assert.Empty(t, again[0].Service)
}
