package grpcclients

import (
	"context"

	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
)

// Status implements ports.Status.
type Status struct {
	C statuslifecyclev1.StatusLifecycleServiceClient
}

func (s Status) ChangeStatus(ctx context.Context, in *statuslifecyclev1.ChangeStatusRequest) (*statuslifecyclev1.ChangeStatusResponse, error) {
	return s.C.ChangeStatus(ctx, in)
}

func (s Status) GetStatus(ctx context.Context, reportID string) (*statuslifecyclev1.GetStatusResponse, error) {
	return s.C.GetStatus(ctx, &statuslifecyclev1.GetStatusRequest{ReportId: reportID})
}
