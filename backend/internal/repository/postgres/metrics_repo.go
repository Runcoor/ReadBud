package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"readbud/internal/domain/metrics"
)

// MetricsSnapshotRepository defines the interface for metrics snapshot data access.
type MetricsSnapshotRepository interface {
	Create(ctx context.Context, snapshot *metrics.MetricsSnapshot) error
	FindByArticleID(ctx context.Context, articleID string) ([]metrics.MetricsSnapshot, error)
	FindByDateRange(ctx context.Context, articleID string, start, end time.Time) ([]metrics.MetricsSnapshot, error)
	GetLatestByArticleID(ctx context.Context, articleID string) (*metrics.MetricsSnapshot, error)
}

type metricsSnapshotRepo struct {
	db *gorm.DB
}

// NewMetricsSnapshotRepository creates a new PostgreSQL-backed metrics snapshot repository.
func NewMetricsSnapshotRepository(db *gorm.DB) MetricsSnapshotRepository {
	return &metricsSnapshotRepo{db: db}
}

func (r *metricsSnapshotRepo) Create(ctx context.Context, snapshot *metrics.MetricsSnapshot) error {
	if err := r.db.WithContext(ctx).Create(snapshot).Error; err != nil {
		return fmt.Errorf("metricsSnapshotRepo.Create: %w", err)
	}
	return nil
}

func (r *metricsSnapshotRepo) FindByArticleID(ctx context.Context, articleID string) ([]metrics.MetricsSnapshot, error) {
	var snapshots []metrics.MetricsSnapshot
	if err := r.db.WithContext(ctx).Where("article_id = ?", articleID).
		Order("metric_date DESC").Find(&snapshots).Error; err != nil {
		return nil, fmt.Errorf("metricsSnapshotRepo.FindByArticleID: %w", err)
	}
	return snapshots, nil
}

func (r *metricsSnapshotRepo) FindByDateRange(ctx context.Context, articleID string, start, end time.Time) ([]metrics.MetricsSnapshot, error) {
	var snapshots []metrics.MetricsSnapshot
	if err := r.db.WithContext(ctx).
		Where("article_id = ? AND metric_date >= ? AND metric_date <= ?", articleID, start, end).
		Order("metric_date ASC").Find(&snapshots).Error; err != nil {
		return nil, fmt.Errorf("metricsSnapshotRepo.FindByDateRange: %w", err)
	}
	return snapshots, nil
}

func (r *metricsSnapshotRepo) GetLatestByArticleID(ctx context.Context, articleID string) (*metrics.MetricsSnapshot, error) {
	var snapshot metrics.MetricsSnapshot
	if err := r.db.WithContext(ctx).Where("article_id = ?", articleID).
		Order("metric_date DESC").First(&snapshot).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("metricsSnapshotRepo.GetLatestByArticleID: %w", err)
	}
	return &snapshot, nil
}
