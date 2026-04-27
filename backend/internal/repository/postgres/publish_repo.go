// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain/publish"
)

// ---------- PublishJob Repository ----------

// PublishJobRepository defines the interface for publish job data access.
type PublishJobRepository interface {
	Create(ctx context.Context, job *publish.PublishJob) error
	FindByID(ctx context.Context, id int64) (*publish.PublishJob, error)
	FindByPublicID(ctx context.Context, publicID string) (*publish.PublishJob, error)
	FindByDraftID(ctx context.Context, draftID int64) ([]publish.PublishJob, error)
	Update(ctx context.Context, job *publish.PublishJob) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]publish.PublishJob, error)
}

type publishJobRepo struct {
	db *gorm.DB
}

// NewPublishJobRepository creates a new PostgreSQL-backed publish job repository.
func NewPublishJobRepository(db *gorm.DB) PublishJobRepository {
	return &publishJobRepo{db: db}
}

func (r *publishJobRepo) Create(ctx context.Context, job *publish.PublishJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("publishJobRepo.Create: %w", err)
	}
	return nil
}

func (r *publishJobRepo) FindByID(ctx context.Context, id int64) (*publish.PublishJob, error) {
	var job publish.PublishJob
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("publishJobRepo.FindByID: %w", err)
	}
	return &job, nil
}

func (r *publishJobRepo) FindByPublicID(ctx context.Context, publicID string) (*publish.PublishJob, error) {
	var job publish.PublishJob
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("publishJobRepo.FindByPublicID: %w", err)
	}
	return &job, nil
}

func (r *publishJobRepo) FindByDraftID(ctx context.Context, draftID int64) ([]publish.PublishJob, error) {
	var jobs []publish.PublishJob
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).
		Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("publishJobRepo.FindByDraftID: %w", err)
	}
	return jobs, nil
}

func (r *publishJobRepo) Update(ctx context.Context, job *publish.PublishJob) error {
	if err := r.db.WithContext(ctx).Save(job).Error; err != nil {
		return fmt.Errorf("publishJobRepo.Update: %w", err)
	}
	return nil
}

func (r *publishJobRepo) ListByStatus(ctx context.Context, status string, limit, offset int) ([]publish.PublishJob, error) {
	var jobs []publish.PublishJob
	if err := r.db.WithContext(ctx).Where("status = ?", status).
		Order("created_at ASC").Limit(limit).Offset(offset).Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("publishJobRepo.ListByStatus: %w", err)
	}
	return jobs, nil
}

// ---------- PublishRecord Repository ----------

// PublishRecordRepository defines the interface for publish record data access.
type PublishRecordRepository interface {
	Create(ctx context.Context, record *publish.PublishRecord) error
	FindByPublishJobID(ctx context.Context, publishJobID int64) ([]publish.PublishRecord, error)
	FindByArticleURL(ctx context.Context, articleURL string) (*publish.PublishRecord, error)
	Update(ctx context.Context, record *publish.PublishRecord) error
}

type publishRecordRepo struct {
	db *gorm.DB
}

// NewPublishRecordRepository creates a new PostgreSQL-backed publish record repository.
func NewPublishRecordRepository(db *gorm.DB) PublishRecordRepository {
	return &publishRecordRepo{db: db}
}

func (r *publishRecordRepo) Create(ctx context.Context, record *publish.PublishRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("publishRecordRepo.Create: %w", err)
	}
	return nil
}

func (r *publishRecordRepo) FindByPublishJobID(ctx context.Context, publishJobID int64) ([]publish.PublishRecord, error) {
	var records []publish.PublishRecord
	if err := r.db.WithContext(ctx).Where("publish_job_id = ?", publishJobID).
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("publishRecordRepo.FindByPublishJobID: %w", err)
	}
	return records, nil
}

func (r *publishRecordRepo) FindByArticleURL(ctx context.Context, articleURL string) (*publish.PublishRecord, error) {
	var record publish.PublishRecord
	if err := r.db.WithContext(ctx).Where("article_url = ?", articleURL).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("publishRecordRepo.FindByArticleURL: %w", err)
	}
	return &record, nil
}

func (r *publishRecordRepo) Update(ctx context.Context, record *publish.PublishRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("publishRecordRepo.Update: %w", err)
	}
	return nil
}
