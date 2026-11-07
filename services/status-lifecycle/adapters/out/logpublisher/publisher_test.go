package logpublisher

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/status-lifecycle/domain"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

func TestPublishStatusChanged_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	p := New(slog.New(slog.NewJSONHandler(&buf, nil)))

	err := p.PublishStatusChanged(context.Background(), ports.StatusChanged{
		ReportID: "r1", From: domain.Open, To: domain.Assigned, Version: 2,
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"event":"status.changed"`)
	assert.Contains(t, buf.String(), `"to_status":"ASSIGNED"`)
}

func TestPublishEscalationCompensated_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	p := New(slog.New(slog.NewJSONHandler(&buf, nil)))

	err := p.PublishEscalationCompensated(context.Background(), ports.EscalationCompensated{
		ReportID: "r1", FailedStep: "status-lifecycle.set-escalated", UndoneStep: "reassignment.escalate",
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"event":"escalation.compensated"`)
	assert.Contains(t, buf.String(), `"undone_step":"reassignment.escalate"`)
}
