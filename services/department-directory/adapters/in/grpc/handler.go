package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	"github.com/jurwaCST/Jurwa/services/department-directory/application"
	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
)

// Handler implements departmentdirectoryv1.DepartmentDirectoryServiceServer.
type Handler struct {
	departmentdirectoryv1.UnimplementedDepartmentDirectoryServiceServer
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListDepartments(ctx context.Context, _ *departmentdirectoryv1.ListDepartmentsRequest) (*departmentdirectoryv1.ListDepartmentsResponse, error) {
	depts, err := h.svc.ListDepartments(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &departmentdirectoryv1.ListDepartmentsResponse{}
	for _, d := range depts {
		resp.Departments = append(resp.Departments, &departmentdirectoryv1.Department{Code: d.Code, Name: d.Name})
	}
	return resp, nil
}

func (h *Handler) GetHoD(ctx context.Context, req *departmentdirectoryv1.GetHoDRequest) (*departmentdirectoryv1.GetHoDResponse, error) {
	p, err := h.svc.GetHoD(ctx, req.GetDepartmentCode())
	if err != nil {
		return nil, toStatus(err)
	}
	return &departmentdirectoryv1.GetHoDResponse{Position: toProtoPosition(p)}, nil
}

func (h *Handler) GetStaff(ctx context.Context, req *departmentdirectoryv1.GetStaffRequest) (*departmentdirectoryv1.GetStaffResponse, error) {
	p, err := h.svc.GetStaff(ctx, req.GetPositionCode())
	if err != nil {
		return nil, toStatus(err)
	}
	return &departmentdirectoryv1.GetStaffResponse{Position: toProtoPosition(p)}, nil
}

func (h *Handler) ListPositions(ctx context.Context, req *departmentdirectoryv1.ListPositionsRequest) (*departmentdirectoryv1.ListPositionsResponse, error) {
	positions, err := h.svc.ListPositions(ctx, req.GetHeldOnly())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &departmentdirectoryv1.ListPositionsResponse{}
	for _, p := range positions {
		resp.Positions = append(resp.Positions, toProtoPosition(p))
	}
	return resp, nil
}

func toProtoPosition(p domain.Position) *departmentdirectoryv1.Position {
	return &departmentdirectoryv1.Position{
		Code:           p.Code,
		Title:          p.Title,
		Kind:           string(p.Kind),
		DepartmentCode: p.DepartmentCode,
		HolderUserId:   p.HolderUserID,
		Active:         p.Active,
	}
}

// toStatus maps domain errors to gRPC status codes, as written in the contract.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCode), errors.Is(err, domain.ErrInvalidPosition):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		slog.Error("department-directory request failed", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}
