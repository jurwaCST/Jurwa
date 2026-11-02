package logsender

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/notification/domain"
)

func TestSendWritesOneJSONLine(t *testing.T) {
	var buf bytes.Buffer
	s := New(slog.New(slog.NewJSONHandler(&buf, nil)))

	err := s.Send(context.Background(), domain.Notification{
		ID:              "11111111-1111-1111-1111-111111111111",
		RecipientUserID: "22222222-2222-2222-2222-222222222222",
		Channel:         domain.InApp,
		EventType:       domain.EventStatusChanged,
		Subject:         "Report updated",
		Body:            "Now In Progress",
	})
	require.NoError(t, err)

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "notification sent", line["msg"])
	assert.Equal(t, "IN_APP", line["channel"])
	assert.Equal(t, "status.changed", line["event_type"])
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", line["recipient_user_id"])
	assert.NotContains(t, buf.String(), "Now In Progress", "the body is not written to the log")
}
