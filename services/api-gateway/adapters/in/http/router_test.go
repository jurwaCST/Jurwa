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
	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
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
	"staff-token":   {Id: "u-staff", Email: "st@rub.edu.bt", Roles: []commonv1.Role{commonv1.Role_STAFF}},
	"admin-token":   {Id: "u-admin", Email: "a@rub.edu.bt", Roles: []commonv1.Role{commonv1.Role_SYSTEM_ADMIN}},
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
	switch action {
	case domain.ActionReportSubmit:
		return role == domain.RoleStudent, nil
	case domain.ActionReportNoteAdd: // same rule as the authorization policy
		return role == domain.RoleCategoryHead || role == domain.RoleStaff, nil
	case domain.ActionReportNoteRead:
		return role == domain.RoleCategoryHead || role == domain.RoleStaff || role == domain.RoleSystemAdmin, nil
	case domain.ActionReportListOwn:
		return role == domain.RoleStudent, nil
	case domain.ActionReportTimelineRead, domain.ActionPhotoRead:
		return role != "PRESIDENT", nil
	case domain.ActionPhotoUpload:
		return role == domain.RoleStudent, nil
	case domain.ActionReportStatusChange:
		return role == domain.RoleCategoryHead || role == domain.RoleStaff, nil
	case domain.ActionReportStatusRead:
		return role != "PRESIDENT", nil
	case domain.ActionDashboardRead: // same rule as the authorization policy
		return role == domain.RoleSystemAdmin || role == "PRESIDENT", nil
	}
	return true, nil
}

// fakeNotes records what the gateway sends to internal-notes.
type fakeNotes struct {
	role    string
	added   *internalnotesv1.AddNoteRequest
	listFor string
}

func (f *fakeNotes) AddNote(_ context.Context, role string, in *internalnotesv1.AddNoteRequest) (*internalnotesv1.AddNoteResponse, error) {
	f.role, f.added = role, in
	if strings.TrimSpace(in.GetBody()) == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid note: body is required")
	}
	return &internalnotesv1.AddNoteResponse{Note: &internalnotesv1.Note{Id: "n1", ReportId: in.GetReportId(), AuthorId: in.GetAuthorId(), Body: in.GetBody()}}, nil
}

func (f *fakeNotes) ListNotes(_ context.Context, role, reportID string) (*internalnotesv1.ListNotesResponse, error) {
	f.role, f.listFor = role, reportID
	return &internalnotesv1.ListNotesResponse{Notes: []*internalnotesv1.Note{{Id: "n1", ReportId: reportID, Body: "Plumber booked"}}}, nil
}

// fakeHistory records what the gateway sends to report-history.
type fakeHistory struct {
	reporter  string
	pageSize  int32
	pageToken string
	timeline  string
	err       error
	empty     bool
}

func (f *fakeHistory) ListReportsByReporter(_ context.Context, reporterID string, pageSize int32, pageToken string) (*reporthistoryv1.ListReportsByReporterResponse, error) {
	f.reporter, f.pageSize, f.pageToken = reporterID, pageSize, pageToken
	if f.empty {
		return &reporthistoryv1.ListReportsByReporterResponse{}, nil
	}
	return &reporthistoryv1.ListReportsByReporterResponse{Reports: []*reporthistoryv1.ReportSummary{{ReportId: "r1", ReporterId: reporterID}}, NextPageToken: "next"}, nil
}

func (f *fakeHistory) GetReportTimeline(_ context.Context, reportID string) (*reporthistoryv1.GetReportTimelineResponse, error) {
	f.timeline = reportID
	if f.err != nil {
		return nil, f.err
	}
	return &reporthistoryv1.GetReportTimelineResponse{Entries: []*reporthistoryv1.TimelineEntry{{Id: "e1", EventType: "report.submitted", Summary: "Report submitted"}}}, nil
}

// ---- helpers ----

