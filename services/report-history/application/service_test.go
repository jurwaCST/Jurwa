package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-history/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-history/domain"
)

const (
	reportA  = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"
	reportB  = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"
	studentA = "1b2c3d4e-5f60-4a7b-8c9d-0e1f2a3b4c5d"
	studentB = "2c3d4e5f-6071-4b8c-9d0e-1f2a3b4c5d6e"
	dsa      = "4e5f6071-8293-4d0e-9f2a-3b4c5d6e7f80"
)

var t0 = time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)

// sequentialIDs returns a new valid UUID on every call.
func sequentialIDs() func() string {
	n := 0
	return func() string {
		n++
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", n)
	}
}

func newTestService() *Service {
	return NewService(memory.NewRepository(), sequentialIDs())
}

func submitted(reportID, reporterID string, at time.Time) SubmittedInput {
	return SubmittedInput{
		ReportID:   reportID,
		ReporterID: reporterID,
		Category:   domain.CategoryHostel,
		Location:   domain.Location{Building: "Hostel A", Wing: "Wing 1", Area: "Bathroom 1"},
		CreatedAt:  at,
	}
}

func statusChange(reportID string, from, to domain.Status, at time.Time) domain.EntryInput {
	return domain.EntryInput{ReportID: reportID, EventType: domain.EventStatusChanged, FromStatus: from, ToStatus: to, ActorID: dsa, Summary: "Status changed to " + string(to), OccurredAt: at}
}

func TestRecordSubmittedStartsTimeline(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	s, err := svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, s.Status)

	timeline, err := svc.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	require.Len(t, timeline, 1)
	assert.Equal(t, domain.EventReportSubmitted, timeline[0].EventType)
	assert.Equal(t, domain.StatusOpen, timeline[0].ToStatus)
	assert.Equal(t, studentA, timeline[0].ActorID)
}

func TestRecordSubmittedTwiceIsRejected(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	_, err := svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
	require.NoError(t, err)

	_, err = svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))

	assert.ErrorIs(t, err, domain.ErrReportExists)
}

func TestRecordSubmittedInvalidInput(t *testing.T) {
	_, err := newTestService().RecordSubmitted(context.Background(), submitted("r-1", studentA, t0))

	assert.ErrorIs(t, err, domain.ErrInvalidReportID)
}

func TestTimelineIsOldestFirstEvenWhenRecordedOutOfOrder(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	_, err := svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
	require.NoError(t, err)

	_, err = svc.Record(ctx, statusChange(reportA, domain.StatusAssigned, domain.StatusInProgress, t0.Add(2*time.Hour)))
	require.NoError(t, err)
	_, err = svc.Record(ctx, statusChange(reportA, domain.StatusOpen, domain.StatusAssigned, t0.Add(time.Hour)))
	require.NoError(t, err)

	timeline, err := svc.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	require.Len(t, timeline, 3)
	assert.Equal(t, domain.StatusOpen, timeline[0].ToStatus)
	assert.Equal(t, domain.StatusAssigned, timeline[1].ToStatus)
	assert.Equal(t, domain.StatusInProgress, timeline[2].ToStatus)

	page, _, err := svc.ListReportsByReporter(ctx, studentA, 0, "")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusInProgress, page[0].Status, "the late OPEN->ASSIGNED event does not move the status back")
}

