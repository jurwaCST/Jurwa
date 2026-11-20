package http

// SCRUM-163: the student who sent a report changes its description or
// withdraws it, while nobody has started work on it (Open or Assigned).

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
)

func (f *fakeReports) UpdateDescription(_ context.Context, in *reportsubmissionv1.UpdateDescriptionRequest) (*reportsubmissionv1.UpdateDescriptionResponse, error) {
	f.updated = in
	return &reportsubmissionv1.UpdateDescriptionResponse{Report: &commonv1.Report{Id: in.GetReportId(), Description: in.GetDescription()}}, nil
}

func (f *fakeReports) WithdrawReport(_ context.Context, in *reportsubmissionv1.WithdrawReportRequest) (*reportsubmissionv1.WithdrawReportResponse, error) {
	f.withdrawn = in
	return &reportsubmissionv1.WithdrawReportResponse{}, nil
}

func withOtherStudent(t *testing.T) {
	t.Helper()
	users["student2-token"] = &identityv1.User{Id: "u-student-2", Roles: []commonv1.Role{commonv1.Role_STUDENT}}
	t.Cleanup(func() { delete(users, "student2-token") })
}

func TestEditDescription(t *testing.T) {
	e := newEnv()
	rec := e.do("PATCH", "/reports/r1", "student-token", `{"description":"Tap leaking, water on the floor"}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, e.reports.updated)
	assert.Equal(t, "r1", e.reports.updated.GetReportId())
	assert.Equal(t, "u-student", e.reports.updated.GetReporterId(), "the student comes from the token")
	assert.Equal(t, "Tap leaking, water on the floor", e.reports.updated.GetDescription())
	assert.Contains(t, rec.Body.String(), "Tap leaking, water on the floor")
}

func TestWithdrawReport(t *testing.T) {
	e := newEnv()
	e.history.holder = "u-sso"
	rec := e.do("DELETE", "/reports/r1", "student-token", "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	require.NotNil(t, e.reports.withdrawn)
	assert.Equal(t, "r1", e.reports.withdrawn.GetReportId())
	assert.Equal(t, "u-student", e.reports.withdrawn.GetReporterId())
	assert.Equal(t, "u-sso", e.reports.withdrawn.GetCurrentAssigneeId(), "so the person holding it is told")
}

// Once someone works on it, it can no longer be changed or withdrawn.
func TestEditOrWithdrawAfterWorkStartedIs422(t *testing.T) {
	for _, now := range []commonv1.Status{commonv1.Status_IN_PROGRESS, commonv1.Status_RESOLVED, commonv1.Status_ESCALATED} {
		e := newEnv()
		e.status.current = now
		assert.Equal(t, http.StatusUnprocessableEntity, e.do("PATCH", "/reports/r1", "student-token", `{"description":"x"}`).Code, now.String())
		assert.Equal(t, http.StatusUnprocessableEntity, e.do("DELETE", "/reports/r1", "student-token", "").Code, now.String())
		assert.Nil(t, e.reports.updated)
		assert.Nil(t, e.reports.withdrawn)
	}
}

func TestOnlyTheStudentWhoSentItCanEditOrWithdraw(t *testing.T) {
	withOtherStudent(t)
	for token, want := range map[string]int{
		"student2-token": http.StatusForbidden,
		"head-token":     http.StatusForbidden,
		"staff-token":    http.StatusForbidden,
		"admin-token":    http.StatusForbidden,
		"":               http.StatusUnauthorized,
	} {
		e := newEnv()
		assert.Equal(t, want, e.do("PATCH", "/reports/r1", token, `{"description":"x"}`).Code, token)
		assert.Equal(t, want, e.do("DELETE", "/reports/r1", token, "").Code, token)
		assert.Nil(t, e.reports.updated, token)
		assert.Nil(t, e.reports.withdrawn, token)
	}
}
