// Package grpcclients implements the gateway ports with the generated gRPC
// clients. cmd/server wraps every connection with pkg/resilience.
package grpcclients

import (
	"context"

	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
)

// Identity implements ports.Identity.
type Identity struct {
	C identityv1.IdentityServiceClient
}

func (i Identity) Register(ctx context.Context, in *identityv1.RegisterRequest) (*identityv1.RegisterResponse, error) {
	return i.C.Register(ctx, in)
}

func (i Identity) Login(ctx context.Context, in *identityv1.LoginRequest) (*identityv1.LoginResponse, error) {
	return i.C.Login(ctx, in)
}

func (i Identity) ValidateToken(ctx context.Context, token string) (*identityv1.User, error) {
	resp, err := i.C.ValidateToken(ctx, &identityv1.ValidateTokenRequest{AccessToken: token})
	if err != nil {
		return nil, err
	}
	return resp.GetUser(), nil
}

// Categories implements ports.Categories.
type Categories struct {
	C categorytaxonomyv1.CategoryTaxonomyServiceClient
}

func (c Categories) ListCategories(ctx context.Context) (*categorytaxonomyv1.ListCategoriesResponse, error) {
	return c.C.ListCategories(ctx, &categorytaxonomyv1.ListCategoriesRequest{})
}

// Locations implements ports.Locations.
type Locations struct {
	C locationtaxonomyv1.LocationTaxonomyServiceClient
}

func (l Locations) ListLocations(ctx context.Context, category commonv1.Category) (*locationtaxonomyv1.ListLocationsResponse, error) {
	return l.C.ListLocations(ctx, &locationtaxonomyv1.ListLocationsRequest{Category: category})
}

// Reports implements ports.Reports.
type Reports struct {
	C reportsubmissionv1.ReportSubmissionServiceClient
}

func (r Reports) SubmitReport(ctx context.Context, in *reportsubmissionv1.SubmitReportRequest) (*reportsubmissionv1.SubmitReportResponse, error) {
	return r.C.SubmitReport(ctx, in)
}

func (r Reports) GetReport(ctx context.Context, reportID string) (*commonv1.Report, error) {
	resp, err := r.C.GetReport(ctx, &reportsubmissionv1.GetReportRequest{ReportId: reportID})
	if err != nil {
		return nil, err
	}
	return resp.GetReport(), nil
}
