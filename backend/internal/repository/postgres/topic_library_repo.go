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

// TopicLibraryRepository defines the interface for topic library data access.
type TopicLibraryRepository interface {
	Create(ctx context.Context, topic *domain.TopicLibrary) error
	FindByID(ctx context.Context, id int64) (*domain.TopicLibrary, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.TopicLibrary, error)
	FindByKeyword(ctx context.Context, keyword string) (*domain.TopicLibrary, error)
	Update(ctx context.Context, topic *domain.TopicLibrary) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int, sortBy string) ([]domain.TopicLibrary, int64, error)
	ListTopRecommended(ctx context.Context, limit int) ([]domain.TopicLibrary, error)
	Search(ctx context.Context, query string, limit, offset int) ([]domain.TopicLibrary, int64, error)
	IncrementUsage(ctx context.Context, id int64) error
	UpdateScoreAndWeight(ctx context.Context, id int64, historicalScore, recommendWeight float64) error
}

type topicLibraryRepo struct {
	db *gorm.DB
}

// NewTopicLibraryRepository creates a new PostgreSQL-backed topic library repository.
func NewTopicLibraryRepository(db *gorm.DB) TopicLibraryRepository {
	return &topicLibraryRepo{db: db}
}

func (r *topicLibraryRepo) Create(ctx context.Context, topic *domain.TopicLibrary) error {
	if err := r.db.WithContext(ctx).Create(topic).Error; err != nil {
		return fmt.Errorf("topicLibraryRepo.Create: %w", err)
	}
	return nil
}

func (r *topicLibraryRepo) FindByID(ctx context.Context, id int64) (*domain.TopicLibrary, error) {
	var topic domain.TopicLibrary
	if err := r.db.WithContext(ctx).First(&topic, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("topicLibraryRepo.FindByID: %w", err)
	}
	return &topic, nil
}

func (r *topicLibraryRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.TopicLibrary, error) {
	var topic domain.TopicLibrary
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&topic).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("topicLibraryRepo.FindByPublicID: %w", err)
	}
	return &topic, nil
}

func (r *topicLibraryRepo) FindByKeyword(ctx context.Context, keyword string) (*domain.TopicLibrary, error) {
	var topic domain.TopicLibrary
	if err := r.db.WithContext(ctx).Where("keyword = ?", keyword).First(&topic).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("topicLibraryRepo.FindByKeyword: %w", err)
	}
	return &topic, nil
}

func (r *topicLibraryRepo) Update(ctx context.Context, topic *domain.TopicLibrary) error {
	if err := r.db.WithContext(ctx).Save(topic).Error; err != nil {
		return fmt.Errorf("topicLibraryRepo.Update: %w", err)
	}
	return nil
}

func (r *topicLibraryRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&domain.TopicLibrary{}, id).Error; err != nil {
		return fmt.Errorf("topicLibraryRepo.Delete: %w", err)
	}
	return nil
}

// List returns topics with pagination and sorting.
// sortBy accepts: "recommend_weight", "historical_score", "last_used_at", "created_at".
func (r *topicLibraryRepo) List(ctx context.Context, limit, offset int, sortBy string) ([]domain.TopicLibrary, int64, error) {
	var topics []domain.TopicLibrary
	var total int64

	orderClause := resolveTopicSortOrder(sortBy)

	if err := r.db.WithContext(ctx).Model(&domain.TopicLibrary{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("topicLibraryRepo.List(count): %w", err)
	}

	if err := r.db.WithContext(ctx).
		Order(orderClause).
		Limit(limit).Offset(offset).
		Find(&topics).Error; err != nil {
		return nil, 0, fmt.Errorf("topicLibraryRepo.List: %w", err)
	}

	return topics, total, nil
}

// ListTopRecommended returns the top-N topics ordered by recommend_weight descending.
func (r *topicLibraryRepo) ListTopRecommended(ctx context.Context, limit int) ([]domain.TopicLibrary, error) {
	var topics []domain.TopicLibrary
	if err := r.db.WithContext(ctx).
		Order("recommend_weight DESC, historical_score DESC").
		Limit(limit).
		Find(&topics).Error; err != nil {
		return nil, fmt.Errorf("topicLibraryRepo.ListTopRecommended: %w", err)
	}
	return topics, nil
}

// Search performs a keyword search using ILIKE on the keyword and audience columns.
func (r *topicLibraryRepo) Search(ctx context.Context, query string, limit, offset int) ([]domain.TopicLibrary, int64, error) {
	var topics []domain.TopicLibrary
	var total int64

	pattern := "%" + query + "%"
	scope := r.db.WithContext(ctx).Model(&domain.TopicLibrary{}).
		Where("keyword ILIKE ? OR audience ILIKE ?", pattern, pattern)

	if err := scope.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("topicLibraryRepo.Search(count): %w", err)
	}

	if err := scope.
		Order("recommend_weight DESC").
		Limit(limit).Offset(offset).
		Find(&topics).Error; err != nil {
		return nil, 0, fmt.Errorf("topicLibraryRepo.Search: %w", err)
	}

	return topics, total, nil
}

// IncrementUsage updates last_used_at to now for the given topic.
func (r *topicLibraryRepo) IncrementUsage(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.TopicLibrary{}).
		Where("id = ?", id).
		Update("last_used_at", gorm.Expr("NOW()")).Error; err != nil {
		return fmt.Errorf("topicLibraryRepo.IncrementUsage: %w", err)
	}
	return nil
}

// UpdateScoreAndWeight updates the historical_score and recommend_weight for a topic.
func (r *topicLibraryRepo) UpdateScoreAndWeight(ctx context.Context, id int64, historicalScore, recommendWeight float64) error {
	updates := map[string]interface{}{
		"historical_score": historicalScore,
		"recommend_weight": recommendWeight,
	}
	if err := r.db.WithContext(ctx).
		Model(&domain.TopicLibrary{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("topicLibraryRepo.UpdateScoreAndWeight: %w", err)
	}
	return nil
}

// resolveTopicSortOrder maps a sort key to a SQL ORDER BY clause.
func resolveTopicSortOrder(sortBy string) string {
	switch sortBy {
	case "historical_score":
		return "historical_score DESC"
	case "last_used_at":
		return "last_used_at DESC NULLS LAST"
	case "created_at":
		return "created_at DESC"
	default:
		return "recommend_weight DESC"
	}
}
