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
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

// ---- status routes (SCRUM-128) ----

// fakeStatus records what the gateway sends to status-lifecycle.
type fakeStatus struct {
	changed   *statuslifecyclev1.ChangeStatusRequest
	asked     string
	changeErr error
	getErr    error
}

func (f *fakeStatus) ChangeStatus(_ context.Context, in *statuslifecyclev1.ChangeStatusRequest) (*statuslifecyclev1.ChangeStatusResponse, error) {
	f.changed = in
	if f.changeErr != nil {
		return nil, f.changeErr
	}
	actor := in.GetActorId()
	return &statuslifecyclev1.ChangeStatusResponse{Status: &statuslifecyclev1.ReportStatus{
		ReportId: in.GetReportId(), Status: in.GetToStatus(), PreviousStatus: commonv1.Status_ASSIGNED, Version: 2, UpdatedBy: &actor,
	}}, nil
}

func (f *fakeStatus) GetStatus(_ context.Context, reportID string) (*statuslifecyclev1.GetStatusResponse, error) {
	f.asked = reportID
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &statuslifecyclev1.GetStatusResponse{Status: &statuslifecyclev1.ReportStatus{ReportId: reportID, Status: commonv1.Status_ASSIGNED, Version: 1}}, nil
}

func TestStatusRoutesNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("POST", "/reports/r1/status", "", `{"toStatus":"IN_PROGRESS"}`).Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/reports/r1/status", "", "").Code)
	assert.Nil(t, e.status.changed)
	assert.Empty(t, e.status.asked)
}

func TestChangeStatusTakesTheReportFromThePathAndTheActorFromTheJWT(t *testing.T) {
	e := newEnv()
	body := `{"reportId":"r-other","actorId":"","toStatus":"IN_PROGRESS","expectedVersion":1,"reason":"Plumber on the way"}`
	rec := e.do("POST", "/reports/r1/status", "head-token", body)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := e.status.changed
	require.NotNil(t, got)
	assert.Equal(t, "r1", got.GetReportId(), "the report comes from the path, never the body")
	assert.Equal(t, "u-head", got.GetActorId(), "the actor comes from the JWT; an empty actor would mean the system")
	assert.Equal(t, commonv1.Status_IN_PROGRESS, got.GetToStatus())
	assert.Equal(t, int32(1), got.GetExpectedVersion())
	assert.Equal(t, "Plumber on the way", got.GetReason())

	out := decode(t, rec)["status"].(map[string]any)
	assert.Equal(t, "IN_PROGRESS", out["status"])
	assert.Equal(t, "u-head", out["updatedBy"])
}

func TestChangeStatusRoles(t *testing.T) {
	for token, want := range map[string]int{
		"head-token":    http.StatusOK,
		"staff-token":   http.StatusOK,
		"student-token": http.StatusForbidden,
		"admin-token":   http.StatusForbidden,
	} {
		e := newEnv()
		rec := e.do("POST", "/reports/r1/status", token, `{"toStatus":"IN_PROGRESS"}`)
		assert.Equal(t, want, rec.Code, token)
		if want == http.StatusForbidden {
			assert.Nil(t, e.status.changed, token+" never reaches status-lifecycle")
		}
	}
}

func TestChangeStatusBadInput(t *testing.T) {
	for name, body := range map[string]string{
		"not JSON":       `toStatus=IN_PROGRESS`,
		"unknown status": `{"toStatus":"DONE"}`,
		"unknown field":  `{"toStatus":"IN_PROGRESS","priority":"high"}`,
	} {
		e := newEnv()
		assert.Equal(t, http.StatusBadRequest, e.do("POST", "/reports/r1/status", "head-token", body).Code, name)
		assert.Nil(t, e.status.changed, name)
	}

	e := newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("POST", "/reports/r1/status", "head-token", `{}`).Code, "no toStatus")
	assert.Nil(t, e.status.changed, "a missing status never reaches status-lifecycle")
}

// status-lifecycle enforces the flow and the version; its answers pass through.
func TestChangeStatusErrorsFromStatusLifecycle(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"transition not allowed, for example to ESCALATED by a person", status.Error(codes.FailedPrecondition, "invalid status transition"), http.StatusUnprocessableEntity},
		{"someone changed it first", status.Error(codes.Aborted, "version mismatch"), http.StatusConflict},
		{"unknown report", status.Error(codes.NotFound, "report not found"), http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			e.status.changeErr = tt.err
			rec := e.do("POST", "/reports/r1/status", "staff-token", `{"toStatus":"ESCALATED"}`)
			assert.Equal(t, tt.want, rec.Code, rec.Body.String())
		})
	}
}

func TestGetStatusOwnership(t *testing.T) {
	// fakeReports says report r1 belongs to u-student.
	for _, token := range []string{"student-token", "head-token", "staff-token", "admin-token"} {
		e := newEnv()
		rec := e.do("GET", "/reports/r1/status", token, "")
		require.Equal(t, http.StatusOK, rec.Code, token)
		assert.Equal(t, "r1", e.status.asked)
		assert.Equal(t, "ASSIGNED", decode(t, rec)["status"].(map[string]any)["status"])
	}

	users["other-student"] = &identityv1.User{Id: "u-other", Roles: []commonv1.Role{commonv1.Role_STUDENT}}
	defer delete(users, "other-student")
	e := newEnv()
	assert.Equal(t, http.StatusForbidden, e.do("GET", "/reports/r1/status", "other-student", "").Code, "another student's report")
	assert.Empty(t, e.status.asked, "the status is never fetched")
}

func TestGetStatusUnknownReport(t *testing.T) {
	e := newEnv()
	e.reports.getErr = status.Error(codes.NotFound, "report not found")
	assert.Equal(t, http.StatusNotFound, e.do("GET", "/reports/r9/status", "student-token", "").Code)

	e = newEnv()
	e.status.getErr = status.Error(codes.NotFound, "report not found")
	assert.Equal(t, http.StatusNotFound, e.do("GET", "/reports/r9/status", "head-token", "").Code)
}
