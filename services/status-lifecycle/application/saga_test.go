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

func lapsed(eventID string) SagaEvent {
	return SagaEvent{EventID: eventID, CorrelationID: "c1", ReportID: reportID}
}

func move(t *testing.T, svc *Service, to ...domain.Status) {
	for _, s := range to {
		_, err := svc.ChangeStatus(context.Background(), ChangeStatusCommand{ReportID: reportID, To: s, ActorID: staffID})
		require.NoError(t, err)
	}
}

func TestEscalateOnLapse_SetsEscalated(t *testing.T) {
	svc, pub := newTestService(t)
	move(t, svc, domain.Assigned)
	pub.events = nil

	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapsed("e1")))
	rs, _ := svc.GetStatus(context.Background(), reportID)
	assert.Equal(t, domain.Escalated, rs.Status)
	assert.Equal(t, domain.Assigned, rs.Previous)
	require.Len(t, pub.events, 1)
	assert.Equal(t, "c1", pub.events[0].CorrelationID, "status.changed carries the Saga run")
	assert.Empty(t, pub.events[0].ActorID, "the system acted")
	assert.Empty(t, pub.compensated)

	// The same event again changes nothing.
	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapsed("e1")))
	rs, _ = svc.GetStatus(context.Background(), reportID)
	assert.Equal(t, int32(3), rs.Version)
}

func TestEscalateOnLapse_AlreadyEscalatedIsDone(t *testing.T) {
	svc, pub := newTestService(t)
	// The sla-escalation-checker function escalated it first.
	_, err := svc.ChangeStatus(context.Background(), ChangeStatusCommand{ReportID: reportID, To: domain.Escalated})
	require.NoError(t, err)

	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapsed("e1")))
	assert.Empty(t, pub.compensated, "nothing to undo")
}

func TestEscalateOnLapse_ResolvedOrUnknownAsksForCompensation(t *testing.T) {
	svc, pub := newTestService(t)
	move(t, svc, domain.Assigned, domain.InProgress, domain.Resolved)

	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapsed("e1")))
	require.Len(t, pub.compensated, 1)
	c := pub.compensated[0]
	assert.Equal(t, reportID, c.ReportID)
	assert.Equal(t, StepSetEscalated, c.FailedStep)
	assert.Equal(t, StepEscalate, c.UndoneStep)
	assert.Equal(t, "c1", c.CorrelationID)
	assert.Contains(t, c.Reason, "RESOLVED to ESCALATED")
	rs, _ := svc.GetStatus(context.Background(), reportID)
	assert.Equal(t, domain.Resolved, rs.Status)

	unknown := SagaEvent{EventID: "e2", CorrelationID: "c2", ReportID: "99999999-9999-9999-9999-999999999999"}
	require.NoError(t, svc.EscalateOnLapse(context.Background(), unknown))
	require.Len(t, pub.compensated, 2)
	assert.Contains(t, pub.compensated[1].Reason, "not found")
}

func TestEscalateOnLapse_Errors(t *testing.T) {
	svc, pub := newTestService(t)
	ctx := context.Background()

	assert.ErrorIs(t, svc.EscalateOnLapse(ctx, SagaEvent{ReportID: reportID}), ErrInvalidArgument, "no event id")
	assert.ErrorIs(t, svc.EscalateOnLapse(ctx, SagaEvent{EventID: "e1", ReportID: "abc"}), ErrInvalidArgument)

	move(t, svc, domain.Assigned, domain.InProgress, domain.Resolved)
	pub.compensErr = errors.New("redis down")
	assert.ErrorContains(t, svc.EscalateOnLapse(ctx, lapsed("e3")), "redis down")
	assert.False(t, svc.seen.has("e3"), "tried again later")
}

// conflictRepo reports a version conflict on every update.
type conflictRepo struct{ *memory.Repository }

func (conflictRepo) Update(context.Context, domain.ReportStatus, int32) error {
	return domain.ErrVersionConflict
}

func TestEscalateOnLapse_VersionConflictIsRetried(t *testing.T) {
	repo := memory.New()
	svc := NewService(conflictRepo{repo}, &fakePublisher{}, time.Now)
	_, err := svc.Register(context.Background(), reportID)
	require.NoError(t, err)

	assert.ErrorIs(t, svc.EscalateOnLapse(context.Background(), lapsed("e1")), domain.ErrVersionConflict)
	assert.False(t, svc.seen.has("e1"))
}
