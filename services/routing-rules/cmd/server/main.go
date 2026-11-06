// Command server starts the routing-rules gRPC server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/routing-rules/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/routing-rules/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/routing-rules/adapters/out/postgres"
	"github.com/jurwaCST/Jurwa/services/routing-rules/application"
	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
	"github.com/jurwaCST/Jurwa/services/routing-rules/ports"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "routing-rules"
	defaultPort        = "9006"
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

	// PostgreSQL when DATABASE_URL is set, otherwise in memory (seeded with the
	// defaults, lost on restart). Unit tests and "make run" need no database.
	repo, closeRepo, err := openRepo(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("cannot open the database", "error", err)
		os.Exit(1)
	}
	defer closeRepo()
	routingrulesv1.RegisterRoutingRulesServiceServer(srv, grpcadapter.NewHandler(application.NewService(repo)))

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

// openRepo returns the PostgreSQL repository when url is set (creating the
// table and seeding it if needed), otherwise the in-memory one.
func openRepo(url string) (ports.RuleRepository, func(), error) {
	if url == "" {
		slog.Info("storage: in memory (DATABASE_URL not set)")
		return memory.NewRepository(domain.DefaultRules()), func() {}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, nil, err
	}
	repo := postgres.New(pool)
	if err := repo.Migrate(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}
	slog.Info("storage: postgres")
	return repo, pool.Close, nil
}
