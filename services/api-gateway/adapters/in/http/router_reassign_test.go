package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
)

// ---- reassign routes ----

const reassignPath = "/reports/6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f/reassign"

// fakeReassignment records what the gateway sends to reassignment.
type fakeReassignment struct {
	role string
	got  *reassignmentv1.ReassignReportRequest
	err  error
}

func (f *fakeReassignment) ReassignReport(_ context.Context, role string, in *reassignmentv1.ReassignReportRequest) (*reassignmentv1.ReassignReportResponse, error) {
	f.role, f.got = role, in
	if f.err != nil {
		return nil, f.err
	}
	return &reassignmentv1.ReassignReportResponse{Assignment: &reassignmentv1.Assignment{
		Id: "a2", ReportId: in.GetReportId(), AssigneeUserId: in.GetToUserId(), Kind: reassignmentv1.AssignmentKind_REASSIGNED,
		AssignedBy: &in.ActorId, Reason: in.GetReason(), Current: true,
	}}, nil
}

// fakeDirectory answers like department-directory with the demo holders.
type fakeDirectory struct {
	calls int
	err   error
}

func (f *fakeDirectory) ListHeldPositions(context.Context) (*departmentdirectoryv1.ListPositionsResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &departmentdirectoryv1.ListPositionsResponse{Positions: []*departmentdirectoryv1.Position{
		{Code: "DSA", Title: "Dean of Student Affairs", Kind: "CATEGORY_HEAD", HolderUserId: "u-head", Active: true},
		{Code: "SSO_MALE", Title: "Student Service Officer (male)", Kind: "STAFF", HolderUserId: "u-staff", Active: true},
	}}, nil
}

func TestReassignRoutesNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("POST", reassignPath, "", `{"toUserId":"u-staff","reason":"x"}`).Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/positions", "", "").Code)
	assert.Nil(t, e.reassign.got)
	assert.Zero(t, e.directory.calls)
}

func TestReassignTakesTheReportFromThePathAndTheActorFromTheJWT(t *testing.T) {
	e := newEnv()
	body := `{"reportId":"r-other","actorId":"u-someone","toUserId":"u-staff","reason":"Plumbing is a staff job"}`
	rec := e.do("POST", reassignPath, "head-token", body)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	got := e.reassign.got
	require.NotNil(t, got)
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", got.GetReportId(), "the report comes from the path, never the body")
	assert.Equal(t, "u-head", got.GetActorId(), "the actor comes from the JWT, never the body")
	assert.Equal(t, "u-staff", got.GetToUserId())
	assert.Equal(t, "Plumbing is a staff job", got.GetReason())
	assert.Equal(t, "CATEGORY_HEAD", e.reassign.role, "sent to reassignment as x-user-role")

	out := decode(t, rec)["assignment"].(map[string]any)
	assert.Equal(t, "u-staff", out["assigneeUserId"])
	assert.Equal(t, "REASSIGNED", out["kind"])
	assert.Equal(t, true, out["current"])
}

func TestReassignRoles(t *testing.T) {
	for token, want := range map[string]int{
		"head-token":    http.StatusOK,
		"staff-token":   http.StatusOK,
		"student-token": http.StatusForbidden,
		"admin-token":   http.StatusForbidden,
	} {
		t.Run(token, func(t *testing.T) {
			e := newEnv()
			assert.Equal(t, want, e.do("POST", reassignPath, token, `{"toUserId":"u-staff","reason":"x"}`).Code)
			assert.Equal(t, want, e.do("GET", "/positions", token, "").Code)
		})
	}
}

func TestReassignErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "missing field", err: status.Error(codes.InvalidArgument, "reason is required"), want: http.StatusBadRequest},
		{name: "not the assignee", err: status.Error(codes.PermissionDenied, "only a Category Head or the current assignee can reassign"), want: http.StatusForbidden},
		{name: "never routed", err: status.Error(codes.NotFound, "report has no assignment"), want: http.StatusNotFound},
		{name: "reassigned meanwhile", err: status.Error(codes.Aborted, "assignment changed"), want: http.StatusConflict},
		{name: "service down", err: status.Error(codes.Unavailable, "dial reassignment:9014"), want: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			e.reassign.err = tt.err
			assert.Equal(t, tt.want, e.do("POST", reassignPath, "staff-token", `{"toUserId":"u-head","reason":"x"}`).Code)
		})
	}
}

func TestReassignRejectsBadJSON(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("POST", reassignPath, "head-token", `{"toUserId":`).Code)
	assert.Nil(t, e.reassign.got, "reassignment is not called")
}

func TestPositionsListsTheHeldPositions(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/positions", "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	list := decode(t, rec)["positions"].([]any)
	require.Len(t, list, 2)
	staff := list[1].(map[string]any)
	assert.Equal(t, "SSO_MALE", staff["code"])
	assert.Equal(t, "STAFF", staff["kind"])
	assert.Equal(t, "u-staff", staff["holderUserId"], "the PWA sends this as toUserId")
	assert.Equal(t, "", list[0].(map[string]any)["departmentCode"], "empty fields are still sent")

	e.directory.err = status.Error(codes.Unavailable, "dial department-directory:9008")
	assert.Equal(t, http.StatusServiceUnavailable, e.do("GET", "/positions", "head-token", "").Code)
}
