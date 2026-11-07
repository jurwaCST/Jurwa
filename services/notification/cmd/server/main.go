// Command server starts the notification gRPC server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/jurwaCST/Jurwa/pkg/observability"
	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/notification/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/notification/adapters/out/logsender"
	"github.com/jurwaCST/Jurwa/services/notification/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "notification"
	defaultPort        = "9018"
)

func main() {
	serviceName := envOr("SERVICE_NAME", defaultServiceName)
	port := envOr("PORT", defaultPort)

	// Metrics on METRICS_ADDR (off when empty) and one JSON log line per request.
	obs, err := observability.Init(serviceName)
	if err != nil {
		slog.Error("cannot start observability", "error", err)
		os.Exit(1)
	}
	logger := obs.Logger()
	slog.SetDefault(logger)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("cannot listen", "port", port, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(obs.Unary()))

	// Standard gRPC health service (grpc.health.v1), used by grpcurl and Kubernetes probes.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	// PostgreSQL when DATABASE_URL is set, otherwise in memory (lost on restart).
	repo, closeRepo, err := openRepo(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("cannot open the database", "error", err)
		os.Exit(1)
	}
	defer closeRepo()
	app := application.NewService(repo, logsender.New(logger))
	notificationv1.RegisterNotificationServiceServer(srv, grpcadapter.NewHandler(app))

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
