package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/repotest"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
	"github.com/jurwaCST/Jurwa/services/analytics/ports"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func TestApplyThenReports(t *testing.T) {
	s := New()
	ctx := context.Background()
	e := domain.Event{ID: "e1", Type: domain.ReportSubmitted, ReportID: "r1", Category: domain.Hostel, OccurredAt: t0}

	applied, err := s.Apply(ctx, e)
	require.NoError(t, err)
	assert.True(t, applied)

	applied, err = s.Apply(ctx, e)
	require.NoError(t, err)
	assert.False(t, applied, "repeat")

	reports, err := s.Reports(ctx)
	require.NoError(t, err)
	require.Len(t, reports, 1)
	assert.Equal(t, domain.Open, reports[0].Status)
}

func TestApplyFromManyGoroutines(t *testing.T) {
	s := New()
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = s.Apply(ctx, domain.Event{ID: fmt.Sprintf("e%d", i), Type: domain.ReportSubmitted, ReportID: fmt.Sprintf("r%d", i), Category: domain.Others, OccurredAt: t0})
			_, _ = s.Reports(ctx)
		}(i)
	}
	wg.Wait()
	reports, err := s.Reports(ctx)
	require.NoError(t, err)
	assert.Len(t, reports, 50)
}

func TestContract(t *testing.T) {
	repotest.Run(t, func(*testing.T) ports.ProjectionStore { return New() })
}
