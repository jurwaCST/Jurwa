// Package directoryclient is the HolderLookup adapter (SCRUM-158): it asks
// department-directory over gRPC who holds the DSA or DAA position, so a late
// report goes up to them.
package directoryclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
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

// GetHolder returns the user id holding the position, or "" when it is vacant,
// inactive or unknown. Any other failure is an error, so the escalation is tried
// again later instead of going to the wrong person.
func (c *Client) GetHolder(ctx context.Context, positionCode string) (string, error) {
	resp, err := c.c.GetStaff(ctx, &departmentdirectoryv1.GetStaffRequest{PositionCode: positionCode})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound, codes.InvalidArgument:
			return "", nil
		default:
			return "", fmt.Errorf("department-directory GetStaff %s: %w", positionCode, err)
		}
	}
	if !resp.GetPosition().GetActive() {
		return "", nil
	}
	return resp.GetPosition().GetHolderUserId(), nil
}
