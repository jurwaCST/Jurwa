package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/reassignment/domain"
	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

const (
	reportID = "11111111-1111-1111-1111-111111111111"
	headID   = "22222222-2222-2222-2222-222222222222"
	staffID  = "33333333-3333-3333-3333-333333333333"
	otherID  = "44444444-4444-4444-4444-444444444444"
)

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

type fakePublisher struct {
	events []ports.ReportReassigned
	err    error
}

func (f *fakePublisher) PublishReportReassigned(_ context.Context, ev ports.ReportReassigned) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, ev)
	return nil
}

func newService(repo ports.AssignmentRepository, pub ports.EventPublisher) *Service {
	n := 0
	newID := func() string { n++; return fmt.Sprintf("a%d", n) }
	return NewService(repo, pub, func() time.Time { return t0 }, newID)
}

// routedService has reportID routed to the Category Head.
func routedService(t *testing.T) (*Service, *fakePublisher) {
	pub := &fakePublisher{}
	svc := newService(memory.New(), pub)
	_, err := svc.RecordRouted(context.Background(), reportID, headID, "DSA")
	require.NoError(t, err)
	return svc, pub
}

func TestRecordRouted(t *testing.T) {
	svc, pub := routedService(t)
	ctx := context.Background()

	list, err := svc.ListReassignments(ctx, domain.RoleSystemAdmin, reportID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, domain.Routed, list[0].Kind)
	assert.Equal(t, "DSA", list[0].AssigneePositionCode)
	assert.Empty(t, pub.events, "routing is announced by routing-engine, not here")

	_, err = svc.RecordRouted(ctx, reportID, headID, "DSA")
	assert.ErrorIs(t, err, domain.ErrAlreadyAssigned)

	_, err = svc.RecordRouted(ctx, "bad", headID, "")
	assert.ErrorIs(t, err, ErrInvalidArgument)
	_, err = svc.RecordRouted(ctx, reportID, "bad", "")
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func TestReassignReport_HeadThenStaff(t *testing.T) {
	svc, pub := routedService(t)
	ctx := context.Background()

	// The Category Head hands it to a staff member.
	a, err := svc.ReassignReport(ctx, ReassignCommand{
		ReportID: reportID, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: domain.RoleCategoryHead,
	})
	require.NoError(t, err)
	assert.Equal(t, "a2", a.ID)
	assert.Equal(t, domain.Reassigned, a.Kind)
	assert.Equal(t, staffID, a.AssigneeUserID)
	assert.Equal(t, headID, a.AssignedBy)

	require.Len(t, pub.events, 1)
	assert.Equal(t, ports.ReportReassigned{
		AssignmentID: "a2", ReportID: reportID, FromUserID: headID, ToUserID: staffID,
		ActorID: headID, Reason: "Plumbing", AssignedAt: t0,
	}, pub.events[0])

	// The staff member, now the assignee, passes it on.
	_, err = svc.ReassignReport(ctx, ReassignCommand{
		ReportID: reportID, ToUserID: otherID, Reason: "Electrical", ActorID: staffID, Role: domain.RoleStaff,
	})
	require.NoError(t, err)

	list, _ := svc.ListReassignments(ctx, domain.RoleStaff, reportID)
	require.Len(t, list, 3)
	assert.Equal(t, []bool{false, false, true}, []bool{list[0].Current, list[1].Current, list[2].Current})

	// The first staff member is no longer the assignee, so cannot reassign again.
	_, err = svc.ReassignReport(ctx, ReassignCommand{
		ReportID: reportID, ToUserID: staffID, Reason: "back", ActorID: staffID, Role: domain.RoleStaff,
	})
	assert.ErrorIs(t, err, domain.ErrPermissionDenied)
}

func TestReassignReport_Errors(t *testing.T) {
	valid := ReassignCommand{ReportID: reportID, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: domain.RoleCategoryHead}
	tests := []struct {
		name    string
		change  func(c *ReassignCommand)
		wantErr error
	}{
		{"report id not a UUID", func(c *ReassignCommand) { c.ReportID = "abc" }, ErrInvalidArgument},
		{"to_user_id missing", func(c *ReassignCommand) { c.ToUserID = "" }, ErrInvalidArgument},
		{"actor_id missing", func(c *ReassignCommand) { c.ActorID = "" }, ErrInvalidArgument},
		{"blank reason", func(c *ReassignCommand) { c.Reason = " " }, domain.ErrInvalidReason},
		{"same assignee", func(c *ReassignCommand) { c.ToUserID = headID }, domain.ErrSameAssignee},
		{"student", func(c *ReassignCommand) { c.Role = domain.RoleStudent }, domain.ErrPermissionDenied},
		{"staff not the assignee", func(c *ReassignCommand) { c.Role = domain.RoleStaff; c.ActorID = otherID }, domain.ErrPermissionDenied},
		{"no assignment", func(c *ReassignCommand) { c.ReportID = otherID }, domain.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, pub := routedService(t)
			cmd := valid
			tt.change(&cmd)
			_, err := svc.ReassignReport(context.Background(), cmd)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Empty(t, pub.events, "no event on error")

			cur, _ := svc.ListReassignments(context.Background(), domain.RoleSystemAdmin, reportID)
			assert.Len(t, cur, 1, "nothing saved on error")
		})
	}
}

func TestListReassignments_Errors(t *testing.T) {
	svc, _ := routedService(t)
	_, err := svc.ListReassignments(context.Background(), domain.RoleStudent, reportID)
	assert.ErrorIs(t, err, domain.ErrPermissionDenied)
	_, err = svc.ListReassignments(context.Background(), domain.RoleStaff, "abc")
	assert.ErrorIs(t, err, ErrInvalidArgument)

	list, err := svc.ListReassignments(context.Background(), domain.RoleStaff, otherID)
	require.NoError(t, err)
	assert.Empty(t, list, "unknown report gives an empty list")
}

func TestReassignReport_PublishFails(t *testing.T) {
	svc, pub := routedService(t)
	pub.err = errors.New("redis down")
	_, err := svc.ReassignReport(context.Background(), ReassignCommand{
		ReportID: reportID, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: domain.RoleCategoryHead,
	})
	assert.ErrorContains(t, err, "redis down")
}

// conflictRepo simulates another reassignment saving first.
type conflictRepo struct{ *memory.Repository }

func (conflictRepo) Replace(_ context.Context, previousID string, _ domain.Assignment) error {
	if previousID == "" {
		return nil
	}
	return domain.ErrConflict
}

func (conflictRepo) Current(context.Context, string) (domain.Assignment, error) {
	return domain.NewRouted("a1", reportID, headID, "", t0), nil
}

type currentFailsRepo struct{ *memory.Repository }

func (currentFailsRepo) Current(context.Context, string) (domain.Assignment, error) {
	return domain.Assignment{}, errors.New("db down")
}

func TestReassignReport_StoreErrors(t *testing.T) {
	cmd := ReassignCommand{ReportID: reportID, ToUserID: staffID, Reason: "Plumbing", ActorID: headID, Role: domain.RoleCategoryHead}

	svc := newService(conflictRepo{memory.New()}, &fakePublisher{})
	_, err := svc.ReassignReport(context.Background(), cmd)
	assert.ErrorIs(t, err, domain.ErrConflict)

	svc = newService(currentFailsRepo{memory.New()}, &fakePublisher{})
	_, err = svc.ReassignReport(context.Background(), cmd)
	assert.ErrorContains(t, err, "db down")
}
