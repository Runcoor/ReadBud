package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain/source"
)

// SourceDocumentRepository defines the interface for source document data access.
type SourceDocumentRepository interface {
	Create(ctx context.Context, doc *source.SourceDocument) error
	CreateBatch(ctx context.Context, docs []source.SourceDocument) error
	FindByID(ctx context.Context, id int64) (*source.SourceDocument, error)
	FindByTaskID(ctx context.Context, taskID int64) ([]source.SourceDocument, error)
	FindByTaskIDOrderByScore(ctx context.Context, taskID int64, limit int) ([]source.SourceDocument, error)
	Update(ctx context.Context, doc *source.SourceDocument) error
	DeleteByTaskID(ctx context.Context, taskID int64) error
}

type sourceRepo struct {
	db *gorm.DB
}

// NewSourceDocumentRepository creates a new PostgreSQL-backed source document repository.
func NewSourceDocumentRepository(db *gorm.DB) SourceDocumentRepository {
	return &sourceRepo{db: db}
}

func (r *sourceRepo) Create(ctx context.Context, doc *source.SourceDocument) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return fmt.Errorf("sourceRepo.Create: %w", err)
	}
	return nil
}

func (r *sourceRepo) CreateBatch(ctx context.Context, docs []source.SourceDocument) error {
	if len(docs) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&docs).Error; err != nil {
		return fmt.Errorf("sourceRepo.CreateBatch: %w", err)
	}
	return nil
}

func (r *sourceRepo) FindByID(ctx context.Context, id int64) (*source.SourceDocument, error) {
	var doc source.SourceDocument
	if err := r.db.WithContext(ctx).First(&doc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("sourceRepo.FindByID: %w", err)
	}
	return &doc, nil
}

func (r *sourceRepo) FindByTaskID(ctx context.Context, taskID int64) ([]source.SourceDocument, error) {
	var docs []source.SourceDocument
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("hot_score DESC").Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("sourceRepo.FindByTaskID: %w", err)
	}
	return docs, nil
}

func (r *sourceRepo) FindByTaskIDOrderByScore(ctx context.Context, taskID int64, limit int) ([]source.SourceDocument, error) {
	var docs []source.SourceDocument
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).
		Order("hot_score DESC").Limit(limit).Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("sourceRepo.FindByTaskIDOrderByScore: %w", err)
	}
	return docs, nil
}

func (r *sourceRepo) Update(ctx context.Context, doc *source.SourceDocument) error {
	if err := r.db.WithContext(ctx).Save(doc).Error; err != nil {
		return fmt.Errorf("sourceRepo.Update: %w", err)
	}
	return nil
}

func (r *sourceRepo) DeleteByTaskID(ctx context.Context, taskID int64) error {
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Delete(&source.SourceDocument{}).Error; err != nil {
		return fmt.Errorf("sourceRepo.DeleteByTaskID: %w", err)
	}
	return nil
}
