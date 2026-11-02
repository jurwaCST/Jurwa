package application

// Interaction tests with gomock. The other tests use hand-written fakes and
// check results; these check which calls are made, how often and in what
// order, which fakes cannot show (quality plan: fakes by default, gomock for
// interaction checks).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports/mocks"
)

type mocked struct {
	svc    *Service
	repo   *mocks.MockTimerRepository
	policy *mocks.MockPolicyClient
	pub    *mocks.MockEventPublisher
}

// newMocked builds the service on gomock doubles. Any call that is not
// expected fails the test.
func newMocked(t *testing.T, now time.Time) mocked {
	ctrl := gomock.NewController(t)
	m := mocked{
		repo:   mocks.NewMockTimerRepository(ctrl),
		policy: mocks.NewMockPolicyClient(ctrl),
		pub:    mocks.NewMockEventPublisher(ctrl),
	}
	m.svc = NewService(m.repo, m.policy, m.pub, func() time.Time { return now })
	return m
}

func runningTimer(id string, deadline time.Time) domain.Timer {
	return domain.Timer{ReportID: id, Category: domain.Hostel, StartedAt: deadline.Add(-time.Hour), Deadline: deadline, State: domain.Running}
}

func lapsedTimer(t domain.Timer) domain.Timer {
	t.State = domain.Lapsed
	return t
}

// Each due timer is published, then saved as LAPSED, one timer at a time.
// Publishing first means a failed save leads to a second event, never a lost one.
func TestListDueTimers_PublishesThenSavesInOrder(t *testing.T) {
	now := t0.Add(2 * time.Hour)
	m := newMocked(t, now)
	ctx := context.Background()
	a := runningTimer(r1, t0)
	b := runningTimer(r2, t0.Add(time.Hour))

	gomock.InOrder(
		m.repo.EXPECT().ListDue(ctx, now, defaultLimit).Return([]domain.Timer{a, b}, nil),
		m.pub.EXPECT().PublishSLALapsed(ctx, ports.SLALapsed{ReportID: r1, Category: domain.Hostel, Deadline: a.Deadline, LapsedAt: now}).Return(nil),
		m.repo.EXPECT().Update(ctx, lapsedTimer(a)).Return(nil),
		m.pub.EXPECT().PublishSLALapsed(ctx, ports.SLALapsed{ReportID: r2, Category: domain.Hostel, Deadline: b.Deadline, LapsedAt: now}).Return(nil),
		m.repo.EXPECT().Update(ctx, lapsedTimer(b)).Return(nil),
	)

	got, err := m.svc.ListDueTimers(ctx, time.Time{}, 0)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

// When publishing fails, the timer is not saved and later timers are not touched,
// so the next run retries them.
func TestListDueTimers_PublishFails_NothingSaved(t *testing.T) {
	now := t0.Add(2 * time.Hour)
	m := newMocked(t, now)
	ctx := context.Background()
	a := runningTimer(r1, t0)
	b := runningTimer(r2, t0)

	m.repo.EXPECT().ListDue(ctx, now, defaultLimit).Return([]domain.Timer{a, b}, nil)
	m.pub.EXPECT().PublishSLALapsed(ctx, gomock.Any()).Return(errors.New("redis down")).Times(1)
	m.repo.EXPECT().Update(gomock.Any(), gomock.Any()).Times(0)

	_, err := m.svc.ListDueTimers(ctx, time.Time{}, 0)
	assert.ErrorContains(t, err, "redis down")
}

// Bad input is rejected before any port is called.
func TestListDueTimers_BadLimit_NoCalls(t *testing.T) {
	m := newMocked(t, t0)
	_, err := m.svc.ListDueTimers(context.Background(), time.Time{}, 501)
	assert.ErrorIs(t, err, ErrInvalidArgument)
	// No EXPECT calls: gomock fails the test if any port was used.
}

// StartTimer asks sla-policy exactly once, then saves the timer with the
// deadline it computed.
func TestStartTimer_AsksPolicyOnceThenCreates(t *testing.T) {
	m := newMocked(t, t0)
	ctx := context.Background()

	gomock.InOrder(
		m.policy.EXPECT().Window(ctx, domain.AcademicBlock).Return(24*time.Hour, nil).Times(1),
		m.repo.EXPECT().Create(ctx, domain.Timer{
			ReportID: r1, Category: domain.AcademicBlock, StartedAt: t0, Deadline: t0.Add(24 * time.Hour), State: domain.Running,
		}).Return(nil),
	)

	_, err := m.svc.StartTimer(ctx, r1, domain.AcademicBlock)
	require.NoError(t, err)
}

// If sla-policy fails, no timer is created.
func TestStartTimer_PolicyDown_NoCreate(t *testing.T) {
	m := newMocked(t, t0)
	ctx := context.Background()

	m.policy.EXPECT().Window(ctx, domain.Hostel).Return(time.Duration(0), errors.New("connection refused"))
	m.repo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	_, err := m.svc.StartTimer(ctx, r1, domain.Hostel)
	assert.ErrorIs(t, err, ErrPolicyUnavailable)
}
