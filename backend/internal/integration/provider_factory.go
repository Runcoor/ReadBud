// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package integration

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain"
	"readbud/internal/integration/image_gen"
	"readbud/internal/integration/llm"
	"readbud/internal/service"
)

// ProviderFactory resolves LLM and ImageGen providers dynamically from DB configuration.
type ProviderFactory struct {
	providerSvc *service.ProviderConfigService
	logger      *zap.Logger
}

// NewProviderFactory creates a new ProviderFactory.
func NewProviderFactory(providerSvc *service.ProviderConfigService, logger *zap.Logger) *ProviderFactory {
	return &ProviderFactory{
		providerSvc: providerSvc,
		logger:      logger,
	}
}

// providerConfigJSON represents the parsed config_json from a provider config.
type providerConfigJSON struct {
	Format    string `json:"format"`     // "openai", "anthropic"
	APIFormat string `json:"api_format"` // Legacy alias for format
	BaseURL   string `json:"base_url"`   // For OpenAI-compatible providers
	Model     string `json:"model"`      // Default model
	ProjectID string `json:"project_id"` // For Vertex AI
	Region    string `json:"region"`     // For Vertex AI
}

// getFormat returns the effective format, checking both format and api_format fields.
func (c providerConfigJSON) getFormat() string {
	if c.Format != "" {
		return c.Format
	}
	return c.APIFormat
}

// ParseAPIKey extracts an API key from a decrypted secret string.
// Supports both JSON format ({"api_key": "sk-xxx"}) and plain string format.
func ParseAPIKey(secret string) string {
	if secret == "" {
		return ""
	}
	var sec struct {
		APIKey string `json:"api_key"`
	}
	if err := json.Unmarshal([]byte(secret), &sec); err == nil && sec.APIKey != "" {
		return sec.APIKey
	}
	// Fallback: treat the whole string as the API key
	return secret
}

