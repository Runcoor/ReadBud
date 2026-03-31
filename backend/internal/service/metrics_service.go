package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"

	"readbud/internal/adapter"
	"readbud/internal/domain/metrics"
	"readbud/internal/integration/wechat"
	"readbud/internal/repository/postgres"
)

// MetricsOverview summarises metrics across articles.
type MetricsOverview struct {
	TotalArticles  int `json:"total_articles"`
	TotalReads     int `json:"total_reads"`
	TotalShares    int `json:"total_shares"`
	TotalFansAdded int `json:"total_fans_added"`
}

// MetricsSyncResult reports the outcome of a manual metrics sync operation.
type MetricsSyncResult struct {
	ArticlesSynced int      `json:"articles_synced"`
	FansSynced     bool     `json:"fans_synced"`
	Errors         []string `json:"errors,omitempty"`
}

// MetricsService handles article metrics business logic.
type MetricsService struct {
	metricsRepo  postgres.MetricsSnapshotRepository
	publishRepo  postgres.PublishRecordRepository
	syncProvider adapter.MetricsSyncProvider
	tokenProv    wechat.TokenProvider
	logger       *zap.Logger
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(
	metricsRepo postgres.MetricsSnapshotRepository,
	publishRepo postgres.PublishRecordRepository,
	syncProvider adapter.MetricsSyncProvider,
	tokenProv wechat.TokenProvider,
	logger *zap.Logger,
) *MetricsService {
	return &MetricsService{
		metricsRepo:  metricsRepo,
		publishRepo:  publishRepo,
		syncProvider: syncProvider,
		tokenProv:    tokenProv,
		logger:       logger,
	}
}

// SyncMetrics triggers a manual metrics sync for a given WeChat account.
// It finds all published articles for the account and fetches their latest metrics
// from the WeChat analytics API, plus fan growth data.
func (s *MetricsService) SyncMetrics(ctx context.Context, wechatAccountID int64, appID string) (*MetricsSyncResult, error) {
	result := &MetricsSyncResult{}

	// Get access token
	accessToken, err := s.tokenProv.GetAccessToken(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("metricsService.SyncMetrics: get access token: %w", err)
	}

	// Find all distinct article IDs for this account
	articleIDs, err := s.metricsRepo.ListDistinctArticleIDs(ctx, wechatAccountID)
	if err != nil {
		return nil, fmt.Errorf("metricsService.SyncMetrics: list articles: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Sync article-level metrics (read/share)
	for _, articleID := range articleIDs {
		articleMetrics, syncErr := s.syncProvider.GetArticleMetrics(ctx, accessToken, articleID)
		if syncErr != nil {
			s.logger.Warn("failed to sync article metrics",
				zap.String("article_id", articleID),
				zap.Error(syncErr),
			)
			result.Errors = append(result.Errors, fmt.Sprintf("article %s: %v", articleID, syncErr))
			continue
		}

		snapshot := &metrics.MetricsSnapshot{
			WechatAccountID: wechatAccountID,
			ArticleID:       articleMetrics.ArticleID,
			MetricDate:      today,
			ReadCount:       intPtr(articleMetrics.ReadCount),
			ReadUserCount:   intPtr(articleMetrics.ReadUserCount),
			ShareCount:      intPtr(articleMetrics.ShareCount),
			ShareUserCount:  intPtr(articleMetrics.ShareUserCount),
			RawJSON:         datatypes.JSON(articleMetrics.RawJSON),
		}

		if upsertErr := s.metricsRepo.Upsert(ctx, snapshot); upsertErr != nil {
			s.logger.Warn("failed to upsert article metrics",
				zap.String("article_id", articleID),
				zap.Error(upsertErr),
			)
			result.Errors = append(result.Errors, fmt.Sprintf("upsert %s: %v", articleID, upsertErr))
			continue
		}

		result.ArticlesSynced++
	}

	// Sync fans metrics for today
	dateStr := today.Format("2006-01-02")
	fansMetrics, err := s.syncProvider.GetFansMetrics(ctx, accessToken, dateStr)
	if err != nil {
		s.logger.Warn("failed to sync fans metrics",
			zap.String("date", dateStr),
			zap.Error(err),
		)
		result.Errors = append(result.Errors, fmt.Sprintf("fans: %v", err))
	} else {
		// Store fans metrics as a special snapshot with a sentinel article_id
		fansSnapshot := &metrics.MetricsSnapshot{
			WechatAccountID: wechatAccountID,
			ArticleID:       fansArticleID(wechatAccountID),
			MetricDate:      today,
			AddFansCount:    intPtr(fansMetrics.AddFansCount),
			CancelFansCount: intPtr(fansMetrics.CancelFansCount),
			NetFansCount:    intPtr(fansMetrics.NetFansCount),
			RawJSON:         datatypes.JSON(fansMetrics.RawJSON),
		}

		if upsertErr := s.metricsRepo.Upsert(ctx, fansSnapshot); upsertErr != nil {
			s.logger.Warn("failed to upsert fans metrics", zap.Error(upsertErr))
			result.Errors = append(result.Errors, fmt.Sprintf("fans upsert: %v", upsertErr))
		} else {
			result.FansSynced = true
		}
	}

	s.logger.Info("metrics sync completed",
		zap.Int64("wechat_account_id", wechatAccountID),
		zap.Int("articles_synced", result.ArticlesSynced),
		zap.Bool("fans_synced", result.FansSynced),
	)

	return result, nil
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

// GetDistinctArticleIDs returns all unique article IDs tracked for a given account.
func (s *MetricsService) GetDistinctArticleIDs(ctx context.Context, accountID int64) ([]string, error) {
	ids, err := s.metricsRepo.ListDistinctArticleIDs(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("metricsService.GetDistinctArticleIDs: %w", err)
	}
	return ids, nil
}

// fansArticleID generates a sentinel article ID for storing account-level fans metrics.
func fansArticleID(accountID int64) string {
	return fmt.Sprintf("__fans__%d", accountID)
}

// intPtr is a helper to create a pointer to an int value.
func intPtr(v int) *int {
	return &v
}
