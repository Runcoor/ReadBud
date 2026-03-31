package image

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

// PexelsProvider implements adapter.ImageSearchProvider using the Pexels API.
type PexelsProvider struct {
	apiKey string
	client *http.Client
	logger *zap.Logger
}

// NewPexelsProvider creates a new Pexels image search provider.
func NewPexelsProvider(apiKey string, logger *zap.Logger) *PexelsProvider {
	return &PexelsProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
		logger: logger,
	}
}

type pexelsResponse struct {
	Photos []pexelsPhoto `json:"photos"`
}

type pexelsPhoto struct {
	ID     int       `json:"id"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	URL    string    `json:"url"`
	Src    pexelsSrc `json:"src"`
}

type pexelsSrc struct {
	Original  string `json:"original"`
	Large2X   string `json:"large2x"`
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Small     string `json:"small"`
	Landscape string `json:"landscape"`
}

// SearchImages searches Pexels for images matching the query.
func (p *PexelsProvider) SearchImages(ctx context.Context, query string, maxResults int) ([]adapter.ImageResult, error) {
	if maxResults <= 0 {
		maxResults = 5
	}

	reqURL := fmt.Sprintf("https://api.pexels.com/v1/search?query=%s&per_page=%d&locale=zh-CN", url.QueryEscape(query), maxResults)

	p.logger.Debug("pexels: searching", zap.String("query", query), zap.Int("max", maxResults))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("pexels.SearchImages: create request: %w", err)
	}
	req.Header.Set("Authorization", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pexels.SearchImages: send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("pexels.SearchImages: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pexels.SearchImages: API error %d: %s", resp.StatusCode, string(body))
	}

	var pResp pexelsResponse
	if err := json.Unmarshal(body, &pResp); err != nil {
		return nil, fmt.Errorf("pexels.SearchImages: unmarshal: %w", err)
	}

	results := make([]adapter.ImageResult, 0, len(pResp.Photos))
	for _, photo := range pResp.Photos {
		results = append(results, adapter.ImageResult{
			URL:       photo.Src.Large,
			Thumbnail: photo.Src.Medium,
			Width:     photo.Width,
			Height:    photo.Height,
			Source:    fmt.Sprintf("https://www.pexels.com/photo/%d/", photo.ID),
		})
	}

	p.logger.Debug("pexels: found images", zap.Int("count", len(results)))
	return results, nil
}

// Compile-time interface check.
var _ adapter.ImageSearchProvider = (*PexelsProvider)(nil)