type env struct {
	h       http.Handler
	cats    *fakeCategories
	locs    *fakeLocations
	reports *fakeReports
	ident   *fakeIdentity
	notes   *fakeNotes
	history *fakeHistory
	photos  *fakePhotos
	status  *fakeStatus
	stats   *fakeAnalytics
	clock   *time.Time
}

func newEnv() *env {
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	e := &env{cats: &fakeCategories{}, locs: &fakeLocations{}, reports: &fakeReports{}, ident: &fakeIdentity{}, notes: &fakeNotes{}, history: &fakeHistory{}, photos: &fakePhotos{}, status: &fakeStatus{}, stats: &fakeAnalytics{}, clock: &now}
	e.h = NewRouter(Deps{
		Auth:       application.NewAuth(e.ident, fakeAuthz{}),
		Identity:   e.ident,
		Categories: e.cats,
		Locations:  e.locs,
		Reports:    e.reports,
		Notes:      e.notes,
		History:    e.history,
		Photos:     e.photos,
		Status:     e.status,
		Analytics:  e.stats,
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

// ---- internal-notes routes (SCRUM-117) ----

const notesPath = "/reports/6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f/notes"

func TestNotesNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("POST", notesPath, "", `{"body":"x"}`).Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", notesPath, "", "").Code)
	assert.Nil(t, e.notes.added)
}

func TestNotesRoles(t *testing.T) {
	tests := []struct {
		token    string
		wantAdd  int
		wantList int
	}{
		{token: "head-token", wantAdd: http.StatusCreated, wantList: http.StatusOK},
		{token: "staff-token", wantAdd: http.StatusCreated, wantList: http.StatusOK},
		{token: "admin-token", wantAdd: http.StatusForbidden, wantList: http.StatusOK},
		{token: "student-token", wantAdd: http.StatusForbidden, wantList: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			e := newEnv()
			assert.Equal(t, tt.wantAdd, e.do("POST", notesPath, tt.token, `{"body":"Plumber booked for Monday"}`).Code)
			assert.Equal(t, tt.wantList, e.do("GET", notesPath, tt.token, "").Code)
		})
	}
}

func TestAddNoteSetsAuthorAndRoleFromTheJWT(t *testing.T) {
	e := newEnv()
	rec := e.do("POST", notesPath, "staff-token", `{"body":"Plumber booked","authorId":"someone-else","reportId":"other"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.Equal(t, "u-staff", e.notes.added.GetAuthorId(), "author id comes from the JWT, never the body")
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", e.notes.added.GetReportId(), "report id comes from the path")
	assert.Equal(t, "STAFF", e.notes.role, "sent to internal-notes as x-user-role")
	assert.Equal(t, "Plumber booked", decode(t, rec)["note"].(map[string]any)["body"])
}

func TestListNotesSendsTheRole(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", notesPath, "admin-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "SYSTEM_ADMIN", e.notes.role)
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", e.notes.listFor)
	assert.Len(t, decode(t, rec)["notes"], 1)
}

func TestAddNoteBadInput(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("POST", notesPath, "staff-token", `{"body":`).Code, "bad JSON")
	assert.Nil(t, e.notes.added, "bad JSON never reaches internal-notes")
	assert.Equal(t, http.StatusBadRequest, e.do("POST", notesPath, "staff-token", `{"body":"  "}`).Code, "empty note (InvalidArgument from internal-notes)")
}

// A user with two roles: the role sent is the one authorization allowed.
func TestNotesSendTheRoleThatWasAllowed(t *testing.T) {
	users["student-staff-token"] = &identityv1.User{Id: "u-two", Roles: []commonv1.Role{commonv1.Role_STUDENT, commonv1.Role_STAFF}}
	defer delete(users, "student-staff-token")
	e := newEnv()
	require.Equal(t, http.StatusCreated, e.do("POST", notesPath, "student-staff-token", `{"body":"x"}`).Code)
	assert.Equal(t, "STAFF", e.notes.role)
}

// ---- report-history routes (SCRUM-119) ----

func TestHistoryRoutesNeedAToken(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/me/reports", "", "").Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/reports/r1/timeline", "", "").Code)
	assert.Empty(t, e.history.reporter)
}

func TestMyReportsUsesTheIDFromTheJWT(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/me/reports?reporterId=u-other&pageSize=5&pageToken=abc", "student-token", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "u-student", e.history.reporter, "the reporter id comes from the JWT, never the request")
	assert.Equal(t, int32(5), e.history.pageSize)
	assert.Equal(t, "abc", e.history.pageToken)
	assert.Equal(t, "next", decode(t, rec)["nextPageToken"])
}

func TestMyReportsDefaultsAndBadPaging(t *testing.T) {
	e := newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/me/reports", "student-token", "").Code)
	assert.Equal(t, int32(0), e.history.pageSize, "0 means the service default (20)")

	e = newEnv()
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/me/reports?pageSize=abc", "student-token", "").Code)
	assert.Empty(t, e.history.reporter, "bad paging never reaches report-history")
}

func TestMyReportsIsForStudentsOnly(t *testing.T) {
	e := newEnv()
	assert.Equal(t, http.StatusForbidden, e.do("GET", "/me/reports", "head-token", "").Code)
	assert.Empty(t, e.history.reporter)
}

func TestTimelineOwnership(t *testing.T) {
	// fakeReports says report r1 belongs to u-student.
	e := newEnv()
	require.Equal(t, http.StatusOK, e.do("GET", "/reports/r1/timeline", "student-token", "").Code, "own report")
	assert.Equal(t, "r1", e.history.timeline)

	e = newEnv()
	assert.Equal(t, http.StatusOK, e.do("GET", "/reports/r1/timeline", "head-token", "").Code, "Category Head reads any report")
	e = newEnv()
	assert.Equal(t, http.StatusOK, e.do("GET", "/reports/r1/timeline", "staff-token", "").Code)
	e = newEnv()
	assert.Equal(t, http.StatusOK, e.do("GET", "/reports/r1/timeline", "admin-token", "").Code)

	users["other-student"] = &identityv1.User{Id: "u-other", Roles: []commonv1.Role{commonv1.Role_STUDENT}}
	defer delete(users, "other-student")
	e = newEnv()
	assert.Equal(t, http.StatusForbidden, e.do("GET", "/reports/r1/timeline", "other-student", "").Code, "another student's report")
	assert.Empty(t, e.history.timeline, "the timeline is never fetched")
}

func TestTimelineUnknownReport(t *testing.T) {
	e := newEnv()
	e.reports.getErr = status.Error(codes.NotFound, "report not found")
	assert.Equal(t, http.StatusNotFound, e.do("GET", "/reports/r9/timeline", "student-token", "").Code, "student: owner lookup says not found")

	e = newEnv()
	e.history.err = status.Error(codes.NotFound, "report not found")
	assert.Equal(t, http.StatusNotFound, e.do("GET", "/reports/r9/timeline", "head-token", "").Code, "staff roles: report-history says not found")
}

// A student with no reports gets an empty list, not a missing field, so the
// PWA can always loop over "reports".
func TestEmptyListsAreSentAsEmptyArrays(t *testing.T) {
	e := newEnv()
	e.history.empty = true
	rec := e.do("GET", "/me/reports", "student-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"reports":[],"nextPageToken":""}`, rec.Body.String())
}

// Plain fields are always sent; an optional field that is not set (photoId)
// stays out instead of being sent as null.
func TestReportJSONFieldsAreAlwaysPresent(t *testing.T) {
	rec := newEnv().do("GET", "/reports/r1", "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	body := decode(t, rec)
	assert.Contains(t, body, "description", "empty text is sent as \"\"")
	assert.NotContains(t, body, "photoId", "a report without a photo has no photoId")
}
