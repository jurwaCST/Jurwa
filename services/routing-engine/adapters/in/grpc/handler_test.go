package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	routingenginev1 "github.com/jurwaCST/Jurwa/proto/gen/routing-engine/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/application"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
	"github.com/jurwaCST/Jurwa/services/routing-engine/ports"
)

// staticRules routes by the agreed defaults, or fails with err.
type staticRules struct{ err error }

func (s staticRules) GetRule(_ context.Context, c domain.Category, _ string) (domain.Rule, error) {
	if s.err != nil {
		return domain.Rule{}, s.err
	}
	pos := map[domain.Category]string{domain.Hostel: "DSA", domain.AcademicBlock: "DAA", domain.Others: "SSO"}[c]
	return domain.Rule{ID: "rule-" + pos, TargetPosition: pos}, nil
}

type staticHolders map[string]string

func (s staticHolders) GetHolder(_ context.Context, p string) (string, error) { return s[p], nil }

type nopPublisher struct{}

func (nopPublisher) PublishReportRouted(context.Context, ports.ReportRouted) error { return nil }

func newHandler(rules staticRules, holders staticHolders) *Handler {
	return NewHandler(application.NewService(rules, holders, nopPublisher{}))
}

var allHolders = staticHolders{"DSA": "user-dsa", "DAA": "user-daa", "SSO": "user-sso"}

func TestRouteReport(t *testing.T) {
	tests := []struct {
		name     string
		category commonv1.Category
		building string
		wantPos  string
		wantHead string
	}{
		{name: "hostel", category: commonv1.Category_HOSTEL, building: "Hostel B", wantPos: "DSA", wantHead: "user-dsa"},
		{name: "academic block", category: commonv1.Category_ACADEMIC_BLOCK, building: "Block 2", wantPos: "DAA", wantHead: "user-daa"},
		{name: "others", category: commonv1.Category_OTHERS, building: "Gate", wantPos: "SSO", wantHead: "user-sso"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := newHandler(staticRules{}, allHolders).RouteReport(context.Background(), &routingenginev1.RouteReportRequest{
				ReportId: "r1", Category: tt.category, Location: &commonv1.Location{Building: tt.building},
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantPos, resp.GetTargetPosition())
			assert.Equal(t, tt.wantHead, resp.GetCategoryHeadId())
			assert.Equal(t, "rule-"+tt.wantPos, resp.GetRuleId())
		})
	}
}

func TestRouteReportErrorCodes(t *testing.T) {
	hostelB := &commonv1.Location{Building: "Hostel B"}
	tests := []struct {
		name    string
		req     *routingenginev1.RouteReportRequest
		rules   staticRules
		holders staticHolders
		want    codes.Code
	}{
		{name: "unspecified category", req: &routingenginev1.RouteReportRequest{Location: hostelB}, holders: allHolders, want: codes.InvalidArgument},
		{name: "unknown category number", req: &routingenginev1.RouteReportRequest{Category: commonv1.Category(99), Location: hostelB}, holders: allHolders, want: codes.InvalidArgument},
		{name: "no location", req: &routingenginev1.RouteReportRequest{Category: commonv1.Category_HOSTEL}, holders: allHolders, want: codes.InvalidArgument},
		{name: "no rule", req: &routingenginev1.RouteReportRequest{Category: commonv1.Category_HOSTEL, Location: hostelB}, rules: staticRules{err: domain.ErrRuleNotFound}, holders: allHolders, want: codes.NotFound},
		{name: "rules down", req: &routingenginev1.RouteReportRequest{Category: commonv1.Category_HOSTEL, Location: hostelB}, rules: staticRules{err: domain.ErrUnavailable}, holders: allHolders, want: codes.Unavailable},
		{name: "no holder", req: &routingenginev1.RouteReportRequest{Category: commonv1.Category_HOSTEL, Location: hostelB}, holders: staticHolders{}, want: codes.FailedPrecondition},
		{name: "unexpected error", req: &routingenginev1.RouteReportRequest{Category: commonv1.Category_HOSTEL, Location: hostelB}, rules: staticRules{err: errors.New("boom")}, holders: allHolders, want: codes.Internal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newHandler(tt.rules, tt.holders).RouteReport(context.Background(), tt.req)
			assert.Equal(t, tt.want, status.Code(err))
		})
	}
}
