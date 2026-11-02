package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

const (
	notifID  = "11111111-1111-1111-1111-111111111111"
	userID   = "22222222-2222-2222-2222-222222222222"
	reportID = "33333333-3333-3333-3333-333333333333"
)

func validInput() Input {
	return Input{
		RecipientUserID: userID,
		Channel:         InApp,
		EventType:       EventStatusChanged,
		ReportID:        reportID,
		Subject:         "Report updated",
		Body:            "Your report is now In Progress.",
	}
}

func TestNewNotificationValid(t *testing.T) {
	n, err := NewNotification(notifID, validInput(), now)
	require.NoError(t, err)
	assert.Equal(t, notifID, n.ID)
	assert.Equal(t, userID, n.RecipientUserID)
	assert.Equal(t, InApp, n.Channel)
	assert.Equal(t, EventStatusChanged, n.EventType)
	assert.Equal(t, reportID, n.ReportID)
	assert.Equal(t, Pending, n.Status, "a new notification is pending until it is sent")
	assert.Equal(t, now, n.CreatedAt)
	assert.Nil(t, n.SentAt)
}

func TestNewNotificationRules(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*Input)
		wantErr error
	}{
		// Recipient
		{"recipient missing", func(in *Input) { in.RecipientUserID = "" }, ErrInvalidRecipient},
		{"recipient not a UUID", func(in *Input) { in.RecipientUserID = "gayley" }, ErrInvalidRecipient},
		// Channel
		{"channel unspecified", func(in *Input) { in.Channel = ChannelUnspecified }, ErrInvalidChannel},
		{"unknown channel", func(in *Input) { in.Channel = Channel(9) }, ErrInvalidChannel},
		{"email channel", func(in *Input) { in.Channel = Email }, nil},
		// Event type: the four from the guide plus reassigned and digest
		{"report submitted", func(in *Input) { in.EventType = EventReportSubmitted }, nil},
		{"assigned (report routed)", func(in *Input) { in.EventType = EventReportRouted }, nil},
		{"status changed", func(in *Input) { in.EventType = EventStatusChanged }, nil},
		{"escalated (sla lapsed)", func(in *Input) { in.EventType = EventSLALapsed }, nil},
		{"reassigned", func(in *Input) { in.EventType = EventReportReassigned }, nil},
		{"daily digest", func(in *Input) { in.EventType = EventDigest }, nil},
		{"event type missing", func(in *Input) { in.EventType = "" }, ErrInvalidEventType},
		{"unknown event type", func(in *Input) { in.EventType = "report.deleted" }, ErrInvalidEventType},
		// Report id is optional, but must be a UUID when given
		{"no report id (digest)", func(in *Input) { in.ReportID = "" }, nil},
		{"report id not a UUID", func(in *Input) { in.ReportID = "abc" }, ErrInvalidReportID},
		// Subject
		{"subject missing", func(in *Input) { in.Subject = "" }, ErrInvalidSubject},
		{"subject blank", func(in *Input) { in.Subject = "   " }, ErrInvalidSubject},
		// Body (message) length boundaries
		{"body empty", func(in *Input) { in.Body = "" }, ErrInvalidBody},
		{"body blank", func(in *Input) { in.Body = "   " }, ErrInvalidBody},
		{"body 1 character", func(in *Input) { in.Body = "a" }, nil},
		{"body 199 characters", func(in *Input) { in.Body = strings.Repeat("a", 199) }, nil},
		{"body 200 characters (limit)", func(in *Input) { in.Body = strings.Repeat("a", 200) }, nil},
		{"body 201 characters", func(in *Input) { in.Body = strings.Repeat("a", 201) }, ErrInvalidBody},
		{"body 200 Dzongkha letters", func(in *Input) { in.Body = strings.Repeat("ཀ", 200) }, nil},
		{"body 201 Dzongkha letters", func(in *Input) { in.Body = strings.Repeat("ཀ", 201) }, ErrInvalidBody},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.change(&in)
			_, err := NewNotification(notifID, in, now)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNewNotificationTrimsSubjectAndBody(t *testing.T) {
	in := validInput()
	in.Subject = "  Report updated  "
	in.Body = "  In Progress  "
	n, err := NewNotification(notifID, in, now)
	require.NoError(t, err)
	assert.Equal(t, "Report updated", n.Subject)
	assert.Equal(t, "In Progress", n.Body)
}

func TestMarkSentAndFailed(t *testing.T) {
	n, err := NewNotification(notifID, validInput(), now)
	require.NoError(t, err)

	sentAt := now.Add(time.Second)
	sent := n.MarkSent(sentAt)
	assert.Equal(t, Sent, sent.Status)
	require.NotNil(t, sent.SentAt)
	assert.Equal(t, sentAt, *sent.SentAt)

	failed := n.MarkFailed()
	assert.Equal(t, Failed, failed.Status)
	assert.Nil(t, failed.SentAt, "a failed notification was never sent")

	assert.Equal(t, Pending, n.Status, "the original is not changed")
}

func TestValidateUserID(t *testing.T) {
	assert.NoError(t, ValidateUserID(userID))
	assert.ErrorIs(t, ValidateUserID(""), ErrInvalidRecipient)
	assert.ErrorIs(t, ValidateUserID("not-a-uuid"), ErrInvalidRecipient)
}

func TestChannelAndStatusString(t *testing.T) {
	assert.Equal(t, "EMAIL", Email.String())
	assert.Equal(t, "IN_APP", InApp.String())
	assert.Equal(t, "CHANNEL_UNSPECIFIED", Channel(9).String())
	assert.Equal(t, "PENDING", Pending.String())
	assert.Equal(t, "SENT", Sent.String())
	assert.Equal(t, "FAILED", Failed.String())
	assert.Equal(t, "DELIVERY_STATUS_UNSPECIFIED", DeliveryStatus(9).String())
}
