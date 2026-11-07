// Package http is the gateway's HTTP adapter: the routes the PWA calls (see
// docs/design/api-routes.md), the rate limit, the JWT check and the mapping of
// gRPC errors to HTTP status codes.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	"github.com/jurwaCST/Jurwa/services/api-gateway/application"
	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
	"github.com/jurwaCST/Jurwa/services/api-gateway/ports"
)

// maxBodyBytes bounds a JSON body.
const maxBodyBytes = 64 << 10

// maxPhotoBodyBytes bounds POST /photos only: a 5 MB photo plus the multipart
// headers. photo-media gives the exact error for a photo between 5 and 6 MB;
// anything bigger is refused here, with the same message (errPhotoTooLarge).
const maxPhotoBodyBytes = 6 << 20

// errPhotoTooLarge is photo-media's own message for a photo over 5 MB, so the
// PWA gets one message whichever side refuses the upload.
const errPhotoTooLarge = "invalid photo: photo is larger than 5 MB"

// Deps are the ports and helpers the routes use.
type Deps struct {
	Auth       *application.Auth
	Identity   ports.Identity
	Categories ports.Categories
	Locations  ports.Locations
	Reports    ports.Reports
	Notes      ports.Notes
	History    ports.History
	Photos     ports.Photos
	Status     ports.Status
	Analytics  ports.Analytics
	Limiter    *domain.Limiter
}

type router struct{ Deps }

// NewRouter returns the gateway's HTTP handler.
func NewRouter(d Deps) http.Handler {
	r := router{d}
	mux := http.NewServeMux()

	// Public routes.
	mux.HandleFunc("GET /health", r.health)
	mux.HandleFunc("POST /auth/register", r.register)
	mux.HandleFunc("POST /auth/login", r.login)

	// Signed-in routes: the JWT is checked, then the action is authorized.
	mux.Handle("GET /me", r.protected(domain.ActionMeRead, r.me))
	mux.Handle("GET /categories", r.protected(domain.ActionCategoryList, r.listCategories))
	mux.Handle("GET /categories/{category}/locations", r.protected(domain.ActionLocationList, r.listLocations))
	mux.Handle("POST /reports", r.protected(domain.ActionReportSubmit, r.submitReport))
	mux.Handle("GET /reports/{reportId}", r.protected(domain.ActionReportRead, r.getReport))
	mux.Handle("POST /reports/{reportId}/notes", r.protected(domain.ActionReportNoteAdd, r.addNote))
	mux.Handle("GET /reports/{reportId}/notes", r.protected(domain.ActionReportNoteRead, r.listNotes))
	mux.Handle("GET /me/reports", r.protected(domain.ActionReportListOwn, r.myReports))
	mux.Handle("POST /reports/{reportId}/status", r.protected(domain.ActionReportStatusChange, r.changeStatus))
	mux.Handle("GET /reports/{reportId}/status", r.protected(domain.ActionReportStatusRead, r.getStatus))
	mux.Handle("GET /dashboard/summary", r.protected(domain.ActionDashboardRead, r.dashboardSummary))
	mux.Handle("GET /dashboard/categories", r.protected(domain.ActionDashboardRead, r.dashboardCategories))
	mux.Handle("POST /photos", r.protected(domain.ActionPhotoUpload, r.uploadPhoto))
	mux.Handle("GET /photos/{photoId}", r.protected(domain.ActionPhotoRead, r.getPhoto))
	mux.Handle("GET /reports/{reportId}/timeline", r.protected(domain.ActionReportTimelineRead, r.timeline))

	return r.rateLimit(mux)
}

// ---- middleware ----

type (
	principalKey struct{}
	roleKey      struct{} // the role that authorization allowed for this route
)

// protected checks the JWT with identity and the role with the authorizer
// before calling next. The principal is put in the request context.
func (r router) protected(action string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p, err := r.Auth.Authenticate(req.Context(), req.Header.Get("Authorization"))
		if err != nil {
			writeError(w, err)
			return
		}
		role, err := r.Auth.AuthorizeAs(req.Context(), p, action, req.URL.Path)
		if err != nil {
			writeError(w, err)
			return
		}
		ctx := context.WithValue(req.Context(), principalKey{}, p)
		next(w, req.WithContext(context.WithValue(ctx, roleKey{}, role)))
	})
}

