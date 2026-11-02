package grpcclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
)

// Fakes of the generated clients. Each embeds the interface for methods not under test.

type fakeIdentityClient struct {
	identityv1.IdentityServiceClient
	token string
}

func (f *fakeIdentityClient) Register(_ context.Context, in *identityv1.RegisterRequest, _ ...grpc.CallOption) (*identityv1.RegisterResponse, error) {
	return &identityv1.RegisterResponse{User: &identityv1.User{Email: in.GetEmail()}}, nil
}

func (f *fakeIdentityClient) Login(_ context.Context, _ *identityv1.LoginRequest, _ ...grpc.CallOption) (*identityv1.LoginResponse, error) {
	return &identityv1.LoginResponse{AccessToken: "t"}, nil
}

func (f *fakeIdentityClient) ValidateToken(_ context.Context, in *identityv1.ValidateTokenRequest, _ ...grpc.CallOption) (*identityv1.ValidateTokenResponse, error) {
	f.token = in.GetAccessToken()
	if f.token != "good" {
		return nil, status.Error(codes.Unauthenticated, "bad token")
	}
	return &identityv1.ValidateTokenResponse{User: &identityv1.User{Id: "u1"}}, nil
}

type fakeCategoryClient struct {
	categorytaxonomyv1.CategoryTaxonomyServiceClient
}

func (fakeCategoryClient) ListCategories(context.Context, *categorytaxonomyv1.ListCategoriesRequest, ...grpc.CallOption) (*categorytaxonomyv1.ListCategoriesResponse, error) {
	return &categorytaxonomyv1.ListCategoriesResponse{Categories: make([]*categorytaxonomyv1.CategoryInfo, 3)}, nil
}

type fakeLocationClient struct {
	locationtaxonomyv1.LocationTaxonomyServiceClient
	got commonv1.Category
}

func (f *fakeLocationClient) ListLocations(_ context.Context, in *locationtaxonomyv1.ListLocationsRequest, _ ...grpc.CallOption) (*locationtaxonomyv1.ListLocationsResponse, error) {
	f.got = in.GetCategory()
	return &locationtaxonomyv1.ListLocationsResponse{}, nil
}

type fakeReportClient struct {
	reportsubmissionv1.ReportSubmissionServiceClient
	getErr error
}

func (f *fakeReportClient) SubmitReport(_ context.Context, in *reportsubmissionv1.SubmitReportRequest, _ ...grpc.CallOption) (*reportsubmissionv1.SubmitReportResponse, error) {
	return &reportsubmissionv1.SubmitReportResponse{Report: &commonv1.Report{ReporterId: in.GetReporterId()}}, nil
}

func (f *fakeReportClient) GetReport(_ context.Context, in *reportsubmissionv1.GetReportRequest, _ ...grpc.CallOption) (*reportsubmissionv1.GetReportResponse, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &reportsubmissionv1.GetReportResponse{Report: &commonv1.Report{Id: in.GetReportId()}}, nil
}

func TestIdentity(t *testing.T) {
	ctx := context.Background()
	fake := &fakeIdentityClient{}
	id := Identity{C: fake}

	u, err := id.ValidateToken(ctx, "good")
	require.NoError(t, err)
	assert.Equal(t, "u1", u.GetId())
	assert.Equal(t, "good", fake.token)

	_, err = id.ValidateToken(ctx, "bad")
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	reg, err := id.Register(ctx, &identityv1.RegisterRequest{Email: "a@rub.edu.bt"})
	require.NoError(t, err)
	assert.Equal(t, "a@rub.edu.bt", reg.GetUser().GetEmail())
	login, err := id.Login(ctx, &identityv1.LoginRequest{})
	require.NoError(t, err)
	assert.Equal(t, "t", login.GetAccessToken())
}

func TestCategoriesAndLocations(t *testing.T) {
	ctx := context.Background()
	cats, err := Categories{C: fakeCategoryClient{}}.ListCategories(ctx)
	require.NoError(t, err)
	assert.Len(t, cats.GetCategories(), 3)

	fake := &fakeLocationClient{}
	_, err = Locations{C: fake}.ListLocations(ctx, commonv1.Category_OTHERS)
	require.NoError(t, err)
	assert.Equal(t, commonv1.Category_OTHERS, fake.got)
}

func TestReports(t *testing.T) {
	ctx := context.Background()
	fake := &fakeReportClient{}
	r := Reports{C: fake}

	sub, err := r.SubmitReport(ctx, &reportsubmissionv1.SubmitReportRequest{ReporterId: "u1"})
	require.NoError(t, err)
	assert.Equal(t, "u1", sub.GetReport().GetReporterId())

	got, err := r.GetReport(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, "r1", got.GetId())

	fake.getErr = status.Error(codes.NotFound, "no report")
	_, err = r.GetReport(ctx, "r2")
	assert.Equal(t, codes.NotFound, status.Code(err))
}
