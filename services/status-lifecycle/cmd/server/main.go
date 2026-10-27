// Command server starts the status-lifecycle gRPC server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "status-lifecycle"
	defaultPort        = "9011"
)

func main() {
	serviceName := envOr("SERVICE_NAME", defaultServiceName)
	port := envOr("PORT", defaultPort)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", serviceName)
	slog.SetDefault(logger)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("cannot listen", "port", port, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()

	// Standard gRPC health service (grpc.health.v1), used by grpcurl and Kubernetes probes.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	// Wire the hexagon: in-memory store (PostgreSQL in Sprint 2) and a log
	// publisher (Redis Streams in Sprint 2).
	app := application.NewService(memory.New(), logpublisher.New(logger), time.Now)
	seedReports(app, os.Getenv("SEED_REPORT_IDS"))
	statuslifecyclev1.RegisterStatusLifecycleServiceServer(srv, grpcadapter.NewHandler(app))

	// Reflection lets grpcurl list and call methods without the .proto files.
	reflection.Register(srv)

	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Stop cleanly on Ctrl+C (SIGINT) or when Kubernetes stops the pod (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		slog.Info("shutting down")
		hs.Shutdown()
		srv.GracefulStop()
	}()

	slog.Info("server started", "port", port)
	if err := srv.Serve(lis); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// seedReports registers each comma-separated report id as OPEN. Until the
// report.submitted event is consumed (Sprint 2), this is how a demo gets reports.
func seedReports(app *application.Service, ids string) {
	for _, id := range strings.Split(ids, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, err := app.Register(context.Background(), id); err != nil {
			slog.Warn("seed report skipped", "report_id", id, "error", err)
			continue
		}
		slog.Info("seed report registered", "report_id", id, "status", "OPEN")
	}
}
