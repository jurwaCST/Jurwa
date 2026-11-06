// Command server starts the report-submission gRPC server.
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

	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/report-submission/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/mongo"
	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/uuid"
	"github.com/jurwaCST/Jurwa/services/report-submission/application"
	"github.com/jurwaCST/Jurwa/services/report-submission/ports"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "report-submission"
	defaultPort        = "9001"
	defaultMongoDB     = "report_submission"
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

	repo, closeRepo, err := newRepository(context.Background())
	if err != nil {
		slog.Error("cannot open storage", "error", err)
		os.Exit(1)
	}
	defer closeRepo()

	app := application.NewService(repo, uuid.Generator{}, time.Now)
	reportsubmissionv1.RegisterReportSubmissionServiceServer(srv, grpcadapter.NewHandler(app))

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

// newRepository uses MongoDB when MONGO_URL is set, otherwise memory (data is
// lost on restart). The database name comes from MONGO_DB.
func newRepository(ctx context.Context) (ports.ReportRepository, func(), error) {
	url := os.Getenv("MONGO_URL")
	if url == "" {
		slog.Info("storage: memory (MONGO_URL not set)")
		return memory.NewRepository(), func() {}, nil
	}
	database := envOr("MONGO_DB", defaultMongoDB)
	repo, err := mongo.Connect(ctx, url, database)
	if err != nil {
		return nil, nil, err
	}
	if err := repo.EnsureIndexes(ctx); err != nil {
		_ = repo.Close(ctx)
		return nil, nil, err
	}
	slog.Info("storage: mongodb", "database", database)
	return repo, func() { _ = repo.Close(context.Background()) }, nil
}

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
