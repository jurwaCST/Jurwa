package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
	"github.com/jurwaCST/Jurwa/services/photo-media/domain"
)

// UseCases is what the handler needs from the application layer.
type UseCases interface {
	UploadPhoto(ctx context.Context, in domain.NewPhotoInput) (domain.Photo, error)
	GetPhoto(ctx context.Context, id string) (domain.Photo, error)
}

// Handler implements photomediav1.PhotoMediaServiceServer.
type Handler struct {
	photomediav1.UnimplementedPhotoMediaServiceServer
	app UseCases
}

// NewHandler returns a gRPC handler backed by the use cases.
func NewHandler(app UseCases) *Handler {
	return &Handler{app: app}
}

// UploadPhoto stores one photo and returns its id. A rule error returns INVALID_ARGUMENT.
func (h *Handler) UploadPhoto(ctx context.Context, req *photomediav1.UploadPhotoRequest) (*photomediav1.UploadPhotoResponse, error) {
	photo, err := h.app.UploadPhoto(ctx, domain.NewPhotoInput{
		UploaderID:  req.GetUploaderId(),
		ContentType: req.GetContentType(),
		Data:        req.GetData(),
		FileName:    req.GetFileName(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &photomediav1.UploadPhotoResponse{PhotoId: photo.ID}, nil
}

// GetPhoto returns one photo. A malformed id returns INVALID_ARGUMENT, an unknown id NOT_FOUND.
func (h *Handler) GetPhoto(ctx context.Context, req *photomediav1.GetPhotoRequest) (*photomediav1.GetPhotoResponse, error) {
	photo, err := h.app.GetPhoto(ctx, req.GetPhotoId())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &photomediav1.GetPhotoResponse{
		PhotoId:     photo.ID,
		ContentType: photo.ContentType,
		Data:        photo.Data,
		SizeBytes:   photo.Size(),
		UploaderId:  photo.UploaderID,
	}
	if photo.ReportID != "" {
		resp.ReportId = &photo.ReportID
	}
	return resp, nil
}

// toStatus maps domain errors to gRPC status codes and hides internal details.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidPhoto):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPhotoNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
