// Command server starts the reassignment gRPC server.
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

	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/reassignment/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/reassignment/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/reassignment/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "reassignment"
	defaultPort        = "9014"
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

	// In-memory store and a log publisher for now; PostgreSQL and Redis Streams later.
	app := application.NewService(memory.New(), logpublisher.New(logger), time.Now, uuid.NewString)
	seedAssignments(app, os.Getenv("SEED_ASSIGNMENTS"))
	reassignmentv1.RegisterReassignmentServiceServer(srv, grpcadapter.NewHandler(app))

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

// seedAssignments records ROUTED assignments from "reportID:userID:position,..."
// (position optional). It stands in for report.routed events until the event bus exists.
func seedAssignments(app *application.Service, spec string) {
	for _, item := range strings.Split(spec, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, ":", 3)
		for len(parts) < 3 {
			parts = append(parts, "")
		}
		a, err := app.RecordRouted(context.Background(), parts[0], parts[1], parts[2])
		if err != nil {
			slog.Warn("seed assignment skipped", "item", item, "error", err)
			continue
		}
		slog.Info("seed assignment recorded", "report_id", a.ReportID, "assignee_user_id", a.AssigneeUserID)
	}
}
