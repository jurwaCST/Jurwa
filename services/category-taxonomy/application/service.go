package application

import (
	"context"

	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/ports"
)

// Service holds the category-taxonomy use cases.
type Service struct {
	repo ports.CategoryRepository
}

func NewService(repo ports.CategoryRepository) *Service {
	return &Service{repo: repo}
}

// ListCategories returns the categories in display order, for the PWA form.
func (s *Service) ListCategories(ctx context.Context) ([]domain.Info, error) {
	return s.repo.List(ctx)
}

// GetCategory returns one category with its display name and Category Head.
func (s *Service) GetCategory(ctx context.Context, c domain.Category) (domain.Info, error) {
	if c == "" {
		return domain.Info{}, domain.ErrInvalidCategory
	}
	all, err := s.repo.List(ctx)
	if err != nil {
		return domain.Info{}, err
	}
	return domain.Find(all, c)
}

// UpsertCommand is an administrator's change to one category (SCRUM-64).
type UpsertCommand struct {
	Category    domain.Category
	Name        string
	Description string
	ActorRole   domain.Role
}

// UpsertCategory renames an existing category. Only a System Administrator may do it.
func (s *Service) UpsertCategory(ctx context.Context, cmd UpsertCommand) (domain.Info, error) {
	current, err := s.GetCategory(ctx, cmd.Category)
	if err != nil {
		return domain.Info{}, err
	}
	updated, err := domain.Rename(current, cmd.Name, cmd.Description, cmd.ActorRole)
	if err != nil {
		return domain.Info{}, err
	}
	if err := s.repo.Save(ctx, updated); err != nil {
		return domain.Info{}, err
	}
	return updated, nil
}
