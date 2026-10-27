package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	reportsubmissionv1 "github.com/jurwaCST/Jurwa/proto/gen/report-submission/v1"
	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

// UseCases is what the handler needs from the application layer.
type UseCases interface {
	SubmitReport(ctx context.Context, in domain.NewReportInput) (domain.Report, error)
	GetReport(ctx context.Context, id string) (domain.Report, error)
}

// Handler implements reportsubmissionv1.ReportSubmissionServiceServer.
type Handler struct {
	reportsubmissionv1.UnimplementedReportSubmissionServiceServer
	app UseCases
}

// NewHandler returns a gRPC handler backed by the use cases.
func NewHandler(app UseCases) *Handler {
	return &Handler{app: app}
}

// SubmitReport validates and stores a report. Form rule errors return INVALID_ARGUMENT.
func (h *Handler) SubmitReport(ctx context.Context, req *reportsubmissionv1.SubmitReportRequest) (*reportsubmissionv1.SubmitReportResponse, error) {
	in := domain.NewReportInput{
		ReporterID:  req.GetReporterId(),
		Category:    categoryFromProto(req.GetCategory()),
		Location:    locationFromProto(req.GetLocation()),
		Description: req.GetDescription(),
	}
	if req.PhotoId != nil {
		photoID := req.GetPhotoId()
		in.PhotoID = &photoID
	}

	report, err := h.app.SubmitReport(ctx, in)
	if err != nil {
		return nil, toStatus(err)
	}
	return &reportsubmissionv1.SubmitReportResponse{Report: reportToProto(report)}, nil
}

// GetReport returns one report. Unknown ids return NOT_FOUND.
func (h *Handler) GetReport(ctx context.Context, req *reportsubmissionv1.GetReportRequest) (*reportsubmissionv1.GetReportResponse, error) {
	if !domain.IsUUID(req.GetReportId()) {
		return nil, status.Error(codes.InvalidArgument, "report_id must be a UUID")
	}
	report, err := h.app.GetReport(ctx, req.GetReportId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &reportsubmissionv1.GetReportResponse{Report: reportToProto(report)}, nil
}

// toStatus maps domain errors to gRPC status codes and hides internal details.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidReport):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrReportNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func categoryFromProto(c commonv1.Category) domain.Category {
	switch c {
	case commonv1.Category_CATEGORY_UNSPECIFIED:
		return ""
	case commonv1.Category_HOSTEL:
		return domain.CategoryHostel
	case commonv1.Category_ACADEMIC_BLOCK:
		return domain.CategoryAcademicBlock
	case commonv1.Category_OTHERS:
		return domain.CategoryOthers
	default:
		// An enum number this version does not know; the domain rejects it as invalid.
		return domain.Category(fmt.Sprintf("UNKNOWN_%d", c))
	}
}

func categoryToProto(c domain.Category) commonv1.Category {
	switch c {
	case domain.CategoryHostel:
		return commonv1.Category_HOSTEL
	case domain.CategoryAcademicBlock:
		return commonv1.Category_ACADEMIC_BLOCK
	case domain.CategoryOthers:
		return commonv1.Category_OTHERS
	default:
		return commonv1.Category_CATEGORY_UNSPECIFIED
	}
}

func statusToProto(s domain.Status) commonv1.Status {
	if s == domain.StatusOpen {
		return commonv1.Status_OPEN
	}
	return commonv1.Status_STATUS_UNSPECIFIED
}

func locationFromProto(l *commonv1.Location) domain.Location {
	return domain.Location{Building: l.GetBuilding(), Wing: l.GetWing(), Area: l.GetArea()}
}

func reportToProto(r domain.Report) *commonv1.Report {
	return &commonv1.Report{
		Id:          r.ID,
		ReporterId:  r.ReporterID,
		Category:    categoryToProto(r.Category),
		Location:    &commonv1.Location{Building: r.Location.Building, Wing: r.Location.Wing, Area: r.Location.Area},
		Description: r.Description,
		PhotoId:     r.PhotoID,
		Status:      statusToProto(r.Status),
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
}
