package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"readbud/internal/domain/metrics"
	"readbud/internal/repository/postgres"
)

// MetricsOverview summarises metrics across articles.
type MetricsOverview struct {
	TotalArticles  int `json:"total_articles"`
	TotalReads     int `json:"total_reads"`
	TotalShares    int `json:"total_shares"`
	TotalFansAdded int `json:"total_fans_added"`
}

// MetricsService handles article metrics business logic.
type MetricsService struct {
	metricsRepo postgres.MetricsSnapshotRepository
	logger      *zap.Logger
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(metricsRepo postgres.MetricsSnapshotRepository, logger *zap.Logger) *MetricsService {
	return &MetricsService{metricsRepo: metricsRepo, logger: logger}
}

// RecordSnapshot persists a new metrics snapshot.
func (s *MetricsService) RecordSnapshot(ctx context.Context, snapshot *metrics.MetricsSnapshot) error {
	if snapshot.ArticleID == "" {
		return fmt.Errorf("metricsService.RecordSnapshot: article_id is required")
	}

	if err := s.metricsRepo.Create(ctx, snapshot); err != nil {
		return fmt.Errorf("metricsService.RecordSnapshot: %w", err)
	}

	s.logger.Info("metrics snapshot recorded",
		zap.String("article_id", snapshot.ArticleID),
		zap.Time("metric_date", snapshot.MetricDate),
	)
	return nil
}

// GetArticleMetrics returns metrics snapshots for an article within a date range.
func (s *MetricsService) GetArticleMetrics(ctx context.Context, articleID string, start, end time.Time) ([]metrics.MetricsSnapshot, error) {
	if articleID == "" {
		return nil, fmt.Errorf("metricsService.GetArticleMetrics: article_id is required")
	}

	snapshots, err := s.metricsRepo.FindByDateRange(ctx, articleID, start, end)
	if err != nil {
		return nil, fmt.Errorf("metricsService.GetArticleMetrics: %w", err)
	}
	return snapshots, nil
}

// GetOverview aggregates the latest metrics across a list of article IDs.
func (s *MetricsService) GetOverview(ctx context.Context, articleIDs []string) (*MetricsOverview, error) {
	overview := &MetricsOverview{}

	for _, articleID := range articleIDs {
		latest, err := s.metricsRepo.GetLatestByArticleID(ctx, articleID)
		if err != nil {
			return nil, fmt.Errorf("metricsService.GetOverview: %w", err)
		}
		if latest == nil {
			continue
		}

		overview.TotalArticles++
		if latest.ReadCount != nil {
			overview.TotalReads += *latest.ReadCount
		}
		if latest.ShareCount != nil {
			overview.TotalShares += *latest.ShareCount
		}
		if latest.AddFansCount != nil {
			overview.TotalFansAdded += *latest.AddFansCount
		}
	}

	return overview, nil
}
