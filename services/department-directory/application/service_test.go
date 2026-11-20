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
	assert.Len(t, got, 6)
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

	got, err = newService().GetStaff(ctx, "SSO_FEMALE")
	require.NoError(t, err)
	assert.Equal(t, "Student Service Officer (female)", got.Title)
	assert.Empty(t, got.HolderUserID, "nobody holds it")

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

func TestListPositions(t *testing.T) {
	all, err := newService().ListPositions(context.Background(), false)
	require.NoError(t, err)
	assert.Len(t, all, 23, "3 Category Heads, 2 staff, 12 councillors, 6 HoDs")

	held, err := newService().ListPositions(context.Background(), true)
	require.NoError(t, err)
	codes := make([]string, 0, len(held))
	for _, p := range held {
		assert.NotEmpty(t, p.HolderUserID)
		codes = append(codes, p.Code)
	}
	assert.Equal(t, []string{
		"BLOCK_COUNCILLOR_HA", "BLOCK_COUNCILLOR_HB", "BLOCK_COUNCILLOR_HC", "BLOCK_COUNCILLOR_HD",
		"BLOCK_COUNCILLOR_HE", "BLOCK_COUNCILLOR_HF", "BLOCK_COUNCILLOR_NK", "BLOCK_COUNCILLOR_RKA",
		"BLOCK_COUNCILLOR_RKB", "DAA", "DEPUTY_CHIEF_COUNCILLOR_MESS", "DSA", "GAMES_HEALTH_COUNCILLOR",
		"HOD_ARCH", "HOD_CIVIL", "HOD_CT", "HOD_EEE", "HOD_MECH", "HOD_SH", "MAINTENANCE_COUNCILLOR",
		"SSO", "SSO_MALE",
	}, codes, "ordered by code")

	_, err = NewService(failingRepo{}).ListPositions(context.Background(), true)
	assert.ErrorIs(t, err, errDB)
}
