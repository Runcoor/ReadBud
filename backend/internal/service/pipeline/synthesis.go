package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"readbud/internal/adapter"
	"readbud/internal/repository/postgres"
)

// SynthesisResult represents the cross-document synthesis output.
type SynthesisResult struct {
	ConsensusViewpoints   []string     `json:"consensus_viewpoints"`
	ConflictingViewpoints []string     `json:"conflicting_viewpoints"`
	FrequentTitlePatterns []string     `json:"frequent_title_patterns"`
	FrequentOpenings      []string     `json:"frequent_openings"`
	FrequentCTAStructures []string     `json:"frequent_cta_structures"`
	DataEvidence          []DataPoint  `json:"data_evidence"`
	ChartCandidates       []ChartCandidate `json:"chart_candidates"`
	ImageSuggestions      []string     `json:"image_suggestions"`
	KeyTakeaways          []string     `json:"key_takeaways"`
}

// ChartCandidate represents a potential chart that can be generated from the data.
type ChartCandidate struct {
	ChartType string      `json:"chart_type"` // line, bar, pie, donut
	Title     string      `json:"title"`
	Labels    []string    `json:"labels"`
	Values    []float64   `json:"values"`
	Unit      string      `json:"unit"`
	Source    string      `json:"source"`
}

// SynthesisService performs cross-document synthesis.
type SynthesisService struct {
	llm        adapter.LLMProvider
	sourceRepo postgres.SourceDocumentRepository
}

// NewSynthesisService creates a new SynthesisService.
func NewSynthesisService(llm adapter.LLMProvider, sourceRepo postgres.SourceDocumentRepository) *SynthesisService {
	return &SynthesisService{llm: llm, sourceRepo: sourceRepo}
}

const synthesisPrompt = `你是一个资深内容策划师。请分析以下多篇文章的结构化摘要，进行跨文档综合分析。

要求：
1. 共识观点（consensus_viewpoints）：多篇文章都提到的观点
2. 冲突观点（conflicting_viewpoints）：文章之间的矛盾之处
3. 高频标题模式（frequent_title_patterns）：常见的标题写法
4. 高频开头模式（frequent_openings）：常见的文章开头方式
5. 高频转化结构（frequent_cta_structures）：常见的CTA设计
6. 数据证据（data_evidence）：可用于文章的数据佐证
7. 图表候选（chart_candidates）：可以制作图表的数据组合（含chart_type/title/labels/values/unit/source）
8. 配图建议（image_suggestions）：建议搜索的图片关键词
9. 核心收获（key_takeaways）：写作时应重点传达的信息

仅输出JSON，不要其他文字。`

// Synthesize performs cross-document synthesis for a task's source documents.
func (s *SynthesisService) Synthesize(ctx context.Context, taskID int64) (*SynthesisResult, error) {
	docs, err := s.sourceRepo.FindByTaskIDOrderByScore(ctx, taskID, 10)
	if err != nil {
		return nil, fmt.Errorf("synthesisService.Synthesize: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("synthesisService.Synthesize: no source documents for task %d", taskID)
	}

	// Collect summaries from all analyzed documents
	var summaries []json.RawMessage
	for _, doc := range docs {
		if len(doc.SummaryJSON) > 0 {
			summaries = append(summaries, json.RawMessage(doc.SummaryJSON))
		}
	}

	if len(summaries) == 0 {
		return nil, fmt.Errorf("synthesisService.Synthesize: no analyzed documents for task %d", taskID)
	}

	// Build user message with all summaries
	summariesJSON, _ := json.Marshal(summaries)
	userMsg := fmt.Sprintf("以下是 %d 篇文章的结构化分析结果：\n\n%s", len(summaries), string(summariesJSON))

	resp, err := s.llm.Chat(ctx, []adapter.LLMMessage{
		{Role: "system", Content: synthesisPrompt},
		{Role: "user", Content: userMsg},
	}, adapter.LLMOptions{
		MaxTokens:   6000,
		Temperature: 0.4,
	})
	if err != nil {
		return nil, fmt.Errorf("synthesisService.Synthesize: LLM call: %w", err)
	}

	var result SynthesisResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("synthesisService.Synthesize: parse response: %w", err)
	}

	return &result, nil
}
