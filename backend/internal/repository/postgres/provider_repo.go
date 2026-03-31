package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// ProviderConfigRepository defines the interface for provider config data access.
type ProviderConfigRepository interface {
	List(ctx context.Context) ([]domain.ProviderConfig, error)
	FindByID(ctx context.Context, id int64) (*domain.ProviderConfig, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.ProviderConfig, error)
	FindByType(ctx context.Context, providerType string) ([]domain.ProviderConfig, error)
	Create(ctx context.Context, cfg *domain.ProviderConfig) error
	Update(ctx context.Context, cfg *domain.ProviderConfig) error
	Delete(ctx context.Context, id int64) error
}

type providerRepo struct {
	db *gorm.DB
}

// NewProviderConfigRepository creates a new PostgreSQL-backed provider config repository.
func NewProviderConfigRepository(db *gorm.DB) ProviderConfigRepository {
	return &providerRepo{db: db}
}

func (r *providerRepo) List(ctx context.Context) ([]domain.ProviderConfig, error) {
	var configs []domain.ProviderConfig
	if err := r.db.WithContext(ctx).Order("provider_type, provider_name").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("providerRepo.List: %w", err)
	}
	return configs, nil
}

func (r *providerRepo) FindByID(ctx context.Context, id int64) (*domain.ProviderConfig, error) {
	var cfg domain.ProviderConfig
	if err := r.db.WithContext(ctx).First(&cfg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("providerRepo.FindByID: %w", err)
	}
	return &cfg, nil
}

func (r *providerRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.ProviderConfig, error) {
	var cfg domain.ProviderConfig
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("providerRepo.FindByPublicID: %w", err)
	}
	return &cfg, nil
}

func (r *providerRepo) FindByType(ctx context.Context, providerType string) ([]domain.ProviderConfig, error) {
	var configs []domain.ProviderConfig
	if err := r.db.WithContext(ctx).Where("provider_type = ? AND status = ?", providerType, domain.StatusActive).Order("is_default DESC, id ASC").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("providerRepo.FindByType: %w", err)
	}
	return configs, nil
}

func (r *providerRepo) Create(ctx context.Context, cfg *domain.ProviderConfig) error {
	if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return fmt.Errorf("providerRepo.Create: %w", err)
	}
	return nil
}

func (r *providerRepo) Update(ctx context.Context, cfg *domain.ProviderConfig) error {
	if err := r.db.WithContext(ctx).Save(cfg).Error; err != nil {
		return fmt.Errorf("providerRepo.Update: %w", err)
	}
	return nil
}

func (r *providerRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&domain.ProviderConfig{}, id).Error; err != nil {
		return fmt.Errorf("providerRepo.Delete: %w", err)
	}
	return nil
}
