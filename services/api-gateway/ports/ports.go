// Package ports holds the interfaces the gateway uses to reach other services.
// The gateway has no business logic, so the ports use the generated proto
// messages directly. adapters/out/grpcclients implements them over gRPC.
package ports

import (
	"context"

	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	feedbackv1 "github.com/jurwaCST/Jurwa/proto/gen/feedback/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
	reassignmentv1 "github.com/jurwaCST/Jurwa/proto/gen/reassignment/v1"
	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

// Identity is the identity service (Gayley).
type Identity interface {
	Register(ctx context.Context, in *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error)
	Login(ctx context.Context, in *identityv1.LoginRequest) (*identityv1.LoginResponse, error)
	// ValidateToken returns the user and roles for a JWT, or an Unauthenticated error.
	ValidateToken(ctx context.Context, token string) (*identityv1.User, error)
}

// Authorizer is the authorization service CheckPermission (Gayley).
type Authorizer interface {
	CheckPermission(ctx context.Context, role, action, resource string) (bool, error)
}

// Categories is category-taxonomy.
type Categories interface {
	ListCategories(ctx context.Context) (*categorytaxonomyv1.ListCategoriesResponse, error)
}

// Locations is location-taxonomy (Choki).
type Locations interface {
	ListLocations(ctx context.Context, category commonv1.Category) (*locationtaxonomyv1.ListLocationsResponse, error)
}

// Reports is report-submission (Choki).
type Reports interface {
	SubmitReport(ctx context.Context, in *reportsubmissionv1.SubmitReportRequest) (*reportsubmissionv1.SubmitReportResponse, error)
	GetReport(ctx context.Context, reportID string) (*commonv1.Report, error)
}

// Notes is internal-notes (Choki). role is the caller's role, sent as the gRPC
// metadata x-user-role because the messages have no role field.
type Notes interface {
	AddNote(ctx context.Context, role string, in *internalnotesv1.AddNoteRequest) (*internalnotesv1.AddNoteResponse, error)
	ListNotes(ctx context.Context, role, reportID string) (*internalnotesv1.ListNotesResponse, error)
}

// History is report-history (Choki): the "My reports" list and report timelines.
type History interface {
	ListReportsByReporter(ctx context.Context, reporterID string, pageSize int32, pageToken string) (*reporthistoryv1.ListReportsByReporterResponse, error)
	ListReportsByAssignee(ctx context.Context, assigneeID string, pageSize int32, pageToken string) (*reporthistoryv1.ListReportsByAssigneeResponse, error)
	GetReportTimeline(ctx context.Context, reportID string) (*reporthistoryv1.GetReportTimelineResponse, error)
}

// Photos is photo-media (Choki): the photo a student attaches to a report.
type Photos interface {
	UploadPhoto(ctx context.Context, in *photomediav1.UploadPhotoRequest) (*photomediav1.UploadPhotoResponse, error)
	GetPhoto(ctx context.Context, photoID string) (*photomediav1.GetPhotoResponse, error)
}

// Status is status-lifecycle (Nyendrak): it owns each report's status and the
// allowed flow. The gateway only sets the report id and the actor.
type Status interface {
	ChangeStatus(ctx context.Context, in *statuslifecyclev1.ChangeStatusRequest) (*statuslifecyclev1.ChangeStatusResponse, error)
	GetStatus(ctx context.Context, reportID string) (*statuslifecyclev1.GetStatusResponse, error)
}

// Analytics is the analytics service (Gayley): dashboard counts. A nil period
// means all time.
type Analytics interface {
	GetSummary(ctx context.Context, period *analyticsv1.Period) (*analyticsv1.GetSummaryResponse, error)
	GetCategoryBreakdown(ctx context.Context, period *analyticsv1.Period) (*analyticsv1.GetCategoryBreakdownResponse, error)
}

// Reassignment is the reassignment service (Nyendrak). role is the caller's
// role, sent as the gRPC metadata x-user-role because the request has no role field.
type Reassignment interface {
	ReassignReport(ctx context.Context, role string, in *reassignmentv1.ReassignReportRequest) (*reassignmentv1.ReassignReportResponse, error)
}

// Directory is department-directory: the positions a report can be reassigned to.
type Directory interface {
	// ListHeldPositions returns the positions that have a holder, ordered by code.
	ListHeldPositions(ctx context.Context) (*departmentdirectoryv1.ListPositionsResponse, error)
}

// Notifications is the notification service (Gayley): the in-app inbox.
type Notifications interface {
	ListNotifications(ctx context.Context, in *notificationv1.ListNotificationsRequest) (*notificationv1.ListNotificationsResponse, error)
}

// Feedback is the feedback service (Gayley): one rating per resolved report.
// It does not check that the student is the reporter; the gateway does.
type Feedback interface {
	SubmitFeedback(ctx context.Context, in *feedbackv1.SubmitFeedbackRequest) (*feedbackv1.SubmitFeedbackResponse, error)
	GetFeedback(ctx context.Context, reportID string) (*feedbackv1.GetFeedbackResponse, error)
}
