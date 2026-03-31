package llm

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// StubLLMProvider is a placeholder implementation of adapter.LLMProvider
// for development and testing. Returns realistic-looking stub responses.
type StubLLMProvider struct {
	logger *zap.Logger
}

// NewStubLLMProvider creates a new stub LLM provider.
func NewStubLLMProvider(logger *zap.Logger) *StubLLMProvider {
	return &StubLLMProvider{logger: logger}
}

// Chat simulates an LLM conversation and returns a stub response.
// The stub detects distribution package prompts and returns structured JSON.
func (s *StubLLMProvider) Chat(ctx context.Context, messages []adapter.LLMMessage, opts adapter.LLMOptions) (*adapter.LLMResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("stubLLMProvider.Chat: empty messages")
	}

	s.logger.Info("stub: LLM chat called",
		zap.Int("message_count", len(messages)),
		zap.String("model", opts.Model),
		zap.Int("max_tokens", opts.MaxTokens),
	)

	// Return a distribution-style JSON response for distribution prompts
	content := `{
  "community_copy": "这篇文章深入剖析了行业最新趋势，从数据到案例都非常扎实。推荐大家花5分钟读一下，特别是第三部分的实操建议，马上就能用起来。欢迎在群里分享你的看法！",
  "moments_copy": "读完这篇，才发现之前的认知都是错的。第三个观点颠覆了我对这个领域的理解，值得反复看。",
  "summary_card": "深度解析行业新趋势，三个核心观点助你把握未来方向。",
  "comment_guide": "你在工作中遇到过类似的问题吗？文中提到的第二个方法你觉得实用吗？欢迎在评论区分享你的经验和看法。",
  "next_topic_suggestion": "1. 深入探讨文中提到的技术方案在不同场景下的应用实践\n2. 对比分析国内外同类解决方案的优劣\n3. 采访行业专家对该趋势的深度解读"
}`

	return &adapter.LLMResponse{
		Content:      content,
		FinishReason: "stop",
		TokensUsed:   350,
	}, nil
}

// Compile-time check that StubLLMProvider satisfies the interface.
var _ adapter.LLMProvider = (*StubLLMProvider)(nil)
