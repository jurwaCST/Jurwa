package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
)

// ---- dashboard routes (SCRUM-129) ----

// fakeAnalytics records the period the gateway asks for.
type fakeAnalytics struct {
	period *analyticsv1.Period
	called int
	err    error
}

func (f *fakeAnalytics) GetSummary(_ context.Context, p *analyticsv1.Period) (*analyticsv1.GetSummaryResponse, error) {
	f.period, f.called = p, f.called+1
	if f.err != nil {
		return nil, f.err
	}
	return &analyticsv1.GetSummaryResponse{Counts: &analyticsv1.StatusCounts{Total: 3, Open: 1, Resolved: 2}, AverageResolutionHours: 5.5}, nil
}

func (f *fakeAnalytics) GetCategoryBreakdown(_ context.Context, p *analyticsv1.Period) (*analyticsv1.GetCategoryBreakdownResponse, error) {
	f.period, f.called = p, f.called+1
	if f.err != nil {
		return nil, f.err
	}
	return &analyticsv1.GetCategoryBreakdownResponse{Categories: []*analyticsv1.CategoryCounts{
		{Category: commonv1.Category_HOSTEL, Counts: &analyticsv1.StatusCounts{Total: 2, Resolved: 2}},
	}}, nil
}

func withPresident(t *testing.T) {
	t.Helper()
	users["president-token"] = &identityv1.User{Id: "u-president", Roles: []commonv1.Role{commonv1.Role_PRESIDENT}}
	t.Cleanup(func() { delete(users, "president-token") })
}

func TestDashboardRoutesNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/dashboard/summary", "", "").Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/dashboard/categories", "", "").Code)
	assert.Zero(t, e.stats.called)
}

func TestDashboardRoles(t *testing.T) {
	withPresident(t)
	for _, path := range []string{"/dashboard/summary", "/dashboard/categories"} {
		for token, want := range map[string]int{
			"admin-token":     http.StatusOK,
			"president-token": http.StatusOK,
			"student-token":   http.StatusForbidden,
			"head-token":      http.StatusForbidden,
			"staff-token":     http.StatusForbidden,
		} {
			e := newEnv()
			assert.Equal(t, want, e.do("GET", path, token, "").Code, path+" "+token)
			if want == http.StatusForbidden {
				assert.Zero(t, e.stats.called, path+" "+token+" never reaches analytics")
			}
		}
	}
}

func TestDashboardSummaryJSON(t *testing.T) {
	rec := newEnv().do("GET", "/dashboard/summary", "admin-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	body := decode(t, rec)
	assert.Equal(t, 5.5, body["averageResolutionHours"])
	counts := body["counts"].(map[string]any)
	assert.Equal(t, "3", counts["total"], "protojson sends int64 as a string")
	assert.Equal(t, "0", counts["escalated"], "zero counts are still sent")
}

func TestDashboardCategoriesJSON(t *testing.T) {
	rec := newEnv().do("GET", "/dashboard/categories", "admin-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	cats := decode(t, rec)["categories"].([]any)
	require.Len(t, cats, 1)
	assert.Equal(t, "HOSTEL", cats[0].(map[string]any)["category"])
}

func TestDashboardPeriod(t *testing.T) {
	e := newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/dashboard/summary", "admin-token", "").Code)
	assert.Nil(t, e.stats.period, "no from or to: all time")

	e = newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/dashboard/categories?from=2026-11-01T00:00:00Z&to=2026-11-08T00:00:00%2B06:00", "admin-token", "").Code)
	require.NotNil(t, e.stats.period)
	assert.True(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Equal(e.stats.period.GetFrom().AsTime()))
	assert.True(t, time.Date(2026, 11, 7, 18, 0, 0, 0, time.UTC).Equal(e.stats.period.GetTo().AsTime()))

	e = newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/dashboard/summary?from=2026-11-01T00:00:00Z", "admin-token", "").Code)
	assert.NotNil(t, e.stats.period.GetFrom())
	assert.Nil(t, e.stats.period.GetTo(), "only from: no upper limit")
}

func TestDashboardBadPeriod(t *testing.T) {
	for name, query := range map[string]string{
		"not a time":    "?from=yesterday",
		"date only":     "?to=2026-11-08",
		"from after to": "?from=2026-11-08T00:00:00Z&to=2026-11-01T00:00:00Z",
	} {
		e := newEnv()
		rec := e.do("GET", "/dashboard/summary"+query, "admin-token", "")
		assert.Equal(t, http.StatusBadRequest, rec.Code, name)
		assert.Zero(t, e.stats.called, name+" never reaches analytics")
	}
}

func TestDashboardAnalyticsDown(t *testing.T) {
	e := newEnv()
	e.stats.err = status.Error(codes.Unavailable, "dial tcp analytics:9020")
	rec := e.do("GET", "/dashboard/summary", "admin-token", "")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotContains(t, rec.Body.String(), "9020")
}
