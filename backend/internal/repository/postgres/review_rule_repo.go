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

// ReviewRuleRepository defines the interface for review rule data access.
type ReviewRuleRepository interface {
	Create(ctx context.Context, rule *domain.ReviewRule) error
	FindByID(ctx context.Context, id int64) (*domain.ReviewRule, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.ReviewRule, error)
	Update(ctx context.Context, rule *domain.ReviewRule) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]domain.ReviewRule, error)
	ListEnabled(ctx context.Context) ([]domain.ReviewRule, error)
	ListByRuleType(ctx context.Context, ruleType string) ([]domain.ReviewRule, error)
}

type reviewRuleRepo struct {
	db *gorm.DB
}

// NewReviewRuleRepository creates a new PostgreSQL-backed review rule repository.
func NewReviewRuleRepository(db *gorm.DB) ReviewRuleRepository {
	return &reviewRuleRepo{db: db}
}

func (r *reviewRuleRepo) Create(ctx context.Context, rule *domain.ReviewRule) error {
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("reviewRuleRepo.Create: %w", err)
	}
	return nil
}

func (r *reviewRuleRepo) FindByID(ctx context.Context, id int64) (*domain.ReviewRule, error) {
	var rule domain.ReviewRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("reviewRuleRepo.FindByID: %w", err)
	}
	return &rule, nil
}

func (r *reviewRuleRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.ReviewRule, error) {
	var rule domain.ReviewRule
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("reviewRuleRepo.FindByPublicID: %w", err)
	}
	return &rule, nil
}

func (r *reviewRuleRepo) Update(ctx context.Context, rule *domain.ReviewRule) error {
	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("reviewRuleRepo.Update: %w", err)
	}
	return nil
}

func (r *reviewRuleRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&domain.ReviewRule{}, id).Error; err != nil {
		return fmt.Errorf("reviewRuleRepo.Delete: %w", err)
	}
	return nil
}

func (r *reviewRuleRepo) List(ctx context.Context) ([]domain.ReviewRule, error) {
	var rules []domain.ReviewRule
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("reviewRuleRepo.List: %w", err)
	}
	return rules, nil
}

func (r *reviewRuleRepo) ListEnabled(ctx context.Context) ([]domain.ReviewRule, error) {
	var rules []domain.ReviewRule
	if err := r.db.WithContext(ctx).Where("is_enabled = ?", 1).Order("id ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("reviewRuleRepo.ListEnabled: %w", err)
	}
	return rules, nil
}

func (r *reviewRuleRepo) ListByRuleType(ctx context.Context, ruleType string) ([]domain.ReviewRule, error) {
	var rules []domain.ReviewRule
	if err := r.db.WithContext(ctx).Where("rule_type = ?", ruleType).Order("id ASC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("reviewRuleRepo.ListByRuleType: %w", err)
	}
	return rules, nil
}
