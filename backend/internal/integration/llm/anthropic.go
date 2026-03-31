package llm

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

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// AnthropicProvider implements adapter.LLMProvider using the Anthropic Messages API.
type AnthropicProvider struct {
	apiKey       string
	defaultModel string
	client       *http.Client
	logger       *zap.Logger
}

// NewAnthropicProvider creates a new Anthropic LLM provider.
func NewAnthropicProvider(apiKey, defaultModel string, logger *zap.Logger) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey:       apiKey,
		defaultModel: defaultModel,
		client:       &http.Client{Timeout: 180 * time.Second},
		logger:       logger,
	}
}

// anthropicRequest is the request body for the Anthropic Messages API.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the response from the Anthropic Messages API.
type anthropicResponse struct {
	Content    []anthropicContent `json:"content"`
	StopReason string             `json:"stop_reason"`
	Usage      anthropicUsage     `json:"usage"`
	Error      *anthropicError    `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// anthropicErrorResponse wraps a top-level error for non-200 responses.
type anthropicErrorResponse struct {
	Type  string          `json:"type"`
	Error *anthropicError `json:"error,omitempty"`
}

// Chat sends a conversation to the Anthropic Messages API and returns the response.
func (p *AnthropicProvider) Chat(ctx context.Context, messages []adapter.LLMMessage, opts adapter.LLMOptions) (*adapter.LLMResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("anthropicProvider.Chat: empty messages")
	}

	// Anthropic requires system messages to be separate. For simplicity,
	// we filter out system messages and only keep user/assistant messages.
	// If the first message is system, it stays as a regular message since
	// the simple API format supports it in the messages array.
	msgs := make([]anthropicMessage, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	model := opts.Model
	if model == "" {
		model = p.defaultModel
	}

	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  msgs,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("anthropicProvider.Chat: marshal request: %w", err)
	}

	p.logger.Debug("anthropic: sending request",
		zap.String("model", opts.Model),
		zap.Int("message_count", len(messages)),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("anthropicProvider.Chat: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropicProvider.Chat: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropicProvider.Chat: read response: %w", err)
	}

	p.logger.Debug("anthropic: received response",
		zap.Int("status", resp.StatusCode),
		zap.Int("body_len", len(respBody)),
	)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("anthropicProvider.Chat: rate limited (429), please retry later")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp anthropicErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != nil {
			return nil, fmt.Errorf("anthropicProvider.Chat: API error %d: %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("anthropicProvider.Chat: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp anthropicResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("anthropicProvider.Chat: unmarshal response: %w", err)
	}

	content := ""
	if len(chatResp.Content) > 0 {
		content = chatResp.Content[0].Text
	}

	return &adapter.LLMResponse{
		Content:      content,
		FinishReason: chatResp.StopReason,
		TokensUsed:   chatResp.Usage.InputTokens + chatResp.Usage.OutputTokens,
	}, nil
}

// Compile-time interface check.
var _ adapter.LLMProvider = (*AnthropicProvider)(nil)