func TestRecordErrors(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()

	_, err := svc.Record(ctx, statusChange(reportA, domain.StatusOpen, domain.StatusAssigned, t0))
	assert.ErrorIs(t, err, domain.ErrReportNotFound)

	_, err = svc.Record(ctx, domain.EntryInput{ReportID: reportA})
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestGetReportTimelineErrors(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetReportTimeline(context.Background(), reportA)
	assert.ErrorIs(t, err, domain.ErrReportNotFound)

	_, err = svc.GetReportTimeline(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidReportID)
}

func TestListReportsByReporterNewestFirstAndOnlyOwn(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	older := "0a000000-0000-4000-8000-000000000001"
	newer := "0a000000-0000-4000-8000-000000000002"
	for _, in := range []SubmittedInput{
		submitted(older, studentA, t0),
		submitted(reportB, studentB, t0.Add(30*time.Minute)),
		submitted(newer, studentA, t0.Add(time.Hour)),
	} {
		_, err := svc.RecordSubmitted(ctx, in)
		require.NoError(t, err)
	}

	pageA, next, err := svc.ListReportsByReporter(ctx, studentA, 0, "")
	require.NoError(t, err)
	assert.Empty(t, next)
	require.Len(t, pageA, 2)
	assert.Equal(t, newer, pageA[0].ReportID)
	assert.Equal(t, older, pageA[1].ReportID)

	pageB, _, err := svc.ListReportsByReporter(ctx, studentB, 0, "")
	require.NoError(t, err)
	require.Len(t, pageB, 1)
	assert.Equal(t, reportB, pageB[0].ReportID, "student B never sees student A's reports")
}

func TestListReportsByReporterPages(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	for i := range 3 {
		_, err := svc.RecordSubmitted(ctx, submitted(fmt.Sprintf("0b000000-0000-4000-8000-%012d", i), studentA, t0.Add(time.Duration(i)*time.Minute)))
		require.NoError(t, err)
	}

	first, next, err := svc.ListReportsByReporter(ctx, studentA, 2, "")
	require.NoError(t, err)
	assert.Len(t, first, 2)
	require.Equal(t, "2", next)

	second, next, err := svc.ListReportsByReporter(ctx, studentA, 2, next)
	require.NoError(t, err)
	assert.Len(t, second, 1)
	assert.Empty(t, next)
}

func TestListReportsByReporterErrors(t *testing.T) {
	svc := newTestService()

	_, _, err := svc.ListReportsByReporter(context.Background(), "", 0, "")
	assert.ErrorIs(t, err, domain.ErrInvalidReporterID)

	_, _, err = svc.ListReportsByReporter(context.Background(), studentA, 500, "")
	assert.ErrorIs(t, err, domain.ErrInvalidPageSize)
}

func routedTo(reportID, assigneeID string, at time.Time) domain.EntryInput {
	return domain.EntryInput{ReportID: reportID, EventType: domain.EventReportRouted, AssigneeID: assigneeID, Summary: "Routed", OccurredAt: at}
}

func TestListReportsByAssigneeNewestFirstAndOnlyTheirs(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	daa := "5f607182-93a4-4e1f-8a3b-4c5d6e7f8091"
	older := "0c000000-0000-4000-8000-000000000001"
	newer := "0c000000-0000-4000-8000-000000000002"
	for _, in := range []SubmittedInput{
		submitted(older, studentA, t0),
		submitted(reportB, studentB, t0.Add(30*time.Minute)),
		submitted(newer, studentB, t0.Add(time.Hour)),
	} {
		_, err := svc.RecordSubmitted(ctx, in)
		require.NoError(t, err)
	}
	for _, in := range []domain.EntryInput{
		routedTo(older, dsa, t0.Add(2*time.Hour)),
		routedTo(reportB, daa, t0.Add(2*time.Hour)),
		routedTo(newer, dsa, t0.Add(2*time.Hour)),
	} {
		_, err := svc.Record(ctx, in)
		require.NoError(t, err)
	}

	page, next, err := svc.ListReportsByAssignee(ctx, dsa, 0, "")
	require.NoError(t, err)
	assert.Empty(t, next)
	require.Len(t, page, 2)
	assert.Equal(t, newer, page[0].ReportID)
	assert.Equal(t, older, page[1].ReportID)

	other, _, err := svc.ListReportsByAssignee(ctx, daa, 0, "")
	require.NoError(t, err)
	require.Len(t, other, 1)
	assert.Equal(t, reportB, other[0].ReportID)
}

func TestListReportsByAssigneeFollowsReassignment(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	hod := "6071a2b3-c4d5-4e6f-9a0b-1c2d3e4f5a6b"
	_, err := svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
	require.NoError(t, err)
	_, err = svc.Record(ctx, routedTo(reportA, dsa, t0.Add(time.Minute)))
	require.NoError(t, err)
	_, err = svc.Record(ctx, domain.EntryInput{ReportID: reportA, EventType: domain.EventReportReassigned, AssigneeID: hod, ActorID: dsa, Summary: "Reassigned to HoD", OccurredAt: t0.Add(time.Hour)})
	require.NoError(t, err)

	old, _, err := svc.ListReportsByAssignee(ctx, dsa, 0, "")
	require.NoError(t, err)
	assert.Empty(t, old, "the old assignee no longer sees it")

	now, _, err := svc.ListReportsByAssignee(ctx, hod, 0, "")
	require.NoError(t, err)
	require.Len(t, now, 1)
	assert.Equal(t, reportA, now[0].ReportID)
}

func TestListReportsByAssigneePages(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	for i := range 3 {
		id := fmt.Sprintf("0d000000-0000-4000-8000-%012d", i)
		_, err := svc.RecordSubmitted(ctx, submitted(id, studentA, t0.Add(time.Duration(i)*time.Minute)))
		require.NoError(t, err)
		_, err = svc.Record(ctx, routedTo(id, dsa, t0.Add(time.Hour)))
		require.NoError(t, err)
	}

	first, next, err := svc.ListReportsByAssignee(ctx, dsa, 2, "")
	require.NoError(t, err)
	assert.Len(t, first, 2)
	require.Equal(t, "2", next)

	second, next, err := svc.ListReportsByAssignee(ctx, dsa, 2, next)
	require.NoError(t, err)
	assert.Len(t, second, 1)
	assert.Empty(t, next)
}

func TestListReportsByAssigneeErrors(t *testing.T) {
	svc := newTestService()

	_, _, err := svc.ListReportsByAssignee(context.Background(), "", 0, "")
	assert.ErrorIs(t, err, domain.ErrInvalidAssigneeID)

	_, _, err = svc.ListReportsByAssignee(context.Background(), dsa, 500, "")
	assert.ErrorIs(t, err, domain.ErrInvalidPageSize)

	_, _, err = NewService(brokenRepo{Repository: memory.NewRepository(), failOn: "ListSummariesByAssignee"}, sequentialIDs()).
		ListReportsByAssignee(context.Background(), dsa, 0, "")
	assert.ErrorIs(t, err, errDB)
}

func TestRecordEventUsesTheEventIDAndIgnoresARepeat(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	_, err := svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
	require.NoError(t, err)
	eventID := "0e000000-0000-4000-8000-000000000001"

	require.NoError(t, svc.RecordEvent(ctx, eventID, routedTo(reportA, dsa, t0.Add(time.Minute))))
	require.NoError(t, svc.RecordEvent(ctx, eventID, routedTo(reportA, dsa, t0.Add(time.Minute))), "a repeated event is not an error")

	timeline, err := svc.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	require.Len(t, timeline, 2, "submitted and routed, the repeat is ignored")
	assert.Equal(t, eventID, timeline[1].ID)

	page, _, err := svc.ListReportsByAssignee(ctx, dsa, 0, "")
	require.NoError(t, err)
	assert.Len(t, page, 1)
}

func TestRecordEventErrors(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	eventID := "0e000000-0000-4000-8000-000000000002"

	assert.ErrorIs(t, svc.RecordEvent(ctx, eventID, routedTo(reportA, dsa, t0)), domain.ErrReportNotFound)
	assert.ErrorIs(t, svc.RecordEvent(ctx, "not-a-uuid", routedTo(reportA, dsa, t0)), domain.ErrInvalidEntryID)

	broken := NewService(brokenRepo{Repository: memory.NewRepository(), failOn: "AppendEntry"}, sequentialIDs())
	_, err := broken.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
	require.Error(t, err)
}

func TestRecordSubmittedUsesTheEventIDForTheFirstEntry(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	in := submitted(reportA, studentA, t0)
	in.EventID = "0e000000-0000-4000-8000-000000000003"

	_, err := svc.RecordSubmitted(ctx, in)
	require.NoError(t, err)

	timeline, err := svc.GetReportTimeline(ctx, reportA)
	require.NoError(t, err)
	assert.Equal(t, in.EventID, timeline[0].ID)
}

// Two consumers (two streams) can record events for the same report at the same
// moment. Neither change may be lost.
func TestConcurrentEventsForOneReportAreAllApplied(t *testing.T) {
	ctx := context.Background()
	hod := "6071a2b3-c4d5-4e6f-9a0b-1c2d3e4f5a6b"
	for i := range 200 {
		svc := newTestService()
		_, err := svc.RecordSubmitted(ctx, submitted(reportA, studentA, t0))
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, svc.RecordEvent(ctx, fmt.Sprintf("0e100000-0000-4000-8000-%012d", i), statusChange(reportA, domain.StatusOpen, domain.StatusAssigned, t0.Add(2*time.Minute))))
		}()
		go func() {
			defer wg.Done()
			assert.NoError(t, svc.RecordEvent(ctx, fmt.Sprintf("0e200000-0000-4000-8000-%012d", i), domain.EntryInput{
				ReportID: reportA, EventType: domain.EventReportReassigned, AssigneeID: hod, Summary: "Reassigned", OccurredAt: t0.Add(time.Hour),
			}))
		}()
		wg.Wait()

		page, _, err := svc.ListReportsByReporter(ctx, studentA, 0, "")
		require.NoError(t, err)
		require.Equal(t, domain.StatusAssigned, page[0].Status, "run %d: status change lost", i)
		require.Equal(t, hod, page[0].AssigneeID, "run %d: reassignment lost", i)
	}
}

