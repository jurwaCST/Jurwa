package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/audit-log/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/audit-log/domain"
)

const reportID = "11111111-1111-1111-1111-111111111111"

var t0 = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func newService() *Service {
	return NewService(memory.New(), func() time.Time { return t0 }, func() string { return "e1" })
}

func TestAppendThenList(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	id, err := svc.Append(ctx, domain.Input{Service: "status-lifecycle", Action: "STATUS_CHANGED", ReportID: reportID})
	require.NoError(t, err)
	assert.Equal(t, "e1", id)

	list, err := svc.ListByReport(ctx, "SYSTEM_ADMIN", reportID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, domain.StatusChanged, list[0].Action)
	assert.Equal(t, t0, list[0].OccurredAt)
}

func TestAppend_Invalid(t *testing.T) {
	svc := newService()
	_, err := svc.Append(context.Background(), domain.Input{Action: "ROUTED"})
	assert.ErrorIs(t, err, domain.ErrInvalidEvent)

	list, _ := svc.ListByReport(context.Background(), "SYSTEM_ADMIN", reportID)
	assert.Empty(t, list, "nothing stored")
}

func TestListByReport_Errors(t *testing.T) {
	svc := newService()
	_, err := svc.ListByReport(context.Background(), "STAFF", reportID)
	assert.ErrorIs(t, err, domain.ErrPermissionDenied)
	_, err = svc.ListByReport(context.Background(), "SYSTEM_ADMIN", "abc")
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

type failingStore struct{}

func (failingStore) Append(context.Context, domain.Event) error { return errors.New("db down") }
func (failingStore) ListByReport(context.Context, string) ([]domain.Event, error) {
	return nil, errors.New("db down")
}

func TestStoreErrors(t *testing.T) {
	svc := NewService(failingStore{}, time.Now, func() string { return "e1" })
	_, err := svc.Append(context.Background(), domain.Input{Service: "s", Action: "ROUTED"})
	assert.ErrorContains(t, err, "db down")
	_, err = svc.ListByReport(context.Background(), "SYSTEM_ADMIN", reportID)
	assert.ErrorContains(t, err, "db down")
}
