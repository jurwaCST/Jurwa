package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
)

// handedToStaff is a report routed to the head and handed to staff.
func handedToStaff(t *testing.T) (*Service, *fakePublisher) {
	svc, pub := routedService(t)
	_, err := svc.ReassignReport(context.Background(), ReassignCommand{
		ReportID: reportID, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: domain.RoleCategoryHead,
	})
	require.NoError(t, err)
	pub.events = nil
	return svc, pub
}

func lapse(eventID, correlation string) SagaEvent {
	return SagaEvent{EventID: eventID, CorrelationID: correlation, ReportID: reportID}
}

func kinds(t *testing.T, svc *Service) []domain.Kind {
	list, err := svc.ListReassignments(context.Background(), domain.RoleSystemAdmin, reportID)
	require.NoError(t, err)
	out := []domain.Kind{}
	for _, a := range list {
		out = append(out, a.Kind)
	}
	return out
}

func TestEscalateOnLapse(t *testing.T) {
	svc, pub := handedToStaff(t)
	ctx := context.Background()

	require.NoError(t, svc.EscalateOnLapse(ctx, lapse("e1", "c1")))
	assert.Equal(t, []domain.Kind{domain.Routed, domain.Reassigned, domain.Escalated}, kinds(t, svc))
	require.Len(t, pub.events, 1)
	ev := pub.events[0]
	assert.Equal(t, staffID, ev.FromUserID)
	assert.Equal(t, headID, ev.ToUserID)
	assert.Empty(t, ev.ActorID, "the system acted")
	assert.Equal(t, "c1", ev.CorrelationID)

	// The same event again changes nothing.
	require.NoError(t, svc.EscalateOnLapse(ctx, lapse("e1", "c1")))
	assert.Len(t, kinds(t, svc), 3)
	assert.Len(t, pub.events, 1)
}

func TestEscalateOnLapse_ReportWithoutAssignment(t *testing.T) {
	svc, pub := handedToStaff(t)
	ev := SagaEvent{EventID: "e1", CorrelationID: "c1", ReportID: otherID}
	assert.NoError(t, svc.EscalateOnLapse(context.Background(), ev), "nothing to hand up")
	assert.Empty(t, pub.events)
}

func TestEscalateOnLapse_InvalidEvent(t *testing.T) {
	svc, _ := handedToStaff(t)
	assert.ErrorIs(t, svc.EscalateOnLapse(context.Background(), SagaEvent{ReportID: reportID}), ErrInvalidArgument)
	assert.ErrorIs(t, svc.EscalateOnLapse(context.Background(), SagaEvent{EventID: "e1", ReportID: "abc"}), ErrInvalidArgument)
	assert.ErrorIs(t, svc.CompensateEscalation(context.Background(), SagaEvent{ReportID: reportID}), ErrInvalidArgument)
}

func TestCompensateEscalation(t *testing.T) {
	svc, pub := handedToStaff(t)
	ctx := context.Background()
	require.NoError(t, svc.EscalateOnLapse(ctx, lapse("e1", "c1")))

	undo := SagaEvent{EventID: "e2", CorrelationID: "c1", ReportID: reportID, Reason: "report already RESOLVED"}
	require.NoError(t, svc.CompensateEscalation(ctx, undo))
	assert.Equal(t, []domain.Kind{domain.Routed, domain.Reassigned, domain.Escalated, domain.Compensated}, kinds(t, svc))
	require.Len(t, pub.events, 2)
	back := pub.events[1]
	assert.Equal(t, headID, back.FromUserID)
	assert.Equal(t, staffID, back.ToUserID, "the person before the escalation")
	assert.Equal(t, "c1", back.CorrelationID)

	// Repeating the undo changes nothing.
	require.NoError(t, svc.CompensateEscalation(ctx, undo))
	assert.Len(t, kinds(t, svc), 4)
}

func TestCompensateEscalation_ArrivesBeforeTheLapse(t *testing.T) {
	svc, pub := handedToStaff(t)
	ctx := context.Background()

	require.NoError(t, svc.CompensateEscalation(ctx, SagaEvent{EventID: "e2", CorrelationID: "c1", ReportID: reportID}))
	require.NoError(t, svc.EscalateOnLapse(ctx, lapse("e1", "c1")), "the late lapse of an undone run is skipped")
	assert.Equal(t, []domain.Kind{domain.Routed, domain.Reassigned}, kinds(t, svc))
	assert.Empty(t, pub.events)

	// A lapse of another run still escalates.
	require.NoError(t, svc.EscalateOnLapse(ctx, lapse("e3", "c2")))
	assert.Equal(t, domain.Escalated, kinds(t, svc)[2])
}

// listFailsRepo fails List; conflictOnReplace fails Replace after a working List.
type listFailsRepo struct{ *memory.Repository }

func (listFailsRepo) List(context.Context, string) ([]domain.Assignment, error) {
	return nil, errors.New("db down")
}

type replaceFailsRepo struct{ *memory.Repository }

func (replaceFailsRepo) Replace(context.Context, string, domain.Assignment) error {
	return domain.ErrConflict
}

func TestSaga_StoreAndPublishErrors(t *testing.T) {
	ctx := context.Background()

	svc := newService(listFailsRepo{memory.New()}, &fakePublisher{})
	assert.ErrorContains(t, svc.EscalateOnLapse(ctx, lapse("e1", "c1")), "db down")
	assert.ErrorContains(t, svc.CompensateEscalation(ctx, lapse("e2", "c1")), "db down")

	// A conflict is returned, so the event bus tries again, and the event is not marked handled.
	repo := memory.New()
	ok := newService(repo, &fakePublisher{})
	_, err := ok.RecordRouted(ctx, reportID, headID, "DSA")
	require.NoError(t, err)
	svc = newService(replaceFailsRepo{repo}, &fakePublisher{})
	assert.ErrorIs(t, svc.EscalateOnLapse(ctx, lapse("e1", "c1")), domain.ErrConflict)
	assert.False(t, svc.saga.handled("e1"))

	// The undo hits the same conflict.
	require.NoError(t, ok.EscalateOnLapse(ctx, lapse("e1", "c1")))
	assert.ErrorIs(t, svc.CompensateEscalation(ctx, lapse("e2", "c1")), domain.ErrConflict)

	// Publish failure is reported.
	pubSvc, pub := handedToStaff(t)
	pub.err = errors.New("redis down")
	assert.ErrorContains(t, pubSvc.EscalateOnLapse(ctx, lapse("e9", "c9")), "redis down")
}
