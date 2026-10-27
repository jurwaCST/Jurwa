package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/routing-rules/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/routing-rules/domain"
)

// failingRepo is a test double whose every call fails.
type failingRepo struct{}

var errDB = errors.New("database down")

func (failingRepo) FindByCategory(context.Context, domain.Category) ([]domain.Rule, error) {
	return nil, errDB
}
func (failingRepo) List(context.Context, bool) ([]domain.Rule, error) { return nil, errDB }
func (failingRepo) Upsert(context.Context, domain.Rule) (domain.Rule, error) {
	return domain.Rule{}, errDB
}

func newService() *Service {
	s := NewService(memory.NewRepository(domain.DefaultRules()))
	s.now = func() time.Time { return time.Date(2026, 10, 27, 9, 0, 0, 0, time.UTC) }
	return s
}

func TestGetRule(t *testing.T) {
	tests := []struct {
		name    string
		query   domain.Query
		want    string
		wantErr error
	}{
		{name: "hostel", query: domain.Query{Category: domain.Hostel, Building: "Hostel B"}, want: "DSA"},
		{name: "academic block", query: domain.Query{Category: domain.AcademicBlock}, want: "DAA"},
		{name: "others", query: domain.Query{Category: domain.Others}, want: "SSO"},
		{name: "unspecified", query: domain.Query{}, wantErr: domain.ErrInvalidCategory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newService().GetRule(context.Background(), tt.query)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.TargetPosition)
		})
	}
}

func TestUpsertRule(t *testing.T) {
	ctx := context.Background()
	s := newService()

	saved, err := s.UpsertRule(ctx, domain.Rule{Category: domain.Hostel, Building: "Hostel B", TargetPosition: "SSO", Active: true})
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultPriority, saved.Priority, "priority 0 becomes the default")
	assert.Equal(t, s.now(), saved.UpdatedAt)

	got, err := s.GetRule(ctx, domain.Query{Category: domain.Hostel, Building: "Hostel B"})
	require.NoError(t, err)
	assert.Equal(t, "SSO", got.TargetPosition, "new building rule is used")

	_, err = s.UpsertRule(ctx, domain.Rule{Category: domain.Hostel})
	assert.ErrorIs(t, err, domain.ErrInvalidRule)

	rules, err := s.ListRules(ctx, false)
	require.NoError(t, err)
	assert.Len(t, rules, 4)
}

func TestRepositoryErrorsArePassedOn(t *testing.T) {
	ctx := context.Background()
	s := NewService(failingRepo{})

	_, err := s.GetRule(ctx, domain.Query{Category: domain.Hostel})
	assert.ErrorIs(t, err, errDB)
	_, err = s.ListRules(ctx, true)
	assert.ErrorIs(t, err, errDB)
	_, err = s.UpsertRule(ctx, domain.Rule{Category: domain.Hostel, TargetPosition: "DSA"})
	assert.ErrorIs(t, err, errDB)
}
