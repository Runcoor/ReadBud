// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

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
	Upsert(ctx context.Context, snapshot *metrics.MetricsSnapshot) error
	FindByArticleID(ctx context.Context, articleID string) ([]metrics.MetricsSnapshot, error)
	FindByDateRange(ctx context.Context, articleID string, start, end time.Time) ([]metrics.MetricsSnapshot, error)
	GetLatestByArticleID(ctx context.Context, articleID string) (*metrics.MetricsSnapshot, error)
	ListByAccountID(ctx context.Context, accountID int64, limit, offset int) ([]metrics.MetricsSnapshot, error)
	ListDistinctArticleIDs(ctx context.Context, accountID int64) ([]string, error)
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

// Upsert creates or updates a metrics snapshot keyed by (article_id, metric_date).
// If a record already exists for the same article and date, it updates the metrics.
func (r *metricsSnapshotRepo) Upsert(ctx context.Context, snapshot *metrics.MetricsSnapshot) error {
	var existing metrics.MetricsSnapshot
	err := r.db.WithContext(ctx).
		Where("article_id = ? AND metric_date = ?", snapshot.ArticleID, snapshot.MetricDate).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		if createErr := r.db.WithContext(ctx).Create(snapshot).Error; createErr != nil {
			return fmt.Errorf("metricsSnapshotRepo.Upsert(create): %w", createErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("metricsSnapshotRepo.Upsert(find): %w", err)
	}

	// Update existing record
	updates := map[string]interface{}{
		"read_count":       snapshot.ReadCount,
		"read_user_count":  snapshot.ReadUserCount,
		"share_count":      snapshot.ShareCount,
		"share_user_count": snapshot.ShareUserCount,
		"add_fans_count":   snapshot.AddFansCount,
		"cancel_fans_count": snapshot.CancelFansCount,
		"net_fans_count":   snapshot.NetFansCount,
		"raw_json":         snapshot.RawJSON,
		"updated_at":       snapshot.UpdatedAt,
	}

	if updateErr := r.db.WithContext(ctx).Model(&existing).Updates(updates).Error; updateErr != nil {
		return fmt.Errorf("metricsSnapshotRepo.Upsert(update): %w", updateErr)
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

func (r *metricsSnapshotRepo) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) ([]metrics.MetricsSnapshot, error) {
	var snapshots []metrics.MetricsSnapshot
	if err := r.db.WithContext(ctx).
		Where("wechat_account_id = ?", accountID).
		Order("metric_date DESC").
		Limit(limit).Offset(offset).
		Find(&snapshots).Error; err != nil {
		return nil, fmt.Errorf("metricsSnapshotRepo.ListByAccountID: %w", err)
	}
	return snapshots, nil
}

func (r *metricsSnapshotRepo) ListDistinctArticleIDs(ctx context.Context, accountID int64) ([]string, error) {
	var articleIDs []string
	if err := r.db.WithContext(ctx).
		Model(&metrics.MetricsSnapshot{}).
		Where("wechat_account_id = ?", accountID).
		Distinct("article_id").
		Pluck("article_id", &articleIDs).Error; err != nil {
		return nil, fmt.Errorf("metricsSnapshotRepo.ListDistinctArticleIDs: %w", err)
	}
	return articleIDs, nil
}
