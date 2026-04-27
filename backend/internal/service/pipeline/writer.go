// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	"readbud/internal/adapter"
	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
)

// ArticleBrief represents the brief/outline generated as Step 1 of article writing.
type ArticleBrief struct {
	CoreViewpoint string        `json:"core_viewpoint"`
	TargetAudience string       `json:"target_audience"`
	ArticleGoal   string        `json:"article_goal"` // 阅读/转发/留资/关注
	TitleCandidates []TitleCandidate `json:"title_candidates"`
	Outline       []OutlineItem `json:"outline"`
	ImageSlots    []string      `json:"image_slots"`
	ChartSlots    []string      `json:"chart_slots"`
	CTADesign     string        `json:"cta_design"`
}

// TitleCandidate is a generated title with its type label.
type TitleCandidate struct {
	Title string `json:"title"`
	Type  string `json:"type"` // 数字型/对比型/反常识型/清单型/问题型/场景型
}

// OutlineItem represents a section in the article outline.
type OutlineItem struct {
	Heading  string `json:"heading"`
	Points   []string `json:"points"`
	HasImage bool   `json:"has_image"`
	HasChart bool   `json:"has_chart"`
}

// BlockContent represents a generated article block.
type BlockContent struct {
	BlockType string  `json:"block_type"`
	Heading   string  `json:"heading,omitempty"`
	TextMD    string  `json:"text_md"`
	ImageQuery string `json:"image_query,omitempty"`
	ChartData  string `json:"chart_data,omitempty"`
}

// WriterService handles article brief generation and block-level writing.
type WriterService struct {
	llm       adapter.LLMProvider
	draftRepo postgres.ArticleDraftRepository
	blockRepo postgres.ArticleBlockRepository
}

// NewWriterService creates a new WriterService.
func NewWriterService(
	llm adapter.LLMProvider,
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
) *WriterService {
	return &WriterService{llm: llm, draftRepo: draftRepo, blockRepo: blockRepo}
}

const briefPrompt = `你是一个资深公众号内容策划师。根据以下综合分析结果和目标参数，生成文章创作大纲（Brief）。

要求：
1. core_viewpoint: 文章核心观点（一句话）
2. target_audience: 目标读者画像
3. article_goal: 文章目标（阅读/转发/留资/关注）
4. title_candidates: 10个标题候选，每个包含title和type（数字型/对比型/反常识型/清单型/问题型/场景型）
5. outline: 文章结构大纲，每个section包含heading、points、has_image、has_chart
6. image_slots: 需要配图的位置描述
7. chart_slots: 需要图表的位置描述
8. cta_design: CTA设计方案

仅输出JSON。`

