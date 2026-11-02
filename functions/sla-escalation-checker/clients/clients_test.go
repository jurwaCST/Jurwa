package clients

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jurwaCST/Jurwa/functions/sla-escalation-checker/checker"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slatrackingv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-tracking/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

var deadline = time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)

type fakeTracking struct {
	slatrackingv1.UnimplementedSlaTrackingServiceServer
	gotLimit int32
}

func (f *fakeTracking) ListDueTimers(_ context.Context, req *slatrackingv1.ListDueTimersRequest) (*slatrackingv1.ListDueTimersResponse, error) {
	f.gotLimit = req.GetLimit()
	if req.GetNow() != nil {
		return nil, status.Error(codes.InvalidArgument, "the function must use server time")
	}
	if req.GetLimit() == 7 {
		return nil, status.Error(codes.Internal, "store down")
	}
	return &slatrackingv1.ListDueTimersResponse{Timers: []*slatrackingv1.Timer{
		{ReportId: "r1", Deadline: timestamppb.New(deadline), State: slatrackingv1.TimerState_LAPSED},
	}}, nil
}

type fakeLifecycle struct {
	statuslifecyclev1.UnimplementedStatusLifecycleServiceServer
	got *statuslifecyclev1.ChangeStatusRequest
}

func (f *fakeLifecycle) ChangeStatus(_ context.Context, req *statuslifecyclev1.ChangeStatusRequest) (*statuslifecyclev1.ChangeStatusResponse, error) {
	f.got = req
	switch req.GetReportId() {
	case "resolved":
		return nil, status.Error(codes.FailedPrecondition, "RESOLVED to ESCALATED is not allowed")
	case "unknown":
		return nil, status.Error(codes.NotFound, "report not found")
	}
	return &statuslifecyclev1.ChangeStatusResponse{}, nil
}

// dial runs the fake servers in memory (bufconn) and returns a connection.
func dial(t *testing.T, tracking *fakeTracking, lifecycle *fakeLifecycle) *grpc.ClientConn {
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	slatrackingv1.RegisterSlaTrackingServiceServer(srv, tracking)
	statuslifecyclev1.RegisterStatusLifecycleServiceServer(srv, lifecycle)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestSlaTracking_LapseDue(t *testing.T) {
	tracking := &fakeTracking{}
	c := NewSlaTracking(slatrackingv1.NewSlaTrackingServiceClient(dial(t, tracking, &fakeLifecycle{})))

	got, err := c.LapseDue(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, []checker.Lapsed{{ReportID: "r1", Deadline: deadline}}, got)
	assert.Equal(t, int32(100), tracking.gotLimit)

	_, err = c.LapseDue(context.Background(), 7)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestStatusLifecycle_Escalate(t *testing.T) {
	lifecycle := &fakeLifecycle{}
	c := NewStatusLifecycle(statuslifecyclev1.NewStatusLifecycleServiceClient(dial(t, &fakeTracking{}, lifecycle)))
	ctx := context.Background()

	require.NoError(t, c.Escalate(ctx, "r1", "SLA lapsed"))
	assert.Equal(t, commonv1.Status_ESCALATED, lifecycle.got.GetToStatus())
	assert.Empty(t, lifecycle.got.GetActorId(), "the system acts")
	assert.Equal(t, "SLA lapsed", lifecycle.got.GetReason())

	err := c.Escalate(ctx, "resolved", "SLA lapsed")
	assert.ErrorIs(t, err, checker.ErrNotEscalatable)
	assert.Contains(t, err.Error(), "RESOLVED to ESCALATED")

	err = c.Escalate(ctx, "unknown", "SLA lapsed")
	assert.NotErrorIs(t, err, checker.ErrNotEscalatable)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
