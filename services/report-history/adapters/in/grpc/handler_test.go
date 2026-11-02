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
	reporthistoryv1 "github.com/jurwaCST/Jurwa/proto/gen/report-history/v1"
	"github.com/jurwaCST/Jurwa/services/report-history/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-history/application"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

const (
	reportA  = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	studentA = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
	dsa      = "4e5f6071-8293-4d0e-9f2a-3b4c5d6e7f80"
	entryID  = "3d4e5f60-7182-4c9d-8e1f-2a3b4c5d6e7f"
)

var t0 = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

// newSeededHandler returns a handler with one hostel report that was assigned an hour later.
func newSeededHandler(t *testing.T) *Handler {
	t.Helper()
	ctx := context.Background()
	app := application.NewService(memory.NewRepository(), func() string { return entryID })
	_, err := app.RecordSubmitted(ctx, application.SubmittedInput{
		ReportID: reportA, ReporterID: studentA, Category: domain.CategoryHostel,
		Location: domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"}, CreatedAt: t0,
	})
	require.NoError(t, err)
	_, err = app.Record(ctx, domain.EntryInput{
		ReportID: reportA, EventType: domain.EventStatusChanged, FromStatus: domain.StatusOpen, ToStatus: domain.StatusAssigned,
		ActorID: dsa, Summary: "Assigned to DSA", OccurredAt: t0.Add(time.Hour),
	})
	require.NoError(t, err)
	return NewHandler(app)
}

// brokenApp fails every call, to check internal errors are hidden.
type brokenApp struct{}

func (brokenApp) ListReportsByReporter(context.Context, string, int32, string) ([]domain.Summary, string, error) {
	return nil, "", errors.New("list summaries: postgres password is wrong")
}

func (brokenApp) GetReportTimeline(context.Context, string) ([]domain.Entry, error) {
	return nil, errors.New("list entries: postgres password is wrong")
}

// fixedApp returns fixed values, for fields the application cannot set yet.
type fixedApp struct {
	summaries []domain.Summary
	entries   []domain.Entry
}

func (a fixedApp) ListReportsByReporter(context.Context, string, int32, string) ([]domain.Summary, string, error) {
	return a.summaries, "", nil
}

func (a fixedApp) GetReportTimeline(context.Context, string) ([]domain.Entry, error) {
	return a.entries, nil
}

func TestListReportsByReporter(t *testing.T) {
	resp, err := newSeededHandler(t).ListReportsByReporter(context.Background(), &reporthistoryv1.ListReportsByReporterRequest{ReporterId: studentA})

	require.NoError(t, err)
	require.Len(t, resp.GetReports(), 1)
	r := resp.GetReports()[0]
	assert.Equal(t, reportA, r.GetReportId())
	assert.Equal(t, studentA, r.GetReporterId())
	assert.Equal(t, commonv1.Category_HOSTEL, r.GetCategory())
	assert.Equal(t, "Hostel A", r.GetLocation().GetBuilding())
	assert.Equal(t, "Wing 1", r.GetLocation().GetWing())
	assert.Equal(t, "Bathroom 1", r.GetLocation().GetArea())
	assert.Equal(t, commonv1.Status_ASSIGNED, r.GetStatus())
	assert.Nil(t, r.CurrentAssigneeId)
	assert.Equal(t, t0, r.GetCreatedAt().AsTime())
	assert.Equal(t, t0.Add(time.Hour), r.GetUpdatedAt().AsTime())
	assert.Empty(t, resp.GetNextPageToken())
}

func TestGetReportTimeline(t *testing.T) {
	resp, err := newSeededHandler(t).GetReportTimeline(context.Background(), &reporthistoryv1.GetReportTimelineRequest{ReportId: reportA})

	require.NoError(t, err)
	require.Len(t, resp.GetEntries(), 2)
	first, second := resp.GetEntries()[0], resp.GetEntries()[1]
	assert.Equal(t, "report.submitted", first.GetEventType())
	assert.Equal(t, commonv1.Status_STATUS_UNSPECIFIED, first.GetFromStatus())
	assert.Equal(t, commonv1.Status_OPEN, first.GetToStatus())
	assert.Equal(t, "status.changed", second.GetEventType())
	assert.Equal(t, commonv1.Status_OPEN, second.GetFromStatus())
	assert.Equal(t, commonv1.Status_ASSIGNED, second.GetToStatus())
	assert.Equal(t, dsa, second.GetActorId())
	assert.Equal(t, "Assigned to DSA", second.GetSummary())
	assert.Equal(t, t0.Add(time.Hour), second.GetOccurredAt().AsTime())
}

func TestErrorCodes(t *testing.T) {
	h := newSeededHandler(t)
	ctx := context.Background()

	_, err := h.ListReportsByReporter(ctx, &reporthistoryv1.ListReportsByReporterRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "empty reporter id")

	_, err = h.ListReportsByReporter(ctx, &reporthistoryv1.ListReportsByReporterRequest{ReporterId: studentA, PageSize: 101})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "page size above 100")

	_, err = h.GetReportTimeline(ctx, &reporthistoryv1.GetReportTimelineRequest{ReportId: "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"})
	assert.Equal(t, codes.NotFound, status.Code(err), "unknown report")

	_, err = h.GetReportTimeline(ctx, &reporthistoryv1.GetReportTimelineRequest{ReportId: "report-1"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "malformed report id")
}

func TestOptionalFieldsAndAllEnums(t *testing.T) {
	statuses := map[domain.Status]commonv1.Status{
		domain.StatusOpen: commonv1.Status_OPEN, domain.StatusAssigned: commonv1.Status_ASSIGNED,
		domain.StatusInProgress: commonv1.Status_IN_PROGRESS, domain.StatusResolved: commonv1.Status_RESOLVED,
		domain.StatusEscalated: commonv1.Status_ESCALATED,
	}
	categories := map[domain.Category]commonv1.Category{
		domain.CategoryHostel: commonv1.Category_HOSTEL, domain.CategoryAcademicBlock: commonv1.Category_ACADEMIC_BLOCK,
		domain.CategoryOthers: commonv1.Category_OTHERS, "": commonv1.Category_CATEGORY_UNSPECIFIED,
	}
	for d, p := range statuses {
		assert.Equal(t, p, statusToProto(d))
	}
	for d, p := range categories {
		assert.Equal(t, p, categoryToProto(d))
	}

	h := NewHandler(fixedApp{
		summaries: []domain.Summary{{ReportID: reportA, AssigneeID: dsa}},
		entries:   []domain.Entry{{ID: entryID, Summary: "Escalated by SLA"}},
	})
	list, err := h.ListReportsByReporter(context.Background(), &reporthistoryv1.ListReportsByReporterRequest{})
	require.NoError(t, err)
	assert.Equal(t, dsa, list.GetReports()[0].GetCurrentAssigneeId())

	timeline, err := h.GetReportTimeline(context.Background(), &reporthistoryv1.GetReportTimelineRequest{})
	require.NoError(t, err)
	assert.Nil(t, timeline.GetEntries()[0].ActorId, "absent when the system acted")
}

func TestInternalErrorsAreHidden(t *testing.T) {
	h := NewHandler(brokenApp{})

	_, err := h.ListReportsByReporter(context.Background(), &reporthistoryv1.ListReportsByReporterRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")

	_, err = h.GetReportTimeline(context.Background(), &reporthistoryv1.GetReportTimelineRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "password")
}
