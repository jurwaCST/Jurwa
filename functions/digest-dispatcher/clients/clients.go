// Package clients holds the gRPC adapters for the dispatcher's three
// interfaces: analytics, department-directory and notification.
package clients

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/functions/digest-dispatcher/dispatcher"
	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
)

// Default addresses, used when the environment variables are not set.
const (
	DefaultAnalyticsAddr    = "localhost:9020"
	DefaultDirectoryAddr    = "localhost:9008"
	DefaultNotificationAddr = "localhost:9018"
)

// DigestEventType is the notification event_type for the digest.
const DigestEventType = "digest.daily"

// Analytics implements dispatcher.CountsSource with GetCategoryBreakdown.
type Analytics struct {
	c analyticsv1.AnalyticsServiceClient
}

func NewAnalytics(c analyticsv1.AnalyticsServiceClient) *Analytics { return &Analytics{c: c} }

// CategoryCounts asks for all reports (no period).
func (a *Analytics) CategoryCounts(ctx context.Context) (map[dispatcher.Category]dispatcher.Counts, error) {
	resp, err := a.c.GetCategoryBreakdown(ctx, &analyticsv1.GetCategoryBreakdownRequest{})
	if err != nil {
		return nil, fmt.Errorf("analytics GetCategoryBreakdown: %w", err)
	}
	out := make(map[dispatcher.Category]dispatcher.Counts, len(resp.GetCategories()))
	for _, row := range resp.GetCategories() {
		c := row.GetCounts()
		// dispatcher.Category uses the same numbers as commonv1.Category.
		out[dispatcher.Category(row.GetCategory())] = dispatcher.Counts{
			Total:      c.GetTotal(),
			Open:       c.GetOpen(),
			Assigned:   c.GetAssigned(),
			InProgress: c.GetInProgress(),
			Resolved:   c.GetResolved(),
			Escalated:  c.GetEscalated(),
		}
	}
	return out, nil
}

// Directory implements dispatcher.HeadDirectory with GetStaff.
type Directory struct {
	c departmentdirectoryv1.DepartmentDirectoryServiceClient
}

func NewDirectory(c departmentdirectoryv1.DepartmentDirectoryServiceClient) *Directory {
	return &Directory{c: c}
}

// HolderOf returns the user holding the position, or "" when nobody does.
func (d *Directory) HolderOf(ctx context.Context, position string) (string, error) {
	resp, err := d.c.GetStaff(ctx, &departmentdirectoryv1.GetStaffRequest{PositionCode: position})
	if err != nil {
		return "", fmt.Errorf("department-directory GetStaff %s: %w", position, err)
	}
	return resp.GetPosition().GetHolderUserId(), nil
}

// Notification implements dispatcher.Sender with SendNotification (in-app).
type Notification struct {
	c notificationv1.NotificationServiceClient
}

func NewNotification(c notificationv1.NotificationServiceClient) *Notification {
	return &Notification{c: c}
}

func (n *Notification) Send(ctx context.Context, userID, subject, body string) error {
	_, err := n.c.SendNotification(ctx, &notificationv1.SendNotificationRequest{
		RecipientUserId: userID,
		Channel:         notificationv1.Channel_IN_APP,
		EventType:       DigestEventType,
		Subject:         subject,
		Body:            body,
	})
	if err != nil {
		return fmt.Errorf("notification SendNotification: %w", err)
	}
	return nil
}