// actingRole is the caller's role that was allowed for the route.
func actingRole(req *http.Request) string {
	role, _ := req.Context().Value(roleKey{}).(string)
	return role
}

func principal(req *http.Request) domain.Principal {
	p, _ := req.Context().Value(principalKey{}).(domain.Principal)
	return p
}

// rateLimit refuses more than the limit per second from one client IP with 429.
// It uses the TCP peer address, not X-Forwarded-For, which a client can forge.
func (r router) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			ip = req.RemoteAddr
		}
		if !r.Limiter.Allow(ip) {
			w.Header().Set("Retry-After", "1")
			writeJSONError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, req)
	})
}

// ---- handlers ----

func (r router) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"status":"ok"}`)
}

func (r router) register(w http.ResponseWriter, req *http.Request) {
	in := &identityv1.RegisterRequest{}
	if !readJSON(w, req, in) {
		return
	}
	resp, err := r.Identity.Register(req.Context(), in)
	reply(w, http.StatusCreated, resp, err)
}

func (r router) login(w http.ResponseWriter, req *http.Request) {
	in := &identityv1.LoginRequest{}
	if !readJSON(w, req, in) {
		return
	}
	resp, err := r.Identity.Login(req.Context(), in)
	reply(w, http.StatusOK, resp, err)
}

func (r router) me(w http.ResponseWriter, req *http.Request) {
	// protected already validated the token; ask again for the full user record.
	token, _ := domain.BearerToken(req.Header.Get("Authorization"))
	u, err := r.Identity.ValidateToken(req.Context(), token)
	reply(w, http.StatusOK, u, err)
}

func (r router) listCategories(w http.ResponseWriter, req *http.Request) {
	resp, err := r.Categories.ListCategories(req.Context())
	reply(w, http.StatusOK, resp, err)
}

func (r router) listLocations(w http.ResponseWriter, req *http.Request) {
	c, ok := commonv1.Category_value[req.PathValue("category")]
	if !ok || c == int32(commonv1.Category_CATEGORY_UNSPECIFIED) {
		writeJSONError(w, http.StatusBadRequest, "unknown category")
		return
	}
	resp, err := r.Locations.ListLocations(req.Context(), commonv1.Category(c))
	reply(w, http.StatusOK, resp, err)
}

func (r router) submitReport(w http.ResponseWriter, req *http.Request) {
	in := &reportsubmissionv1.SubmitReportRequest{}
	if !readJSON(w, req, in) {
		return
	}
	in.ReporterId = principal(req).UserID // from the JWT, never from the body (rule 2)
	resp, err := r.Reports.SubmitReport(req.Context(), in)
	reply(w, http.StatusCreated, resp, err)
}

func (r router) getReport(w http.ResponseWriter, req *http.Request) {
	report, err := r.Reports.GetReport(req.Context(), req.PathValue("reportId"))
	if err == nil && !domain.CanReadReport(principal(req), report.GetReporterId()) {
		err = domain.ErrForbidden
	}
	reply(w, http.StatusOK, report, err)
}

func (r router) addNote(w http.ResponseWriter, req *http.Request) {
	in := &internalnotesv1.AddNoteRequest{}
	if !readJSON(w, req, in) {
		return
	}
	in.ReportId = req.PathValue("reportId") // from the path, never the body
	in.AuthorId = principal(req).UserID     // from the JWT, never the body (rule 2)
	resp, err := r.Notes.AddNote(req.Context(), actingRole(req), in)
	reply(w, http.StatusCreated, resp, err)
}

func (r router) listNotes(w http.ResponseWriter, req *http.Request) {
	resp, err := r.Notes.ListNotes(req.Context(), actingRole(req), req.PathValue("reportId"))
	reply(w, http.StatusOK, resp, err)
}

func (r router) myReports(w http.ResponseWriter, req *http.Request) {
	var pageSize int64
	if v := req.URL.Query().Get("pageSize"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "pageSize must be a number")
			return
		}
		pageSize = n
	}
	// The reporter is always the signed-in student, never a query parameter (rule 2).
	resp, err := r.History.ListReportsByReporter(req.Context(), principal(req).UserID, int32(pageSize), req.URL.Query().Get("pageToken"))
	reply(w, http.StatusOK, resp, err)
}

// timeline applies "Student (own)": the timeline has no reporter, so for a
// student the owner is looked up in report-submission first, the same check
// as GET /reports/{reportId}.
func (r router) timeline(w http.ResponseWriter, req *http.Request) {
	reportID := req.PathValue("reportId")
	if err := r.checkOwnReport(req, reportID); err != nil {
		writeError(w, err)
		return
	}
	resp, err := r.History.GetReportTimeline(req.Context(), reportID)
	reply(w, http.StatusOK, resp, err)
}

// checkOwnReport applies "Student (own)" for routes whose answer has no
// reporter: when the allowed role is Student, the report must be theirs
// (ErrForbidden otherwise, or report-submission's NOT_FOUND). Other roles pass.
func (r router) checkOwnReport(req *http.Request, reportID string) error {
	if actingRole(req) != domain.RoleStudent {
		return nil
	}
	report, err := r.Reports.GetReport(req.Context(), reportID)
	if err != nil {
		return err
	}
	if !domain.CanReadReport(principal(req), report.GetReporterId()) {
		return domain.ErrForbidden
	}
	return nil
}

// changeStatus sends the new status to status-lifecycle, which checks the flow
// (FAILED_PRECONDITION, 422) and the version (ABORTED, 409).
func (r router) changeStatus(w http.ResponseWriter, req *http.Request) {
	in := &statuslifecyclev1.ChangeStatusRequest{}
	if !readJSON(w, req, in) {
		return
	}
	if in.GetToStatus() == commonv1.Status_STATUS_UNSPECIFIED {
		writeJSONError(w, http.StatusBadRequest, "toStatus is required")
		return
	}
	in.ReportId = req.PathValue("reportId") // from the path, never the body
	// From the JWT, never the body (rule 2). It is never empty here, because an
	// empty actor means the system, which may set ESCALATED.
	in.ActorId = principal(req).UserID
	resp, err := r.Status.ChangeStatus(req.Context(), in)
	reply(w, http.StatusOK, resp, err)
}

// getStatus applies "Student (own)" like timeline: a student's ownership is
// checked with report-submission first.
func (r router) getStatus(w http.ResponseWriter, req *http.Request) {
	reportID := req.PathValue("reportId")
	if err := r.checkOwnReport(req, reportID); err != nil {
		writeError(w, err)
		return
	}
	resp, err := r.Status.GetStatus(req.Context(), reportID)
	reply(w, http.StatusOK, resp, err)
}

func (r router) dashboardSummary(w http.ResponseWriter, req *http.Request) {
	period, ok := readPeriod(w, req)
	if !ok {
		return
	}
	resp, err := r.Analytics.GetSummary(req.Context(), period)
	reply(w, http.StatusOK, resp, err)
}

func (r router) dashboardCategories(w http.ResponseWriter, req *http.Request) {
	period, ok := readPeriod(w, req)
	if !ok {
		return
	}
	resp, err := r.Analytics.GetCategoryBreakdown(req.Context(), period)
	reply(w, http.StatusOK, resp, err)
}

// readPeriod reads the optional ?from= and ?to= (RFC 3339, for example
// 2026-11-01T00:00:00+06:00). Neither given: nil, meaning all time. A bad time,
// or from after to, writes 400 and returns false.
func readPeriod(w http.ResponseWriter, req *http.Request) (*analyticsv1.Period, bool) {
	from, ok := queryTime(w, req, "from")
	if !ok {
		return nil, false
	}
	to, ok := queryTime(w, req, "to")
	if !ok {
		return nil, false
	}
	switch {
	case from == nil && to == nil:
		return nil, true
	case from != nil && to != nil && from.AsTime().After(to.AsTime()):
		writeJSONError(w, http.StatusBadRequest, "from must not be after to")
		return nil, false
	}
	return &analyticsv1.Period{From: from, To: to}, true
}

// queryTime reads one RFC 3339 query value: nil when it is not given. A bad
// value writes 400 and returns false.
func queryTime(w http.ResponseWriter, req *http.Request, name string) (*timestamppb.Timestamp, bool) {
	v := req.URL.Query().Get(name)
	if v == "" {
		return nil, true
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, name+" must be an RFC 3339 time, for example 2026-11-01T00:00:00+06:00")
		return nil, false
	}
	return timestamppb.New(t), true
}

// uploadPhoto reads the multipart field "photo" and sends it to photo-media.
// The gateway does no type check: photo-media checks the real bytes.
func (r router) uploadPhoto(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, maxPhotoBodyBytes)
	mr, err := req.MultipartReader()
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, `invalid photo: send multipart/form-data with the file in the field "photo"`)
		return
	}
	in, err := readPhotoPart(mr)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeJSONError(w, http.StatusBadRequest, errPhotoTooLarge)
			return
		}
		writeJSONError(w, http.StatusBadRequest, `invalid photo: send multipart/form-data with the file in the field "photo"`)
		return
	}
	in.UploaderId = principal(req).UserID // from the JWT, never the form (rule 2)
	resp, err := r.Photos.UploadPhoto(req.Context(), in)
	reply(w, http.StatusCreated, resp, err)
}

// readPhotoPart returns the first part named "photo". Other fields are skipped.
func readPhotoPart(mr *multipart.Reader) (*photomediav1.UploadPhotoRequest, error) {
	for {
		part, err := mr.NextPart()
		if err != nil {
			return nil, err // io.EOF: no photo field
		}
		if part.FormName() != "photo" {
			if _, err := io.Copy(io.Discard, part); err != nil {
				return nil, err
			}
			continue
		}
		data, err := io.ReadAll(part)
		if err != nil {
			return nil, err
		}
		return &photomediav1.UploadPhotoRequest{
			ContentType: part.Header.Get("Content-Type"),
			FileName:    part.FileName(),
			Data:        data,
		}, nil
	}
}

// imageTypes are the only types sent with their own Content-Type.
var imageTypes = map[string]bool{"image/jpeg": true, "image/png": true}

// getPhoto sends the raw image, so the PWA can show it from a blob URL.
func (r router) getPhoto(w http.ResponseWriter, req *http.Request) {
	photo, err := r.Photos.GetPhoto(req.Context(), req.PathValue("photoId"))
	if err == nil && !domain.CanReadReport(principal(req), photo.GetUploaderId()) {
		err = domain.ErrForbidden
	}
	if err != nil {
		writeError(w, err)
		return
	}
	ct := photo.GetContentType()
	if !imageTypes[ct] {
		ct = "application/octet-stream" // never let a browser run it
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(photo.GetData())
}

// ---- JSON and errors ----

var (
	// Strict: an unknown field or enum value (for example "category":"NOPE") is a 400,
	// not silently dropped.
	unmarshal = protojson.UnmarshalOptions{}
	// Every field is sent, so the PWA can rely on it: an empty list is [] and
	// an empty text is "". Optional fields that are not set (photoId) stay out.
	marshal = protojson.MarshalOptions{EmitUnpopulated: true}
)

// readJSON decodes the body into m. On bad JSON it writes 400 and returns false.
func readJSON(w http.ResponseWriter, req *http.Request, m proto.Message) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, req.Body, maxBodyBytes))
	if err == nil {
		err = unmarshal.Unmarshal(body, m)
	}
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

// reply writes m as JSON with code, or the error.
func reply(w http.ResponseWriter, code int, m proto.Message, err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	b, err := marshal.Marshal(m)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(b)
}

// httpStatus maps gRPC codes to HTTP (docs/design/api-routes.md, rule 4).
var httpStatus = map[codes.Code]int{
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.Aborted:            http.StatusConflict,
	codes.FailedPrecondition: http.StatusUnprocessableEntity,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrMissingToken), errors.Is(err, domain.ErrMalformedToken):
		writeJSONError(w, http.StatusUnauthorized, err.Error())
		return
	case errors.Is(err, domain.ErrForbidden):
		writeJSONError(w, http.StatusForbidden, "not allowed")
		return
	}
	st := status.Convert(err)
	code, ok := httpStatus[st.Code()]
	if !ok {
		// Never pass internal details (database errors, stack traces) to the browser.
		slog.Error("api-gateway request failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if code == http.StatusServiceUnavailable || code == http.StatusGatewayTimeout {
		// The message names internal addresses; keep it in the log only.
		slog.Warn("upstream service unavailable", "error", err)
		writeJSONError(w, code, "service temporarily unavailable, try again")
		return
	}
	writeJSONError(w, code, st.Message())
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