// brokenRepo wraps the memory repository and fails the chosen method.
type brokenRepo struct {
	*memory.Repository
	failOn string
}

var errDB = errors.New("database is down")

func (r brokenRepo) fail(method string) error {
	if r.failOn == method {
		return errDB
	}
	return nil
}

func (r brokenRepo) CreateSummary(ctx context.Context, s domain.Summary) error {
	if err := r.fail("CreateSummary"); err != nil {
		return err
	}
	return r.Repository.CreateSummary(ctx, s)
}

func (r brokenRepo) UpdateSummary(ctx context.Context, s domain.Summary) error {
	if err := r.fail("UpdateSummary"); err != nil {
		return err
	}
	return r.Repository.UpdateSummary(ctx, s)
}

func (r brokenRepo) AppendEntry(ctx context.Context, e domain.Entry) error {
	if err := r.fail("AppendEntry"); err != nil {
		return err
	}
	return r.Repository.AppendEntry(ctx, e)
}

func (r brokenRepo) ListSummariesByReporter(ctx context.Context, id string) ([]domain.Summary, error) {
	if err := r.fail("ListSummariesByReporter"); err != nil {
		return nil, err
	}
	return r.Repository.ListSummariesByReporter(ctx, id)
}

func (r brokenRepo) ListSummariesByAssignee(ctx context.Context, id string) ([]domain.Summary, error) {
	if err := r.fail("ListSummariesByAssignee"); err != nil {
		return nil, err
	}
	return r.Repository.ListSummariesByAssignee(ctx, id)
}

