package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Errors for bad input. The gRPC adapter maps them to InvalidArgument.
var (
	ErrInvalidRecipient = errors.New("recipient_user_id must be a UUID")
	ErrInvalidChannel   = errors.New("channel must be EMAIL or IN_APP")
	ErrInvalidEventType = errors.New("unknown event_type")
	ErrInvalidReportID  = errors.New("report_id must be a UUID")
	ErrInvalidSubject   = errors.New("subject is required")
	ErrInvalidBody      = errors.New("body is required and at most 200 characters")
)

// MaxBodyChars is the longest body accepted, counted in characters (runes).
const MaxBodyChars = 200

// Channel matches jurwa.notification.v1.Channel.
type Channel int32

const (
	ChannelUnspecified Channel = 0
	Email              Channel = 1
	InApp              Channel = 2
)

func (c Channel) String() string {
	switch c {
	case Email:
		return "EMAIL"
	case InApp:
		return "IN_APP"
	default:
		return "CHANNEL_UNSPECIFIED"
	}
}

// DeliveryStatus matches jurwa.notification.v1.DeliveryStatus.
type DeliveryStatus int32

const (
	StatusUnspecified DeliveryStatus = 0
	Pending           DeliveryStatus = 1
	Sent              DeliveryStatus = 2
	Failed            DeliveryStatus = 3
)

func (s DeliveryStatus) String() string {
	switch s {
	case Pending:
		return "PENDING"
	case Sent:
		return "SENT"
	case Failed:
		return "FAILED"
	default:
		return "DELIVERY_STATUS_UNSPECIFIED"
	}
}

// Event types that can cause a notification. They are the stream names in
// proto/events/v1/events.proto, plus the daily digest.
const (
	EventReportSubmitted  = "report.submitted"  // student: report received
	EventReportRouted     = "report.routed"     // category head: report assigned
	EventStatusChanged    = "status.changed"    // student: status changed
	EventReportReassigned = "report.reassigned" // new assignee
	EventSLALapsed        = "sla.lapsed"        // escalated
	EventDigest           = "digest.daily"      // digest-dispatcher
)

var eventTypes = map[string]bool{
	EventReportSubmitted:  true,
	EventReportRouted:     true,
	EventStatusChanged:    true,
	EventReportReassigned: true,
	EventSLALapsed:        true,
	EventDigest:           true,
}

// Input is what a caller asks to send.
type Input struct {
	RecipientUserID string
	Channel         Channel
	EventType       string
	ReportID        string // optional, "" when not about one report
	Subject         string
	Body            string
}

// Notification is one message in the outbox.
type Notification struct {
	ID              string
	RecipientUserID string
	Channel         Channel
	EventType       string
	ReportID        string
	Subject         string
	Body            string
	Status          DeliveryStatus
	CreatedAt       time.Time
	SentAt          *time.Time
	ReadAt          *time.Time
}

// ValidateUserID checks that id is a UUID.
func ValidateUserID(id string) error {
	if uuid.Validate(id) != nil {
		return ErrInvalidRecipient
	}
	return nil
}

// NewNotification checks the input and returns a pending notification.
func NewNotification(id string, in Input, now time.Time) (Notification, error) {
	if err := ValidateUserID(in.RecipientUserID); err != nil {
		return Notification{}, err
	}
	if in.Channel != Email && in.Channel != InApp {
		return Notification{}, ErrInvalidChannel
	}
	if !eventTypes[in.EventType] {
		return Notification{}, ErrInvalidEventType
	}
	if in.ReportID != "" && uuid.Validate(in.ReportID) != nil {
		return Notification{}, ErrInvalidReportID
	}
	subject := strings.TrimSpace(in.Subject)
	if subject == "" {
		return Notification{}, ErrInvalidSubject
	}
	body := strings.TrimSpace(in.Body)
	if body == "" || utf8.RuneCountInString(body) > MaxBodyChars {
		return Notification{}, ErrInvalidBody
	}
	return Notification{
		ID:              id,
		RecipientUserID: in.RecipientUserID,
		Channel:         in.Channel,
		EventType:       in.EventType,
		ReportID:        in.ReportID,
		Subject:         subject,
		Body:            body,
		Status:          Pending,
		CreatedAt:       now,
	}, nil
}

// MarkSent returns a copy marked as sent at the given time.
func (n Notification) MarkSent(at time.Time) Notification {
	n.Status = Sent
	n.SentAt = &at
	return n
}

// MarkFailed returns a copy marked as failed.
func (n Notification) MarkFailed() Notification {
	n.Status = Failed
	return n
}
