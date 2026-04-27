// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// TavilySearchProvider implements adapter.SearchProvider using Tavily Search API.
type TavilySearchProvider struct {
	apiKey string
	client *http.Client
	logger *zap.Logger
}

// NewTavilySearchProvider creates a new TavilySearchProvider.
func NewTavilySearchProvider(apiKey string, logger *zap.Logger) *TavilySearchProvider {
	return &TavilySearchProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}
}

// Tavily API request/response structs.
type tavilySearchRequest struct {
	APIKey     string `json:"api_key"`
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type tavilySearchResponse struct {
	Results []tavilyResult `json:"results"`
}

type tavilyResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// Search performs a Tavily search and returns results.
func (p *TavilySearchProvider) Search(ctx context.Context, query string, opts adapter.SearchOptions) ([]adapter.SearchResult, error) {
	maxResults := opts.MaxResults
	if maxResults <= 0 || maxResults > 10 {
		maxResults = 5
	}

	reqBody := tavilySearchRequest{
		APIKey:     p.apiKey,
		Query:      query,
		MaxResults: maxResults,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("tavilySearch: marshal request: %w", err)
	}

	p.logger.Debug("tavily search", zap.String("query", query))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("tavilySearch: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavilySearch: send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tavilySearch: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavilySearch: API error %d: %s", resp.StatusCode, string(body))
	}

	var tResp tavilySearchResponse
	if err := json.Unmarshal(body, &tResp); err != nil {
		return nil, fmt.Errorf("tavilySearch: unmarshal: %w", err)
	}

	results := make([]adapter.SearchResult, 0, len(tResp.Results))
	for _, item := range tResp.Results {
		results = append(results, adapter.SearchResult{
			Title:   item.Title,
			URL:     item.URL,
			Snippet: item.Content,
		})
	}

	p.logger.Debug("tavily search: results", zap.Int("count", len(results)))
	return results, nil
}

// Compile-time interface check.
var _ adapter.SearchProvider = (*TavilySearchProvider)(nil)
