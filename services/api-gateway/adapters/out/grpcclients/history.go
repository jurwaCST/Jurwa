package grpcclients

import (
	"context"

	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
)

// History implements ports.History.
type History struct {
	C reporthistoryv1.ReportHistoryServiceClient
}

func (h History) ListReportsByReporter(ctx context.Context, reporterID string, pageSize int32, pageToken string) (*reporthistoryv1.ListReportsByReporterResponse, error) {
	return h.C.ListReportsByReporter(ctx, &reporthistoryv1.ListReportsByReporterRequest{
		ReporterId: reporterID,
		PageSize:   pageSize,
		PageToken:  pageToken,
	})
}

func (h History) GetReportTimeline(ctx context.Context, reportID string) (*reporthistoryv1.GetReportTimelineResponse, error) {
	return h.C.GetReportTimeline(ctx, &reporthistoryv1.GetReportTimelineRequest{ReportId: reportID})
}
