package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const method = "/identity.v1.IdentityService/Login"

var info = &grpc.UnaryServerInfo{FullMethod: method}

// newTest makes an Obs that logs into buf and serves no metrics.
func newTest(t *testing.T, buf *bytes.Buffer) *Obs {
	t.Helper()
	o, err := New(Config{Service: "identity", LogOutput: buf})
	require.NoError(t, err)
	t.Cleanup(func() { _ = o.Close(context.Background()) })
	return o
}

func okHandler(ctx context.Context, req any) (any, error) { return "reply", nil }

// logLines returns each JSON log line in buf as a map.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var lines []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(l), &m), "not one JSON object per line: %q", l)
		lines = append(lines, m)
	}
	return lines
}

func TestNewNeedsAServiceName(t *testing.T) {
	_, err := New(Config{Service: ""})
	assert.Error(t, err)
}

func TestUnaryCountsOneRequestWithItsLabels(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)

	reply, err := o.Unary()(context.Background(), "req", info, okHandler)

	require.NoError(t, err)
	assert.Equal(t, "reply", reply, "the handler's reply is passed through")
	assert.Equal(t, 1.0, testutil.ToFloat64(o.requests.WithLabelValues("identity", method, "OK")))
	assert.Equal(t, 1, testutil.CollectAndCount(o.requests), "one label set only")
}

func TestUnaryRecordsADuration(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)

	_, _ = o.Unary()(context.Background(), "req", info, okHandler)
	_, _ = o.Unary()(context.Background(), "req", info, okHandler)

	assert.Equal(t, 1, testutil.CollectAndCount(o.durations), "one histogram for service and method")
	// Only the number of durations is checked, the time itself changes on every run.
	assert.Equal(t, uint64(2), sampleCount(t, o))
}

func TestUnaryCountsAHandlerErrorWithItsCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{name: "gRPC status error", err: status.Error(codes.NotFound, "no such user"), wantCode: "NotFound"},
		{name: "unauthenticated", err: status.Error(codes.Unauthenticated, "wrong password"), wantCode: "Unauthenticated"},
		{name: "plain Go error is Unknown", err: errors.New("boom"), wantCode: "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			o := newTest(t, &buf)
			failing := func(ctx context.Context, req any) (any, error) { return nil, tt.err }

			_, err := o.Unary()(context.Background(), "req", info, failing)

			assert.Equal(t, tt.err, err, "the handler's error is passed through unchanged")
			assert.Equal(t, 1.0, testutil.ToFloat64(o.requests.WithLabelValues("identity", method, tt.wantCode)))
			assert.Equal(t, 0.0, testutil.ToFloat64(o.requests.WithLabelValues("identity", method, "OK")))
		})
	}
}

func TestUnaryWritesOneJSONLogLine(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-correlation-id", "corr-123"))

	_, _ = o.Unary()(ctx, "req", info, func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.PermissionDenied, "no")
	})

	lines := logLines(t, &buf)
	require.Len(t, lines, 1)
	line := lines[0]
	assert.Equal(t, "identity", line["service"])
	assert.Equal(t, method, line["method"])
	assert.Equal(t, "PermissionDenied", line["code"])
	assert.Equal(t, "corr-123", line["correlation_id"])
	assert.Contains(t, line, "duration_ms")
}

// Kubernetes probes call the health check every few seconds: count them, but
// do not fill the logs with them.
func TestUnaryCountsHealthChecksWithoutALogLine(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)
	health := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}

	_, _ = o.Unary()(context.Background(), "req", health, okHandler)

	assert.Equal(t, 1.0, testutil.ToFloat64(o.requests.WithLabelValues("identity", health.FullMethod, "OK")))
	assert.Empty(t, buf.String(), "no log line for a health check")
}

func TestUnaryLogsAFailedHealthCheck(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)
	health := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}

	_, _ = o.Unary()(context.Background(), "req", health, func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "unknown service")
	})

	assert.Len(t, logLines(t, &buf), 1, "a failing health check is worth a line")
}

