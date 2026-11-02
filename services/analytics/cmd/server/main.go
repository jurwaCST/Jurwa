// Command server starts the analytics gRPC server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/analytics/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/analytics/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/analytics/application"
	"github.com/jurwaCST/Jurwa/services/analytics/domain"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "analytics"
	defaultPort        = "9020"
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

	// Built from events only. In-memory for now; the Redis Streams consumer
	// that calls app.Record and a database come in Sprint 3.
	app := application.NewService(memory.New())
	if os.Getenv("SEED_DEMO_DATA") == "true" {
		n := seed(app, demoEvents(time.Now().UTC()))
		slog.Info("demo data loaded", "events", n)
	}
	analyticsv1.RegisterAnalyticsServiceServer(srv, grpcadapter.NewHandler(app))

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

// seed records the events and returns how many were applied.
func seed(app *application.Service, events []domain.Event) int {
	n := 0
	for _, e := range events {
		if ok, _ := app.Record(context.Background(), e); ok {
			n++
		}
	}
	return n
}

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
