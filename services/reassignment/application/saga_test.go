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

// ---- SCRUM-158: escalation goes up to the DSA or the DAA ----

const (
	dsaID = "55555555-5555-5555-5555-555555555555"
	daaID = "66666666-6666-6666-6666-666666666666"
)

// fakeHolders is a HolderLookup: position code to holder user id.
type fakeHolders struct {
	holders map[string]string
	err     error
	asked   []string
}

func (f *fakeHolders) GetHolder(_ context.Context, code string) (string, error) {
	f.asked = append(f.asked, code)
	return f.holders[code], f.err
}

// routedTo has reportID routed to position (held by headID) and handed to staff
// by the Category Head, with escalation targets looked up in holders.
func routedTo(t *testing.T, position string, holders *fakeHolders) (*Service, *fakePublisher) {
	pub := &fakePublisher{}
	svc := newService(memory.New(), pub).UseEscalationTargets(holders)
	ctx := context.Background()
	_, err := svc.RecordRouted(ctx, reportID, headID, position)
	require.NoError(t, err)
	if position == "SSO" {
		_, err = svc.ReassignReport(ctx, ReassignCommand{
			ReportID: reportID, ToUserID: staffID, Reason: "Block HA", ActorID: headID, Role: domain.RoleCategoryHead,
		})
		require.NoError(t, err)
	}
	pub.events = nil
	return svc, pub
}

func current(t *testing.T, svc *Service) domain.Assignment {
	list, err := svc.ListReassignments(context.Background(), domain.RoleSystemAdmin, reportID)
	require.NoError(t, err)
	return list[len(list)-1]
}

func TestEscalateOnLapse_SSOReportGoesUpToTheDSA(t *testing.T) {
	holders := &fakeHolders{holders: map[string]string{"DSA": dsaID, "DAA": daaID}}
	svc, pub := routedTo(t, "SSO", holders)

	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapse("e1", "c1")))
	got := current(t, svc)
	assert.Equal(t, domain.Escalated, got.Kind)
	assert.Equal(t, dsaID, got.AssigneeUserID)
	assert.Equal(t, "DSA", got.AssigneePositionCode)
	assert.Equal(t, "SLA lapsed: escalated to the DSA", got.Reason)
	assert.Equal(t, []string{"DSA"}, holders.asked)
	require.Len(t, pub.events, 1)
	assert.Equal(t, staffID, pub.events[0].FromUserID, "taken from the councillor")
	assert.Equal(t, dsaID, pub.events[0].ToUserID)
}

func TestEscalateOnLapse_HoDReportGoesUpToTheDAA(t *testing.T) {
	holders := &fakeHolders{holders: map[string]string{"DSA": dsaID, "DAA": daaID}}
	svc, pub := routedTo(t, "HOD_CT", holders)

	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapse("e1", "c1")))
	got := current(t, svc)
	assert.Equal(t, daaID, got.AssigneeUserID)
	assert.Equal(t, "DAA", got.AssigneePositionCode)
	require.Len(t, pub.events, 1)
	assert.Equal(t, headID, pub.events[0].FromUserID, "taken from the HoD")
	assert.Equal(t, daaID, pub.events[0].ToUserID)
}

func TestEscalateOnLapse_NobodyAboveGoesBackToTheRoutedHead(t *testing.T) {
	// The DSA position is vacant, so the report goes back to the SSO.
	holders := &fakeHolders{holders: map[string]string{}}
	svc, _ := routedTo(t, "SSO", holders)
	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapse("e1", "c1")))
	assert.Equal(t, headID, current(t, svc).AssigneeUserID)
	assert.Equal(t, "SSO", current(t, svc).AssigneePositionCode)

	// Without a lookup (as before SCRUM-158) it also goes back to the routed head.
	svc, _ = handedToStaff(t)
	require.NoError(t, svc.EscalateOnLapse(context.Background(), lapse("e1", "c1")))
	assert.Equal(t, headID, current(t, svc).AssigneeUserID)
}

func TestEscalateOnLapse_DirectoryDownIsTriedAgain(t *testing.T) {
	holders := &fakeHolders{err: errors.New("department-directory unavailable")}
	svc, pub := routedTo(t, "SSO", holders)

	err := svc.EscalateOnLapse(context.Background(), lapse("e1", "c1"))
	assert.ErrorContains(t, err, "unavailable", "the event bus tries the event again")
	assert.Equal(t, domain.Reassigned, current(t, svc).Kind, "nothing changed")
	assert.Empty(t, pub.events)
	assert.False(t, svc.saga.handled("e1"))
}
