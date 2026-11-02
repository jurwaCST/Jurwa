package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
	"github.com/jurwaCST/Jurwa/services/feedback/application"
	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

// Handler implements FeedbackService from proto/feedback/v1.
type Handler struct {
	feedbackv1.UnimplementedFeedbackServiceServer
	app *application.Service
}

// NewHandler returns a handler that calls the given use cases.
func NewHandler(app *application.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) SubmitFeedback(ctx context.Context, req *feedbackv1.SubmitFeedbackRequest) (*feedbackv1.SubmitFeedbackResponse, error) {
	f, err := h.app.Submit(ctx, req.GetReportId(), req.GetStudentUserId(), req.GetRating(), req.GetComment())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &feedbackv1.SubmitFeedbackResponse{Feedback: toProto(f)}, nil
}

func (h *Handler) GetFeedback(ctx context.Context, req *feedbackv1.GetFeedbackRequest) (*feedbackv1.GetFeedbackResponse, error) {
	f, err := h.app.Get(ctx, req.GetReportId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return &feedbackv1.GetFeedbackResponse{Feedback: toProto(f)}, nil
}

func toProto(f domain.Feedback) *feedbackv1.Feedback {
	return &feedbackv1.Feedback{
		Id:            f.ID,
		ReportId:      f.ReportID,
		StudentUserId: f.StudentUserID,
		Rating:        f.Rating,
		Comment:       f.Comment,
		CreatedAt:     timestamppb.New(f.CreatedAt),
	}
}

// toStatusError maps domain errors to the gRPC codes promised in the .proto file.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidReportID),
		errors.Is(err, domain.ErrInvalidStudentID),
		errors.Is(err, domain.ErrInvalidRating),
		errors.Is(err, domain.ErrCommentTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrNotResolved):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUnavailable):
		return status.Error(codes.Unavailable, domain.ErrUnavailable.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
