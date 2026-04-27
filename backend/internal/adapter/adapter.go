// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Package adapter defines interfaces for all external capabilities.
// Implementations are injected via dependency injection at startup.
// This follows the Adapter pattern per the ReadBud architecture spec.
package adapter

import (
	"context"
)

// ---------- Search Provider (HY-291) ----------

// SearchResult represents a single search result from a web search.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// SearchProvider abstracts web search APIs (e.g., SerpAPI, Bing, Google Custom Search).
type SearchProvider interface {
	// Search performs a web search and returns a list of results.
	Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)
}

// SearchOptions configures a search request.
type SearchOptions struct {
	MaxResults int    `json:"max_results"`
	Language   string `json:"language"`
	Region     string `json:"region"`
}

// ---------- Crawler Provider (HY-292) ----------

// CrawledPage represents content extracted from a web page.
type CrawledPage struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Content     string `json:"content"`      // Cleaned text content
	HTML        string `json:"html"`         // Raw HTML
	PublishDate string `json:"publish_date"` // ISO 8601 if available
	Author      string `json:"author"`
	WordCount   int    `json:"word_count"`
}

// CrawlerProvider abstracts web page crawling and content extraction.
type CrawlerProvider interface {
	// Crawl fetches and extracts content from the given URL.
	Crawl(ctx context.Context, url string) (*CrawledPage, error)
}

// ---------- LLM Provider (HY-290) ----------

// LLMMessage represents a single message in a conversation.
type LLMMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// LLMResponse represents the LLM's response.
type LLMResponse struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	TokensUsed   int    `json:"tokens_used"`
}

// LLMOptions configures an LLM request.
type LLMOptions struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// LLMProvider abstracts LLM APIs for text generation (keyword expansion, article writing, etc.).
type LLMProvider interface {
	// Chat sends a conversation and returns the LLM response.
	Chat(ctx context.Context, messages []LLMMessage, opts LLMOptions) (*LLMResponse, error)
}

// ---------- Image Search Provider ----------

// ImageResult represents a single image search result.
type ImageResult struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Source    string `json:"source"`
}

// ImageSearchProvider abstracts image search APIs.
type ImageSearchProvider interface {
	// SearchImages returns relevant images for the given query.
	SearchImages(ctx context.Context, query string, maxResults int) ([]ImageResult, error)
}

// ---------- Image Generation Provider ----------

// GeneratedImage represents a generated image.
type GeneratedImage struct {
	URL    string `json:"url"`
	Base64 string `json:"base64,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// ImageGenProvider abstracts image generation APIs (e.g., DALL-E, Midjourney).
type ImageGenProvider interface {
	// Generate creates an image based on the prompt.
	Generate(ctx context.Context, prompt string, opts ImageGenOptions) (*GeneratedImage, error)
}

// ImageGenOptions configures an image generation request.
type ImageGenOptions struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Style  string `json:"style"`
}

// ---------- Chart Provider ----------

// ChartData represents data for chart generation.
type ChartData struct {
	Type     string    `json:"type"` // "bar", "line", "pie", etc.
	Title    string    `json:"title"`
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset represents a single data series in a chart.
type Dataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
}

// ChartProvider abstracts server-side chart rendering.
type ChartProvider interface {
	// RenderChart generates a chart image and returns its URL or base64 data.
	RenderChart(ctx context.Context, data ChartData) (string, error)
}

// ---------- Storage Provider ----------

// StorageProvider abstracts object storage (local FS, S3, etc.).
type StorageProvider interface {
	// Upload stores data and returns the object URL.
	Upload(ctx context.Context, bucket, key string, data []byte, contentType string) (string, error)
	// Download retrieves the bytes for a previously uploaded object.
	// Returns an error wrapping os.ErrNotExist when the object is missing.
	Download(ctx context.Context, bucket, key string) ([]byte, error)
	// GetURL returns a presigned URL for the given object.
	GetURL(ctx context.Context, bucket, key string) (string, error)
	// Delete removes an object.
	Delete(ctx context.Context, bucket, key string) error
}

// ---------- WeChat Publisher ----------

// WeChatArticle represents an article ready for WeChat publishing.
type WeChatArticle struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	Content   string `json:"content"`    // WeChat-compatible HTML
	Digest    string `json:"digest"`     // Summary for card display
	ThumbURL  string `json:"thumb_url"`  // Cover image URL
	SourceURL string `json:"source_url"` // "Read original" link
}

// Content image upload constraints per WeChat API specification.
const (
	// ContentImageMaxBytes is the maximum size for a single content image (1 MB).
	ContentImageMaxBytes = 1 * 1024 * 1024
	// ContentImageMIMEJPEG is the MIME type for JPEG images.
	ContentImageMIMEJPEG = "image/jpeg"
	// ContentImageMIMEPNG is the MIME type for PNG images.
	ContentImageMIMEPNG = "image/png"
)

// WeChatPublishResult contains the result of a WeChat publish operation.
type WeChatPublishResult struct {
	MediaID   string `json:"media_id"`
	MsgID     string `json:"msg_id"`
	PublishID string `json:"publish_id"`
}

// ---------- WeChat Metrics Sync Provider ----------

// ArticleMetrics represents metrics data for a single article from WeChat analytics API.
type ArticleMetrics struct {
	ArticleID      string `json:"article_id"`
	ReadCount      int    `json:"read_count"`
	ReadUserCount  int    `json:"read_user_count"`
	ShareCount     int    `json:"share_count"`
	ShareUserCount int    `json:"share_user_count"`
	RawJSON        []byte `json:"raw_json,omitempty"` // Original API response
}

// FansMetrics represents fan growth data from WeChat analytics API.
type FansMetrics struct {
	Date            string `json:"date"` // YYYY-MM-DD
	AddFansCount    int    `json:"add_fans_count"`
	CancelFansCount int    `json:"cancel_fans_count"`
	NetFansCount    int    `json:"net_fans_count"`
	RawJSON         []byte `json:"raw_json,omitempty"`
}

// MetricsSyncProvider abstracts WeChat data analytics APIs for syncing article metrics.
type MetricsSyncProvider interface {
	// GetArticleMetrics fetches read/share metrics for a published article.
	GetArticleMetrics(ctx context.Context, accessToken string, articleID string) (*ArticleMetrics, error)
	// GetFansMetrics fetches fan growth data for a date range.
	GetFansMetrics(ctx context.Context, accessToken string, date string) (*FansMetrics, error)
}

// WeChatPublisher abstracts WeChat Official Account article publishing.
type WeChatPublisher interface {
	// UploadImage uploads an image to WeChat as permanent material and returns media_id.
	UploadImage(ctx context.Context, accessToken string, imageData []byte, filename string) (string, error)
	// UploadContentImage uploads an in-article image (正文图片) to WeChat.
	// Unlike UploadImage, this returns a WeChat-accessible URL (not a media_id)
	// and does NOT consume the permanent material quota.
	// Only jpg/png allowed, max 1MB per image.
	UploadContentImage(ctx context.Context, accessToken string, imageData []byte, filename string) (string, error)
	// CreateDraft creates a draft article on WeChat.
	CreateDraft(ctx context.Context, accessToken string, article WeChatArticle) (string, error)
	// Publish publishes a draft article.
	Publish(ctx context.Context, accessToken string, mediaID string) (*WeChatPublishResult, error)
}
