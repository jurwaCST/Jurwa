package slapolicy

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

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slapolicyv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-policy/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

// fakeServer answers HOSTEL with 48 hours, OTHERS with an internal error,
// and every other category with NOT_FOUND.
type fakeServer struct {
	slapolicyv1.UnimplementedSlaPolicyServiceServer
}

func (fakeServer) GetPolicy(_ context.Context, req *slapolicyv1.GetPolicyRequest) (*slapolicyv1.GetPolicyResponse, error) {
	switch req.GetCategory() {
	case commonv1.Category_HOSTEL:
		return &slapolicyv1.GetPolicyResponse{Policy: &slapolicyv1.Policy{Category: req.GetCategory(), WindowHours: 48}}, nil
	case commonv1.Category_OTHERS:
		return nil, status.Error(codes.Internal, "db down")
	default:
		return nil, status.Error(codes.NotFound, "no policy")
	}
}

// newClient runs fakeServer in memory (bufconn) and returns a Client for it.
func newClient(t *testing.T) *Client {
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	slapolicyv1.RegisterSlaPolicyServiceServer(srv, fakeServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return New(slapolicyv1.NewSlaPolicyServiceClient(conn))
}

func TestWindow(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	w, err := c.Window(ctx, domain.Hostel)
	require.NoError(t, err)
	assert.Equal(t, 48*time.Hour, w)

	_, err = c.Window(ctx, domain.AcademicBlock)
	assert.ErrorIs(t, err, ports.ErrNoPolicy, "NOT_FOUND becomes ErrNoPolicy")

	_, err = c.Window(ctx, domain.Others)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ports.ErrNoPolicy, "other errors are passed on")
	assert.Contains(t, err.Error(), "db down")
}
