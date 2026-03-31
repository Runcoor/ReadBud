package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// BrandProfileRepository defines the interface for brand profile data access.
type BrandProfileRepository interface {
	Create(ctx context.Context, bp *domain.BrandProfile) error
	FindByID(ctx context.Context, id int64) (*domain.BrandProfile, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.BrandProfile, error)
	FindDefault(ctx context.Context) (*domain.BrandProfile, error)
	Update(ctx context.Context, bp *domain.BrandProfile) error
	List(ctx context.Context) ([]domain.BrandProfile, error)
}

type brandRepo struct {
	db *gorm.DB
}

// NewBrandProfileRepository creates a new PostgreSQL-backed brand profile repository.
func NewBrandProfileRepository(db *gorm.DB) BrandProfileRepository {
	return &brandRepo{db: db}
}

func (r *brandRepo) Create(ctx context.Context, bp *domain.BrandProfile) error {
	if err := r.db.WithContext(ctx).Create(bp).Error; err != nil {
		return fmt.Errorf("brandRepo.Create: %w", err)
	}
	return nil
}

func (r *brandRepo) FindByID(ctx context.Context, id int64) (*domain.BrandProfile, error) {
	var bp domain.BrandProfile
	if err := r.db.WithContext(ctx).First(&bp, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("brandRepo.FindByID: %w", err)
	}
	return &bp, nil
}

func (r *brandRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.BrandProfile, error) {
	var bp domain.BrandProfile
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&bp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("brandRepo.FindByPublicID: %w", err)
	}
	return &bp, nil
}

// FindDefault returns the first brand profile ordered by ID (earliest created).
func (r *brandRepo) FindDefault(ctx context.Context) (*domain.BrandProfile, error) {
	var bp domain.BrandProfile
	if err := r.db.WithContext(ctx).Order("id ASC").First(&bp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("brandRepo.FindDefault: %w", err)
	}
	return &bp, nil
}

func (r *brandRepo) Update(ctx context.Context, bp *domain.BrandProfile) error {
	if err := r.db.WithContext(ctx).Save(bp).Error; err != nil {
		return fmt.Errorf("brandRepo.Update: %w", err)
	}
	return nil
}

func (r *brandRepo) List(ctx context.Context) ([]domain.BrandProfile, error) {
	var profiles []domain.BrandProfile
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("brandRepo.List: %w", err)
	}
	return profiles, nil
}
