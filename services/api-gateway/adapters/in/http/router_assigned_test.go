package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
)

// ---- GET /me/assigned-reports ----

func (f *fakeHistory) ListReportsByAssignee(_ context.Context, assigneeID string, pageSize int32, pageToken string) (*reporthistoryv1.ListReportsByAssigneeResponse, error) {
	f.assignee, f.pageSize, f.pageToken = assigneeID, pageSize, pageToken
	if f.empty {
		return &reporthistoryv1.ListReportsByAssigneeResponse{}, nil
	}
	id := assigneeID
	return &reporthistoryv1.ListReportsByAssigneeResponse{
		Reports:       []*reporthistoryv1.ReportSummary{{ReportId: "r1", ReporterId: "u-student", CurrentAssigneeId: &id}},
		NextPageToken: "next",
	}, nil
}

func TestAssignedReportsNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/me/assigned-reports", "", "").Code)
	assert.Empty(t, e.history.assignee)
}

func TestAssignedReportsUseTheIDFromTheJWT(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/me/assigned-reports?assigneeId=u-other&pageSize=5&pageToken=abc", "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "u-head", e.history.assignee, "the assignee is the signed-in user, never a query value")
	assert.Equal(t, int32(5), e.history.pageSize)
	assert.Equal(t, "abc", e.history.pageToken)
	body := decode(t, rec)
	assert.Equal(t, "next", body["nextPageToken"])
	assert.Equal(t, "u-head", body["reports"].([]any)[0].(map[string]any)["currentAssigneeId"])
}

func TestAssignedReportsRoles(t *testing.T) {
	for token, want := range map[string]int{
		"head-token":    http.StatusOK,
		"staff-token":   http.StatusOK,
		"student-token": http.StatusForbidden,
		"admin-token":   http.StatusForbidden,
	} {
		e := newEnv()
		assert.Equal(t, want, e.do("GET", "/me/assigned-reports", token, "").Code, token)
		if want == http.StatusForbidden {
			assert.Empty(t, e.history.assignee, token+" never reaches report-history")
		}
	}
}

func TestAssignedReportsPaging(t *testing.T) {
	e := newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/me/assigned-reports", "staff-token", "").Code)
	assert.Equal(t, int32(0), e.history.pageSize, "0 means the service default (20)")

	e = newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/me/assigned-reports?pageSize=ten", "staff-token", "").Code)
	assert.Empty(t, e.history.assignee)
}

func TestAssignedReportsEmptyList(t *testing.T) {
	e := newEnv()
	e.history.empty = true
	rec := e.do("GET", "/me/assigned-reports", "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"reports":[],"nextPageToken":""}`, rec.Body.String())
}
