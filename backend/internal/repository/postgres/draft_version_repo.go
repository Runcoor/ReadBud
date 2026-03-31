package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// DraftVersionRepository defines the interface for draft version data access.
type DraftVersionRepository interface {
	Create(ctx context.Context, v *domain.DraftVersion) error
	FindByID(ctx context.Context, id int64) (*domain.DraftVersion, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.DraftVersion, error)
	ListByDraftID(ctx context.Context, draftID int64, limit int) ([]domain.DraftVersion, error)
	GetLatestVersion(ctx context.Context, draftID int64) (*domain.DraftVersion, error)
	CountByDraftID(ctx context.Context, draftID int64) (int64, error)
	DeleteOldVersions(ctx context.Context, draftID int64, keepLast int) error
}

type draftVersionRepo struct {
	db *gorm.DB
}

// NewDraftVersionRepository creates a new PostgreSQL-backed draft version repository.
func NewDraftVersionRepository(db *gorm.DB) DraftVersionRepository {
	return &draftVersionRepo{db: db}
}

func (r *draftVersionRepo) Create(ctx context.Context, v *domain.DraftVersion) error {
	if err := r.db.WithContext(ctx).Create(v).Error; err != nil {
		return fmt.Errorf("draftVersionRepo.Create: %w", err)
	}
	return nil
}

func (r *draftVersionRepo) FindByID(ctx context.Context, id int64) (*domain.DraftVersion, error) {
	var v domain.DraftVersion
	if err := r.db.WithContext(ctx).First(&v, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("draftVersionRepo.FindByID: %w", err)
	}
	return &v, nil
}

func (r *draftVersionRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.DraftVersion, error) {
	var v domain.DraftVersion
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("draftVersionRepo.FindByPublicID: %w", err)
	}
	return &v, nil
}

func (r *draftVersionRepo) ListByDraftID(ctx context.Context, draftID int64, limit int) ([]domain.DraftVersion, error) {
	var versions []domain.DraftVersion
	q := r.db.WithContext(ctx).Where("draft_id = ?", draftID).Order("version_no DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("draftVersionRepo.ListByDraftID: %w", err)
	}
	return versions, nil
}

func (r *draftVersionRepo) GetLatestVersion(ctx context.Context, draftID int64) (*domain.DraftVersion, error) {
	var v domain.DraftVersion
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).
		Order("version_no DESC").First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("draftVersionRepo.GetLatestVersion: %w", err)
	}
	return &v, nil
}

func (r *draftVersionRepo) CountByDraftID(ctx context.Context, draftID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.DraftVersion{}).
		Where("draft_id = ?", draftID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("draftVersionRepo.CountByDraftID: %w", err)
	}
	return count, nil
}

func (r *draftVersionRepo) DeleteOldVersions(ctx context.Context, draftID int64, keepLast int) error {
	// Delete versions older than the most recent `keepLast` versions.
	subQuery := r.db.WithContext(ctx).Model(&domain.DraftVersion{}).
		Select("id").
		Where("draft_id = ?", draftID).
		Order("version_no DESC").
		Limit(keepLast)

	if err := r.db.WithContext(ctx).
		Where("draft_id = ? AND id NOT IN (?)", draftID, subQuery).
		Delete(&domain.DraftVersion{}).Error; err != nil {
		return fmt.Errorf("draftVersionRepo.DeleteOldVersions: %w", err)
	}
	return nil
}
