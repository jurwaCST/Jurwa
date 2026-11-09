package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	"github.com/jurwaCST/Jurwa/services/department-directory/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/department-directory/application"
	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
)

func newHandler() *Handler {
	repo := memory.NewRepository(domain.DefaultDepartments(), domain.WithDemoHolders(domain.DefaultPositions()))
	return NewHandler(application.NewService(repo))
}

func TestListDepartments(t *testing.T) {
	resp, err := newHandler().ListDepartments(context.Background(), &departmentdirectoryv1.ListDepartmentsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetDepartments(), 6)
	assert.Equal(t, "ARCH", resp.GetDepartments()[0].GetCode())
}

func TestGetHoD(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		want     string
		wantCode codes.Code
	}{
		{name: "known department", code: "EEE", want: "HOD_EEE"},
		{name: "empty id", code: "", wantCode: codes.InvalidArgument},
		{name: "unknown department", code: "PHYSICS", wantCode: codes.NotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := newHandler().GetHoD(context.Background(), &departmentdirectoryv1.GetHoDRequest{DepartmentCode: tt.code})
			if tt.wantCode != codes.OK {
				assert.Equal(t, tt.wantCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, resp.GetPosition().GetCode())
			assert.Equal(t, "HOD", resp.GetPosition().GetKind())
			assert.Equal(t, tt.code, resp.GetPosition().GetDepartmentCode())
		})
	}
}

func TestGetStaff(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantKind string
		wantCode codes.Code
	}{
		{name: "DSA", code: "DSA", wantKind: "CATEGORY_HEAD"},
		{name: "maintenance councillor", code: "MAINTENANCE_COUNCILLOR", wantKind: "COUNCILLOR"},
		{name: "empty", code: "", wantCode: codes.InvalidArgument},
		{name: "unknown", code: "PRINCIPAL", wantCode: codes.NotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := newHandler().GetStaff(context.Background(), &departmentdirectoryv1.GetStaffRequest{PositionCode: tt.code})
			if tt.wantCode != codes.OK {
				assert.Equal(t, tt.wantCode, status.Code(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantKind, resp.GetPosition().GetKind())
			assert.True(t, resp.GetPosition().GetActive())
		})
	}
}

func TestToStatus(t *testing.T) {
	assert.Equal(t, codes.InvalidArgument, status.Code(toStatus(domain.ErrInvalidPosition)))
	assert.Equal(t, codes.Internal, status.Code(toStatus(errors.New("database down"))))
}

func TestListPositions(t *testing.T) {
	resp, err := newHandler().ListPositions(context.Background(), &departmentdirectoryv1.ListPositionsRequest{HeldOnly: true})
	require.NoError(t, err)
	require.Len(t, resp.GetPositions(), 4)
	staff := resp.GetPositions()[3]
	assert.Equal(t, "SSO_MALE", staff.GetCode())
	assert.Equal(t, "STAFF", staff.GetKind())
	assert.Equal(t, "00000000-0000-4000-8000-000000000010", staff.GetHolderUserId())

	all, err := newHandler().ListPositions(context.Background(), &departmentdirectoryv1.ListPositionsRequest{})
	require.NoError(t, err)
	assert.Len(t, all.GetPositions(), 15)

	_, err = NewHandler(application.NewService(failRepo{})).ListPositions(context.Background(), &departmentdirectoryv1.ListPositionsRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
}

// failRepo is a repository whose every call fails.
type failRepo struct{}

func (failRepo) ListDepartments(context.Context) ([]domain.Department, error) {
	return nil, errors.New("database down")
}
func (failRepo) ListPositions(context.Context) ([]domain.Position, error) {
	return nil, errors.New("database down")
}
