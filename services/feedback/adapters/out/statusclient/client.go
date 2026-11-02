// Package statusclient is the StatusChecker adapter: it calls
// status-lifecycle GetStatus over gRPC and says whether a report is RESOLVED.
package statusclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	"github.com/jurwaCST/Jurwa/services/feedback/domain"
)

// DefaultAddr is used when STATUS_LIFECYCLE_ADDR is not set.
const DefaultAddr = "localhost:9011"

// Client implements ports.StatusChecker.
type Client struct {
	c statuslifecyclev1.StatusLifecycleServiceClient
}

func New(c statuslifecyclev1.StatusLifecycleServiceClient) *Client {
	return &Client{c: c}
}

// IsResolved is true only when the report's status is RESOLVED. A report
// status-lifecycle does not know is not resolved. "status-lifecycle is down"
// is kept apart from "not resolved", so a student is never told the wrong thing.
func (c *Client) IsResolved(ctx context.Context, reportID string) (bool, error) {
	resp, err := c.c.GetStatus(ctx, &statuslifecyclev1.GetStatusRequest{ReportId: reportID})
	if err != nil {
		st := status.Convert(err)
		switch st.Code() {
		case codes.NotFound:
			return false, nil
		case codes.Unavailable, codes.DeadlineExceeded:
			return false, fmt.Errorf("%w: %s", domain.ErrUnavailable, st.Message())
		default:
			return false, fmt.Errorf("status-lifecycle GetStatus: %w", err)
		}
	}
	return resp.GetStatus().GetStatus() == commonv1.Status_RESOLVED, nil
}
