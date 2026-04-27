// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// ContentCitationRepository defines the interface for content citation data access.
type ContentCitationRepository interface {
	Create(ctx context.Context, c *domain.ContentCitation) error
	CreateBatch(ctx context.Context, citations []domain.ContentCitation) error
	FindByDraftID(ctx context.Context, draftID int64) ([]domain.ContentCitation, error)
	FindByBlockID(ctx context.Context, blockID int64) ([]domain.ContentCitation, error)
	DeleteByDraftID(ctx context.Context, draftID int64) error
}

type citationRepo struct {
	db *gorm.DB
}

// NewContentCitationRepository creates a new PostgreSQL-backed content citation repository.
func NewContentCitationRepository(db *gorm.DB) ContentCitationRepository {
	return &citationRepo{db: db}
}

func (r *citationRepo) Create(ctx context.Context, c *domain.ContentCitation) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("citationRepo.Create: %w", err)
	}
	return nil
}

func (r *citationRepo) CreateBatch(ctx context.Context, citations []domain.ContentCitation) error {
	if len(citations) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&citations).Error; err != nil {
		return fmt.Errorf("citationRepo.CreateBatch: %w", err)
	}
	return nil
}

func (r *citationRepo) FindByDraftID(ctx context.Context, draftID int64) ([]domain.ContentCitation, error) {
	var citations []domain.ContentCitation
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).
		Order("id ASC").Find(&citations).Error; err != nil {
		return nil, fmt.Errorf("citationRepo.FindByDraftID: %w", err)
	}
	return citations, nil
}

func (r *citationRepo) FindByBlockID(ctx context.Context, blockID int64) ([]domain.ContentCitation, error) {
	var citations []domain.ContentCitation
	if err := r.db.WithContext(ctx).Where("block_id = ?", blockID).
		Order("id ASC").Find(&citations).Error; err != nil {
		return nil, fmt.Errorf("citationRepo.FindByBlockID: %w", err)
	}
	return citations, nil
}

func (r *citationRepo) DeleteByDraftID(ctx context.Context, draftID int64) error {
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).
		Delete(&domain.ContentCitation{}).Error; err != nil {
		return fmt.Errorf("citationRepo.DeleteByDraftID: %w", err)
	}
	return nil
}
