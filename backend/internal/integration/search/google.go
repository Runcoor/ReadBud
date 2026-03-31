package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// GoogleSearchProvider implements adapter.SearchProvider using Google Custom Search JSON API.
type GoogleSearchProvider struct {
	apiKey         string
	searchEngineID string
	client         *http.Client
	logger         *zap.Logger
}

// NewGoogleSearchProvider creates a new GoogleSearchProvider.
func NewGoogleSearchProvider(apiKey, searchEngineID string, logger *zap.Logger) *GoogleSearchProvider {
	return &GoogleSearchProvider{
		apiKey:         apiKey,
		searchEngineID: searchEngineID,
		client:         &http.Client{Timeout: 15 * time.Second},
		logger:         logger,
	}
}

// Google CSE response structs.
type googleSearchResponse struct {
	Items []googleSearchItem `json:"items"`
}

type googleSearchItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// Search performs a Google Custom Search and returns results.
func (p *GoogleSearchProvider) Search(ctx context.Context, query string, opts adapter.SearchOptions) ([]adapter.SearchResult, error) {
	maxResults := opts.MaxResults
	if maxResults <= 0 || maxResults > 10 {
		maxResults = 5
	}

	reqURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&num=%d",
		p.apiKey, p.searchEngineID, url.QueryEscape(query), maxResults)

	p.logger.Debug("google search", zap.String("query", query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("googleSearch: create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("googleSearch: send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("googleSearch: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("googleSearch: API error %d: %s", resp.StatusCode, string(body))
	}

	var gResp googleSearchResponse
	if err := json.Unmarshal(body, &gResp); err != nil {
		return nil, fmt.Errorf("googleSearch: unmarshal: %w", err)
	}

	results := make([]adapter.SearchResult, 0, len(gResp.Items))
	for _, item := range gResp.Items {
		results = append(results, adapter.SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
		})
	}

	p.logger.Debug("google search: results", zap.Int("count", len(results)))
	return results, nil
}

// Compile-time interface check.
var _ adapter.SearchProvider = (*GoogleSearchProvider)(nil)
