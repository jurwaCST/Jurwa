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

// GetRatingSummary is the count and average rating (SCRUM-164).
func (f Feedback) GetRatingSummary(ctx context.Context) (*feedbackv1.GetRatingSummaryResponse, error) {
	return f.C.GetRatingSummary(ctx, &feedbackv1.GetRatingSummaryRequest{})
}
