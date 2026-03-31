package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain"
	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
)

// DistributionVO is the view-object returned by distribution package API endpoints.
type DistributionVO struct {
	PublicID            string    `json:"public_id"`
	DraftID             int64     `json:"draft_id"`
	CommunityCopy       string    `json:"community_copy"`
	MomentsCopy         string    `json:"moments_copy"`
	SummaryCard         string    `json:"summary_card"`
	CommentGuide        string    `json:"comment_guide"`
	NextTopicSuggestion string    `json:"next_topic_suggestion"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GenerateDistributionRequest is the DTO for triggering distribution generation.
type GenerateDistributionRequest struct {
	DraftPublicID string `json:"draft_public_id" binding:"required"`
}

// distributionLLMOutput is the structured output expected from the LLM.
// NextTopicSuggestion uses json.RawMessage because LLM may return a string or array.
type distributionLLMOutput struct {
	CommunityCopy       string          `json:"community_copy"`
	MomentsCopy         string          `json:"moments_copy"`
	SummaryCard         string          `json:"summary_card"`
	CommentGuide        string          `json:"comment_guide"`
	NextTopicSuggestion json.RawMessage `json:"next_topic_suggestion"`
}

// nextTopicAsString normalises the LLM output to a single string.
func (o *distributionLLMOutput) nextTopicAsString() string {
	if len(o.NextTopicSuggestion) == 0 {
		return ""
	}
	// Try string first
	var s string
	if json.Unmarshal(o.NextTopicSuggestion, &s) == nil {
		return s
	}
	// Try array of strings
	var arr []string
	if json.Unmarshal(o.NextTopicSuggestion, &arr) == nil {
		return strings.Join(arr, "\n")
	}
	// Fallback: raw text
	return string(o.NextTopicSuggestion)
}

// DistributionService handles distribution package generation and retrieval.
type DistributionService struct {
	distRepo  postgres.DistributionPackageRepository
	draftRepo postgres.ArticleDraftRepository
	blockRepo postgres.ArticleBlockRepository
	llm       adapter.LLMProvider
	logger    *zap.Logger
}

// NewDistributionService creates a new DistributionService.
func NewDistributionService(
	distRepo postgres.DistributionPackageRepository,
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
	llm adapter.LLMProvider,
	logger *zap.Logger,
) *DistributionService {
	return &DistributionService{
		distRepo:  distRepo,
		draftRepo: draftRepo,
		blockRepo: blockRepo,
		llm:       llm,
		logger:    logger,
	}
}

// Generate creates distribution materials for a given draft using LLM.
func (s *DistributionService) Generate(ctx context.Context, draftPublicID string) (*DistributionVO, error) {
	// 1. Find the draft
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("distributionService.Generate: find draft: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	// 2. Load the article blocks to build content
	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("distributionService.Generate: find blocks: %w", err)
	}

	articleContent := buildArticleText(d, blocks)

	// 3. Call LLM to generate distribution materials
	output, err := s.callLLM(ctx, d.Title, d.Digest, articleContent)
	if err != nil {
		return nil, fmt.Errorf("distributionService.Generate: llm: %w", err)
	}

	// 4. Upsert distribution package
	pkg := &domain.DistributionPackage{
		DraftID:             d.ID,
		CommunityCopy:       output.CommunityCopy,
		MomentsCopy:         output.MomentsCopy,
		SummaryCard:         output.SummaryCard,
		CommentGuide:        output.CommentGuide,
		NextTopicSuggestion: output.nextTopicAsString(),
	}

	if err := s.distRepo.Upsert(ctx, pkg); err != nil {
		return nil, fmt.Errorf("distributionService.Generate: upsert: %w", err)
	}

	// Re-fetch to get the complete record with public_id and timestamps
	saved, err := s.distRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("distributionService.Generate: refetch: %w", err)
	}

	s.logger.Info("distribution package generated",
		zap.String("draft_public_id", draftPublicID),
		zap.String("package_public_id", saved.PublicID),
	)

	return toDistributionVO(saved), nil
}

// GetByDraftPublicID retrieves a distribution package by draft public ID.
func (s *DistributionService) GetByDraftPublicID(ctx context.Context, draftPublicID string) (*DistributionVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("distributionService.GetByDraftPublicID: find draft: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	pkg, err := s.distRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("distributionService.GetByDraftPublicID: find package: %w", err)
	}
	if pkg == nil {
		return nil, ErrNotFound
	}

	return toDistributionVO(pkg), nil
}

// GetByPublicID retrieves a distribution package by its own public ID.
func (s *DistributionService) GetByPublicID(ctx context.Context, publicID string) (*DistributionVO, error) {
	pkg, err := s.distRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("distributionService.GetByPublicID: %w", err)
	}
	if pkg == nil {
		return nil, ErrNotFound
	}

	return toDistributionVO(pkg), nil
}

// Delete removes a distribution package by public ID.
func (s *DistributionService) Delete(ctx context.Context, publicID string) error {
	pkg, err := s.distRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("distributionService.Delete: find: %w", err)
	}
	if pkg == nil {
		return ErrNotFound
	}

	if err := s.distRepo.Delete(ctx, pkg.ID); err != nil {
		return fmt.Errorf("distributionService.Delete: %w", err)
	}

	s.logger.Info("distribution package deleted",
		zap.String("public_id", publicID),
	)
	return nil
}

// callLLM sends article content to LLM and parses the structured response.
func (s *DistributionService) callLLM(ctx context.Context, title, digest, content string) (*distributionLLMOutput, error) {
	systemPrompt := `你是一位资深的微信公众号运营专家。根据提供的文章标题、摘要和正文内容，生成以下分发素材包。

要求：
1. 社群文案（community_copy）：100字以内，适合微信群转发，包含核心观点和互动引导
2. 朋友圈文案（moments_copy）：简短精炼，适合朋友圈分享，带情感共鸣或悬念
3. 摘要卡片（summary_card）：50字以内，提炼文章核心价值
4. 评论区引导语（comment_guide）：引导读者参与评论互动，提出开放性问题
5. 下篇选题建议（next_topic_suggestion）：基于本文内容延伸，推荐 2-3 个后续选题方向

请以 JSON 格式返回，字段名为：community_copy, moments_copy, summary_card, comment_guide, next_topic_suggestion`

	userPrompt := fmt.Sprintf("文章标题：%s\n\n文章摘要：%s\n\n文章正文：\n%s", title, digest, content)

	messages := []adapter.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	opts := adapter.LLMOptions{
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	resp, err := s.llm.Chat(ctx, messages, opts)
	if err != nil {
		return nil, fmt.Errorf("callLLM: chat: %w", err)
	}

	// Parse the JSON response
	output, err := parseLLMDistributionOutput(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("callLLM: parse: %w", err)
	}

	return output, nil
}

// parseLLMDistributionOutput extracts JSON from LLM response which may contain markdown fences.
func parseLLMDistributionOutput(raw string) (*distributionLLMOutput, error) {
	cleaned := raw

	// Strip markdown code fences if present
	if idx := strings.Index(cleaned, "```json"); idx != -1 {
		cleaned = cleaned[idx+7:]
	} else if idx := strings.Index(cleaned, "```"); idx != -1 {
		cleaned = cleaned[idx+3:]
	}
	if idx := strings.LastIndex(cleaned, "```"); idx != -1 {
		cleaned = cleaned[:idx]
	}
	cleaned = strings.TrimSpace(cleaned)

	var output distributionLLMOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, fmt.Errorf("parseLLMDistributionOutput: %w", err)
	}

	return &output, nil
}

