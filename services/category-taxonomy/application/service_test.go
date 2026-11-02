package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

// failingRepo is a test double whose every call fails.
type failingRepo struct{}

var errDB = errors.New("database down")

func (failingRepo) List(context.Context) ([]domain.Info, error) { return nil, errDB }
func (failingRepo) Save(context.Context, domain.Info) error     { return errDB }

// saveFailsRepo lists normally but cannot save.
type saveFailsRepo struct{ *memory.Repository }

func (saveFailsRepo) Save(context.Context, domain.Info) error { return errDB }

func newService() *Service {
	return NewService(memory.NewRepository(domain.DefaultCategories()))
}

func TestListCategories(t *testing.T) {
	got, err := newService().ListCategories(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, domain.Hostel, got[0].Category)
	assert.Equal(t, domain.Others, got[2].Category)
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		category domain.Category
		wantHead string
		wantErr  error
	}{
		{name: "hostel", category: domain.Hostel, wantHead: "DSA"},
		{name: "academic block", category: domain.AcademicBlock, wantHead: "DAA"},
		{name: "others", category: domain.Others, wantHead: "SSO"},
		{name: "unspecified", category: "", wantErr: domain.ErrInvalidCategory},
		{name: "unknown", category: "LIBRARY", wantErr: domain.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newService().GetCategory(context.Background(), tt.category)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantHead, got.HeadPosition)
		})
	}
}

func TestUpsertCategory(t *testing.T) {
	ctx := context.Background()
	s := newService()

	got, err := s.UpsertCategory(ctx, UpsertCommand{Category: domain.Hostel, Name: "Hostels", ActorRole: domain.SystemAdmin})
	require.NoError(t, err)
	assert.Equal(t, "Hostels", got.Name)

	stored, err := s.GetCategory(ctx, domain.Hostel)
	require.NoError(t, err)
	assert.Equal(t, "Hostels", stored.Name, "change is saved")

	_, err = s.UpsertCategory(ctx, UpsertCommand{Category: domain.Hostel, Name: "x", ActorRole: "STUDENT"})
	assert.ErrorIs(t, err, domain.ErrPermissionDenied)
	_, err = s.UpsertCategory(ctx, UpsertCommand{Category: domain.Hostel, Name: " ", ActorRole: domain.SystemAdmin})
	assert.ErrorIs(t, err, domain.ErrInvalidName)
	_, err = s.UpsertCategory(ctx, UpsertCommand{Category: "LIBRARY", Name: "Library", ActorRole: domain.SystemAdmin})
	assert.ErrorIs(t, err, domain.ErrNotFound, "new categories cannot be added")
}

func TestRepositoryErrorsArePassedOn(t *testing.T) {
	ctx := context.Background()
	s := NewService(failingRepo{})
	_, err := s.ListCategories(ctx)
	assert.ErrorIs(t, err, errDB)
	_, err = s.GetCategory(ctx, domain.Hostel)
	assert.ErrorIs(t, err, errDB)

	s = NewService(saveFailsRepo{memory.NewRepository(domain.DefaultCategories())})
	_, err = s.UpsertCategory(ctx, UpsertCommand{Category: domain.Hostel, Name: "Hostels", ActorRole: domain.SystemAdmin})
	assert.ErrorIs(t, err, errDB)
}
