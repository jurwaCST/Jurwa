package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/api-gateway/application"
	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
)

// ---- fakes for the ports ----

// fakeIdentity knows two tokens: "student-token" and "head-token".
type fakeIdentity struct{ loginErr error }

var users = map[string]*identityv1.User{
	"student-token": {Id: "u-student", Email: "s@rub.edu.bt", Roles: []commonv1.Role{commonv1.Role_STUDENT}},
	"head-token":    {Id: "u-head", Email: "h@rub.edu.bt", Roles: []commonv1.Role{commonv1.Role_CATEGORY_HEAD}},
}

func (f *fakeIdentity) Register(_ context.Context, in *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	return &identityv1.RegisterResponse{User: &identityv1.User{Id: "u-new", Email: in.GetEmail()}}, nil
}

func (f *fakeIdentity) Login(_ context.Context, in *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	return &identityv1.LoginResponse{AccessToken: "student-token", User: users["student-token"]}, nil
}

func (f *fakeIdentity) ValidateToken(_ context.Context, token string) (*identityv1.User, error) {
	if u, ok := users[token]; ok {
		return u, nil
	}
	return nil, status.Error(codes.Unauthenticated, "invalid token")
}

type fakeCategories struct{ calls int }

func (f *fakeCategories) ListCategories(context.Context) (*categorytaxonomyv1.ListCategoriesResponse, error) {
	f.calls++
	return &categorytaxonomyv1.ListCategoriesResponse{Categories: []*categorytaxonomyv1.CategoryInfo{
		{Category: commonv1.Category_HOSTEL, Name: "Hostel"},
		{Category: commonv1.Category_ACADEMIC_BLOCK, Name: "Academic Block"},
		{Category: commonv1.Category_OTHERS, Name: "Others"},
	}}, nil
}

type fakeLocations struct{ got commonv1.Category }

func (f *fakeLocations) ListLocations(_ context.Context, c commonv1.Category) (*locationtaxonomyv1.ListLocationsResponse, error) {
	f.got = c
	return &locationtaxonomyv1.ListLocationsResponse{Locations: []*locationtaxonomyv1.LocationEntry{
		{Id: "l1", Category: c, Location: &commonv1.Location{Building: "Hostel B"}},
	}}, nil
}

type fakeReports struct {
	submitted *reportsubmissionv1.SubmitReportRequest
	getErr    error
}

func (f *fakeReports) SubmitReport(_ context.Context, in *reportsubmissionv1.SubmitReportRequest) (*reportsubmissionv1.SubmitReportResponse, error) {
	f.submitted = in
	return &reportsubmissionv1.SubmitReportResponse{Report: &commonv1.Report{Id: "r1", ReporterId: in.GetReporterId(), Category: in.GetCategory()}}, nil
}

func (f *fakeReports) GetReport(_ context.Context, id string) (*commonv1.Report, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &commonv1.Report{Id: id, ReporterId: "u-student", Category: commonv1.Category_HOSTEL}, nil
}

// fakeAuthz is the role table for the routes under test: only students submit,
// any signed-in role lists and reads (ownership is checked by the router).
type fakeAuthz struct{}

func (fakeAuthz) CheckPermission(_ context.Context, role, action, _ string) (bool, error) {
	if action == domain.ActionReportSubmit {
		return role == domain.RoleStudent, nil
	}
	return true, nil
}

// ---- helpers ----

type env struct {
	h       http.Handler
	cats    *fakeCategories
	locs    *fakeLocations
	reports *fakeReports
	ident   *fakeIdentity
	clock   *time.Time
}

func newEnv() *env {
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	e := &env{cats: &fakeCategories{}, locs: &fakeLocations{}, reports: &fakeReports{}, ident: &fakeIdentity{}, clock: &now}
	e.h = NewRouter(Deps{
		Auth:       application.NewAuth(e.ident, fakeAuthz{}),
		Identity:   e.ident,
		Categories: e.cats,
		Locations:  e.locs,
		Reports:    e.reports,
		Limiter:    domain.NewLimiter(20, func() time.Time { return *e.clock }),
	})
	return e
}

func (e *env) do(method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "10.0.0.1:5555"
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out), rec.Body.String())
	return out
}

// ---- tests ----

