package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slatrackingv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-tracking/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/application"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

const reportID = "11111111-1111-1111-1111-111111111111"

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

type fakePolicy struct{ err error }

func (f fakePolicy) Window(_ context.Context, c domain.Category) (time.Duration, error) {
	if f.err != nil {
		return 0, f.err
	}
	if c == domain.Others {
		return 0, ports.ErrNoPolicy
	}
	return 24 * time.Hour, nil
}

type nopPublisher struct{}

func (nopPublisher) PublishSLALapsed(context.Context, ports.SLALapsed) error { return nil }

func newHandler(policy ports.PolicyClient, now *time.Time) *Handler {
	return NewHandler(application.NewService(memory.New(), policy, nopPublisher{}, func() time.Time { return *now }))
}

func TestTimerLifecycle(t *testing.T) {
	now := t0
	h := newHandler(fakePolicy{}, &now)
	ctx := context.Background()

	start, err := h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID, Category: commonv1.Category_HOSTEL})
	require.NoError(t, err)
	assert.Equal(t, slatrackingv1.TimerState_RUNNING, start.GetTimer().GetState())
	assert.Equal(t, commonv1.Category_HOSTEL, start.GetTimer().GetCategory())
	assert.Equal(t, t0.Add(24*time.Hour), start.GetTimer().GetDeadline().AsTime())

	got, err := h.GetTimer(ctx, &slatrackingv1.GetTimerRequest{ReportId: reportID})
	require.NoError(t, err)
	assert.Equal(t, reportID, got.GetTimer().GetReportId())

	// Server time: 25 hours later, so the timer has lapsed.
	now = t0.Add(25 * time.Hour)
	due, err := h.ListDueTimers(ctx, &slatrackingv1.ListDueTimersRequest{})
	require.NoError(t, err)
	require.Len(t, due.GetTimers(), 1)
	assert.Equal(t, slatrackingv1.TimerState_LAPSED, due.GetTimers()[0].GetState())

	stop, err := h.StopTimer(ctx, &slatrackingv1.StopTimerRequest{ReportId: reportID})
	require.NoError(t, err)
	assert.Equal(t, slatrackingv1.TimerState_STOPPED, stop.GetTimer().GetState())
}

func TestListDueTimers_ExplicitNow(t *testing.T) {
	now := t0.Add(100 * time.Hour)
	h := newHandler(fakePolicy{}, &now)
	ctx := context.Background()
	_, err := h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID, Category: commonv1.Category_HOSTEL})
	require.NoError(t, err)

	due, err := h.ListDueTimers(ctx, &slatrackingv1.ListDueTimersRequest{Now: timestamppb.New(now.Add(time.Hour))})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "now in the future")
	assert.Nil(t, due)

	due, err = h.ListDueTimers(ctx, &slatrackingv1.ListDueTimersRequest{Now: timestamppb.New(now.Add(-time.Hour)), Limit: 5})
	require.NoError(t, err)
	assert.Empty(t, due.GetTimers(), "deadline is now+24h, so nothing is due")
}

func TestErrorCodes(t *testing.T) {
	now := t0
	h := newHandler(fakePolicy{}, &now)
	ctx := context.Background()

	_, err := h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: "abc", Category: commonv1.Category_HOSTEL})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "report id not a UUID")

	_, err = h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "category unspecified")

	_, err = h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID, Category: commonv1.Category_OTHERS})
	assert.Equal(t, codes.FailedPrecondition, status.Code(err), "no SLA window for OTHERS")

	_, err = h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID, Category: commonv1.Category_HOSTEL})
	require.NoError(t, err)
	_, err = h.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID, Category: commonv1.Category_HOSTEL})
	assert.Equal(t, codes.AlreadyExists, status.Code(err), "second start")

	other := "22222222-2222-2222-2222-222222222222"
	_, err = h.StopTimer(ctx, &slatrackingv1.StopTimerRequest{ReportId: other})
	assert.Equal(t, codes.NotFound, status.Code(err), "stop without timer")
	_, err = h.GetTimer(ctx, &slatrackingv1.GetTimerRequest{ReportId: other})
	assert.Equal(t, codes.NotFound, status.Code(err), "get without timer")

	_, err = h.ListDueTimers(ctx, &slatrackingv1.ListDueTimersRequest{Limit: 501})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "limit above 500")

	down := newHandler(fakePolicy{err: errors.New("connection refused")}, &now)
	_, err = down.StartTimer(ctx, &slatrackingv1.StartTimerRequest{ReportId: reportID, Category: commonv1.Category_HOSTEL})
	assert.Equal(t, codes.Unavailable, status.Code(err), "sla-policy down")
}

func TestToStatusError(t *testing.T) {
	assert.Equal(t, codes.FailedPrecondition, status.Code(toStatusError(domain.ErrInvalidWindow)))
	assert.Equal(t, codes.Internal, status.Code(toStatusError(errors.New("store down"))))
}
