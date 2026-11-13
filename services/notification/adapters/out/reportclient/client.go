// Package reportclient is the Reports adapter: it calls report-submission
// GetReport over gRPC to find who sent a report and where it is (SCRUM-155).
package reportclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/notification/ports"
)

// DefaultAddr is used when REPORT_SUBMISSION_ADDR is not set.
const DefaultAddr = "localhost:9001"

// Client implements ports.Reports.
type Client struct {
	c reportsubmissionv1.ReportSubmissionServiceClient
}

func New(c reportsubmissionv1.ReportSubmissionServiceClient) *Client {
	return &Client{c: c}
}

// GetReport returns the report's student and place. A report that does not
// exist (or a bad id) is ErrReportNotFound, which trying again cannot fix.
// Every other error is ErrReportsUnavailable, so the event is tried again.
func (c *Client) GetReport(ctx context.Context, reportID string) (ports.Report, error) {
	resp, err := c.c.GetReport(ctx, &reportsubmissionv1.GetReportRequest{ReportId: reportID})
	if err != nil {
		st := status.Convert(err)
		switch st.Code() {
		case codes.NotFound, codes.InvalidArgument:
			return ports.Report{}, fmt.Errorf("%w: %s", ports.ErrReportNotFound, st.Message())
		default:
			return ports.Report{}, fmt.Errorf("%w: %s", ports.ErrReportsUnavailable, st.Message())
		}
	}
	r := resp.GetReport()
	return ports.Report{
		ID:         r.GetId(),
		ReporterID: r.GetReporterId(),
		Building:   r.GetLocation().GetBuilding(),
		Wing:       r.GetLocation().GetWing(),
		Area:       r.GetLocation().GetArea(),
	}, nil
}