func (r brokenRepo) ListEntries(ctx context.Context, id string) ([]domain.Entry, error) {
	if err := r.fail("ListEntries"); err != nil {
		return nil, err
	}
	return r.Repository.ListEntries(ctx, id)
}

func TestStorageErrorsArePassedOn(t *testing.T) {
	tests := []struct {
		failOn string
		call   func(ctx context.Context, svc *Service) error
	}{
		{failOn: "CreateSummary", call: func(ctx context.Context, svc *Service) error {
			_, err := svc.RecordSubmitted(ctx, submitted(reportB, studentA, t0))
			return err
		}},
		{failOn: "AppendEntry", call: func(ctx context.Context, svc *Service) error {
			_, err := svc.RecordSubmitted(ctx, submitted(reportB, studentA, t0))
			return err
		}},
		{failOn: "AppendEntry", call: func(ctx context.Context, svc *Service) error {
			_, err := svc.Record(ctx, statusChange(reportA, domain.StatusOpen, domain.StatusAssigned, t0.Add(time.Hour)))
			return err
		}},
		{failOn: "UpdateSummary", call: func(ctx context.Context, svc *Service) error {
			_, err := svc.Record(ctx, statusChange(reportA, domain.StatusOpen, domain.StatusAssigned, t0.Add(time.Hour)))
			return err
		}},
		{failOn: "ListSummariesByReporter", call: func(ctx context.Context, svc *Service) error {
			_, _, err := svc.ListReportsByReporter(ctx, studentA, 0, "")
			return err
		}},
		{failOn: "ListEntries", call: func(ctx context.Context, svc *Service) error {
			_, err := svc.GetReportTimeline(ctx, reportA)
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.failOn, func(t *testing.T) {
			ctx := context.Background()
			mem := memory.NewRepository()
			ids := sequentialIDs() // one generator, so entry ids stay unique
			_, err := NewService(mem, ids).RecordSubmitted(ctx, submitted(reportA, studentA, t0))
			require.NoError(t, err)
			svc := NewService(brokenRepo{Repository: mem, failOn: tt.failOn}, ids)

			err = tt.call(ctx, svc)

			assert.ErrorIs(t, err, errDB)
		})
	}
}
