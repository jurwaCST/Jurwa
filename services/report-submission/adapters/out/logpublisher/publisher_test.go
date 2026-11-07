package logpublisher

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

func TestReportSubmittedWritesOneJSONLine(t *testing.T) {
	var buf bytes.Buffer
	pub := New(slog.New(slog.NewJSONHandler(&buf, nil)))

	err := pub.ReportSubmitted(context.Background(), domain.Report{
		ID: "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", Category: domain.CategoryHostel,
		Location: domain.Location{Building: "Hostel B"}, Status: domain.StatusOpen, CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "event published", line["msg"])
	assert.Equal(t, "report.submitted", line["event"])
	assert.Equal(t, "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f", line["report_id"])
	assert.Equal(t, "HOSTEL", line["category"])
	assert.Equal(t, "Hostel B", line["building"])
}
