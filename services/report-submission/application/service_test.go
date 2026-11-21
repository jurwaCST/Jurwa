package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/report-submission/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

const fixedID = "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f"

type fixedIDs struct{}

func (fixedIDs) NewID() string { return fixedID }

// failingRepo returns an error on Save, to check the use case reports it.
type failingRepo struct{ memory.Repository }

func (*failingRepo) Save(context.Context, domain.Report) error { return errors.New("disk full") }

var fixedNow = time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)

// recordingPublisher remembers every published report, or fails when err is set.
type recordingPublisher struct {
	published []domain.Report
	withdrawn []withdrawn
	err       error
}

// withdrawn is one report.withdrawn the service published (SCRUM-163).
type withdrawn struct {
	report   domain.Report
	assignee string
}

func (p *recordingPublisher) ReportWithdrawn(_ context.Context, r domain.Report, assignee string) error {
	if p.err != nil {
		return p.err
	}
	p.withdrawn = append(p.withdrawn, withdrawn{r, assignee})
	return nil
}

func (p *recordingPublisher) ReportSubmitted(_ context.Context, r domain.Report) error {
	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, r)
	return nil
}

func newTestService(repo *memory.Repository) *Service {
	return NewService(repo, fixedIDs{}, &recordingPublisher{}, func() time.Time { return fixedNow })
}

func validForm() domain.NewReportInput {
	return domain.NewReportInput{
		ReporterID:  "0b9e8d7c-6a5b-4c3d-9e2f-1a0b9c8d7e6f",
		Category:    domain.CategoryAcademicBlock,
		Location:    domain.Location{Building: "Academic Block 2", Area: "Lab 3"},
		Description: "Projector does not turn on.",
	}
}

func TestSubmitReportStoresOpenReport(t *testing.T) {
	repo := memory.NewRepository()
	svc := newTestService(repo)

	report, err := svc.SubmitReport(context.Background(), validForm())

	require.NoError(t, err)
	assert.Equal(t, fixedID, report.ID)
	assert.Equal(t, domain.StatusOpen, report.Status)
	assert.Equal(t, fixedNow, report.CreatedAt)

	stored, err := repo.FindByID(context.Background(), fixedID)
	require.NoError(t, err)
	assert.Equal(t, report, stored)
}

func TestSubmitReportInvalidIsNotStored(t *testing.T) {
	repo := memory.NewRepository()
	svc := newTestService(repo)
	form := validForm()
	form.Description = "   "

	_, err := svc.SubmitReport(context.Background(), form)

	require.ErrorIs(t, err, domain.ErrMissingDescription)
	_, err = repo.FindByID(context.Background(), fixedID)
	assert.ErrorIs(t, err, domain.ErrReportNotFound)
}

func TestSubmitReportSaveError(t *testing.T) {
	pub := &recordingPublisher{}
	svc := NewService(&failingRepo{}, fixedIDs{}, pub, func() time.Time { return fixedNow })

	_, err := svc.SubmitReport(context.Background(), validForm())

	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidReport)
	assert.Empty(t, pub.published, "nothing is published when the report was not stored")
}

func TestSubmitReportPublishesReportSubmitted(t *testing.T) {
	pub := &recordingPublisher{}
	svc := NewService(memory.NewRepository(), fixedIDs{}, pub, func() time.Time { return fixedNow })

	report, err := svc.SubmitReport(context.Background(), validForm())

	require.NoError(t, err)
	require.Len(t, pub.published, 1)
	assert.Equal(t, report, pub.published[0])
}

func TestSubmitReportInvalidIsNotPublished(t *testing.T) {
	pub := &recordingPublisher{}
	svc := NewService(memory.NewRepository(), fixedIDs{}, pub, func() time.Time { return fixedNow })
	form := validForm()
	form.Category = ""

	_, err := svc.SubmitReport(context.Background(), form)

	require.ErrorIs(t, err, domain.ErrMissingCategory)
	assert.Empty(t, pub.published)
}

func TestSubmitReportPublishErrorIsReturned(t *testing.T) {
	pub := &recordingPublisher{err: errors.New("redis down")}
	svc := NewService(memory.NewRepository(), fixedIDs{}, pub, func() time.Time { return fixedNow })

	_, err := svc.SubmitReport(context.Background(), validForm())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "publish report.submitted")
}

func TestGetReport(t *testing.T) {
	repo := memory.NewRepository()
	svc := newTestService(repo)
	created, err := svc.SubmitReport(context.Background(), validForm())
	require.NoError(t, err)

	got, err := svc.GetReport(context.Background(), fixedID)
	require.NoError(t, err)
	assert.Equal(t, created, got)

	_, err = svc.GetReport(context.Background(), "11111111-1111-4111-8111-111111111111")
	assert.ErrorIs(t, err, domain.ErrReportNotFound)
}

// ---- SCRUM-163: the student edits or withdraws their own report ----

const otherStudent = "11111111-1111-4111-8111-111111111111"

func submittedReport(t *testing.T) (*Service, *recordingPublisher, domain.Report) {
	t.Helper()
	pub := &recordingPublisher{}
	svc := NewService(memory.NewRepository(), fixedIDs{}, pub, func() time.Time { return fixedNow })
	r, err := svc.SubmitReport(context.Background(), validForm())
	require.NoError(t, err)
	return svc, pub, r
}

func TestUpdateDescription(t *testing.T) {
	svc, _, r := submittedReport(t)
	ctx := context.Background()

	got, err := svc.UpdateDescription(ctx, r.ID, r.ReporterID, "Projector and speakers do not turn on.")
	require.NoError(t, err)
	assert.Equal(t, "Projector and speakers do not turn on.", got.Description)
	stored, err := svc.GetReport(ctx, r.ID)
	require.NoError(t, err)
	assert.Equal(t, "Projector and speakers do not turn on.", stored.Description)
	assert.Equal(t, r.Location, stored.Location, "only the description changes")

	_, err = svc.UpdateDescription(ctx, r.ID, otherStudent, "mine now")
	assert.ErrorIs(t, err, domain.ErrNotReporter)
	_, err = svc.UpdateDescription(ctx, r.ID, r.ReporterID, " ")
	assert.ErrorIs(t, err, domain.ErrMissingDescription)
	_, err = svc.UpdateDescription(ctx, "6f1c2a7e-3b4d-4c5e-8f90-000000000000", r.ReporterID, "x")
	assert.ErrorIs(t, err, domain.ErrReportNotFound)
}

func TestWithdrawReport(t *testing.T) {
	svc, pub, r := submittedReport(t)
	ctx := context.Background()

	assert.ErrorIs(t, svc.WithdrawReport(ctx, r.ID, otherStudent, ""), domain.ErrNotReporter)
	assert.Empty(t, pub.withdrawn, "nothing published for a refused withdrawal")

	require.NoError(t, svc.WithdrawReport(ctx, r.ID, r.ReporterID, "user-sso"))
	_, err := svc.GetReport(ctx, r.ID)
	assert.ErrorIs(t, err, domain.ErrReportNotFound, "the report is gone")
	require.Len(t, pub.withdrawn, 1)
	assert.Equal(t, r.ID, pub.withdrawn[0].report.ID)
	assert.Equal(t, "user-sso", pub.withdrawn[0].assignee, "so the person holding it is told")

	assert.ErrorIs(t, svc.WithdrawReport(ctx, r.ID, r.ReporterID, ""), domain.ErrReportNotFound)
}
