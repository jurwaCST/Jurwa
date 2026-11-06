// Package observability gives every Jurwa gRPC service the same metrics and
// request logs. A service calls Init once and adds Unary to its server:
//
//	obs, err := observability.Init("identity")
//	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(obs.Unary()))
//
// Metrics are served for Prometheus on METRICS_ADDR (for example ":9100"); when
// it is empty no port is opened, so local runs are unchanged. See README.md.
package observability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Config is what New needs. Init fills it from the environment.
type Config struct {
	Service     string    // required, the "service" label and log field
	MetricsAddr string    // where /metrics is served, "" means off
	LogOutput   io.Writer // where log lines go, os.Stdout when nil
}

// Obs holds one service's metrics, logger and metrics server.
type Obs struct {
	service   string
	logger    *slog.Logger
	registry  *prometheus.Registry
	requests  *prometheus.CounterVec
	durations *prometheus.HistogramVec
	server    *http.Server
	addr      string
}

// Init is New with METRICS_ADDR from the environment and logs on stdout.
func Init(service string) (*Obs, error) {
	return New(Config{Service: service, MetricsAddr: os.Getenv("METRICS_ADDR")})
}

// New makes the metrics and the logger and, when cfg.MetricsAddr is set, starts
// the metrics server. A wrong address is an error, so the service does not
// start without the metrics it was asked for.
func New(cfg Config) (*Obs, error) {
	if cfg.Service == "" {
		return nil, errors.New("observability: service name is required")
	}
	out := cfg.LogOutput
	if out == nil {
		out = os.Stdout
	}
	o := &Obs{
		service:  cfg.Service,
		logger:   slog.New(slog.NewJSONHandler(out, nil)).With("service", cfg.Service),
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "gRPC requests handled, by service, method and result code.",
		}, []string{"service", "method", "code"}),
		durations: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "How long a gRPC request took, in seconds.",
			Buckets: prometheus.DefBuckets, // 5 ms to 10 s
		}, []string{"service", "method"}),
	}
	// Each Obs has its own registry, so tests and services never share counters.
	o.registry.MustRegister(o.requests, o.durations,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	if cfg.MetricsAddr != "" {
		if err := o.serveMetrics(cfg.MetricsAddr); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func (o *Obs) serveMetrics(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("observability: metrics address %q: %w", addr, err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(o.registry, promhttp.HandlerOpts{}))
	o.server = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	o.addr = lis.Addr().String()
	go func() {
		if err := o.server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			o.logger.Error("metrics server stopped", "error", err)
		}
	}()
	return nil
}

// Logger is a JSON logger that adds "service" to every line. Use it for the
// service's own log lines, for example slog.SetDefault(obs.Logger()).
func (o *Obs) Logger() *slog.Logger { return o.logger }

// MetricsAddr is the address /metrics is served on, "" when off.
func (o *Obs) MetricsAddr() string { return o.addr }

// Close stops the metrics server, if there is one.
func (o *Obs) Close(ctx context.Context) error {
	if o.server == nil {
		return nil
	}
	return o.server.Shutdown(ctx)
}

// Unary is the server interceptor: for every request it counts the result,
// records the duration and writes one log line with the correlation id. The
// handler's reply and error are returned unchanged.
func (o *Obs) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		ctx, id := withCorrelationID(ctx)

		resp, err := handler(ctx, req)

		took := time.Since(start)
		code := status.Code(err).String() // OK when err is nil, Unknown for a plain Go error
		o.requests.WithLabelValues(o.service, info.FullMethod, code).Inc()
		o.durations.WithLabelValues(o.service, info.FullMethod).Observe(took.Seconds())

		level := slog.LevelInfo
		if err != nil {
			level = slog.LevelWarn
		}
		o.logger.LogAttrs(ctx, level, "grpc request",
			slog.String("method", info.FullMethod),
			slog.String("code", code),
			slog.Float64("duration_ms", float64(took.Microseconds())/1000),
			slog.String("correlation_id", id))
		return resp, err
	}
}
