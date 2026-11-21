// Package directoryclient is the Positions adapter (SCRUM-162): it asks
// department-directory which post a user holds, so a message can say who has
// the report, for example "the Block Councillor (HC)".
package directoryclient

import (
	"context"
	"fmt"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
)

// DefaultAddr is used when DEPARTMENT_DIRECTORY_ADDR is not set.
const DefaultAddr = "localhost:9008"

// Client implements ports.Positions.
type Client struct {
	c departmentdirectoryv1.DepartmentDirectoryServiceClient
}

func New(c departmentdirectoryv1.DepartmentDirectoryServiceClient) *Client {
	return &Client{c: c}
}

// TitleOf returns the title of the post userID holds, or "" when none.
func (c *Client) TitleOf(ctx context.Context, userID string) (string, error) {
	resp, err := c.c.ListPositions(ctx, &departmentdirectoryv1.ListPositionsRequest{HeldOnly: true})
	if err != nil {
		return "", fmt.Errorf("department-directory ListPositions: %w", err)
	}
	for _, p := range resp.GetPositions() {
		if p.GetHolderUserId() == userID && p.GetActive() {
			return p.GetTitle(), nil
		}
	}
	return "", nil
}
