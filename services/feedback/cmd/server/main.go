// Command server starts the feedback gRPC server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/jurwaCST/Jurwa/pkg/resilience"
	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/feedback/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/feedback/adapters/out/statusclient"
	"github.com/jurwaCST/Jurwa/services/feedback/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "feedback"
	defaultPort        = "9019"
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

	// feedback asks status-lifecycle whether a report is RESOLVED.
	statusConn, err := dial(envOr("STATUS_LIFECYCLE_ADDR", statusclient.DefaultAddr))
	if err != nil {
		slog.Error("cannot create status-lifecycle client", "error", err)
		os.Exit(1)
	}
	defer statusConn.Close()

	// PostgreSQL when DATABASE_URL is set, otherwise in memory (lost on restart).
	repo, closeRepo, err := openRepo(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("cannot open the database", "error", err)
		os.Exit(1)
	}
	defer closeRepo()
	app := application.NewService(
		repo,
		statusclient.New(statuslifecyclev1.NewStatusLifecycleServiceClient(statusConn)),
	)
	feedbackv1.RegisterFeedbackServiceServer(srv, grpcadapter.NewHandler(app))

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

// dial creates a client connection. gRPC connects lazily, so if
// status-lifecycle is not running yet, calls fail with Unavailable instead of
// stopping the server. Every call gets a timeout, retries and a circuit
// breaker (pkg/resilience).
func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(resilience.UnaryClientInterceptor(resilience.DefaultConfig())),
	)
}

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
