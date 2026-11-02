package logpublisher

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

func TestPublishReportRouted_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	p := New(slog.New(slog.NewJSONHandler(&buf, nil)))

	err := p.PublishReportRouted(context.Background(), ports.ReportRouted{
		ReportID: "r1", Category: domain.Hostel, CategoryHeadID: "user-dsa", TargetPosition: "DSA", RuleID: "rule-1",
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"event":"report.routed"`)
	assert.Contains(t, buf.String(), `"category":"HOSTEL"`)
	assert.Contains(t, buf.String(), `"category_head_id":"user-dsa"`)
}
