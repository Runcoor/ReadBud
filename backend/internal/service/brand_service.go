package service

import (
	"context"
	"fmt"

	"readbud/internal/domain"
	"readbud/internal/repository/postgres"
)

// BrandProfileService handles brand profile business logic.
type BrandProfileService struct {
	repo postgres.BrandProfileRepository
}

// NewBrandProfileService creates a new BrandProfileService.
func NewBrandProfileService(repo postgres.BrandProfileRepository) *BrandProfileService {
	return &BrandProfileService{repo: repo}
}

// Create creates a new brand profile.
func (s *BrandProfileService) Create(ctx context.Context, bp *domain.BrandProfile) error {
	if err := s.repo.Create(ctx, bp); err != nil {
		return fmt.Errorf("brandProfileService.Create: %w", err)
	}
	return nil
}

// Get returns a brand profile by its public ID.
func (s *BrandProfileService) Get(ctx context.Context, publicID string) (*domain.BrandProfile, error) {
	bp, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("brandProfileService.Get: %w", err)
	}
	if bp == nil {
		return nil, ErrNotFound
	}
	return bp, nil
}

// Update updates an existing brand profile identified by public ID.
func (s *BrandProfileService) Update(ctx context.Context, publicID string, updates *domain.BrandProfile) (*domain.BrandProfile, error) {
	bp, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("brandProfileService.Update: %w", err)
	}
	if bp == nil {
		return nil, ErrNotFound
	}

	bp.Name = updates.Name
	bp.BrandTone = updates.BrandTone
	bp.ForbiddenWords = updates.ForbiddenWords
	bp.PreferredWords = updates.PreferredWords
	bp.CTARules = updates.CTARules
	bp.CoverStyleRules = updates.CoverStyleRules
	bp.ImageStyleRules = updates.ImageStyleRules

	if err := s.repo.Update(ctx, bp); err != nil {
		return nil, fmt.Errorf("brandProfileService.Update: %w", err)
	}
	return bp, nil
}

// List returns all brand profiles.
func (s *BrandProfileService) List(ctx context.Context) ([]domain.BrandProfile, error) {
	profiles, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("brandProfileService.List: %w", err)
	}
	return profiles, nil
}

// GetDefault returns the default brand profile.
func (s *BrandProfileService) GetDefault(ctx context.Context) (*domain.BrandProfile, error) {
	bp, err := s.repo.FindDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("brandProfileService.GetDefault: %w", err)
	}
	if bp == nil {
		return nil, ErrNotFound
	}
	return bp, nil
}
