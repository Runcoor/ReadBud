// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Package pipeline implements the content generation pipeline services.
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	"readbud/internal/adapter"
	"readbud/internal/domain/source"
	"readbud/internal/repository/postgres"
)

// AnalysisResult represents the structured analysis of a single source document.
type AnalysisResult struct {
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	CoreViewpoints  []string `json:"core_viewpoints"`
	PainPoints      []string `json:"pain_points"`
	UseCases        []string `json:"use_cases"`
	DataPoints      []DataPoint `json:"data_points"`
	Quotes          []string `json:"quotes"`
	ContentOutline  []string `json:"content_outline"`
	ImageClues      []string `json:"image_clues"`
	CTAStyle        string   `json:"cta_style"`
	Tone            string   `json:"tone"`
	SourceURL       string   `json:"source_url"`
}

// DataPoint represents a single numeric data point extracted from content.
type DataPoint struct {
	Label     string  `json:"label"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Source    string  `json:"source"`
	Dimension string  `json:"dimension"`
}

// AnalysisService performs structured content analysis on source documents.
type AnalysisService struct {
	llm       adapter.LLMProvider
	sourceRepo postgres.SourceDocumentRepository
}

// NewAnalysisService creates a new AnalysisService.
func NewAnalysisService(llm adapter.LLMProvider, sourceRepo postgres.SourceDocumentRepository) *AnalysisService {
	return &AnalysisService{llm: llm, sourceRepo: sourceRepo}
}

// analysisPrompt is the system prompt for content analysis.
const analysisPrompt = `你是一个专业的内容分析师。请分析以下文章内容，输出结构化JSON。

要求：
1. 提取核心观点（core_viewpoints）：文章的主要论点
2. 提取痛点（pain_points）：文章描述的用户痛点或行业问题
3. 提取使用场景（use_cases）：文章涉及的应用场景
4. 提取数据点（data_points）：所有数值数据，包含标签、数值、单位、来源、维度
5. 提取引言（quotes）：有价值的引用句子
6. 提取内容大纲（content_outline）：文章的结构层次
7. 提取图片线索（image_clues）：可用于配图搜索的关键词
8. 判断CTA风格（cta_style）和文章语调（tone）

仅输出JSON，不要其他文字。`

// AnalyzeDocument performs structured analysis on a single source document.
func (s *AnalysisService) AnalyzeDocument(ctx context.Context, doc *source.SourceDocument) (*AnalysisResult, error) {
	if doc.PlainText == "" {
		return nil, fmt.Errorf("analysisService: document %d has no plain text", doc.ID)
	}

	// Truncate content to avoid exceeding token limits
	content := doc.PlainText
	if len(content) > 12000 {
		content = content[:12000]
	}

	userMsg := fmt.Sprintf("标题: %s\n来源: %s\n\n正文:\n%s", doc.Title, doc.SourceURL, content)

	resp, err := s.llm.Chat(ctx, []adapter.LLMMessage{
		{Role: "system", Content: analysisPrompt},
		{Role: "user", Content: userMsg},
	}, adapter.LLMOptions{
		MaxTokens:   4000,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("analysisService.AnalyzeDocument: LLM call: %w", err)
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("analysisService.AnalyzeDocument: parse response: %w", err)
	}
	result.SourceURL = doc.SourceURL

	return &result, nil
}

// AnalyzeAndStore analyzes a document and stores the result in the database.
func (s *AnalysisService) AnalyzeAndStore(ctx context.Context, doc *source.SourceDocument) error {
	result, err := s.AnalyzeDocument(ctx, doc)
	if err != nil {
		return err
	}

	summaryBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("analysisService: marshal summary: %w", err)
	}
	doc.SummaryJSON = datatypes.JSON(summaryBytes)

	// Extract and store data points separately for chart detection
	if len(result.DataPoints) > 0 {
		dpBytes, _ := json.Marshal(result.DataPoints)
		doc.DataPointsJSON = datatypes.JSON(dpBytes)
	}

	// Store image clues for image matching
	if len(result.ImageClues) > 0 {
		icBytes, _ := json.Marshal(result.ImageClues)
		doc.ImageCluesJSON = datatypes.JSON(icBytes)
	}

	return s.sourceRepo.Update(ctx, doc)
}

// AnalyzeAllForTask analyzes all source documents for a given task.
func (s *AnalysisService) AnalyzeAllForTask(ctx context.Context, taskID int64) error {
	docs, err := s.sourceRepo.FindByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("analysisService.AnalyzeAllForTask: %w", err)
	}

	for i := range docs {
		if err := s.AnalyzeAndStore(ctx, &docs[i]); err != nil {
			// Log error but continue with other documents
			fmt.Printf("[analysis] failed to analyze doc %d: %v\n", docs[i].ID, err)
			continue
		}
	}
	return nil
}
