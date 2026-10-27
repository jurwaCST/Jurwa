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
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

const (
	reportID = "11111111-1111-1111-1111-111111111111"
	staffID  = "22222222-2222-2222-2222-222222222222"
)

// fakePublisher records events so tests can check them.
type fakePublisher struct {
	events []ports.StatusChanged
	err    error
}

func (f *fakePublisher) PublishStatusChanged(_ context.Context, ev ports.StatusChanged) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, ev)
	return nil
}

var fixedNow = time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC)

func newTestService(t *testing.T) (*Service, *fakePublisher) {
	t.Helper()
	pub := &fakePublisher{}
	svc := NewService(memory.New(), pub, func() time.Time { return fixedNow })
	_, err := svc.Register(context.Background(), reportID)
	require.NoError(t, err)
	return svc, pub
}

func TestChangeStatus_ValidChangeIsSavedAndPublished(t *testing.T) {
	svc, pub := newTestService(t)
	ctx := context.Background()

	rs, err := svc.ChangeStatus(ctx, ChangeStatusCommand{ReportID: reportID, To: domain.Assigned, Reason: "routed to DSA"})
	require.NoError(t, err)
	assert.Equal(t, domain.Assigned, rs.Status)
	assert.Equal(t, int32(2), rs.Version)

	got, err := svc.GetStatus(ctx, reportID)
	require.NoError(t, err)
	assert.Equal(t, domain.Assigned, got.Status)

	require.Len(t, pub.events, 1)
	assert.Equal(t, ports.StatusChanged{
		ReportID: reportID, From: domain.Open, To: domain.Assigned, Version: 2, Reason: "routed to DSA",
	}, pub.events[0])
}

func TestChangeStatus_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cmd     ChangeStatusCommand
		wantErr error
	}{
		{"report id not a UUID", ChangeStatusCommand{ReportID: "abc", To: domain.Assigned}, ErrInvalidArgument},
		{"actor id not a UUID", ChangeStatusCommand{ReportID: reportID, To: domain.Assigned, ActorID: "bob"}, ErrInvalidArgument},
		{"unknown report", ChangeStatusCommand{ReportID: staffID, To: domain.Assigned}, domain.ErrNotFound},
		{"OPEN straight to RESOLVED", ChangeStatusCommand{ReportID: reportID, To: domain.Resolved}, domain.ErrInvalidTransition},
		{"user cannot escalate", ChangeStatusCommand{ReportID: reportID, To: domain.Escalated, ActorID: staffID}, domain.ErrInvalidTransition},
		{"status unspecified", ChangeStatusCommand{ReportID: reportID, To: domain.StatusUnspecified}, domain.ErrInvalidStatus},
		{"stale expected version", ChangeStatusCommand{ReportID: reportID, To: domain.Assigned, ExpectedVersion: 7}, domain.ErrVersionConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, pub := newTestService(t)
			_, err := svc.ChangeStatus(context.Background(), tt.cmd)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Empty(t, pub.events, "nothing is published when the change fails")
		})
	}
}

func TestChangeStatus_EscalationAndCompensation(t *testing.T) {
	svc, pub := newTestService(t)
	ctx := context.Background()

	_, err := svc.ChangeStatus(ctx, ChangeStatusCommand{ReportID: reportID, To: domain.Assigned})
	require.NoError(t, err)
	rs, err := svc.ChangeStatus(ctx, ChangeStatusCommand{ReportID: reportID, To: domain.Escalated, Reason: "SLA lapsed"})
	require.NoError(t, err)
	assert.Equal(t, domain.Assigned, rs.Previous)

	// A later Saga step failed: undo the escalation.
	rs, err = svc.ChangeStatus(ctx, ChangeStatusCommand{ReportID: reportID, To: rs.Previous, Reason: "compensation"})
	require.NoError(t, err)
	assert.Equal(t, domain.Assigned, rs.Status)
	assert.Len(t, pub.events, 3)
}

func TestChangeStatus_MatchingExpectedVersion(t *testing.T) {
	svc, _ := newTestService(t)
	rs, err := svc.ChangeStatus(context.Background(), ChangeStatusCommand{ReportID: reportID, To: domain.Assigned, ExpectedVersion: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(2), rs.Version)
}

func TestChangeStatus_PublishError(t *testing.T) {
	svc, pub := newTestService(t)
	pub.err = errors.New("redis down")
	_, err := svc.ChangeStatus(context.Background(), ChangeStatusCommand{ReportID: reportID, To: domain.Assigned})
	assert.ErrorContains(t, err, "redis down")
}

func TestRegister(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, reportID)
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)

	_, err = svc.Register(ctx, "not-a-uuid")
	assert.ErrorIs(t, err, ErrInvalidArgument)

	_, err = svc.GetStatus(ctx, "not-a-uuid")
	assert.ErrorIs(t, err, ErrInvalidArgument)
}
