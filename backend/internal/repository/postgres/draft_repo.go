package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain/draft"
)

// ArticleDraftRepository defines the interface for article draft data access.
type ArticleDraftRepository interface {
	Create(ctx context.Context, d *draft.ArticleDraft) error
	FindByID(ctx context.Context, id int64) (*draft.ArticleDraft, error)
	FindByPublicID(ctx context.Context, publicID string) (*draft.ArticleDraft, error)
	FindByTaskID(ctx context.Context, taskID int64) ([]draft.ArticleDraft, error)
	FindLatestByTaskID(ctx context.Context, taskID int64) (*draft.ArticleDraft, error)
	Update(ctx context.Context, d *draft.ArticleDraft) error
}

// ArticleBlockRepository defines the interface for article block data access.
type ArticleBlockRepository interface {
	Create(ctx context.Context, b *draft.ArticleBlock) error
	CreateBatch(ctx context.Context, blocks []draft.ArticleBlock) error
	FindByDraftID(ctx context.Context, draftID int64) ([]draft.ArticleBlock, error)
	Update(ctx context.Context, b *draft.ArticleBlock) error
	DeleteByDraftID(ctx context.Context, draftID int64) error
}

// --- ArticleDraft repo impl ---

type draftRepo struct {
	db *gorm.DB
}

// NewArticleDraftRepository creates a new PostgreSQL-backed article draft repository.
func NewArticleDraftRepository(db *gorm.DB) ArticleDraftRepository {
	return &draftRepo{db: db}
}

func (r *draftRepo) Create(ctx context.Context, d *draft.ArticleDraft) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("draftRepo.Create: %w", err)
	}
	return nil
}

func (r *draftRepo) FindByID(ctx context.Context, id int64) (*draft.ArticleDraft, error) {
	var d draft.ArticleDraft
	if err := r.db.WithContext(ctx).First(&d, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("draftRepo.FindByID: %w", err)
	}
	return &d, nil
}

func (r *draftRepo) FindByPublicID(ctx context.Context, publicID string) (*draft.ArticleDraft, error) {
	var d draft.ArticleDraft
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("draftRepo.FindByPublicID: %w", err)
	}
	return &d, nil
}

func (r *draftRepo) FindByTaskID(ctx context.Context, taskID int64) ([]draft.ArticleDraft, error) {
	var drafts []draft.ArticleDraft
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).
		Order("version DESC").Find(&drafts).Error; err != nil {
		return nil, fmt.Errorf("draftRepo.FindByTaskID: %w", err)
	}
	return drafts, nil
}

func (r *draftRepo) FindLatestByTaskID(ctx context.Context, taskID int64) (*draft.ArticleDraft, error) {
	var d draft.ArticleDraft
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).
		Order("version DESC").First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("draftRepo.FindLatestByTaskID: %w", err)
	}
	return &d, nil
}

func (r *draftRepo) Update(ctx context.Context, d *draft.ArticleDraft) error {
	if err := r.db.WithContext(ctx).Save(d).Error; err != nil {
		return fmt.Errorf("draftRepo.Update: %w", err)
	}
	return nil
}

// --- ArticleBlock repo impl ---

type blockRepo struct {
	db *gorm.DB
}

// NewArticleBlockRepository creates a new PostgreSQL-backed article block repository.
func NewArticleBlockRepository(db *gorm.DB) ArticleBlockRepository {
	return &blockRepo{db: db}
}

func (r *blockRepo) CreateBatch(ctx context.Context, blocks []draft.ArticleBlock) error {
	if len(blocks) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&blocks).Error; err != nil {
		return fmt.Errorf("blockRepo.CreateBatch: %w", err)
	}
	return nil
}

func (r *blockRepo) Create(ctx context.Context, b *draft.ArticleBlock) error {
	if err := r.db.WithContext(ctx).Create(b).Error; err != nil {
		return fmt.Errorf("blockRepo.Create: %w", err)
	}
	return nil
}

func (r *blockRepo) FindByDraftID(ctx context.Context, draftID int64) ([]draft.ArticleBlock, error) {
	var blocks []draft.ArticleBlock
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).
		Order("sort_no ASC").Find(&blocks).Error; err != nil {
		return nil, fmt.Errorf("blockRepo.FindByDraftID: %w", err)
	}
	return blocks, nil
}

func (r *blockRepo) Update(ctx context.Context, b *draft.ArticleBlock) error {
	if err := r.db.WithContext(ctx).Save(b).Error; err != nil {
		return fmt.Errorf("blockRepo.Update: %w", err)
	}
	return nil
}

func (r *blockRepo) DeleteByDraftID(ctx context.Context, draftID int64) error {
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).Delete(&draft.ArticleBlock{}).Error; err != nil {
		return fmt.Errorf("blockRepo.DeleteByDraftID: %w", err)
	}
	return nil
}
