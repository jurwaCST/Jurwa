// Command server starts the photo-media gRPC server.
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/photo-media/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/photo-media/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/photo-media/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "photo-media"
	defaultPort        = "9002"
)

// maxRecvMsgBytes lets one request carry a 5 MB photo plus the other fields.
// gRPC accepts only 4 MB by default (see proto/photo-media/v1/photo-media.proto).
const maxRecvMsgBytes = 6 << 20

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

	srv := grpc.NewServer(grpc.MaxRecvMsgSize(maxRecvMsgBytes))

	// Standard gRPC health service (grpc.health.v1), used by grpcurl and Kubernetes probes.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	// In-memory store for Sprint 2; the MongoDB GridFS adapter replaces it in Sprint 3.
	app := application.NewService(memory.NewRepository(), uuid.NewString, time.Now)
	photomediav1.RegisterPhotoMediaServiceServer(srv, grpcadapter.NewHandler(app))

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