func TestUnaryCreatesACorrelationIDWhenMissing(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)
	var seen string

	_, _ = o.Unary()(context.Background(), "req", info, func(ctx context.Context, req any) (any, error) {
		seen = CorrelationID(ctx)
		return "ok", nil
	})

	lines := logLines(t, &buf)
	require.Len(t, lines, 1)
	id, _ := lines[0]["correlation_id"].(string)
	assert.NotEmpty(t, id, "a new id is created")
	assert.Equal(t, id, seen, "the handler sees the same id as the log line")
}

func TestUnaryGivesTheHandlerTheIncomingCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-correlation-id", "from-gateway"))
	var seen string

	_, _ = o.Unary()(ctx, "req", info, func(ctx context.Context, req any) (any, error) {
		seen = CorrelationID(ctx)
		return "ok", nil
	})

	assert.Equal(t, "from-gateway", seen)
}

func TestCorrelationIDIsEmptyOutsideARequest(t *testing.T) {
	assert.Equal(t, "", CorrelationID(context.Background()))
}

func TestLoggerAddsTheServiceName(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)

	o.Logger().Info("server started", "port", "9016")

	lines := logLines(t, &buf)
	require.Len(t, lines, 1)
	assert.Equal(t, "identity", lines[0]["service"])
	assert.Equal(t, "server started", lines[0]["msg"])
}

func TestMetricsServerIsOffWithoutAnAddress(t *testing.T) {
	var buf bytes.Buffer
	o := newTest(t, &buf)
	assert.Equal(t, "", o.MetricsAddr())
}

func TestMetricsServerServesTheCounters(t *testing.T) {
	var buf bytes.Buffer
	o, err := New(Config{Service: "identity", MetricsAddr: "127.0.0.1:0", LogOutput: &buf})
	require.NoError(t, err)
	t.Cleanup(func() { _ = o.Close(context.Background()) })
	require.NotEmpty(t, o.MetricsAddr())

	_, _ = o.Unary()(context.Background(), "req", info, okHandler)

	resp, err := http.Get("http://" + o.MetricsAddr() + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `grpc_requests_total{code="OK",method="`+method+`",service="identity"} 1`)
	assert.Contains(t, string(body), `grpc_request_duration_seconds_count{method="`+method+`",service="identity"} 1`)
}

func TestMetricsServerFailsOnABadAddress(t *testing.T) {
	_, err := New(Config{Service: "identity", MetricsAddr: "not-an-address", LogOutput: io.Discard})
	assert.Error(t, err)
}

func TestInitReadsMetricsAddrFromTheEnvironment(t *testing.T) {
	t.Setenv("METRICS_ADDR", "127.0.0.1:0")
	o, err := Init("authorization")
	require.NoError(t, err)
	t.Cleanup(func() { _ = o.Close(context.Background()) })
	assert.NotEmpty(t, o.MetricsAddr())
}

func TestInitWithoutMetricsAddrServesNothing(t *testing.T) {
	t.Setenv("METRICS_ADDR", "")
	o, err := Init("authorization")
	require.NoError(t, err)
	t.Cleanup(func() { _ = o.Close(context.Background()) })
	assert.Equal(t, "", o.MetricsAddr())
}

func TestTwoObsDoNotShareCounters(t *testing.T) {
	var a, b bytes.Buffer
	first := newTest(t, &a)
	second := newTest(t, &b)

	_, _ = first.Unary()(context.Background(), "req", info, okHandler)

	assert.Equal(t, 0.0, testutil.ToFloat64(second.requests.WithLabelValues("identity", method, "OK")))
}

// sampleCount reads how many durations the histogram holds.
func sampleCount(t *testing.T, o *Obs) uint64 {
	t.Helper()
	families, err := o.registry.Gather()
	require.NoError(t, err)
	for _, f := range families {
		if f.GetName() == "grpc_request_duration_seconds" {
			return f.GetMetric()[0].GetHistogram().GetSampleCount()
		}
	}
	t.Fatal("no grpc_request_duration_seconds metric")
	return 0
}
