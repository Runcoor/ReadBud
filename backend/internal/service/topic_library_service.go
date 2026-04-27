// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"readbud/internal/domain"
	"readbud/internal/repository/postgres"
)

// TopicVO is the view-object returned by topic library API endpoints.
type TopicVO struct {
	PublicID        string    `json:"public_id"`
	Keyword         string    `json:"keyword"`
	Audience        string    `json:"audience"`
	ArticleGoal     string    `json:"article_goal"`
	HistoricalScore float64   `json:"historical_score"`
	RecommendWeight float64   `json:"recommend_weight"`
	LastUsedAt      string    `json:"last_used_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TopicListResponse wraps a paginated list of topics.
type TopicListResponse struct {
	Items []TopicVO `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

// CreateTopicRequest is the DTO for creating a topic.
type CreateTopicRequest struct {
	Keyword     string `json:"keyword" binding:"required,max=255"`
	Audience    string `json:"audience" binding:"max=255"`
	ArticleGoal string `json:"article_goal" binding:"max=64"`
}

// UpdateTopicRequest is the DTO for updating a topic.
type UpdateTopicRequest struct {
	Keyword     *string `json:"keyword" binding:"omitempty,max=255"`
	Audience    *string `json:"audience" binding:"omitempty,max=255"`
	ArticleGoal *string `json:"article_goal" binding:"omitempty,max=64"`
}

// PerformanceFeedback carries metrics data used to update a topic's scores.
type PerformanceFeedback struct {
	ReadCount  int `json:"read_count"`
	ShareCount int `json:"share_count"`
	FansGained int `json:"fans_gained"`
}

// TopicLibraryService handles topic library business logic.
type TopicLibraryService struct {
	topicRepo   postgres.TopicLibraryRepository
	taskRepo    postgres.TaskRepository
	metricsRepo postgres.MetricsSnapshotRepository
	logger      *zap.Logger
}

// NewTopicLibraryService creates a new TopicLibraryService.
func NewTopicLibraryService(
	topicRepo postgres.TopicLibraryRepository,
	taskRepo postgres.TaskRepository,
	metricsRepo postgres.MetricsSnapshotRepository,
	logger *zap.Logger,
) *TopicLibraryService {
	return &TopicLibraryService{
		topicRepo:   topicRepo,
		taskRepo:    taskRepo,
		metricsRepo: metricsRepo,
		logger:      logger,
	}
}

// Create adds a new topic to the library. If the keyword already exists, it returns an error.
func (s *TopicLibraryService) Create(ctx context.Context, req CreateTopicRequest) (*TopicVO, error) {
	existing, err := s.topicRepo.FindByKeyword(ctx, req.Keyword)
	if err != nil {
		return nil, fmt.Errorf("topicLibraryService.Create: check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("topicLibraryService.Create: keyword '%s' already exists", req.Keyword)
	}

	topic := &domain.TopicLibrary{
		Keyword:     req.Keyword,
		Audience:    req.Audience,
		ArticleGoal: req.ArticleGoal,
	}

	if err := s.topicRepo.Create(ctx, topic); err != nil {
		return nil, fmt.Errorf("topicLibraryService.Create: %w", err)
	}

	s.logger.Info("topic created",
		zap.String("keyword", topic.Keyword),
		zap.String("public_id", topic.PublicID),
	)

	return toTopicVO(topic), nil
}

// GetByPublicID retrieves a single topic by its public ID.
func (s *TopicLibraryService) GetByPublicID(ctx context.Context, publicID string) (*TopicVO, error) {
	topic, err := s.topicRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("topicLibraryService.GetByPublicID: %w", err)
	}
	if topic == nil {
		return nil, ErrNotFound
	}
	return toTopicVO(topic), nil
}

// Update modifies an existing topic's editable fields.
func (s *TopicLibraryService) Update(ctx context.Context, publicID string, req UpdateTopicRequest) (*TopicVO, error) {
	topic, err := s.topicRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("topicLibraryService.Update: find: %w", err)
	}
	if topic == nil {
		return nil, ErrNotFound
	}

	if req.Keyword != nil && *req.Keyword != topic.Keyword {
		// Check keyword uniqueness
		existing, findErr := s.topicRepo.FindByKeyword(ctx, *req.Keyword)
		if findErr != nil {
			return nil, fmt.Errorf("topicLibraryService.Update: check keyword: %w", findErr)
		}
		if existing != nil && existing.ID != topic.ID {
			return nil, fmt.Errorf("topicLibraryService.Update: keyword '%s' already in use", *req.Keyword)
		}
		topic.Keyword = *req.Keyword
	}
	if req.Audience != nil {
		topic.Audience = *req.Audience
	}
	if req.ArticleGoal != nil {
		topic.ArticleGoal = *req.ArticleGoal
	}

	if err := s.topicRepo.Update(ctx, topic); err != nil {
		return nil, fmt.Errorf("topicLibraryService.Update: %w", err)
	}

	return toTopicVO(topic), nil
}

// Delete removes a topic from the library by its public ID.
func (s *TopicLibraryService) Delete(ctx context.Context, publicID string) error {
	topic, err := s.topicRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("topicLibraryService.Delete: find: %w", err)
	}
	if topic == nil {
		return ErrNotFound
	}
	if err := s.topicRepo.Delete(ctx, topic.ID); err != nil {
		return fmt.Errorf("topicLibraryService.Delete: %w", err)
	}

	s.logger.Info("topic deleted",
		zap.String("public_id", publicID),
		zap.String("keyword", topic.Keyword),
	)
	return nil
}

