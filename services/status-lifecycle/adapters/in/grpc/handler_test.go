package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	statuslifecyclev1 "github.com/jurwaCST/Jurwa/proto/gen/status-lifecycle/v1"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/application"
	"github.com/jurwaCST/Jurwa/services/status-lifecycle/ports"
)

const (
	reportID = "11111111-1111-1111-1111-111111111111"
	staffID  = "22222222-2222-2222-2222-222222222222"
)

type nopPublisher struct{ err error }

func (p nopPublisher) PublishStatusChanged(context.Context, ports.StatusChanged) error { return p.err }

func newHandler(t *testing.T, pubErr error) *Handler {
	t.Helper()
	app := application.NewService(memory.New(), nopPublisher{err: pubErr}, time.Now)
	_, err := app.Register(context.Background(), reportID)
	require.NoError(t, err)
	return NewHandler(app)
}

func TestChangeStatus_OpenToAssigned(t *testing.T) {
	h := newHandler(t, nil)

	resp, err := h.ChangeStatus(context.Background(), &statuslifecyclev1.ChangeStatusRequest{
		ReportId: reportID, ToStatus: commonv1.Status_ASSIGNED,
	})
	require.NoError(t, err)
	assert.Equal(t, commonv1.Status_ASSIGNED, resp.GetStatus().GetStatus())
	assert.Equal(t, commonv1.Status_OPEN, resp.GetStatus().GetPreviousStatus())
	assert.Equal(t, int32(2), resp.GetStatus().GetVersion())
	assert.Nil(t, resp.GetStatus().UpdatedBy, "absent when the system acted")

	resp, err = h.ChangeStatus(context.Background(), &statuslifecyclev1.ChangeStatusRequest{
		ReportId: reportID, ToStatus: commonv1.Status_IN_PROGRESS, ActorId: staffID,
	})
	require.NoError(t, err)
	assert.Equal(t, staffID, resp.GetStatus().GetUpdatedBy())

	got, err := h.GetStatus(context.Background(), &statuslifecyclev1.GetStatusRequest{ReportId: reportID})
	require.NoError(t, err)
	assert.Equal(t, commonv1.Status_IN_PROGRESS, got.GetStatus().GetStatus())
}

func TestChangeStatus_ErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		req      *statuslifecyclev1.ChangeStatusRequest
		pubErr   error
		wantCode codes.Code
	}{
		{"OPEN to RESOLVED", &statuslifecyclev1.ChangeStatusRequest{ReportId: reportID, ToStatus: commonv1.Status_RESOLVED}, nil, codes.FailedPrecondition},
		{"bad report id", &statuslifecyclev1.ChangeStatusRequest{ReportId: "x", ToStatus: commonv1.Status_ASSIGNED}, nil, codes.InvalidArgument},
		{"status unspecified", &statuslifecyclev1.ChangeStatusRequest{ReportId: reportID}, nil, codes.InvalidArgument},
		{"unknown report", &statuslifecyclev1.ChangeStatusRequest{ReportId: staffID, ToStatus: commonv1.Status_ASSIGNED}, nil, codes.NotFound},
		{"stale version", &statuslifecyclev1.ChangeStatusRequest{ReportId: reportID, ToStatus: commonv1.Status_ASSIGNED, ExpectedVersion: 9}, nil, codes.Aborted},
		{"publisher fails", &statuslifecyclev1.ChangeStatusRequest{ReportId: reportID, ToStatus: commonv1.Status_ASSIGNED}, errors.New("down"), codes.Internal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newHandler(t, tt.pubErr).ChangeStatus(context.Background(), tt.req)
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestGetStatus_NotFound(t *testing.T) {
	_, err := newHandler(t, nil).GetStatus(context.Background(), &statuslifecyclev1.GetStatusRequest{ReportId: staffID})
	assert.Equal(t, codes.NotFound, status.Code(err))
}
