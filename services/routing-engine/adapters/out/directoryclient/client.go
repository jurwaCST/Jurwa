// Package directoryclient is the HolderLookup adapter: it calls
// department-directory GetStaff over gRPC.
package directoryclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	"github.com/jurwaCST/Jurwa/services/routing-engine/domain"
)

// DefaultAddr is used when DEPARTMENT_DIRECTORY_ADDR is not set.
const DefaultAddr = "localhost:9008"

// Client implements ports.HolderLookup.
type Client struct {
	c departmentdirectoryv1.DepartmentDirectoryServiceClient
}

func New(c departmentdirectoryv1.DepartmentDirectoryServiceClient) *Client {
	return &Client{c: c}
}

// GetHolder returns the user id holding the position, or "" when it is vacant
// or inactive.
func (c *Client) GetHolder(ctx context.Context, positionCode string) (string, error) {
	resp, err := c.c.GetStaff(ctx, &departmentdirectoryv1.GetStaffRequest{PositionCode: positionCode})
	if err != nil {
		st := status.Convert(err)
		switch st.Code() {
		case codes.NotFound, codes.InvalidArgument:
			// The rule points at a position the directory does not know.
			return "", fmt.Errorf("%w: %s", domain.ErrNoHolder, st.Message())
		case codes.Unavailable, codes.DeadlineExceeded:
			return "", fmt.Errorf("%w: department-directory: %s", domain.ErrUnavailable, st.Message())
		default:
			return "", fmt.Errorf("department-directory GetStaff: %w", err)
		}
	}
	if !resp.GetPosition().GetActive() {
		return "", nil
	}
	return resp.GetPosition().GetHolderUserId(), nil
}
