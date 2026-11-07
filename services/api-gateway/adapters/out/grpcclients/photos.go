package grpcclients

import (
	"context"

	"google.golang.org/grpc"

	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
)

// MaxPhotoMessageBytes is the largest gRPC message to and from photo-media:
// a 5 MB photo plus room for the other fields. photo-media's server accepts
// the same size (grpc.MaxRecvMsgSize(6 MB)).
const MaxPhotoMessageBytes = 6 << 20

// PhotoMessageLimits raises the client limits for the photo-media connection.
// The grpc-go default for a received message is 4 MB, which would block
// GET /photos for a photo between 4 and 5 MB.
func PhotoMessageLimits() grpc.DialOption {
	return grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(MaxPhotoMessageBytes),
		grpc.MaxCallSendMsgSize(MaxPhotoMessageBytes),
	)
}

// Photos implements ports.Photos.
type Photos struct {
	C photomediav1.PhotoMediaServiceClient
}

func (p Photos) UploadPhoto(ctx context.Context, in *photomediav1.UploadPhotoRequest) (*photomediav1.UploadPhotoResponse, error) {
	return p.C.UploadPhoto(ctx, in)
}

func (p Photos) GetPhoto(ctx context.Context, photoID string) (*photomediav1.GetPhotoResponse, error) {
	return p.C.GetPhoto(ctx, &photomediav1.GetPhotoRequest{PhotoId: photoID})
}
