package rulesclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// fakeRulesClient answers GetRule with a fixed response or error. The embedded
// interface covers the methods routing-engine never calls.
type fakeRulesClient struct {
	routingrulesv1.RoutingRulesServiceClient
	got  *routingrulesv1.GetRuleRequest
	resp *routingrulesv1.GetRuleResponse
	err  error
}

func (f *fakeRulesClient) GetRule(_ context.Context, in *routingrulesv1.GetRuleRequest, _ ...grpc.CallOption) (*routingrulesv1.GetRuleResponse, error) {
	f.got = in
	return f.resp, f.err
}

func TestGetRule(t *testing.T) {
	fake := &fakeRulesClient{resp: &routingrulesv1.GetRuleResponse{Rule: &routingrulesv1.Rule{Id: "rule-1", TargetPosition: "DSA"}}}
	got, err := New(fake).GetRule(context.Background(), domain.Hostel, "Hostel B")
	require.NoError(t, err)
	assert.Equal(t, domain.Rule{ID: "rule-1", TargetPosition: "DSA"}, got)
	assert.Equal(t, commonv1.Category_HOSTEL, fake.got.GetCategory())
	assert.Equal(t, "Hostel B", fake.got.GetLocation().GetBuilding())
}

func TestGetRuleErrors(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want error
	}{
		{name: "not found", code: codes.NotFound, want: domain.ErrRuleNotFound},
		{name: "invalid", code: codes.InvalidArgument, want: domain.ErrInvalidCategory},
		{name: "down", code: codes.Unavailable, want: domain.ErrUnavailable},
		{name: "too slow", code: codes.DeadlineExceeded, want: domain.ErrUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&fakeRulesClient{err: status.Error(tt.code, "x")}).GetRule(context.Background(), domain.Hostel, "B")
			assert.ErrorIs(t, err, tt.want)
		})
	}

	_, err := New(&fakeRulesClient{err: status.Error(codes.Internal, "boom")}).GetRule(context.Background(), domain.Hostel, "B")
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrUnavailable)
	assert.NotErrorIs(t, err, domain.ErrRuleNotFound)
}