// ResolveLLM reads the active LLM config from DB and creates the appropriate provider.
func (f *ProviderFactory) ResolveLLM(ctx context.Context) (adapter.LLMProvider, error) {
	cfg, err := f.providerSvc.GetActiveByType(ctx, domain.ProviderTypeLLM)
	if err != nil {
		return nil, fmt.Errorf("providerFactory.ResolveLLM: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("providerFactory.ResolveLLM: no active LLM provider configured")
	}

	var config providerConfigJSON
	if err := json.Unmarshal(cfg.ConfigJSON, &config); err != nil {
		return nil, fmt.Errorf("providerFactory.ResolveLLM: parse config_json: %w", err)
	}

	secret, err := f.providerSvc.DecryptSecret(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("providerFactory.ResolveLLM: decrypt secret: %w", err)
	}

	apiKey := ParseAPIKey(secret)

	format := config.getFormat()
	if format == "" {
		// Infer format from provider name
		switch cfg.ProviderName {
		case "anthropic", "claude":
			format = "anthropic"
		default:
			format = "openai"
		}
	}

	f.logger.Info("resolved LLM provider",
		zap.String("provider_name", cfg.ProviderName),
		zap.String("format", format),
	)

	switch format {
	case "anthropic":
		return llm.NewAnthropicProvider(apiKey, config.Model, f.logger), nil
	case "openai":
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return llm.NewOpenAICompatProvider(baseURL, apiKey, config.Model, f.logger), nil
	default:
		return nil, fmt.Errorf("providerFactory.ResolveLLM: unknown format %q", format)
	}
}

// ResolveImageGen reads the active image_gen config from DB and creates the appropriate provider.
func (f *ProviderFactory) ResolveImageGen(ctx context.Context) (adapter.ImageGenProvider, error) {
	cfg, err := f.providerSvc.GetActiveByType(ctx, domain.ProviderTypeImageGen)
	if err != nil {
		return nil, fmt.Errorf("providerFactory.ResolveImageGen: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("providerFactory.ResolveImageGen: no active image_gen provider configured")
	}

	var config providerConfigJSON
	if err := json.Unmarshal(cfg.ConfigJSON, &config); err != nil {
		return nil, fmt.Errorf("providerFactory.ResolveImageGen: parse config_json: %w", err)
	}

	secret, err := f.providerSvc.DecryptSecret(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("providerFactory.ResolveImageGen: decrypt secret: %w", err)
	}

	apiKey := ParseAPIKey(secret)

	format := config.Format
	if format == "" {
		switch cfg.ProviderName {
		case "vertex", "vertex_ai", "google", "imagen":
			format = "vertex"
		default:
			format = "openai"
		}
	}

	f.logger.Info("resolved image_gen provider",
		zap.String("provider_name", cfg.ProviderName),
		zap.String("format", format),
	)

	switch format {
	case "vertex":
		region := config.Region
		if region == "" {
			region = "us-central1"
		}
		return image_gen.NewVertexImageGen(config.ProjectID, region, apiKey, config.Model, f.logger), nil
	case "openai":
		baseURL := config.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return image_gen.NewOpenAICompatImageGen(baseURL, apiKey, config.Model, f.logger), nil
	default:
		return nil, fmt.Errorf("providerFactory.ResolveImageGen: unknown format %q", format)
	}
}

// ---------------------------------------------------------------------------
// Test connection helper
// ---------------------------------------------------------------------------

// TestProviderConnection makes a minimal test API call to verify the provider works.
func TestProviderConnection(ctx context.Context, providerType, providerName, format, baseURL, apiKey, model string, logger *zap.Logger) error {
	switch providerType {
	case domain.ProviderTypeLLM:
		return testLLMConnection(ctx, providerName, format, baseURL, apiKey, model, logger)
	case domain.ProviderTypeImageGen:
		return testImageGenConnection(ctx, providerName, format, baseURL, apiKey, logger)
	default:
		return fmt.Errorf("test connection not supported for provider type %q", providerType)
	}
}

func testLLMConnection(ctx context.Context, providerName, format, baseURL, apiKey, model string, logger *zap.Logger) error {
	if format == "" {
		switch providerName {
		case "anthropic", "claude":
			format = "anthropic"
		default:
			format = "openai"
		}
	}

	messages := []adapter.LLMMessage{
		{Role: "user", Content: "Hello, respond with just the word OK."},
	}
	opts := adapter.LLMOptions{Model: model, MaxTokens: 16}

	var provider adapter.LLMProvider
	switch format {
	case "anthropic":
		provider = llm.NewAnthropicProvider(apiKey, "", logger)
	case "openai":
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		provider = llm.NewOpenAICompatProvider(baseURL, apiKey, "", logger)
	default:
		return fmt.Errorf("unknown LLM format %q", format)
	}

	_, err := provider.Chat(ctx, messages, opts)
	return err
}

func testImageGenConnection(ctx context.Context, providerName, format, baseURL, apiKey string, logger *zap.Logger) error {
	// For image gen, we just verify we can reach the endpoint.
	// A full generation would be expensive, so we send a minimal request
	// and check that the auth is accepted (even if the generation itself might fail for other reasons).
	if format == "" {
		switch providerName {
		case "vertex", "vertex_ai", "google", "imagen":
			format = "vertex"
		default:
			format = "openai"
		}
	}

	prompt := "A simple red dot on white background"
	opts := adapter.ImageGenOptions{Width: 256, Height: 256}

	var provider adapter.ImageGenProvider
	switch format {
	case "vertex":
		// Vertex requires project_id and region which we don't have here easily.
		// Return a message indicating test is not fully supported.
		return fmt.Errorf("vertex AI image gen test: please verify credentials via the Vertex AI console")
	case "openai":
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		provider = image_gen.NewOpenAICompatImageGen(baseURL, apiKey, "", logger)
	default:
		return fmt.Errorf("unknown image_gen format %q", format)
	}

	_, err := provider.Generate(ctx, prompt, opts)
	return err
}

// ---------------------------------------------------------------------------
// Lazy wrappers — implement adapter interfaces, resolve real provider per call
// ---------------------------------------------------------------------------

// LazyLLMProvider implements adapter.LLMProvider, resolving from DB on each call.
// Falls back to stub if no provider is configured.
type LazyLLMProvider struct {
	factory  *ProviderFactory
	fallback adapter.LLMProvider
	logger   *zap.Logger
}

// NewLazyLLMProvider creates a new LazyLLMProvider with an optional stub fallback.
func NewLazyLLMProvider(factory *ProviderFactory, fallback adapter.LLMProvider) *LazyLLMProvider {
	return &LazyLLMProvider{
		factory:  factory,
		fallback: fallback,
		logger:   factory.logger,
	}
}

// Chat resolves the LLM provider from DB and delegates the call.
func (l *LazyLLMProvider) Chat(ctx context.Context, messages []adapter.LLMMessage, opts adapter.LLMOptions) (*adapter.LLMResponse, error) {
	provider, err := l.factory.ResolveLLM(ctx)
	if err != nil {
		if l.fallback != nil {
			l.logger.Warn("LLM provider resolution failed, using fallback stub", zap.Error(err))
			return l.fallback.Chat(ctx, messages, opts)
		}
		return nil, err
	}
	return provider.Chat(ctx, messages, opts)
}

// Compile-time interface check.
var _ adapter.LLMProvider = (*LazyLLMProvider)(nil)

// LazyImageGenProvider implements adapter.ImageGenProvider, resolving from DB on each call.
// Falls back to stub if no provider is configured.
type LazyImageGenProvider struct {
	factory  *ProviderFactory
	fallback adapter.ImageGenProvider
	logger   *zap.Logger
}

// NewLazyImageGenProvider creates a new LazyImageGenProvider with an optional stub fallback.
func NewLazyImageGenProvider(factory *ProviderFactory, fallback adapter.ImageGenProvider) *LazyImageGenProvider {
	return &LazyImageGenProvider{
		factory:  factory,
		fallback: fallback,
		logger:   factory.logger,
	}
}

// Generate resolves the ImageGen provider from DB and delegates the call.
func (l *LazyImageGenProvider) Generate(ctx context.Context, prompt string, opts adapter.ImageGenOptions) (*adapter.GeneratedImage, error) {
	provider, err := l.factory.ResolveImageGen(ctx)
	if err != nil {
		if l.fallback != nil {
			l.logger.Warn("ImageGen provider resolution failed, using fallback stub", zap.Error(err))
			return l.fallback.Generate(ctx, prompt, opts)
		}
		return nil, err
	}
	return provider.Generate(ctx, prompt, opts)
}

// Compile-time interface check.
var _ adapter.ImageGenProvider = (*LazyImageGenProvider)(nil)
