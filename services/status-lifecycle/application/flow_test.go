package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
)

const otherReport = "33333333-3333-3333-3333-333333333333"

func TestOnReportSubmitted(t *testing.T) {
	svc, _ := newTestService(t) // reportID already registered
	ctx := context.Background()

	require.NoError(t, svc.OnReportSubmitted(ctx, otherReport))
	rs, err := svc.GetStatus(ctx, otherReport)
	require.NoError(t, err)
	assert.Equal(t, domain.Open, rs.Status)

	assert.NoError(t, svc.OnReportSubmitted(ctx, otherReport), "a repeated event counts as done")
	assert.ErrorIs(t, svc.OnReportSubmitted(ctx, "abc"), ErrInvalidArgument)
}

func TestOnReportRouted_SetsAssignedAsTheSystem(t *testing.T) {
	svc, pub := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.OnReportRouted(ctx, reportID, "DSA", "c1"))
	rs, _ := svc.GetStatus(ctx, reportID)
	assert.Equal(t, domain.Assigned, rs.Status)
	assert.Empty(t, rs.UpdatedBy, "the system acted")
	require.Len(t, pub.events, 1)
	assert.Equal(t, "routed to DSA", pub.events[0].Reason)
	assert.Equal(t, "c1", pub.events[0].CorrelationID)

	// Repeated: the report is past OPEN, nothing changes.
	require.NoError(t, svc.OnReportRouted(ctx, reportID, "DSA", "c1"))
	rs, _ = svc.GetStatus(ctx, reportID)
	assert.Equal(t, int32(2), rs.Version)
}

func TestOnReportRouted_BeforeSubmitted(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.OnReportRouted(ctx, otherReport, "SSO", "c1"))
	rs, err := svc.GetStatus(ctx, otherReport)
	require.NoError(t, err)
	assert.Equal(t, domain.Assigned, rs.Status, "registered, then assigned")
	require.NoError(t, svc.OnReportSubmitted(ctx, otherReport), "the late report.submitted counts as done")
	rs, _ = svc.GetStatus(ctx, otherReport)
	assert.Equal(t, domain.Assigned, rs.Status)
}

func TestOnReportRouted_Errors(t *testing.T) {
	svc, _ := newTestService(t)
	assert.ErrorIs(t, svc.OnReportRouted(context.Background(), "abc", "DSA", ""), ErrInvalidArgument)

	repo := memory.New()
	conflict := NewService(conflictRepo{repo}, &fakePublisher{}, time.Now)
	require.NoError(t, conflict.OnReportSubmitted(context.Background(), reportID))
	assert.ErrorIs(t, conflict.OnReportRouted(context.Background(), reportID, "DSA", ""), domain.ErrVersionConflict, "retried by the bus")

	broken := NewService(getFailsRepo{repo}, &fakePublisher{}, time.Now)
	assert.ErrorContains(t, broken.OnReportRouted(context.Background(), reportID, "DSA", ""), "db down")
}

// getFailsRepo can create but not read.
type getFailsRepo struct{ *memory.Repository }

func (getFailsRepo) Get(context.Context, string) (domain.ReportStatus, error) {
	return domain.ReportStatus{}, errors.New("db down")
}

// staleFirstRead answers OPEN to the first read only, like a report a staff
// member moved on right after OnReportRouted looked at it.
type staleFirstRead struct {
	*memory.Repository
	reads *int
}

func (r staleFirstRead) Get(ctx context.Context, id string) (domain.ReportStatus, error) {
	rs, err := r.Repository.Get(ctx, id)
	*r.reads++
	if *r.reads == 1 {
		rs.Status = domain.Open
	}
	return rs, err
}

func TestOnReportRouted_ReportMovedOnMeanwhile(t *testing.T) {
	repo := memory.New()
	svc := NewService(repo, &fakePublisher{}, time.Now)
	ctx := context.Background()
	require.NoError(t, svc.OnReportSubmitted(ctx, reportID))
	for _, s := range []domain.Status{domain.Assigned, domain.InProgress} {
		_, err := svc.ChangeStatus(ctx, ChangeStatusCommand{ReportID: reportID, To: s, ActorID: staffID})
		require.NoError(t, err)
	}

	reads := 0
	stale := NewService(staleFirstRead{repo, &reads}, &fakePublisher{}, time.Now)
	assert.NoError(t, stale.OnReportRouted(ctx, reportID, "DSA", ""), "IN_PROGRESS cannot go to ASSIGNED: counts as done")
	rs, _ := svc.GetStatus(ctx, reportID)
	assert.Equal(t, domain.InProgress, rs.Status)
}
