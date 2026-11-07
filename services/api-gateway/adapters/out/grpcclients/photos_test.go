package grpcclients

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
)

// photoServer stands in for photo-media: it keeps the last upload and returns
// a photo of the size it is told.
type photoServer struct {
	photomediav1.UnimplementedPhotoMediaServiceServer
	got      *photomediav1.UploadPhotoRequest
	sendSize int
}

func (s *photoServer) UploadPhoto(_ context.Context, in *photomediav1.UploadPhotoRequest) (*photomediav1.UploadPhotoResponse, error) {
	s.got = in
	return &photomediav1.UploadPhotoResponse{PhotoId: "p1"}, nil
}

func (s *photoServer) GetPhoto(_ context.Context, in *photomediav1.GetPhotoRequest) (*photomediav1.GetPhotoResponse, error) {
	return &photomediav1.GetPhotoResponse{PhotoId: in.GetPhotoId(), ContentType: "image/jpeg", Data: bytes.Repeat([]byte{1}, s.sendSize)}, nil
}

// startPhotoServer runs a real gRPC server in memory, set up like photo-media
// (it accepts 6 MB), and returns a Photos client dialled with opts.
func startPhotoServer(t *testing.T, srv *photoServer, opts ...grpc.DialOption) Photos {
	t.Helper()
	lis := bufconn.Listen(16 << 20)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(MaxPhotoMessageBytes))
	photomediav1.RegisterPhotoMediaServiceServer(s, srv)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	opts = append(opts,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	conn, err := grpc.NewClient("passthrough:///photo-media", opts...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return Photos{C: photomediav1.NewPhotoMediaServiceClient(conn)}
}

func TestPhotosSendsTheRequest(t *testing.T) {
	srv := &photoServer{}
	p := startPhotoServer(t, srv, PhotoMessageLimits())
	in := &photomediav1.UploadPhotoRequest{UploaderId: "u1", ContentType: "image/png", Data: []byte{1, 2}, FileName: "a.png"}
	resp, err := p.UploadPhoto(context.Background(), in)
	require.NoError(t, err)
	assert.Equal(t, "p1", resp.GetPhotoId())
	assert.Equal(t, "u1", srv.got.GetUploaderId())
	assert.Equal(t, "a.png", srv.got.GetFileName())

	got, err := p.GetPhoto(context.Background(), "p7")
	require.NoError(t, err)
	assert.Equal(t, "p7", got.GetPhotoId())
}

// A 5 MB photo is bigger than gRPC's 4 MB default for a received message, so
// without PhotoMessageLimits the gateway could not read a photo back.
func TestPhotoMessageLimitsAllowAFiveMBPhoto(t *testing.T) {
	const fiveMB = 5 << 20

	p := startPhotoServer(t, &photoServer{sendSize: fiveMB}, PhotoMessageLimits())
	got, err := p.GetPhoto(context.Background(), "p1")
	require.NoError(t, err)
	assert.Len(t, got.GetData(), fiveMB)

	_, err = p.UploadPhoto(context.Background(), &photomediav1.UploadPhotoRequest{Data: make([]byte, fiveMB+fiveMB/10)})
	require.NoError(t, err, "5.5 MB reaches photo-media, which gives the real size error")

	plain := startPhotoServer(t, &photoServer{sendSize: fiveMB})
	_, err = plain.GetPhoto(context.Background(), "p1")
	assert.Equal(t, codes.ResourceExhausted, status.Code(err), "the 4 MB default would block GET /photos")
}

func TestPhotoMessageLimitsStopAnythingOverSixMB(t *testing.T) {
	p := startPhotoServer(t, &photoServer{}, PhotoMessageLimits())
	_, err := p.UploadPhoto(context.Background(), &photomediav1.UploadPhotoRequest{Data: make([]byte, MaxPhotoMessageBytes+1)})
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}
