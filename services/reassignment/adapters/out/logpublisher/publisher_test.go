package logpublisher

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/reassignment/ports"
)

func TestPublishReportReassigned_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	p := New(slog.New(slog.NewJSONHandler(&buf, nil)))

	err := p.PublishReportReassigned(context.Background(), ports.ReportReassigned{
		ReportID: "r1", FromUserID: "u1", ToUserID: "u2", Reason: "Plumbing",
		AssignedAt: time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"event":"report.reassigned"`)
	assert.Contains(t, buf.String(), `"to_user_id":"u2"`)
	assert.Contains(t, buf.String(), `"assigned_at":"2026-11-02T09:00:00Z"`)
}
