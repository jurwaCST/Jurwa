package application

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/department-directory/domain"
	"github.com/jurwaCST/Jurwa/services/department-directory/ports"
)

// Service holds the department-directory use cases.
type Service struct {
	repo ports.DirectoryRepository
}

func NewService(repo ports.DirectoryRepository) *Service {
	return &Service{repo: repo}
}

// ListDepartments returns the 5 academic departments.
func (s *Service) ListDepartments(ctx context.Context) ([]domain.Department, error) {
	return s.repo.ListDepartments(ctx)
}

// GetHoD returns the Head of Department position for one department.
func (s *Service) GetHoD(ctx context.Context, deptCode string) (domain.Position, error) {
	depts, err := s.repo.ListDepartments(ctx)
	if err != nil {
		return domain.Position{}, err
	}
	positions, err := s.repo.ListPositions(ctx)
	if err != nil {
		return domain.Position{}, err
	}
	return domain.FindHoD(depts, positions, deptCode)
}

// GetStaff returns one position and its holder, for example "DSA".
func (s *Service) GetStaff(ctx context.Context, positionCode string) (domain.Position, error) {
	positions, err := s.repo.ListPositions(ctx)
	if err != nil {
		return domain.Position{}, err
	}
	return domain.FindPosition(positions, positionCode)
}

// ListPositions returns every position ordered by code, or with heldOnly only
// the ones that have a holder.
func (s *Service) ListPositions(ctx context.Context, heldOnly bool) ([]domain.Position, error) {
	positions, err := s.repo.ListPositions(ctx)
	if err != nil {
		return nil, err
	}
	if heldOnly {
		return domain.HeldPositions(positions), nil
	}
	return positions, nil
}
