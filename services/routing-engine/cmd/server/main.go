// Command server starts the routing-engine gRPC server.
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

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	routingenginev1 "github.com/jurwaCST/Jurwa/proto/gen/routing-engine/v1"
	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	grpcadapter "github.com/jurwaCST/Jurwa/services/routing-engine/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/directoryclient"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/rulesclient"
	"github.com/jurwaCST/Jurwa/services/routing-engine/application"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "routing-engine"
	defaultPort        = "9007"
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

	// routing-engine is stateless: it asks routing-rules for the rule and
	// department-directory for the holder. Events go to the log until Redis
	// Streams arrives in Sprint 3.
	rulesConn, err := dial(envOr("ROUTING_RULES_ADDR", rulesclient.DefaultAddr))
	if err != nil {
		slog.Error("cannot create routing-rules client", "error", err)
		os.Exit(1)
	}
	defer rulesConn.Close()
	dirConn, err := dial(envOr("DEPARTMENT_DIRECTORY_ADDR", directoryclient.DefaultAddr))
	if err != nil {
		slog.Error("cannot create department-directory client", "error", err)
		os.Exit(1)
	}
	defer dirConn.Close()

	app := application.NewService(
		rulesclient.New(routingrulesv1.NewRoutingRulesServiceClient(rulesConn)),
		directoryclient.New(departmentdirectoryv1.NewDepartmentDirectoryServiceClient(dirConn)),
		logpublisher.New(logger),
	)
	routingenginev1.RegisterRoutingEngineServiceServer(srv, grpcadapter.NewHandler(app))

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

// dial creates a client connection. gRPC connects lazily, so a service that is
// not running yet makes calls fail with Unavailable instead of stopping the server.
func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