// List returns paginated topics with optional sorting.
func (s *TopicLibraryService) List(ctx context.Context, page, size int, sortBy string) (*TopicListResponse, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	topics, total, err := s.topicRepo.List(ctx, size, offset, sortBy)
	if err != nil {
		return nil, fmt.Errorf("topicLibraryService.List: %w", err)
	}

	items := make([]TopicVO, 0, len(topics))
	for i := range topics {
		items = append(items, *toTopicVO(&topics[i]))
	}

	return &TopicListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// Search queries topics by keyword or audience text.
func (s *TopicLibraryService) Search(ctx context.Context, query string, page, size int) (*TopicListResponse, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	topics, total, err := s.topicRepo.Search(ctx, query, size, offset)
	if err != nil {
		return nil, fmt.Errorf("topicLibraryService.Search: %w", err)
	}

	items := make([]TopicVO, 0, len(topics))
	for i := range topics {
		items = append(items, *toTopicVO(&topics[i]))
	}

	return &TopicListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// GetRecommendations returns the top-N recommended topics by weight.
func (s *TopicLibraryService) GetRecommendations(ctx context.Context, limit int) ([]TopicVO, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	topics, err := s.topicRepo.ListTopRecommended(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("topicLibraryService.GetRecommendations: %w", err)
	}

	items := make([]TopicVO, 0, len(topics))
	for i := range topics {
		items = append(items, *toTopicVO(&topics[i]))
	}
	return items, nil
}

// RecordUsage marks a topic as recently used (called when a task uses this keyword).
func (s *TopicLibraryService) RecordUsage(ctx context.Context, keyword string) error {
	topic, err := s.topicRepo.FindByKeyword(ctx, keyword)
	if err != nil {
		return fmt.Errorf("topicLibraryService.RecordUsage: find: %w", err)
	}
	if topic == nil {
		// Auto-create the topic entry if it doesn't exist yet
		topic = &domain.TopicLibrary{
			Keyword: keyword,
		}
		if createErr := s.topicRepo.Create(ctx, topic); createErr != nil {
			return fmt.Errorf("topicLibraryService.RecordUsage: auto-create: %w", createErr)
		}
		s.logger.Info("topic auto-created from usage",
			zap.String("keyword", keyword),
		)
		return nil
	}

	if err := s.topicRepo.IncrementUsage(ctx, topic.ID); err != nil {
		return fmt.Errorf("topicLibraryService.RecordUsage: %w", err)
	}
	return nil
}

// UpdatePerformance recalculates a topic's historical_score and recommend_weight
// based on article performance metrics. This implements the performance feedback loop.
//
// Score formula: historical_score = 0.4 * reads_norm + 0.3 * shares_norm + 0.3 * fans_norm
// Weight formula: recommend_weight = historical_score * recency_factor
// Recency factor decays over time: 1.0 for today, 0.5 after 30 days.
func (s *TopicLibraryService) UpdatePerformance(ctx context.Context, publicID string, feedback PerformanceFeedback) error {
	topic, err := s.topicRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("topicLibraryService.UpdatePerformance: find: %w", err)
	}
	if topic == nil {
		return ErrNotFound
	}

	score := computeHistoricalScore(feedback)
	// Blend with existing score using exponential moving average (alpha = 0.3)
	const alpha = 0.3
	blendedScore := alpha*score + (1-alpha)*topic.HistoricalScore

	recencyFactor := computeRecencyFactor(topic.LastUsedAt)
	weight := blendedScore * recencyFactor

	if err := s.topicRepo.UpdateScoreAndWeight(ctx, topic.ID, blendedScore, weight); err != nil {
		return fmt.Errorf("topicLibraryService.UpdatePerformance: update: %w", err)
	}

	s.logger.Info("topic performance updated",
		zap.String("keyword", topic.Keyword),
		zap.Float64("new_score", blendedScore),
		zap.Float64("new_weight", weight),
	)
	return nil
}

// computeHistoricalScore calculates a normalised score from raw metrics.
// Uses sigmoid-like normalisation to keep values in [0, 100].
func computeHistoricalScore(fb PerformanceFeedback) float64 {
	readsNorm := normalise(float64(fb.ReadCount), 10000)
	sharesNorm := normalise(float64(fb.ShareCount), 1000)
	fansNorm := normalise(float64(fb.FansGained), 500)

	return 0.4*readsNorm + 0.3*sharesNorm + 0.3*fansNorm
}

// normalise maps a value to [0, 100] using a sigmoid curve centred at halfPoint.
func normalise(value, halfPoint float64) float64 {
	if value <= 0 {
		return 0
	}
	return 100 * (1 - math.Exp(-value/halfPoint))
}

// computeRecencyFactor returns a decay factor [0.1, 1.0] based on how recently
// the topic was used. Decays exponentially with a 30-day half-life.
func computeRecencyFactor(lastUsed *time.Time) float64 {
	if lastUsed == nil {
		return 0.1
	}
	daysSince := time.Since(*lastUsed).Hours() / 24
	if daysSince < 0 {
		daysSince = 0
	}
	const halfLife = 30.0
	factor := math.Exp(-0.693 * daysSince / halfLife) // ln(2) ≈ 0.693
	if factor < 0.1 {
		return 0.1
	}
	return factor
}

// toTopicVO converts a domain TopicLibrary to a TopicVO.
func toTopicVO(t *domain.TopicLibrary) *TopicVO {
	vo := &TopicVO{
		PublicID:        t.PublicID,
		Keyword:         t.Keyword,
		Audience:        t.Audience,
		ArticleGoal:     t.ArticleGoal,
		HistoricalScore: t.HistoricalScore,
		RecommendWeight: t.RecommendWeight,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
	if t.LastUsedAt != nil {
		vo.LastUsedAt = t.LastUsedAt.Format(time.RFC3339)
	}
	return vo
}
