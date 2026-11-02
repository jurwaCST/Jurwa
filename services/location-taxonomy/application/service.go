package application

import (
	"context"
	"fmt"

	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
	"github.com/jurwaCST/Jurwa/services/location-taxonomy/ports"
)

// Service holds the location-taxonomy use cases.
type Service struct {
	repo ports.LocationRepository
}

// NewService wires the use cases to their port.
func NewService(repo ports.LocationRepository) *Service {
	return &Service{repo: repo}
}

// ListLocations returns the active locations of one category, for the PWA drop-downs.
// A bad category wraps domain.ErrInvalidInput.
func (s *Service) ListLocations(ctx context.Context, category domain.Category) ([]domain.Entry, error) {
	if err := domain.ValidateCategory(category); err != nil {
		return nil, err
	}
	entries, err := s.repo.ListByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	return domain.ActiveOnly(entries), nil
}

// ValidateLocation says whether loc is an active location of the category.
// reason is "" when valid. A bad category or empty building wraps domain.ErrInvalidInput.
func (s *Service) ValidateLocation(ctx context.Context, category domain.Category, loc domain.Location) (bool, string, error) {
	if err := domain.ValidateCategory(category); err != nil {
		return false, "", err
	}
	entries, err := s.repo.ListByCategory(ctx, category)
	if err != nil {
		return false, "", fmt.Errorf("list locations: %w", err)
	}
	return domain.CheckLocation(entries, loc)
}
