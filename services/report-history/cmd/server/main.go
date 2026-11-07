// Command server starts the report-history gRPC server.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
	"github.com/jurwaCST/Jurwa/services/report-history/adapters/in/events"
	grpcadapter "github.com/jurwaCST/Jurwa/services/report-history/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/report-history/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-history/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "report-history"
	defaultPort        = "9003"
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

	// In-memory store. The data is rebuilt from the event streams on start
	// (see startConsumers); a PostgreSQL adapter would keep it across restarts.
	app := application.NewService(memory.NewRepository(), uuid.NewString)
	if os.Getenv("SEED_SAMPLE") == "true" {
		if err := seedSample(context.Background(), app); err != nil {
			slog.Error("cannot load sample data", "error", err)
			os.Exit(1)
		}
		slog.Info("sample data loaded", "reporter_id", sampleStudentID, "assignee_id", sampleDSAID)
	}
	reporthistoryv1.RegisterReportHistoryServiceServer(srv, grpcadapter.NewHandler(app))

	// Reflection lets grpcurl list and call methods without the .proto files.
	reflection.Register(srv)

	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Stop cleanly on Ctrl+C (SIGINT) or when Kubernetes stops the pod (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := startConsumers(ctx, app, serviceName); err != nil {
		slog.Error("cannot start event consumers", "error", err)
		os.Exit(1)
	}

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

// startConsumers reads the event streams when REDIS_URL is set. The store is in
// memory, so every stream is read again from its first event on start
// (Rewind); events seen before are ignored by their event id.
func startConsumers(ctx context.Context, app *application.Service, serviceName string) error {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		slog.Info("events: not read (REDIS_URL not set)")
		return nil
	}
	bus, err := eventbus.NewRedis(url)
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := bus.Ping(pingCtx); err != nil {
		_ = bus.Close()
		return err
	}
	for _, stream := range events.Streams() {
		if err := bus.Rewind(ctx, stream, events.Group); err != nil {
			_ = bus.Close()
			return err
		}
	}
	consumer, _ := os.Hostname()
	if consumer == "" {
		consumer = serviceName
	}
	go func() {
		defer bus.Close()
		if err := events.NewConsumer(app).Run(ctx, bus, consumer); err != nil {
			slog.Error("event consumer stopped", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("events: reading", "streams", events.Streams(), "group", events.Group, "consumer", consumer)
	return nil
}

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
