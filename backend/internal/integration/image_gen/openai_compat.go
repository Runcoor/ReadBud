// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package image_gen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// OpenAICompatImageGen implements adapter.ImageGenProvider using the OpenAI-compatible
// images/generations API. Works with OpenAI DALL-E, and compatible services.
type OpenAICompatImageGen struct {
	baseURL      string
	apiKey       string
	defaultModel string
	client       *http.Client
	logger       *zap.Logger
}

// NewOpenAICompatImageGen creates a new OpenAI-compatible image generation provider.
func NewOpenAICompatImageGen(baseURL, apiKey, defaultModel string, logger *zap.Logger) *OpenAICompatImageGen {
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAICompatImageGen{
		baseURL:      baseURL,
		apiKey:       apiKey,
		defaultModel: defaultModel,
		client:       &http.Client{Timeout: 120 * time.Second},
		logger:       logger,
	}
}

type imageGenRequest struct {
	Model          string `json:"model,omitempty"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n"`
	ResponseFormat string `json:"response_format"`
}

type imageGenResponse struct {
	Data  []imageGenData `json:"data"`
	Error *imageGenError `json:"error,omitempty"`
}

type imageGenData struct {
	URL       string `json:"url"`
	B64JSON   string `json:"b64_json"`
}

type imageGenError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Generate creates an image using the OpenAI-compatible images/generations API.
func (p *OpenAICompatImageGen) Generate(ctx context.Context, prompt string, opts adapter.ImageGenOptions) (*adapter.GeneratedImage, error) {
	width := opts.Width
	height := opts.Height
	if width <= 0 {
		width = 1024
	}
	if height <= 0 {
		height = 1024
	}

	reqBody := imageGenRequest{
		Model:          p.defaultModel,
		Prompt:         prompt,
		Size:           fmt.Sprintf("%dx%d", width, height),
		N:              1,
		ResponseFormat: "url",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: marshal request: %w", err)
	}

	url := p.baseURL + "/images/generations"
	p.logger.Debug("openai-compat-image: sending request",
		zap.String("url", url),
		zap.String("size", reqBody.Size),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: read response: %w", err)
	}

	p.logger.Debug("openai-compat-image: received response",
		zap.Int("status", resp.StatusCode),
		zap.Int("body_len", len(respBody)),
	)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: rate limited (429), please retry later")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp imageGenResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != nil {
			return nil, fmt.Errorf("openAICompatImageGen.Generate: API error %d: %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("openAICompatImageGen.Generate: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var genResp imageGenResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: unmarshal response: %w", err)
	}

	if len(genResp.Data) == 0 {
		return nil, fmt.Errorf("openAICompatImageGen.Generate: no images in response")
	}

	return &adapter.GeneratedImage{
		URL:    genResp.Data[0].URL,
		Base64: genResp.Data[0].B64JSON,
		Width:  width,
		Height: height,
	}, nil
}

// Compile-time interface check.
var _ adapter.ImageGenProvider = (*OpenAICompatImageGen)(nil)
