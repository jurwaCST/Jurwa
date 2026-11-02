// Package rulesclient is the RuleLookup adapter: it calls routing-rules GetRule
// over gRPC and turns gRPC status codes into domain errors.
package rulesclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	routingrulesv1 "github.com/jurwaCST/Jurwa/proto/gen/routing-rules/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// DefaultAddr is used when ROUTING_RULES_ADDR is not set.
const DefaultAddr = "localhost:9006"

// Client implements ports.RuleLookup.
type Client struct {
	c routingrulesv1.RoutingRulesServiceClient
}

func New(c routingrulesv1.RoutingRulesServiceClient) *Client {
	return &Client{c: c}
}

func (c *Client) GetRule(ctx context.Context, category domain.Category, building string) (domain.Rule, error) {
	resp, err := c.c.GetRule(ctx, &routingrulesv1.GetRuleRequest{
		Category: commonv1.Category(commonv1.Category_value[string(category)]),
		Location: &commonv1.Location{Building: building},
	})
	if err != nil {
		return domain.Rule{}, toDomainError(err)
	}
	return domain.Rule{ID: resp.GetRule().GetId(), TargetPosition: resp.GetRule().GetTargetPosition()}, nil
}

// toDomainError keeps "the rules service is down" apart from "no rule", so the
// caller never gets a wrong answer when routing-rules cannot be reached.
func toDomainError(err error) error {
	st := status.Convert(err)
	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %s", domain.ErrRuleNotFound, st.Message())
	case codes.InvalidArgument:
		return fmt.Errorf("%w: %s", domain.ErrInvalidCategory, st.Message())
	case codes.Unavailable, codes.DeadlineExceeded:
		return fmt.Errorf("%w: routing-rules: %s", domain.ErrUnavailable, st.Message())
	default:
		return fmt.Errorf("routing-rules GetRule: %w", err)
	}
}
