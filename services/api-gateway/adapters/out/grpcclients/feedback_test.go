package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
)

type fakeFeedbackClient struct {
	feedbackv1.FeedbackServiceClient
	submitted *feedbackv1.SubmitFeedbackRequest
	asked     string
}

func (f *fakeFeedbackClient) SubmitFeedback(_ context.Context, in *feedbackv1.SubmitFeedbackRequest, _ ...grpc.CallOption) (*feedbackv1.SubmitFeedbackResponse, error) {
	f.submitted = in
	return &feedbackv1.SubmitFeedbackResponse{Feedback: &feedbackv1.Feedback{Rating: in.GetRating()}}, nil
}

func (f *fakeFeedbackClient) GetFeedback(_ context.Context, in *feedbackv1.GetFeedbackRequest, _ ...grpc.CallOption) (*feedbackv1.GetFeedbackResponse, error) {
	f.asked = in.GetReportId()
	return &feedbackv1.GetFeedbackResponse{Feedback: &feedbackv1.Feedback{ReportId: in.GetReportId()}}, nil
}

func TestFeedbackClient(t *testing.T) {
	fake := &fakeFeedbackClient{}
	c := Feedback{C: fake}

	resp, err := c.SubmitFeedback(context.Background(), &feedbackv1.SubmitFeedbackRequest{ReportId: "r1", StudentUserId: "u1", Rating: 4})
	require.NoError(t, err)
	assert.Equal(t, int32(4), resp.GetFeedback().GetRating())
	assert.Equal(t, "u1", fake.submitted.GetStudentUserId())

	got, err := c.GetFeedback(context.Background(), "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", fake.asked)
	assert.Equal(t, "r1", got.GetFeedback().GetReportId())
}