// GenerateBrief creates an article brief from synthesis results (Step 1 of 3-stage generation).
func (s *WriterService) GenerateBrief(ctx context.Context, taskID int64, keyword string, audience string, tone string, targetWords int, synthesis *SynthesisResult) (*ArticleBrief, *draft.ArticleDraft, error) {
	synthJSON, _ := json.Marshal(synthesis)
	userMsg := fmt.Sprintf("关键词: %s\n目标读者: %s\n写作风格: %s\n目标字数: %d\n\n综合分析结果:\n%s",
		keyword, audience, tone, targetWords, string(synthJSON))

	resp, err := s.llm.Chat(ctx, []adapter.LLMMessage{
		{Role: "system", Content: briefPrompt},
		{Role: "user", Content: userMsg},
	}, adapter.LLMOptions{
		MaxTokens:   4000,
		Temperature: 0.5,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("writerService.GenerateBrief: LLM call: %w", err)
	}

	var brief ArticleBrief
	if err := json.Unmarshal([]byte(resp.Content), &brief); err != nil {
		return nil, nil, fmt.Errorf("writerService.GenerateBrief: parse response: %w", err)
	}

	// Pick the first title candidate as the working title
	title := keyword
	if len(brief.TitleCandidates) > 0 {
		title = brief.TitleCandidates[0].Title
	}

	outlineJSON, _ := json.Marshal(brief)

	// Create the article draft
	d := &draft.ArticleDraft{
		TaskID:      taskID,
		Title:       title,
		Digest:      brief.CoreViewpoint,
		AuthorName:  "阅芽",
		OutlineJSON: datatypes.JSON(outlineJSON),
		Version:     1,
	}

	if err := s.draftRepo.Create(ctx, d); err != nil {
		return nil, nil, fmt.Errorf("writerService.GenerateBrief: save draft: %w", err)
	}

	return &brief, d, nil
}

const blockWritePrompt = `你是一个专业的公众号文章作者。根据大纲中的章节信息，撰写该章节的内容。

要求：
1. 输出Markdown格式
2. 语言生动、有观点、有数据支撑
3. 段落精炼，每段不超过150字
4. 如果需要配图，在image_query中给出搜索关键词
5. 如果需要图表，在chart_data中描述数据

输出JSON格式: {"block_type":"section","heading":"","text_md":"","image_query":"","chart_data":""}`

// GenerateBlocks generates article blocks from the brief (Step 2 of 3-stage generation).
func (s *WriterService) GenerateBlocks(ctx context.Context, d *draft.ArticleDraft, brief *ArticleBrief) ([]draft.ArticleBlock, error) {
	var blocks []draft.ArticleBlock
	sortNo := 1

	// Title block
	titleMD := fmt.Sprintf("# %s", d.Title)
	blocks = append(blocks, draft.ArticleBlock{
		DraftID:   d.ID,
		SortNo:    sortNo,
		BlockType: draft.BlockTypeTitle,
		Heading:   &d.Title,
		TextMD:    &titleMD,
		Status:    "active",
	})
	sortNo++

	// Generate each section via LLM
	for _, section := range brief.Outline {
		sectionJSON, _ := json.Marshal(section)
		userMsg := fmt.Sprintf("文章主题: %s\n核心观点: %s\n\n当前章节:\n%s",
			d.Title, brief.CoreViewpoint, string(sectionJSON))

		resp, err := s.llm.Chat(ctx, []adapter.LLMMessage{
			{Role: "system", Content: blockWritePrompt},
			{Role: "user", Content: userMsg},
		}, adapter.LLMOptions{
			MaxTokens:   2000,
			Temperature: 0.6,
		})
		if err != nil {
			return nil, fmt.Errorf("writerService.GenerateBlocks: LLM call for section %q: %w", section.Heading, err)
		}

		var bc BlockContent
		if err := json.Unmarshal([]byte(resp.Content), &bc); err != nil {
			// Fallback: treat entire response as markdown text
			bc = BlockContent{
				BlockType: draft.BlockTypeSection,
				Heading:   section.Heading,
				TextMD:    resp.Content,
			}
		}

		heading := bc.Heading
		if heading == "" {
			heading = section.Heading
		}
		textMD := bc.TextMD

		blocks = append(blocks, draft.ArticleBlock{
			DraftID:   d.ID,
			SortNo:    sortNo,
			BlockType: draft.BlockTypeSection,
			Heading:   &heading,
			TextMD:    &textMD,
			Status:    "active",
		})
		sortNo++

		// Add image block if this section needs one
		if section.HasImage && bc.ImageQuery != "" {
			blocks = append(blocks, draft.ArticleBlock{
				DraftID:    d.ID,
				SortNo:     sortNo,
				BlockType:  draft.BlockTypeImage,
				PromptText: &bc.ImageQuery,
				Status:     "active",
			})
			sortNo++
		}

		// Add chart block if this section needs one
		if section.HasChart && bc.ChartData != "" {
			blocks = append(blocks, draft.ArticleBlock{
				DraftID:    d.ID,
				SortNo:     sortNo,
				BlockType:  draft.BlockTypeChart,
				PromptText: &bc.ChartData,
				Status:     "active",
			})
			sortNo++
		}
	}

	// CTA block
	if brief.CTADesign != "" {
		ctaMD := brief.CTADesign
		blocks = append(blocks, draft.ArticleBlock{
			DraftID:   d.ID,
			SortNo:    sortNo,
			BlockType: draft.BlockTypeCTA,
			TextMD:    &ctaMD,
			Status:    "active",
		})
	}

	// Batch save all blocks
	if err := s.blockRepo.CreateBatch(ctx, blocks); err != nil {
		return nil, fmt.Errorf("writerService.GenerateBlocks: save blocks: %w", err)
	}

	return blocks, nil
}
