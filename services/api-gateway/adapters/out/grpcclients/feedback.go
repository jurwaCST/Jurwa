package grpcclients

import (
	"context"

	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
)

// Feedback implements ports.Feedback.
type Feedback struct {
	C feedbackv1.FeedbackServiceClient
}

func (f Feedback) SubmitFeedback(ctx context.Context, in *feedbackv1.SubmitFeedbackRequest) (*feedbackv1.SubmitFeedbackResponse, error) {
	return f.C.SubmitFeedback(ctx, in)
}

func (f Feedback) GetFeedback(ctx context.Context, reportID string) (*feedbackv1.GetFeedbackResponse, error) {
	return f.C.GetFeedback(ctx, &feedbackv1.GetFeedbackRequest{ReportId: reportID})
}
