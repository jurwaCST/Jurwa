// Command server starts the api-gateway: HTTP for the PWA on HTTP_PORT (8080)
// and the gRPC health check on PORT (9010).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/jurwaCST/Jurwa/pkg/resilience"
	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	httpadapter "github.com/jurwaCST/Jurwa/services/api-gateway/adapters/in/http"
	"github.com/jurwaCST/Jurwa/services/api-gateway/adapters/out/grpcclients"
	"github.com/jurwaCST/Jurwa/services/api-gateway/adapters/out/staticauthz"
	"github.com/jurwaCST/Jurwa/services/api-gateway/application"
	"github.com/jurwaCST/Jurwa/services/api-gateway/domain"
)

// Defaults used when the environment variables are not set.
// scripts/new-service.sh replaces the placeholders.
const (
	defaultServiceName = "api-gateway"
	defaultPort        = "9010"
	defaultHTTPPort    = "8080"
	defaultRateLimit   = 20 // requests per second per client IP
)

func main() {
	serviceName := envOr("SERVICE_NAME", defaultServiceName)
	port := envOr("PORT", defaultPort)
	httpPort := envOr("HTTP_PORT", defaultHTTPPort)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", serviceName)
	slog.SetDefault(logger)

	rateLimit, err := strconv.Atoi(envOr("RATE_LIMIT_PER_SECOND", strconv.Itoa(defaultRateLimit)))
	if err != nil || rateLimit < 1 {
		slog.Error("RATE_LIMIT_PER_SECOND must be a positive number")
		os.Exit(1)
	}

	// One connection per service, each with its own timeout, retry and circuit breaker.
	conns := map[string]*grpc.ClientConn{}
	for name, addr := range map[string]string{
		"identity":          envOr("IDENTITY_ADDR", "localhost:9016"),
		"category-taxonomy": envOr("CATEGORY_TAXONOMY_ADDR", "localhost:9009"),
		"location-taxonomy": envOr("LOCATION_TAXONOMY_ADDR", "localhost:9004"),
		"report-submission": envOr("REPORT_SUBMISSION_ADDR", "localhost:9001"),
	} {
		conn, err := dial(addr)
		if err != nil {
			slog.Error("cannot create client", "service", name, "addr", addr, "error", err)
			os.Exit(1)
		}
		defer conn.Close()
		conns[name] = conn
	}

	identity := grpcclients.Identity{C: identityv1.NewIdentityServiceClient(conns["identity"])}
	router := httpadapter.NewRouter(httpadapter.Deps{
		// TODO(SCRUM-38): use authorization CheckPermission (AUTHORIZATION_ADDR,
		// localhost:9017) instead of staticauthz once Gayley's service is merged.
		Auth:       application.NewAuth(identity, staticauthz.New()),
		Identity:   identity,
		Categories: grpcclients.Categories{C: categorytaxonomyv1.NewCategoryTaxonomyServiceClient(conns["category-taxonomy"])},
		Locations:  grpcclients.Locations{C: locationtaxonomyv1.NewLocationTaxonomyServiceClient(conns["location-taxonomy"])},
		Reports:    grpcclients.Reports{C: reportsubmissionv1.NewReportSubmissionServiceClient(conns["report-submission"])},
		Limiter:    domain.NewLimiter(rateLimit, time.Now),
	})
	httpSrv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("cannot listen", "port", port, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()

	// Standard gRPC health service (grpc.health.v1), used by grpcurl and Kubernetes probes.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(srv, hs)

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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
		srv.GracefulStop()
	}()

	go func() {
		slog.Info("http server started", "port", httpPort)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped", "error", err)
			stop()
		}
	}()

	slog.Info("server started", "port", port)
	if err := srv.Serve(lis); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

// dial creates a lazy client connection wrapped with pkg/resilience.
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
