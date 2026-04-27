// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package search

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// StubSearchProvider returns fake search results for testing.
type StubSearchProvider struct {
	logger *zap.Logger
}

// NewStubSearchProvider creates a new StubSearchProvider.
func NewStubSearchProvider(logger *zap.Logger) *StubSearchProvider {
	return &StubSearchProvider{logger: logger}
}

// Search returns mock search results.
func (s *StubSearchProvider) Search(_ context.Context, query string, opts adapter.SearchOptions) ([]adapter.SearchResult, error) {
	s.logger.Info("stub search", zap.String("query", query), zap.Int("max_results", opts.MaxResults))

	max := opts.MaxResults
	if max <= 0 {
		max = 5
	}

	results := make([]adapter.SearchResult, 0, max)
	for i := 1; i <= max; i++ {
		results = append(results, adapter.SearchResult{
			Title:   fmt.Sprintf("%s - 相关文章 %d", query, i),
			URL:     fmt.Sprintf("https://example.com/article/%s/%d", query, i),
			Snippet: fmt.Sprintf("这是关于「%s」的一篇优质文章，涵盖了核心要点和实用建议。", query),
		})
	}

	return results, nil
}

// StubCrawlerProvider returns fake crawled content for testing.
type StubCrawlerProvider struct {
	logger *zap.Logger
}

// NewStubCrawlerProvider creates a new StubCrawlerProvider.
func NewStubCrawlerProvider(logger *zap.Logger) *StubCrawlerProvider {
	return &StubCrawlerProvider{logger: logger}
}

// Crawl returns mock crawled page content.
func (s *StubCrawlerProvider) Crawl(_ context.Context, url string) (*adapter.CrawledPage, error) {
	s.logger.Info("stub crawl", zap.String("url", url))

	return &adapter.CrawledPage{
		URL:         url,
		Title:       "测试文章标题 — 深度解析与实践指南",
		Content:     "这是一篇关于技术趋势的深度分析文章。内容涵盖了行业背景、核心观点、数据支撑和实践建议。文章结构清晰，论据充分，适合作为参考素材。总字数约1500字，包含3个核心观点和5组数据支撑。",
		HTML:        "<h1>测试文章标题</h1><p>这是一篇关于技术趋势的深度分析文章。</p>",
		PublishDate: "2026-03-28",
		Author:      "张三",
		WordCount:   1500,
	}, nil
}
