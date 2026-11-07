// Command server starts the routing-engine gRPC server.
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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/jurwaCST/Jurwa/pkg/eventbus"
	"github.com/jurwaCST/Jurwa/pkg/resilience"
	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	routingenginev1 "github.com/jurwaCST/Jurwa/proto/gen/routing-engine/v1"
	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/in/events"
	grpcadapter "github.com/jurwaCST/Jurwa/services/routing-engine/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/directoryclient"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/eventpublisher"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/routing-engine/adapters/out/rulesclient"
	"github.com/jurwaCST/Jurwa/services/routing-engine/application"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
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
	// department-directory for the holder.
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

	// Stop cleanly on Ctrl+C (SIGINT) or when Kubernetes stops the pod (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// With REDIS_URL: publish report.routed on Redis Streams and route every
	// report.submitted. Without it: events go to the log and only the RPC routes.
	bus, err := openBus(ctx, os.Getenv("REDIS_URL"))
	if err != nil {
		slog.Error("cannot connect to the event bus", "error", err)
		os.Exit(1)
	}
	var pub ports.EventPublisher = logpublisher.New(logger)
	if bus != nil {
		defer bus.Close()
		pub = eventpublisher.New(bus, uuid.NewString, time.Now)
	}

	app := application.NewService(
		rulesclient.New(routingrulesv1.NewRoutingRulesServiceClient(rulesConn)),
		directoryclient.New(departmentdirectoryv1.NewDepartmentDirectoryServiceClient(dirConn)),
		pub,
	)

	if bus != nil {
		consumer, _ := os.Hostname() // one consumer name per pod
		go func() {
			if err := bus.Subscribe(ctx, eventbus.StreamReportSubmitted, events.Group, consumer, events.New(app).Handle); err != nil {
				slog.Error("report.submitted consumer stopped", "error", err)
				stop()
			}
		}()
		slog.Info("events: redis streams", "reads", eventbus.StreamReportSubmitted, "publishes", eventbus.StreamReportRouted)
	} else {
		slog.Info("events: log only (REDIS_URL not set)")
	}
	routingenginev1.RegisterRoutingEngineServiceServer(srv, grpcadapter.NewHandler(app))

	// Reflection lets grpcurl list and call methods without the .proto files.
	reflection.Register(srv)

	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

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

// openBus connects to Redis Streams when url is set, and nil when it is not.
// A Redis that does not answer within 10 seconds is an error.
func openBus(ctx context.Context, url string) (*eventbus.Redis, error) {
	if url == "" {
		return nil, nil
	}
	bus, err := eventbus.NewRedis(url)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := bus.Ping(pingCtx); err != nil {
		_ = bus.Close()
		return nil, err
	}
	return bus, nil
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
// Every call gets a timeout, retries and its own circuit breaker (pkg/resilience).
func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(resilience.UnaryClientInterceptor(resilience.DefaultConfig())),
	)
}
