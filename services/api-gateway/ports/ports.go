// Package ports holds the interfaces the gateway uses to reach other services.
// The gateway has no business logic, so the ports use the generated proto
// messages directly. adapters/out/grpcclients implements them over gRPC.
package ports

import (
	"context"

	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	internalnotesv1 "github.com/jurwaCST/Jurwa/proto/gen/internal-notes/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
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
