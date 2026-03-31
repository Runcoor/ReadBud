package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"readbud/internal/adapter"
	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
)

// ReviewResult holds the scores returned by the LLM quality assessment.
type ReviewResult struct {
	SimilarityScore float64 `json:"similarity_score"`
	QualityScore    float64 `json:"quality_score"`
	RiskLevel       string  `json:"risk_level"`
	Issues          []string `json:"issues,omitempty"`
}

// ReviewService assesses article quality, similarity, and risk.
type ReviewService struct {
	llm       adapter.LLMProvider
	draftRepo postgres.ArticleDraftRepository
	blockRepo postgres.ArticleBlockRepository
	sourceRepo postgres.SourceDocumentRepository
	brandRepo  postgres.BrandProfileRepository
}

// NewReviewService creates a new ReviewService.
func NewReviewService(
	llm adapter.LLMProvider,
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
	sourceRepo postgres.SourceDocumentRepository,
	brandRepo postgres.BrandProfileRepository,
) *ReviewService {
	return &ReviewService{
		llm:        llm,
		draftRepo:  draftRepo,
		blockRepo:  blockRepo,
		sourceRepo: sourceRepo,
		brandRepo:  brandRepo,
	}
}

const reviewPrompt = `你是一位专业的内容审核员。请对以下文章草稿进行质量审核。

审核维度：
1. similarity_score (0-100): 与源文档的相似度，越低越好（<30 优秀，30-60 可接受，>60 需修改）
2. quality_score (0-100): 内容质量评分（结构、逻辑、可读性、数据支撑）
3. risk_level: 风险等级 "low" / "medium" / "high"
   - high: 包含敏感词、违禁词、侵权风险
   - medium: 相似度偏高或质量偏低
   - low: 通过审核
4. issues: 具体问题列表（如有）

额外要求：检查是否包含以下禁用词，如包含则标记为 high 风险：
%s

仅输出JSON格式。`

// Review performs quality review on an article draft.
func (s *ReviewService) Review(ctx context.Context, draftID int64, taskID int64) (*ReviewResult, error) {
	// Load draft
	d, err := s.draftRepo.FindByID(ctx, draftID)
	if err != nil {
		return nil, fmt.Errorf("reviewService.Review: load draft: %w", err)
	}
	if d == nil {
		return nil, fmt.Errorf("reviewService.Review: draft %d not found", draftID)
	}

	// Load blocks
	blocks, err := s.blockRepo.FindByDraftID(ctx, draftID)
	if err != nil {
		return nil, fmt.Errorf("reviewService.Review: load blocks: %w", err)
	}

	// Load source documents for comparison
	sources, err := s.sourceRepo.FindByTaskIDOrderByScore(ctx, taskID, 5)
	if err != nil {
		return nil, fmt.Errorf("reviewService.Review: load sources: %w", err)
	}

	// Load forbidden words from default brand profile
	forbiddenWords := "（无禁用词列表）"
	brandProfile, err := s.brandRepo.FindDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("reviewService.Review: load brand profile: %w", err)
	}
	if brandProfile != nil && len(brandProfile.ForbiddenWords) > 0 {
		var words []string
		if jsonErr := json.Unmarshal(brandProfile.ForbiddenWords, &words); jsonErr == nil && len(words) > 0 {
			forbiddenWords = strings.Join(words, "、")
		}
	}

	// Build article text from blocks
	var articleText strings.Builder
	articleText.WriteString(fmt.Sprintf("标题: %s\n摘要: %s\n\n", d.Title, d.Digest))
	for _, b := range blocks {
		if b.Status != "active" {
			continue
		}
		if b.Heading != nil && *b.Heading != "" {
			articleText.WriteString(fmt.Sprintf("## %s\n", *b.Heading))
		}
		if b.TextMD != nil && *b.TextMD != "" {
			articleText.WriteString(*b.TextMD)
			articleText.WriteString("\n\n")
		}
	}

	// Build source summaries
	var sourceText strings.Builder
	for i, src := range sources {
		sourceText.WriteString(fmt.Sprintf("源文档%d: %s\n", i+1, src.Title))
		if len(src.SummaryJSON) > 0 {
			sourceText.WriteString(string(src.SummaryJSON))
			sourceText.WriteString("\n")
		}
	}

	systemMsg := fmt.Sprintf(reviewPrompt, forbiddenWords)
	userMsg := fmt.Sprintf("=== 文章草稿 ===\n%s\n\n=== 源文档摘要 ===\n%s",
		articleText.String(), sourceText.String())

	resp, err := s.llm.Chat(ctx, []adapter.LLMMessage{
		{Role: "system", Content: systemMsg},
		{Role: "user", Content: userMsg},
	}, adapter.LLMOptions{
		MaxTokens:   2000,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("reviewService.Review: LLM call: %w", err)
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("reviewService.Review: parse response: %w", err)
	}

	// Validate risk level
	switch result.RiskLevel {
	case draft.RiskLow, draft.RiskMedium, draft.RiskHigh:
		// valid
	default:
		result.RiskLevel = draft.RiskMedium
	}

	// Update draft review fields
	d.QualityScore = result.QualityScore
	d.SimilarityScore = result.SimilarityScore
	d.RiskLevel = result.RiskLevel

	if result.RiskLevel == draft.RiskHigh {
		d.ReviewStatus = draft.ReviewReject
	} else {
		d.ReviewStatus = draft.ReviewPass
	}

	if err := s.draftRepo.Update(ctx, d); err != nil {
		return nil, fmt.Errorf("reviewService.Review: update draft: %w", err)
	}

	return &result, nil
}
