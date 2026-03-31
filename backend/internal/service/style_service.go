package service

import (
	"context"
	"fmt"

	"readbud/internal/domain"
	"readbud/internal/repository/postgres"
)

// StyleProfileService handles style profile business logic.
type StyleProfileService struct {
	repo postgres.StyleProfileRepository
}

// NewStyleProfileService creates a new StyleProfileService.
func NewStyleProfileService(repo postgres.StyleProfileRepository) *StyleProfileService {
	return &StyleProfileService{repo: repo}
}

// Create creates a new style profile.
func (s *StyleProfileService) Create(ctx context.Context, sp *domain.StyleProfile) error {
	if err := s.repo.Create(ctx, sp); err != nil {
		return fmt.Errorf("styleProfileService.Create: %w", err)
	}
	return nil
}

// Get returns a style profile by its public ID.
func (s *StyleProfileService) Get(ctx context.Context, publicID string) (*domain.StyleProfile, error) {
	sp, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("styleProfileService.Get: %w", err)
	}
	if sp == nil {
		return nil, ErrNotFound
	}
	return sp, nil
}

// Update updates an existing style profile identified by public ID.
func (s *StyleProfileService) Update(ctx context.Context, publicID string, updates *domain.StyleProfile) (*domain.StyleProfile, error) {
	sp, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("styleProfileService.Update: %w", err)
	}
	if sp == nil {
		return nil, ErrNotFound
	}

	sp.Name = updates.Name
	sp.ApplicableScene = updates.ApplicableScene
	sp.OpeningTemplate = updates.OpeningTemplate
	sp.StructureTemplate = updates.StructureTemplate
	sp.ClosingTemplate = updates.ClosingTemplate
	sp.SentencePreference = updates.SentencePreference
	sp.TitlePreference = updates.TitlePreference

	if err := s.repo.Update(ctx, sp); err != nil {
		return nil, fmt.Errorf("styleProfileService.Update: %w", err)
	}
	return sp, nil
}

// List returns all style profiles.
func (s *StyleProfileService) List(ctx context.Context) ([]domain.StyleProfile, error) {
	profiles, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("styleProfileService.List: %w", err)
	}
	return profiles, nil
}

// GetDefault returns the default style profile.
func (s *StyleProfileService) GetDefault(ctx context.Context) (*domain.StyleProfile, error) {
	sp, err := s.repo.FindDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("styleProfileService.GetDefault: %w", err)
	}
	if sp == nil {
		return nil, ErrNotFound
	}
	return sp, nil
}
