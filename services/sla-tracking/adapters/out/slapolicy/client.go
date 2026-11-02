// Package slapolicy is the PolicyClient that calls the sla-policy service over gRPC.
package slapolicy

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	slapolicyv1 "github.com/jurwaCST/Jurwa/proto/gen/sla-policy/v1"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/domain"
	"github.com/jurwaCST/Jurwa/services/sla-tracking/ports"
)

// callTimeout bounds each GetPolicy call so StartTimer never hangs.
const callTimeout = 3 * time.Second

// Client wraps the generated sla-policy client.
type Client struct {
	c slapolicyv1.SlaPolicyServiceClient
}

// New returns a Client that uses c.
func New(c slapolicyv1.SlaPolicyServiceClient) *Client {
	return &Client{c: c}
}

// Window returns the SLA window for the category, or ports.ErrNoPolicy when
// sla-policy answers NOT_FOUND.
func (cl *Client) Window(ctx context.Context, category domain.Category) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	resp, err := cl.c.GetPolicy(ctx, &slapolicyv1.GetPolicyRequest{Category: commonv1.Category(category)})
	if status.Code(err) == codes.NotFound {
		return 0, ports.ErrNoPolicy
	}
	if err != nil {
		return 0, fmt.Errorf("sla-policy GetPolicy: %w", err)
	}
	return time.Duration(resp.GetPolicy().GetWindowHours()) * time.Hour, nil
}
