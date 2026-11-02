// Package authzclient implements ports.Authorizer with authorization
// CheckPermission over gRPC (Gayley's service, port 9017).
package authzclient

import (
	"context"

	authorizationv1 "github.com/jurwaCST/Jurwa/proto/gen/authorization/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
)

// DefaultAddr is used when AUTHORIZATION_ADDR is not set.
const DefaultAddr = "localhost:9017"

// Client implements ports.Authorizer.
type Client struct {
	c authorizationv1.AuthorizationServiceClient
}

func New(c authorizationv1.AuthorizationServiceClient) *Client {
	return &Client{c: c}
}

// CheckPermission asks authorization whether role may do action on resource.
// A role the proto does not know is denied without a call. Errors are returned
// as they are, so the gateway refuses the request when authorization is down.
func (c *Client) CheckPermission(ctx context.Context, role, action, resource string) (bool, error) {
	r, ok := commonv1.Role_value[role]
	if !ok || r == int32(commonv1.Role_ROLE_UNSPECIFIED) {
		return false, nil
	}
	resp, err := c.c.CheckPermission(ctx, &authorizationv1.CheckPermissionRequest{
		Role:     commonv1.Role(r),
		Action:   action,
		Resource: resource,
	})
	if err != nil {
		return false, err
	}
	return resp.GetAllowed(), nil
}