func TestHealthIsPublic(t *testing.T) {
	rec := newEnv().do("GET", "/health", "", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestProtectedRouteNeedsToken(t *testing.T) {
	e := newEnv()
	for _, r := range []struct{ method, path string }{
		{"GET", "/categories"}, {"GET", "/categories/HOSTEL/locations"}, {"POST", "/reports"}, {"GET", "/reports/r1"}, {"GET", "/me"},
	} {
		rec := e.do(r.method, r.path, "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s %s", r.method, r.path)
	}
	assert.Equal(t, 0, e.cats.calls, "the service is never reached without a token")
}

func TestInvalidTokenIs401(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/categories", "forged-token", "").Code)
	req := httptest.NewRequest("GET", "/categories", nil)
	req.Header.Set("Authorization", "Basic abc")
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestValidTokenReachesService(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/categories", "student-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	cats := decode(t, rec)["categories"].([]any)
	assert.Len(t, cats, 3)
	assert.Equal(t, "HOSTEL", cats[0].(map[string]any)["category"])
	assert.Equal(t, 1, e.cats.calls)
}

func TestMe(t *testing.T) {
	rec := newEnv().do("GET", "/me", "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "u-head", decode(t, rec)["id"])
}

func TestLocations(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/categories/HOSTEL/locations", "student-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, commonv1.Category_HOSTEL, e.locs.got)

	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/categories/LIBRARY/locations", "student-token", "").Code)
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/categories/CATEGORY_UNSPECIFIED/locations", "student-token", "").Code)
}

func TestLoginAndRegisterArePublic(t *testing.T) {
	e := newEnv()
	rec := e.do("POST", "/auth/login", "", `{"email":"s@rub.edu.bt","password":"secret123"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "student-token", decode(t, rec)["accessToken"])

	rec = e.do("POST", "/auth/register", "", `{"email":"n@rub.edu.bt","fullName":"New","password":"secret123"}`)
	require.Equal(t, http.StatusCreated, rec.Code)

	e.ident.loginErr = status.Error(codes.Unauthenticated, "wrong email or password")
	assert.Equal(t, http.StatusUnauthorized, e.do("POST", "/auth/login", "", `{"email":"s@rub.edu.bt","password":"bad"}`).Code)
}

func TestSubmitReport(t *testing.T) {
	e := newEnv()
	body := `{"reporterId":"someone-else","category":"HOSTEL","location":{"building":"Hostel B"},"description":"Tap leaking"}`
	rec := e.do("POST", "/reports", "student-token", body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "u-student", e.reports.submitted.GetReporterId(), "reporter id comes from the JWT, never the body")
	assert.Equal(t, "Tap leaking", e.reports.submitted.GetDescription())
}

func TestSubmitReportBadJSON(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("POST", "/reports", "student-token", `{"category":`).Code)
	assert.Equal(t, http.StatusBadRequest, e.do("POST", "/reports", "student-token", `{"category":"NOPE"}`).Code)
	assert.Equal(t, http.StatusBadRequest, e.do("POST", "/auth/login", "", `not json`).Code)
	assert.Nil(t, e.reports.submitted)
}

func TestSubmitReportOnlyStudents(t *testing.T) {
	e := newEnv()
	rec := e.do("POST", "/reports", "head-token", `{"category":"HOSTEL","location":{"building":"B"},"description":"x"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Nil(t, e.reports.submitted)
}

func TestGetReportOwnership(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusOK, e.do("GET", "/reports/r1", "student-token", "").Code, "own report")
	assert.Equal(t, http.StatusOK, e.do("GET", "/reports/r1", "head-token", "").Code, "category head reads any")

	users["other-student"] = &identityv1.User{Id: "u-other", Roles: []commonv1.Role{commonv1.Role_STUDENT}}
	defer delete(users, "other-student")
	assert.Equal(t, http.StatusForbidden, e.do("GET", "/reports/r1", "other-student", "").Code, "another student's report")
}

func TestGRPCErrorsMapToHTTP(t *testing.T) {
	tests := []struct {
		code codes.Code
		want int
	}{
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.NotFound, http.StatusNotFound},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.Aborted, http.StatusConflict},
		{codes.FailedPrecondition, http.StatusUnprocessableEntity},
		{codes.ResourceExhausted, http.StatusTooManyRequests},
		{codes.Unavailable, http.StatusServiceUnavailable},
		{codes.DeadlineExceeded, http.StatusGatewayTimeout},
		{codes.Internal, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			e := newEnv()
			e.reports.getErr = status.Error(tt.code, "from service")
			rec := e.do("GET", "/reports/r1", "head-token", "")
			assert.Equal(t, tt.want, rec.Code)
			assert.Contains(t, decode(t, rec), "error")
		})
	}
}

func TestInternalErrorsAreHidden(t *testing.T) {
	e := newEnv()
	e.reports.getErr = status.Error(codes.Internal, "pq: password authentication failed")
	rec := e.do("GET", "/reports/r1", "head-token", "")
	body, _ := io.ReadAll(rec.Body)
	assert.NotContains(t, string(body), "password")
}

func TestRateLimit(t *testing.T) {
	e := newEnv()
	for i := 1; i <= 20; i++ {
		require.Equal(t, http.StatusOK, e.do("GET", "/health", "", "").Code, "request %d", i)
	}
	rec := e.do("GET", "/health", "", "")
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "21st request in one second from one IP")
	assert.Equal(t, "1", rec.Header().Get("Retry-After"))

	*e.clock = e.clock.Add(time.Second)
	assert.Equal(t, http.StatusOK, e.do("GET", "/health", "", "").Code, "allowed again after a second")
}

func TestUnknownRouteIs404(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, newEnv().do("GET", "/nope", "student-token", "").Code)
}

func TestUnavailableHidesAddress(t *testing.T) {
	e := newEnv()
	e.reports.getErr = status.Error(codes.Unavailable, "dial tcp 127.0.0.1:9001: connection refused")
	rec := e.do("GET", "/reports/r1", "head-token", "")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotContains(t, rec.Body.String(), "127.0.0.1")
}

type downAuthz struct{}

func (downAuthz) CheckPermission(context.Context, string, string, string) (bool, error) {
	return false, status.Error(codes.Unavailable, "authorization down")
}

// Fail closed: when authorization cannot answer, the request is refused.
func TestAuthorizationDownRefusesRequest(t *testing.T) {
	e := newEnv()
	e.h = NewRouter(Deps{
		Auth:       application.NewAuth(e.ident, downAuthz{}),
		Identity:   e.ident,
		Categories: e.cats,
		Locations:  e.locs,
		Reports:    e.reports,
		Limiter:    domain.NewLimiter(20, func() time.Time { return *e.clock }),
	})
	rec := e.do("POST", "/reports", "student-token", `{"category":"HOSTEL","location":{"building":"B"},"description":"x"}`)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Nil(t, e.reports.submitted, "the report service is never reached")
	assert.Equal(t, 0, e.cats.calls)
}
