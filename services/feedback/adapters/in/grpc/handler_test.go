package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
	"github.com/jurwaCST/Jurwa/services/feedback/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/feedback/application"
	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

const (
	reportID  = "33333333-3333-3333-3333-333333333333"
	studentID = "22222222-2222-2222-2222-222222222222"
)

type fakeStatus struct {
	resolved bool
	err      error
}

func (f fakeStatus) IsResolved(context.Context, string) (bool, error) { return f.resolved, f.err }

func newHandler(s fakeStatus) *Handler {
	return NewHandler(application.NewService(memory.New(), s))
}

func TestSubmitThenGet(t *testing.T) {
	h := newHandler(fakeStatus{resolved: true})
	ctx := context.Background()

	sub, err := h.SubmitFeedback(ctx, &feedbackv1.SubmitFeedbackRequest{
		ReportId: reportID, StudentUserId: studentID, Rating: 4, Comment: "Fixed in a day",
	})
	require.NoError(t, err)
	f := sub.GetFeedback()
	assert.NotEmpty(t, f.GetId())
	assert.Equal(t, reportID, f.GetReportId())
	assert.Equal(t, studentID, f.GetStudentUserId())
	assert.Equal(t, int32(4), f.GetRating())
	assert.Equal(t, "Fixed in a day", f.GetComment())
	assert.NotNil(t, f.GetCreatedAt())

	got, err := h.GetFeedback(ctx, &feedbackv1.GetFeedbackRequest{ReportId: reportID})
	require.NoError(t, err)
	assert.Equal(t, f.GetId(), got.GetFeedback().GetId())
}

func TestErrorCodes(t *testing.T) {
	ctx := context.Background()
	valid := func() *feedbackv1.SubmitFeedbackRequest {
		return &feedbackv1.SubmitFeedbackRequest{ReportId: reportID, StudentUserId: studentID, Rating: 5}
	}

	h := newHandler(fakeStatus{resolved: true})
	bad := valid()
	bad.Rating = 0
	_, err := h.SubmitFeedback(ctx, bad)
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "rating 0")

	_, err = h.SubmitFeedback(ctx, valid())
	require.NoError(t, err)
	_, err = h.SubmitFeedback(ctx, valid())
	assert.Equal(t, codes.AlreadyExists, status.Code(err), "second feedback")

	_, err = newHandler(fakeStatus{resolved: false}).SubmitFeedback(ctx, valid())
	assert.Equal(t, codes.FailedPrecondition, status.Code(err), "not resolved")

	_, err = newHandler(fakeStatus{err: fmt.Errorf("%w: down", domain.ErrUnavailable)}).SubmitFeedback(ctx, valid())
	assert.Equal(t, codes.Unavailable, status.Code(err), "status-lifecycle down")

	_, err = h.GetFeedback(ctx, &feedbackv1.GetFeedbackRequest{ReportId: "44444444-4444-4444-4444-444444444444"})
	assert.Equal(t, codes.NotFound, status.Code(err), "no feedback")

	_, err = h.GetFeedback(ctx, &feedbackv1.GetFeedbackRequest{ReportId: "x"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "bad report id")
}

func TestToStatusErrorInternal(t *testing.T) {
	err := toStatusError(errors.New("db down"))
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "db down")
}
