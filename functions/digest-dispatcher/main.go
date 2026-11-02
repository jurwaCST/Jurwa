// Command digest-dispatcher is the serverless function that sends the daily
// digest. It is a small HTTP program: every request to "/" does one run and
// returns {"digests":N,"errors":M}. On Knative a PingSource calls it on a
// schedule (Sprint 3, see README).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jurwaCST/Jurwa/functions/digest-dispatcher/clients"
	"github.com/jurwaCST/Jurwa/functions/digest-dispatcher/dispatcher"
	"github.com/jurwaCST/Jurwa/pkg/resilience"
	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
)

func main() {
	port := envOr("PORT", "8080") // Knative sets PORT
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "digest-dispatcher"))

	analyticsConn := mustDial("analytics", envOr("ANALYTICS_ADDR", clients.DefaultAnalyticsAddr))
	defer analyticsConn.Close()
	directoryConn := mustDial("department-directory", envOr("DEPARTMENT_DIRECTORY_ADDR", clients.DefaultDirectoryAddr))
	defer directoryConn.Close()
	notificationConn := mustDial("notification", envOr("NOTIFICATION_ADDR", clients.DefaultNotificationAddr))
	defer notificationConn.Close()

	d := dispatcher.New(
		clients.NewAnalytics(analyticsv1.NewAnalyticsServiceClient(analyticsConn)),
		clients.NewDirectory(departmentdirectoryv1.NewDepartmentDirectoryServiceClient(directoryConn)),
		clients.NewNotification(notificationv1.NewNotificationServiceClient(notificationConn)),
	)

	srv := &http.Server{Addr: ":" + port, Handler: newHandler(d), ReadHeaderTimeout: 5 * time.Second}

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

// newHandler runs the digest on GET or POST "/". PingSource sends a POST.
func newHandler(d *dispatcher.Dispatcher) http.Handler {
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
		res, err := d.Run(r.Context())
		code := http.StatusOK
		if err != nil {
			slog.ErrorContext(r.Context(), "digest run failed", "error", err)
			res = dispatcher.Result{Errors: 1}
			code = http.StatusServiceUnavailable
		} else {
			slog.InfoContext(r.Context(), "digest run finished", "digests", res.Digests, "errors", res.Errors)
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
