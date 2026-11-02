package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/sla-policy/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/sla-policy/domain"
)

const adminID = "22222222-2222-2222-2222-222222222222"

var fixedNow = time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)

func newTestService() *Service {
	return NewService(memory.New(), func() time.Time { return fixedNow })
}

func TestUpsertThenGet(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	p, err := svc.UpsertPolicy(ctx, domain.Hostel, 48, adminID)
	require.NoError(t, err)
	assert.Equal(t, fixedNow, p.UpdatedAt)

	got, err := svc.GetPolicy(ctx, domain.Hostel)
	require.NoError(t, err)
	assert.Equal(t, int32(48), got.WindowHours)
	assert.Equal(t, adminID, got.UpdatedBy)

	// Upsert replaces the window.
	_, err = svc.UpsertPolicy(ctx, domain.Hostel, 24, adminID)
	require.NoError(t, err)
	got, _ = svc.GetPolicy(ctx, domain.Hostel)
	assert.Equal(t, int32(24), got.WindowHours)
}

func TestUpsertPolicy_Errors(t *testing.T) {
	tests := []struct {
		name      string
		category  domain.Category
		hours     int32
		updatedBy string
		wantErr   error
	}{
		{"updated_by missing", domain.Hostel, 24, "", ErrInvalidArgument},
		{"updated_by not a UUID", domain.Hostel, 24, "admin", ErrInvalidArgument},
		{"category unspecified", domain.CategoryUnspecified, 24, adminID, domain.ErrInvalidCategory},
		{"zero window", domain.Hostel, 0, adminID, domain.ErrInvalidWindow},
		{"negative window", domain.Hostel, -5, adminID, domain.ErrInvalidWindow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			_, err := svc.UpsertPolicy(context.Background(), tt.category, tt.hours, tt.updatedBy)
			assert.ErrorIs(t, err, tt.wantErr)

			// A failed upsert saves nothing.
			_, err = svc.GetPolicy(context.Background(), domain.Hostel)
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func TestGetPolicy_Errors(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPolicy(context.Background(), domain.CategoryUnspecified)
	assert.ErrorIs(t, err, domain.ErrInvalidCategory)

	_, err = svc.GetPolicy(context.Background(), domain.Others)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

type failingRepo struct{}

func (failingRepo) Get(context.Context, domain.Category) (domain.Policy, error) {
	return domain.Policy{}, errors.New("db down")
}
func (failingRepo) Save(context.Context, domain.Policy) error { return errors.New("db down") }

func TestUpsertPolicy_SaveError(t *testing.T) {
	svc := NewService(failingRepo{}, time.Now)
	_, err := svc.UpsertPolicy(context.Background(), domain.Hostel, 24, adminID)
	assert.ErrorContains(t, err, "db down")
}
