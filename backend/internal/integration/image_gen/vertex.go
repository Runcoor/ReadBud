package image_gen

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

// VertexImageGen implements adapter.ImageGenProvider using Google Vertex AI Imagen API.
type VertexImageGen struct {
	projectID string
	region    string
	apiKey    string
	model     string
	client    *http.Client
	logger    *zap.Logger
}

// NewVertexImageGen creates a new Vertex AI Imagen provider.
func NewVertexImageGen(projectID, region, apiKey, model string, logger *zap.Logger) *VertexImageGen {
	if model == "" {
		model = "imagen-4.0-ultra-generate-001"
	}
	return &VertexImageGen{
		projectID: projectID,
		region:    region,
		apiKey:    apiKey,
		model:     model,
		client:    &http.Client{Timeout: 120 * time.Second},
		logger:    logger,
	}
}

type vertexRequest struct {
	Instances  []vertexInstance  `json:"instances"`
	Parameters vertexParameters  `json:"parameters"`
}

type vertexInstance struct {
	Prompt string `json:"prompt"`
}

type vertexParameters struct {
	SampleCount int    `json:"sampleCount"`
	AspectRatio string `json:"aspectRatio,omitempty"`
}

type vertexResponse struct {
	Predictions []vertexPrediction `json:"predictions"`
	Error       *vertexError       `json:"error,omitempty"`
}

type vertexPrediction struct {
	BytesBase64Encoded string `json:"bytesBase64Encoded"`
	MimeType           string `json:"mimeType"`
}

type vertexError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// aspectRatioFromDimensions determines the Vertex AI aspect ratio string.
func aspectRatioFromDimensions(w, h int) string {
	if w <= 0 || h <= 0 {
		return "1:1"
	}
	ratio := float64(w) / float64(h)
	switch {
	case ratio > 1.7:
		return "16:9"
	case ratio > 1.4:
		return "3:2"
	case ratio > 1.1:
		return "4:3"
	case ratio < 0.6:
		return "9:16"
	case ratio < 0.7:
		return "2:3"
	case ratio < 0.9:
		return "3:4"
	default:
		return "1:1"
	}
}

// Generate creates an image using the Vertex AI Imagen API.
func (p *VertexImageGen) Generate(ctx context.Context, prompt string, opts adapter.ImageGenOptions) (*adapter.GeneratedImage, error) {
	aspectRatio := aspectRatioFromDimensions(opts.Width, opts.Height)

	reqBody := vertexRequest{
		Instances: []vertexInstance{
			{Prompt: prompt},
		},
		Parameters: vertexParameters{
			SampleCount: 1,
			AspectRatio: aspectRatio,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("vertexImageGen.Generate: marshal request: %w", err)
	}

	url := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		p.region, p.projectID, p.region, p.model,
	)

	p.logger.Debug("vertex-imagen: sending request",
		zap.String("url", url),
		zap.String("aspect_ratio", aspectRatio),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("vertexImageGen.Generate: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vertexImageGen.Generate: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("vertexImageGen.Generate: read response: %w", err)
	}

	p.logger.Debug("vertex-imagen: received response",
		zap.Int("status", resp.StatusCode),
		zap.Int("body_len", len(respBody)),
	)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("vertexImageGen.Generate: rate limited (429), please retry later")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp vertexResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != nil {
			return nil, fmt.Errorf("vertexImageGen.Generate: API error %d: %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("vertexImageGen.Generate: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var genResp vertexResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return nil, fmt.Errorf("vertexImageGen.Generate: unmarshal response: %w", err)
	}

	if len(genResp.Predictions) == 0 {
		return nil, fmt.Errorf("vertexImageGen.Generate: no predictions in response")
	}

	width := opts.Width
	height := opts.Height
	if width <= 0 {
		width = 1024
	}
	if height <= 0 {
		height = 1024
	}

	return &adapter.GeneratedImage{
		Base64: genResp.Predictions[0].BytesBase64Encoded,
		Width:  width,
		Height: height,
	}, nil
}

// Compile-time interface check.
var _ adapter.ImageGenProvider = (*VertexImageGen)(nil)
