// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package pipeline

import (
	"context"
	"fmt"

	"readbud/internal/adapter"
)

// ImageResult wraps the outcome of an image search or generation operation.
type ImageResult struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Source    string `json:"source"`     // "search" or "generated"
	SourceURL string `json:"source_url"` // original page URL for attribution
}

// ImageService handles image search, generation, and processing for article blocks.
type ImageService struct {
	search adapter.ImageSearchProvider
	gen    adapter.ImageGenProvider
}

// NewImageService creates a new ImageService.
func NewImageService(search adapter.ImageSearchProvider, gen adapter.ImageGenProvider) *ImageService {
	return &ImageService{search: search, gen: gen}
}

// SearchAndMatch uses the image search provider to find images matching a query.
func (s *ImageService) SearchAndMatch(ctx context.Context, query string, maxResults int) ([]ImageResult, error) {
	if maxResults <= 0 {
		maxResults = 5
	}

	results, err := s.search.SearchImages(ctx, query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("imageService.SearchAndMatch: %w", err)
	}

	images := make([]ImageResult, 0, len(results))
	for _, r := range results {
		images = append(images, ImageResult{
			URL:       r.URL,
			Width:     r.Width,
			Height:    r.Height,
			Source:    "search",
			SourceURL: r.Source,
		})
	}
	return images, nil
}

// GenerateFallback uses the image generation provider as a fallback when search yields no results.
func (s *ImageService) GenerateFallback(ctx context.Context, prompt string, width, height int) (*ImageResult, error) {
	if width <= 0 {
		width = 1024
	}
	if height <= 0 {
		height = 768
	}

	generated, err := s.gen.Generate(ctx, prompt, adapter.ImageGenOptions{
		Width:  width,
		Height: height,
		Style:  "professional",
	})
	if err != nil {
		return nil, fmt.Errorf("imageService.GenerateFallback: %w", err)
	}

	return &ImageResult{
		URL:    generated.URL,
		Width:  generated.Width,
		Height: generated.Height,
		Source: "generated",
	}, nil
}

// ProcessForBlock searches for an image first; if no results are found, generates one as fallback.
func (s *ImageService) ProcessForBlock(ctx context.Context, query string) (*ImageResult, error) {
	// Try search first
	results, err := s.SearchAndMatch(ctx, query, 3)
	if err == nil && len(results) > 0 {
		return &results[0], nil
	}

	// Fallback to generation
	img, err := s.GenerateFallback(ctx, query, 1024, 768)
	if err != nil {
		return nil, fmt.Errorf("imageService.ProcessForBlock: search and generation both failed: %w", err)
	}

	return img, nil
}
