package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jurwaCST/Jurwa/services/department-directory/adapters/out/memory"
	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
)

var errDB = errors.New("database down")

// failingRepo is a test double whose every call fails.
type failingRepo struct{}

func (failingRepo) ListDepartments(context.Context) ([]domain.Department, error) { return nil, errDB }
func (failingRepo) ListPositions(context.Context) ([]domain.Position, error)     { return nil, errDB }

// positionsFailRepo lists departments but not positions.
type positionsFailRepo struct{ failingRepo }

func (positionsFailRepo) ListDepartments(context.Context) ([]domain.Department, error) {
	return domain.DefaultDepartments(), nil
}

func newService() *Service {
	return NewService(memory.NewRepository(domain.DefaultDepartments(), domain.WithDemoHolders(domain.DefaultPositions())))
}

func TestListDepartments(t *testing.T) {
	got, err := newService().ListDepartments(context.Background())
	require.NoError(t, err)
	assert.Len(t, got, 5)
}

func TestGetHoD(t *testing.T) {
	ctx := context.Background()
	got, err := newService().GetHoD(ctx, "CT")
	require.NoError(t, err)
	assert.Equal(t, "HOD_CT", got.Code)

	_, err = newService().GetHoD(ctx, "")
	assert.ErrorIs(t, err, domain.ErrInvalidCode)
	_, err = newService().GetHoD(ctx, "PHYSICS")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetStaff(t *testing.T) {
	ctx := context.Background()
	got, err := newService().GetStaff(ctx, "DSA")
	require.NoError(t, err)
	assert.Equal(t, domain.KindCategoryHead, got.Kind)
	assert.NotEmpty(t, got.HolderUserID)

	got, err = newService().GetStaff(ctx, "CHIEF_COUNCILLOR")
	require.NoError(t, err)
	assert.Equal(t, "Chief Councillor", got.Title)

	_, err = newService().GetStaff(ctx, "PRINCIPAL")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRepositoryErrorsArePassedOn(t *testing.T) {
	ctx := context.Background()
	s := NewService(failingRepo{})
	_, err := s.ListDepartments(ctx)
	assert.ErrorIs(t, err, errDB)
	_, err = s.GetHoD(ctx, "ARCH")
	assert.ErrorIs(t, err, errDB)
	_, err = s.GetStaff(ctx, "DSA")
	assert.ErrorIs(t, err, errDB)

	_, err = NewService(positionsFailRepo{}).GetHoD(ctx, "ARCH")
	assert.ErrorIs(t, err, errDB)
}
