// Command server starts the identity gRPC server.
package main

import (
	"context"
	"errors"
	"fmt"
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

	"github.com/jurwaCST/Jurwa/pkg/observability"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/identity/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/identity/application"
	"github.com/jurwaCST/Jurwa/services/identity/domain"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
// JWT_SECRET has no default on purpose: the server will not start without it.
const (
	defaultServiceName = "identity"
	defaultPort        = "9016"
	defaultJWTTTL      = "15m"
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

	tokens, err := newTokenIssuer(os.Getenv("JWT_SECRET"), envOr("JWT_TTL", defaultJWTTTL))
	if err != nil {
		slog.Error("bad token config", "error", err)
		os.Exit(1)
	}
	// PostgreSQL when DATABASE_URL is set, otherwise in memory (lost on restart).
	users, closeUsers, err := openRepo(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("cannot open the database", "error", err)
		os.Exit(1)
	}
	defer closeUsers()
	app := application.NewService(users, tokens)

	// Demo accounts for the roles that cannot register (SEED_DEMO_USERS=true).
	created, err := seedFromEnv(context.Background(), app, os.Getenv)
	if err != nil {
		slog.Error("cannot seed demo users", "error", err)
		os.Exit(1)
	}
	if created > 0 {
		slog.Warn("demo users seeded, for the demo only", "created", created)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("cannot listen", "port", port, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(obs.Unary()))

	// Standard gRPC health service (grpc.health.v1), used by grpcurl and Kubernetes probes.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	identityv1.RegisterIdentityServiceServer(srv, grpcadapter.NewHandler(app))

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

// newTokenIssuer reads the JWT settings. The secret must be set and at least
// 32 bytes; the expiry is a Go duration such as "15m".
func newTokenIssuer(secret, ttl string) (*domain.TokenIssuer, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	d, err := time.ParseDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("JWT_TTL %q is not a duration such as 15m: %w", ttl, err)
	}
	return domain.NewTokenIssuer([]byte(secret), d)
}