// buildArticleText concatenates draft title, digest, and block content into plain text.
func buildArticleText(d *draft.ArticleDraft, blocks []draft.ArticleBlock) string {
	var sb strings.Builder
	sb.WriteString(d.Title)
	sb.WriteString("\n\n")
	sb.WriteString(d.Digest)
	sb.WriteString("\n\n")

	for _, b := range blocks {
		if b.Heading != nil && *b.Heading != "" {
			sb.WriteString(*b.Heading)
			sb.WriteString("\n")
		}
		if b.TextMD != nil && *b.TextMD != "" {
			sb.WriteString(*b.TextMD)
			sb.WriteString("\n\n")
		}
	}

	// Truncate to avoid exceeding LLM context limits
	const maxContentLen = 8000
	text := sb.String()
	if len(text) > maxContentLen {
		text = text[:maxContentLen] + "\n...(内容已截断)"
	}

	return text
}

// toDistributionVO converts a domain DistributionPackage to a DistributionVO.
func toDistributionVO(pkg *domain.DistributionPackage) *DistributionVO {
	return &DistributionVO{
		PublicID:            pkg.PublicID,
		DraftID:             pkg.DraftID,
		CommunityCopy:       pkg.CommunityCopy,
		MomentsCopy:         pkg.MomentsCopy,
		SummaryCard:         pkg.SummaryCard,
		CommentGuide:        pkg.CommentGuide,
		NextTopicSuggestion: pkg.NextTopicSuggestion,
		CreatedAt:           pkg.CreatedAt,
		UpdatedAt:           pkg.UpdatedAt,
	}
}
