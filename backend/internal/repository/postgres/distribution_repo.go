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

// DistributionPackageRepository defines the interface for distribution package data access.
type DistributionPackageRepository interface {
	Create(ctx context.Context, pkg *domain.DistributionPackage) error
	FindByID(ctx context.Context, id int64) (*domain.DistributionPackage, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.DistributionPackage, error)
	FindByDraftID(ctx context.Context, draftID int64) (*domain.DistributionPackage, error)
	Update(ctx context.Context, pkg *domain.DistributionPackage) error
	Upsert(ctx context.Context, pkg *domain.DistributionPackage) error
	Delete(ctx context.Context, id int64) error
}

type distributionPackageRepo struct {
	db *gorm.DB
}

// NewDistributionPackageRepository creates a new PostgreSQL-backed distribution package repository.
func NewDistributionPackageRepository(db *gorm.DB) DistributionPackageRepository {
	return &distributionPackageRepo{db: db}
}

func (r *distributionPackageRepo) Create(ctx context.Context, pkg *domain.DistributionPackage) error {
	if err := r.db.WithContext(ctx).Create(pkg).Error; err != nil {
		return fmt.Errorf("distributionPackageRepo.Create: %w", err)
	}
	return nil
}

func (r *distributionPackageRepo) FindByID(ctx context.Context, id int64) (*domain.DistributionPackage, error) {
	var pkg domain.DistributionPackage
	if err := r.db.WithContext(ctx).First(&pkg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("distributionPackageRepo.FindByID: %w", err)
	}
	return &pkg, nil
}

func (r *distributionPackageRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.DistributionPackage, error) {
	var pkg domain.DistributionPackage
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&pkg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("distributionPackageRepo.FindByPublicID: %w", err)
	}
	return &pkg, nil
}

func (r *distributionPackageRepo) FindByDraftID(ctx context.Context, draftID int64) (*domain.DistributionPackage, error) {
	var pkg domain.DistributionPackage
	if err := r.db.WithContext(ctx).Where("draft_id = ?", draftID).First(&pkg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("distributionPackageRepo.FindByDraftID: %w", err)
	}
	return &pkg, nil
}

func (r *distributionPackageRepo) Update(ctx context.Context, pkg *domain.DistributionPackage) error {
	if err := r.db.WithContext(ctx).Save(pkg).Error; err != nil {
		return fmt.Errorf("distributionPackageRepo.Update: %w", err)
	}
	return nil
}

// Upsert creates or updates a distribution package by draft_id.
func (r *distributionPackageRepo) Upsert(ctx context.Context, pkg *domain.DistributionPackage) error {
	existing, err := r.FindByDraftID(ctx, pkg.DraftID)
	if err != nil {
		return fmt.Errorf("distributionPackageRepo.Upsert: find: %w", err)
	}

	if existing != nil {
		existing.CommunityCopy = pkg.CommunityCopy
		existing.MomentsCopy = pkg.MomentsCopy
		existing.SummaryCard = pkg.SummaryCard
		existing.CommentGuide = pkg.CommentGuide
		existing.NextTopicSuggestion = pkg.NextTopicSuggestion
		return r.Update(ctx, existing)
	}

	return r.Create(ctx, pkg)
}

func (r *distributionPackageRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&domain.DistributionPackage{}, id).Error; err != nil {
		return fmt.Errorf("distributionPackageRepo.Delete: %w", err)
	}
	return nil
}
