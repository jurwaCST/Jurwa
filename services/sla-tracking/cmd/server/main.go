// Command server starts the sla-tracking gRPC server.
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

	slapolicyv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-policy/v1"
	slatrackingv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-tracking/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/in/events"
	grpcadapter "github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/in/grpc"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/eventpublisher"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/logpublisher"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/adapters/out/slapolicy"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/application"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "sla-tracking"
	defaultPort        = "9012"
	defaultPolicyAddr  = "localhost:9013"
)

func main() {
	serviceName := envOr("SERVICE_NAME", defaultServiceName)
	port := envOr("PORT", defaultPort)
	policyAddr := envOr("SLA_POLICY_ADDR", defaultPolicyAddr)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", serviceName)
	slog.SetDefault(logger)

	// Stop cleanly on Ctrl+C (SIGINT) or when Kubernetes stops the pod (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("cannot listen", "port", port, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()

	// Standard gRPC health service (grpc.health.v1), used by grpcurl and Kubernetes probes.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

	// sla-policy gives the SLA window. NewClient does not connect yet, so the
	// server starts even when sla-policy is down; StartTimer then gives UNAVAILABLE.
	policyConn, err := grpc.NewClient(policyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("bad SLA_POLICY_ADDR", "addr", policyAddr, "error", err)
		os.Exit(1)
	}
	defer policyConn.Close()

	// In-memory timers. sla.lapsed goes to Redis Streams when REDIS_URL is set
	// (it starts the escalation Saga), otherwise it is only logged.
	var pub ports.EventPublisher = logpublisher.New(logger)
	bus, err := eventBus(ctx)
	if err != nil {
		slog.Error("cannot connect to the event bus", "error", err)
		os.Exit(1)
	}
	if bus != nil {
		defer bus.Close()
		pub = eventpublisher.New(bus, uuid.NewString, time.Now)
	}
	app := application.NewService(
		memory.New(),
		slapolicy.New(slapolicyv1.NewSlaPolicyServiceClient(policyConn)),
		pub,
		time.Now,
	)
	if bus != nil {
		// report.routed starts the timer, status.changed to RESOLVED stops it.
		events.Start(ctx, bus, consumerName(serviceName), app)
		slog.Info("event bus connected", "group", events.Group)
	}
	slatrackingv1.RegisterSlaTrackingServiceServer(srv, grpcadapter.NewHandler(app))

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

// envOr returns the environment variable key, or fallback when it is empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
