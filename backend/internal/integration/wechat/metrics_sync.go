package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// StubMetricsSyncProvider is a placeholder implementation of adapter.MetricsSyncProvider
// for development and testing. Returns realistic-looking stub data.
// Will be replaced with real WeChat data analytics API calls in production.
type StubMetricsSyncProvider struct {
	logger *zap.Logger
}

// NewStubMetricsSyncProvider creates a new stub metrics sync provider.
func NewStubMetricsSyncProvider(logger *zap.Logger) *StubMetricsSyncProvider {
	return &StubMetricsSyncProvider{logger: logger}
}

// GetArticleMetrics returns stub article metrics for development.
func (p *StubMetricsSyncProvider) GetArticleMetrics(
	ctx context.Context, accessToken string, articleID string,
) (*adapter.ArticleMetrics, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("StubMetricsSyncProvider.GetArticleMetrics: access token is required")
	}
	if articleID == "" {
		return nil, fmt.Errorf("StubMetricsSyncProvider.GetArticleMetrics: article_id is required")
	}

	// Generate deterministic-looking stub data based on articleID hash
	readCount := 100 + rand.Intn(9900)
	readUserCount := readCount - rand.Intn(readCount/5+1)
	shareCount := rand.Intn(readCount / 10)
	shareUserCount := shareCount - rand.Intn(shareCount/3+1)
	if shareUserCount < 0 {
		shareUserCount = 0
	}

	rawData := map[string]interface{}{
		"article_id":      articleID,
		"read_num":        readCount,
		"read_unique_num": readUserCount,
		"share_num":       shareCount,
		"share_unique_num": shareUserCount,
		"stub":            true,
	}
	rawJSON, _ := json.Marshal(rawData)

	result := &adapter.ArticleMetrics{
		ArticleID:      articleID,
		ReadCount:      readCount,
		ReadUserCount:  readUserCount,
		ShareCount:     shareCount,
		ShareUserCount: shareUserCount,
		RawJSON:        rawJSON,
	}

	p.logger.Info("stub: fetched article metrics",
		zap.String("article_id", articleID),
		zap.Int("read_count", readCount),
		zap.Int("share_count", shareCount),
	)

	return result, nil
}

// GetFansMetrics returns stub fan growth data for development.
func (p *StubMetricsSyncProvider) GetFansMetrics(
	ctx context.Context, accessToken string, date string,
) (*adapter.FansMetrics, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("StubMetricsSyncProvider.GetFansMetrics: access token is required")
	}
	if date == "" {
		return nil, fmt.Errorf("StubMetricsSyncProvider.GetFansMetrics: date is required")
	}

	addFans := 5 + rand.Intn(50)
	cancelFans := rand.Intn(addFans/2 + 1)

	rawData := map[string]interface{}{
		"ref_date":         date,
		"new_user":         addFans,
		"cancel_user":      cancelFans,
		"cumulate_user":    10000 + rand.Intn(5000),
		"stub":             true,
	}
	rawJSON, _ := json.Marshal(rawData)

	result := &adapter.FansMetrics{
		Date:            date,
		AddFansCount:    addFans,
		CancelFansCount: cancelFans,
		NetFansCount:    addFans - cancelFans,
		RawJSON:         rawJSON,
	}

	p.logger.Info("stub: fetched fans metrics",
		zap.String("date", date),
		zap.Int("add_fans", addFans),
		zap.Int("cancel_fans", cancelFans),
	)

	return result, nil
}

// Compile-time check that StubMetricsSyncProvider satisfies the interface.
var _ adapter.MetricsSyncProvider = (*StubMetricsSyncProvider)(nil)
