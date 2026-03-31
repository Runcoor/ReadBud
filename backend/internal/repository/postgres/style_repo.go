package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// StyleProfileRepository defines the interface for style profile data access.
type StyleProfileRepository interface {
	Create(ctx context.Context, sp *domain.StyleProfile) error
	FindByID(ctx context.Context, id int64) (*domain.StyleProfile, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.StyleProfile, error)
	FindDefault(ctx context.Context) (*domain.StyleProfile, error)
	Update(ctx context.Context, sp *domain.StyleProfile) error
	List(ctx context.Context) ([]domain.StyleProfile, error)
}

type styleRepo struct {
	db *gorm.DB
}

// NewStyleProfileRepository creates a new PostgreSQL-backed style profile repository.
func NewStyleProfileRepository(db *gorm.DB) StyleProfileRepository {
	return &styleRepo{db: db}
}

func (r *styleRepo) Create(ctx context.Context, sp *domain.StyleProfile) error {
	if err := r.db.WithContext(ctx).Create(sp).Error; err != nil {
		return fmt.Errorf("styleRepo.Create: %w", err)
	}
	return nil
}

func (r *styleRepo) FindByID(ctx context.Context, id int64) (*domain.StyleProfile, error) {
	var sp domain.StyleProfile
	if err := r.db.WithContext(ctx).First(&sp, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("styleRepo.FindByID: %w", err)
	}
	return &sp, nil
}

func (r *styleRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.StyleProfile, error) {
	var sp domain.StyleProfile
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&sp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("styleRepo.FindByPublicID: %w", err)
	}
	return &sp, nil
}

// FindDefault returns the first style profile ordered by ID (earliest created).
func (r *styleRepo) FindDefault(ctx context.Context) (*domain.StyleProfile, error) {
	var sp domain.StyleProfile
	if err := r.db.WithContext(ctx).Order("id ASC").First(&sp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("styleRepo.FindDefault: %w", err)
	}
	return &sp, nil
}

func (r *styleRepo) Update(ctx context.Context, sp *domain.StyleProfile) error {
	if err := r.db.WithContext(ctx).Save(sp).Error; err != nil {
		return fmt.Errorf("styleRepo.Update: %w", err)
	}
	return nil
}

func (r *styleRepo) List(ctx context.Context) ([]domain.StyleProfile, error) {
	var profiles []domain.StyleProfile
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("styleRepo.List: %w", err)
	}
	return profiles, nil
}
