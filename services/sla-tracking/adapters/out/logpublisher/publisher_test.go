package logpublisher

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

func TestPublishSLALapsed_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	p := New(slog.New(slog.NewJSONHandler(&buf, nil)))
	deadline := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)

	err := p.PublishSLALapsed(context.Background(), ports.SLALapsed{
		ReportID: "r1", Category: domain.Hostel, Deadline: deadline, LapsedAt: deadline.Add(time.Minute),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"event":"sla.lapsed"`)
	assert.Contains(t, buf.String(), `"category":"HOSTEL"`)
	assert.Contains(t, buf.String(), `"deadline":"2026-11-03T09:00:00Z"`)
}
