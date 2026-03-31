package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain/asset"
)

// AssetRepository defines the interface for asset data access.
type AssetRepository interface {
	Create(ctx context.Context, a *asset.Asset) error
	FindByID(ctx context.Context, id int64) (*asset.Asset, error)
	FindByPublicID(ctx context.Context, publicID string) (*asset.Asset, error)
	FindBySHA256(ctx context.Context, sha256 string) (*asset.Asset, error)
	FindByDraftID(ctx context.Context, draftID int64) ([]asset.Asset, error)
	Update(ctx context.Context, a *asset.Asset) error
	Delete(ctx context.Context, id int64) error
}

type assetRepo struct {
	db *gorm.DB
}

// NewAssetRepository creates a new PostgreSQL-backed asset repository.
func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepo{db: db}
}

func (r *assetRepo) Create(ctx context.Context, a *asset.Asset) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("assetRepo.Create: %w", err)
	}
	return nil
}

func (r *assetRepo) FindByID(ctx context.Context, id int64) (*asset.Asset, error) {
	var a asset.Asset
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("assetRepo.FindByID: %w", err)
	}
	return &a, nil
}

func (r *assetRepo) FindByPublicID(ctx context.Context, publicID string) (*asset.Asset, error) {
	var a asset.Asset
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("assetRepo.FindByPublicID: %w", err)
	}
	return &a, nil
}

func (r *assetRepo) FindBySHA256(ctx context.Context, sha256 string) (*asset.Asset, error) {
	var a asset.Asset
	if err := r.db.WithContext(ctx).Where("sha256 = ?", sha256).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("assetRepo.FindBySHA256: %w", err)
	}
	return &a, nil
}

func (r *assetRepo) FindByDraftID(ctx context.Context, draftID int64) ([]asset.Asset, error) {
	var assets []asset.Asset
	// Find assets linked to this draft via article_blocks
	if err := r.db.WithContext(ctx).
		Where("id IN (SELECT asset_id FROM article_blocks WHERE draft_id = ? AND asset_id IS NOT NULL)", draftID).
		Find(&assets).Error; err != nil {
		return nil, fmt.Errorf("assetRepo.FindByDraftID: %w", err)
	}
	return assets, nil
}

func (r *assetRepo) Update(ctx context.Context, a *asset.Asset) error {
	if err := r.db.WithContext(ctx).Save(a).Error; err != nil {
		return fmt.Errorf("assetRepo.Update: %w", err)
	}
	return nil
}

func (r *assetRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&asset.Asset{}, id).Error; err != nil {
		return fmt.Errorf("assetRepo.Delete: %w", err)
	}
	return nil
}
