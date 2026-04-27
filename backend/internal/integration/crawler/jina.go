// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// JinaReaderProvider implements adapter.CrawlerProvider using the Jina Reader API.
type JinaReaderProvider struct {
	apiKey string
	client *http.Client
	logger *zap.Logger
}

// NewJinaReaderProvider creates a new JinaReaderProvider.
func NewJinaReaderProvider(apiKey string, logger *zap.Logger) *JinaReaderProvider {
	return &JinaReaderProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}
}

// Crawl fetches and extracts content from the given URL using Jina Reader.
func (p *JinaReaderProvider) Crawl(ctx context.Context, targetURL string) (*adapter.CrawledPage, error) {
	reqURL := "https://r.jina.ai/" + targetURL

	p.logger.Debug("jina reader: crawling", zap.String("url", targetURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("jinaReader: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/plain")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jinaReader: send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jinaReader: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("jinaReader: error %d: %s", resp.StatusCode, preview)
	}

	content := string(body)

	// Extract title from first markdown heading if present
	title := targetURL
	if idx := strings.Index(content, "\n"); idx > 0 {
		firstLine := strings.TrimSpace(content[:idx])
		if strings.HasPrefix(firstLine, "# ") {
			title = strings.TrimPrefix(firstLine, "# ")
		}
	}

	// Truncate very long content (keep first ~3000 chars for LLM context)
	wordCount := len([]rune(content))
	if len(content) > 3000 {
		content = content[:3000]
	}

	p.logger.Debug("jina reader: crawled", zap.Int("content_len", len(content)))
	return &adapter.CrawledPage{
		URL:       targetURL,
		Title:     title,
		Content:   content,
		WordCount: wordCount,
	}, nil
}

// Compile-time interface check.
var _ adapter.CrawlerProvider = (*JinaReaderProvider)(nil)
