package image

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// StubImageSearchProvider returns fake image search results for testing.
type StubImageSearchProvider struct {
	logger *zap.Logger
}

// NewStubImageSearchProvider creates a new StubImageSearchProvider.
func NewStubImageSearchProvider(logger *zap.Logger) *StubImageSearchProvider {
	return &StubImageSearchProvider{logger: logger}
}

// SearchImages returns mock image results.
func (s *StubImageSearchProvider) SearchImages(_ context.Context, query string, maxResults int) ([]adapter.ImageResult, error) {
	s.logger.Info("stub image search", zap.String("query", query), zap.Int("max_results", maxResults))

	if maxResults <= 0 {
		maxResults = 3
	}

	results := make([]adapter.ImageResult, 0, maxResults)
	for i := 1; i <= maxResults; i++ {
		results = append(results, adapter.ImageResult{
			URL:       fmt.Sprintf("https://images.example.com/%s/%d.jpg", query, i),
			Thumbnail: fmt.Sprintf("https://images.example.com/%s/%d_thumb.jpg", query, i),
			Width:     1200,
			Height:    800,
			Source:    "example.com",
		})
	}

	return results, nil
}

// StubImageGenProvider returns fake generated images for testing.
type StubImageGenProvider struct {
	logger *zap.Logger
}

// NewStubImageGenProvider creates a new StubImageGenProvider.
func NewStubImageGenProvider(logger *zap.Logger) *StubImageGenProvider {
	return &StubImageGenProvider{logger: logger}
}

// Generate returns a mock generated image.
func (s *StubImageGenProvider) Generate(_ context.Context, prompt string, opts adapter.ImageGenOptions) (*adapter.GeneratedImage, error) {
	s.logger.Info("stub image gen", zap.String("prompt", prompt))

	width := opts.Width
	height := opts.Height
	if width <= 0 {
		width = 1024
	}
	if height <= 0 {
		height = 1024
	}

	return &adapter.GeneratedImage{
		URL:    fmt.Sprintf("https://gen.example.com/image_%s_%dx%d.png", "stub", width, height),
		Width:  width,
		Height: height,
	}, nil
}
