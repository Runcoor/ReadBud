package llm

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

// OpenAICompatProvider implements adapter.LLMProvider using the OpenAI-compatible
// chat completions API. Works with OpenAI, DeepSeek, Moonshot, Qwen, OpenRouter, etc.
type OpenAICompatProvider struct {
	baseURL      string
	apiKey       string
	defaultModel string
	client       *http.Client
	logger       *zap.Logger
}

// NewOpenAICompatProvider creates a new OpenAI-compatible LLM provider.
func NewOpenAICompatProvider(baseURL, apiKey, defaultModel string, logger *zap.Logger) *OpenAICompatProvider {
	// Normalize base URL: strip trailing slash
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAICompatProvider{
		baseURL:      baseURL,
		apiKey:       apiKey,
		defaultModel: defaultModel,
		client:       &http.Client{Timeout: 180 * time.Second},
		logger:       logger,
	}
}

// openAIChatRequest is the request body for the chat completions API.
type openAIChatRequest struct {
	Model       string           `json:"model"`
	Messages    []openAIMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse is the response from the chat completions API.
type openAIChatResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIUsage struct {
	TotalTokens int `json:"total_tokens"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Chat sends a conversation to the OpenAI-compatible API and returns the response.
func (p *OpenAICompatProvider) Chat(ctx context.Context, messages []adapter.LLMMessage, opts adapter.LLMOptions) (*adapter.LLMResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("openAICompatProvider.Chat: empty messages")
	}

	// Build request
	msgs := make([]openAIMessage, len(messages))
	for i, m := range messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	model := opts.Model
	if model == "" {
		model = p.defaultModel
	}

	reqBody := openAIChatRequest{
		Model:    model,
		Messages: msgs,
	}
	if opts.MaxTokens > 0 {
		reqBody.MaxTokens = opts.MaxTokens
	}
	if opts.Temperature > 0 {
		temp := opts.Temperature
		reqBody.Temperature = &temp
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openAICompatProvider.Chat: marshal request: %w", err)
	}

	url := p.baseURL + "/chat/completions"
	p.logger.Debug("openai-compat: sending request",
		zap.String("url", url),
		zap.String("model", opts.Model),
		zap.Int("message_count", len(messages)),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("openAICompatProvider.Chat: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openAICompatProvider.Chat: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openAICompatProvider.Chat: read response: %w", err)
	}

	p.logger.Debug("openai-compat: received response",
		zap.Int("status", resp.StatusCode),
		zap.Int("body_len", len(respBody)),
	)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("openAICompatProvider.Chat: rate limited (429), please retry later")
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error message
		var errResp openAIChatResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != nil {
			return nil, fmt.Errorf("openAICompatProvider.Chat: API error %d: %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("openAICompatProvider.Chat: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("openAICompatProvider.Chat: unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openAICompatProvider.Chat: no choices in response")
	}

	return &adapter.LLMResponse{
		Content:      chatResp.Choices[0].Message.Content,
		FinishReason: chatResp.Choices[0].FinishReason,
		TokensUsed:   chatResp.Usage.TotalTokens,
	}, nil
}

// Compile-time interface check.
var _ adapter.LLMProvider = (*OpenAICompatProvider)(nil)
