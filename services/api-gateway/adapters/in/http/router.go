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
	"net"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/api-gateway/application"
	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
	"github.com/jurwaCST/Jurwa/services/api-gateway/ports"
)

// maxBodyBytes bounds a JSON body. Photos go to photo-media, not through here.
const maxBodyBytes = 64 << 10

// Deps are the ports and helpers the routes use.
type Deps struct {
	Auth       *application.Auth
	Identity   ports.Identity
	Categories ports.Categories
	Locations  ports.Locations
	Reports    ports.Reports
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

	return r.rateLimit(mux)
}

// ---- middleware ----

type principalKey struct{}

// protected checks the JWT with identity and the role with the authorizer
// before calling next. The principal is put in the request context.
func (r router) protected(action string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p, err := r.Auth.Authenticate(req.Context(), req.Header.Get("Authorization"))
		if err != nil {
			writeError(w, err)
			return
		}
		if err := r.Auth.Authorize(req.Context(), p, action, req.URL.Path); err != nil {
			writeError(w, err)
			return
		}
		next(w, req.WithContext(context.WithValue(req.Context(), principalKey{}, p)))
	})
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

// ---- JSON and errors ----

var (
	// Strict: an unknown field or enum value (for example "category":"NOPE") is a 400,
	// not silently dropped.
	unmarshal = protojson.UnmarshalOptions{}
	marshal   = protojson.MarshalOptions{}
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
	writeJSONError(w, code, st.Message())
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
