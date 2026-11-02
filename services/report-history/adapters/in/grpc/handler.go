package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

// UseCases is what the handler needs from the application layer.
type UseCases interface {
	ListReportsByReporter(ctx context.Context, reporterID string, pageSize int32, pageToken string) ([]domain.Summary, string, error)
	GetReportTimeline(ctx context.Context, reportID string) ([]domain.Entry, error)
}

// Handler implements reporthistoryv1.ReportHistoryServiceServer.
type Handler struct {
	reporthistoryv1.UnimplementedReportHistoryServiceServer
	app UseCases
}

// NewHandler returns a gRPC handler backed by the use cases.
func NewHandler(app UseCases) *Handler {
	return &Handler{app: app}
}

// ListReportsByReporter returns one page of a student's reports, newest first.
// A bad reporter id, page size or page token returns INVALID_ARGUMENT.
func (h *Handler) ListReportsByReporter(ctx context.Context, req *reporthistoryv1.ListReportsByReporterRequest) (*reporthistoryv1.ListReportsByReporterResponse, error) {
	summaries, next, err := h.app.ListReportsByReporter(ctx, req.GetReporterId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &reporthistoryv1.ListReportsByReporterResponse{
		Reports:       make([]*reporthistoryv1.ReportSummary, 0, len(summaries)),
		NextPageToken: next,
	}
	for _, s := range summaries {
		resp.Reports = append(resp.Reports, summaryToProto(s))
	}
	return resp, nil
}

// GetReportTimeline returns every entry of one report, oldest first.
// An unknown report returns NOT_FOUND.
func (h *Handler) GetReportTimeline(ctx context.Context, req *reporthistoryv1.GetReportTimelineRequest) (*reporthistoryv1.GetReportTimelineResponse, error) {
	entries, err := h.app.GetReportTimeline(ctx, req.GetReportId())
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &reporthistoryv1.GetReportTimelineResponse{Entries: make([]*reporthistoryv1.TimelineEntry, 0, len(entries))}
	for _, e := range entries {
		resp.Entries = append(resp.Entries, entryToProto(e))
	}
	return resp, nil
}

// toStatus maps domain errors to gRPC status codes and hides internal details.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrReportNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func summaryToProto(s domain.Summary) *reporthistoryv1.ReportSummary {
	out := &reporthistoryv1.ReportSummary{
		ReportId:   s.ReportID,
		ReporterId: s.ReporterID,
		Category:   categoryToProto(s.Category),
		Location:   &commonv1.Location{Building: s.Location.Building, Wing: s.Location.Wing, Area: s.Location.Area},
		Status:     statusToProto(s.Status),
		CreatedAt:  timestamppb.New(s.CreatedAt),
		UpdatedAt:  timestamppb.New(s.UpdatedAt),
	}
	if s.AssigneeID != "" {
		out.CurrentAssigneeId = &s.AssigneeID
	}
	return out
}

func entryToProto(e domain.Entry) *reporthistoryv1.TimelineEntry {
	out := &reporthistoryv1.TimelineEntry{
		Id:         e.ID,
		EventType:  e.EventType,
		FromStatus: statusToProto(e.FromStatus),
		ToStatus:   statusToProto(e.ToStatus),
		Summary:    e.Summary,
		OccurredAt: timestamppb.New(e.OccurredAt),
	}
	if e.ActorID != "" {
		out.ActorId = &e.ActorID
	}
	return out
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
	switch s {
	case domain.StatusOpen:
		return commonv1.Status_OPEN
	case domain.StatusAssigned:
		return commonv1.Status_ASSIGNED
	case domain.StatusInProgress:
		return commonv1.Status_IN_PROGRESS
	case domain.StatusResolved:
		return commonv1.Status_RESOLVED
	case domain.StatusEscalated:
		return commonv1.Status_ESCALATED
	default:
		// "" means the entry is not a status change.
		return commonv1.Status_STATUS_UNSPECIFIED
	}
}
