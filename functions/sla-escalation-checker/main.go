// Command sla-escalation-checker is the serverless function that escalates
// reports past their SLA (FR-15). It is a small HTTP program: every request to
// "/" does one run and returns {"lapsed":N,"escalated":N,"skipped":N,"errors":N}.
// On Knative a PingSource calls it on a schedule (SCRUM-104, see README).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jurwaCST/Jurwa/functions/sla-escalation-checker/checker"
	"github.com/jurwaCST/Jurwa/functions/sla-escalation-checker/clients"
	"github.com/jurwaCST/Jurwa/pkg/resilience"
	slatrackingv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-tracking/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

func main() {
	port := envOr("PORT", "8080") // Knative sets PORT
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "sla-escalation-checker"))

	trackingConn := mustDial("sla-tracking", envOr("SLA_TRACKING_ADDR", clients.DefaultSlaTrackingAddr))
	defer trackingConn.Close()
	lifecycleConn := mustDial("status-lifecycle", envOr("STATUS_LIFECYCLE_ADDR", clients.DefaultStatusLifecycleAddr))
	defer lifecycleConn.Close()

	c := checker.New(
		clients.NewSlaTracking(slatrackingv1.NewSlaTrackingServiceClient(trackingConn)),
		clients.NewStatusLifecycle(statuslifecyclev1.NewStatusLifecycleServiceClient(lifecycleConn)),
	)

	// RUN_ONCE=true: one check, then exit. The Kubernetes CronJob fallback
	// (deploy/knative/cronjob-fallback.yaml) runs the image this way.
	if runOnceEnabled() {
		ctx, cancel := context.WithTimeout(context.Background(), runOnceTimeout)
		code := runOnce(ctx, c, os.Stdout)
		cancel()
		trackingConn.Close()
		lifecycleConn.Close()
		os.Exit(code)
	}

	srv := &http.Server{Addr: ":" + port, Handler: newHandler(c), ReadHeaderTimeout: 5 * time.Second}

	// Stop cleanly on Ctrl+C or when Knative scales the pod down (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("function started", "port", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// runner is what the handler needs; *checker.Checker implements it.
type runner interface {
	Run(ctx context.Context) (checker.Result, error)
}

// newHandler runs the check on GET or POST "/". PingSource sends a POST.
// If sla-tracking cannot be reached the answer is 503, with the counts of
// anything already done in this run.
func newHandler(c runner) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		res, err := c.Run(r.Context())
		code := http.StatusOK
		if err != nil {
			slog.ErrorContext(r.Context(), "escalation run failed", "error", err)
			res.Errors++
			code = http.StatusServiceUnavailable
		} else {
			slog.InfoContext(r.Context(), "escalation run finished",
				"lapsed", res.Lapsed, "escalated", res.Escalated, "skipped", res.Skipped, "errors", res.Errors)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(res)
	})
	return mux
}

// mustDial creates a client connection. gRPC connects lazily, so a service
// that is down only fails the run, not the start. Every call gets a timeout,
// retries and a circuit breaker (pkg/resilience).
func mustDial(name, addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(resilience.UnaryClientInterceptor(resilience.DefaultConfig())),
	)
	if err != nil {
		slog.Error("cannot create client", "service", name, "addr", addr, "error", err)
		os.Exit(1)
	}
	return conn
}

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// runOnceTimeout bounds one CronJob run, well inside its 1 minute schedule.
const runOnceTimeout = 45 * time.Second

// runOnceEnabled reports whether RUN_ONCE asks for a single check.
func runOnceEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("RUN_ONCE")))
	return v == "true" || v == "1"
}

// runOnce does one check and writes the counts as JSON, like the HTTP answer.
// It returns the exit code: 0 when the check ran, 1 when sla-tracking could not
// be asked, so the Kubernetes Job is marked failed.
func runOnce(ctx context.Context, c runner, w io.Writer) int {
	res, err := c.Run(ctx)
	code := 0
	if err != nil {
		slog.ErrorContext(ctx, "escalation run failed", "error", err)
		res.Errors++
		code = 1
	} else {
		slog.InfoContext(ctx, "escalation run finished",
			"lapsed", res.Lapsed, "escalated", res.Escalated, "skipped", res.Skipped, "errors", res.Errors)
	}
	_ = json.NewEncoder(w).Encode(res)
	return code
}
