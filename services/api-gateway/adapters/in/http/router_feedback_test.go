package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
)

// ---- feedback routes (SCRUM-147) ----

const feedbackPath = "/reports/6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f/feedback"

// fakeFeedback records what the gateway sends to feedback.
type fakeFeedback struct {
	submitted *feedbackv1.SubmitFeedbackRequest
	asked     string
	submitErr error
	getErr    error
}

func (f *fakeFeedback) SubmitFeedback(_ context.Context, in *feedbackv1.SubmitFeedbackRequest) (*feedbackv1.SubmitFeedbackResponse, error) {
	f.submitted = in
	if f.submitErr != nil {
		return nil, f.submitErr
	}
	return &feedbackv1.SubmitFeedbackResponse{Feedback: &feedbackv1.Feedback{
		Id: "f1", ReportId: in.GetReportId(), StudentUserId: in.GetStudentUserId(), Rating: in.GetRating(), Comment: in.GetComment(),
	}}, nil
}

func (f *fakeFeedback) GetFeedback(_ context.Context, reportID string) (*feedbackv1.GetFeedbackResponse, error) {
	f.asked = reportID
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &feedbackv1.GetFeedbackResponse{Feedback: &feedbackv1.Feedback{Id: "f1", ReportId: reportID, StudentUserId: "u-student", Rating: 5}}, nil
}

func TestFeedbackNeedsAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("POST", feedbackPath, "", `{"rating":5}`).Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", feedbackPath, "", "").Code)
	assert.Nil(t, e.feedback.submitted)
	assert.Empty(t, e.feedback.asked)
}

func TestSubmitFeedbackTakesTheReportFromThePathAndTheStudentFromTheJWT(t *testing.T) {
	e := newEnv()
	body := `{"reportId":"r-other","studentUserId":"u-other","rating":4,"comment":"Fixed in two days"}`
	rec := e.do("POST", feedbackPath, "student-token", body)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	got := e.feedback.submitted
	require.NotNil(t, got)
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", got.GetReportId(), "the report comes from the path, never the body")
	assert.Equal(t, "u-student", got.GetStudentUserId(), "the student comes from the JWT, never the body")
	assert.Equal(t, int32(4), got.GetRating())
	assert.Equal(t, "Fixed in two days", got.GetComment())

	out := decode(t, rec)["feedback"].(map[string]any)
	assert.Equal(t, "u-student", out["studentUserId"], "lowerCamelCase JSON")
	assert.Equal(t, float64(4), out["rating"])
}

// feedback trusts the gateway to check that the student is the reporter.
func TestFeedbackOnlyOnTheStudentsOwnReport(t *testing.T) {
	e := newEnv()
	users["other-student"] = &identityv1.User{Id: "u-other", Roles: []commonv1.Role{commonv1.Role_STUDENT}}
	defer delete(users, "other-student")

	assert.Equal(t, http.StatusForbidden, e.do("POST", feedbackPath, "other-student", `{"rating":5}`).Code)
	assert.Equal(t, http.StatusForbidden, e.do("GET", feedbackPath, "other-student", "").Code)
	assert.Nil(t, e.feedback.submitted, "feedback is not called for another student's report")
	assert.Empty(t, e.feedback.asked)

	e.reports.getErr = status.Error(codes.NotFound, "report not found")
	assert.Equal(t, http.StatusNotFound, e.do("POST", feedbackPath, "student-token", `{"rating":5}`).Code)
}

func TestFeedbackRoles(t *testing.T) {
	for token, want := range map[string][2]int{ // POST, GET
		"student-token": {http.StatusCreated, http.StatusOK},
		"head-token":    {http.StatusForbidden, http.StatusOK},
		"admin-token":   {http.StatusForbidden, http.StatusOK},
		"staff-token":   {http.StatusForbidden, http.StatusForbidden},
	} {
		t.Run(token, func(t *testing.T) {
			e := newEnv()
			assert.Equal(t, want[0], e.do("POST", feedbackPath, token, `{"rating":5}`).Code)
			assert.Equal(t, want[1], e.do("GET", feedbackPath, token, "").Code)
		})
	}
}

func TestGetFeedback(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", feedbackPath, "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", e.feedback.asked)
	out := decode(t, rec)["feedback"].(map[string]any)
	assert.Equal(t, float64(5), out["rating"])
	assert.Equal(t, "", out["comment"], "empty fields are still sent")

	e.feedback.getErr = status.Error(codes.NotFound, "no feedback for this report")
	assert.Equal(t, http.StatusNotFound, e.do("GET", feedbackPath, "student-token", "").Code)
}

func TestSubmitFeedbackErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "rating 0 or 6", err: status.Error(codes.InvalidArgument, "rating must be 1 to 5"), want: http.StatusBadRequest},
		{name: "not resolved", err: status.Error(codes.FailedPrecondition, "report is not resolved"), want: http.StatusUnprocessableEntity},
		{name: "second feedback", err: status.Error(codes.AlreadyExists, "report already has feedback"), want: http.StatusConflict},
		{name: "service down", err: status.Error(codes.Unavailable, "dial feedback:9019"), want: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			e.feedback.submitErr = tt.err
			assert.Equal(t, tt.want, e.do("POST", feedbackPath, "student-token", `{"rating":6}`).Code)
		})
	}
}

func TestSubmitFeedbackRejectsBadJSON(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("POST", feedbackPath, "student-token", `{"rating":`).Code)
	assert.Nil(t, e.feedback.submitted)
}
