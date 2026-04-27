// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"context"
	"fmt"

	"readbud/internal/adapter"
	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
)

// DraftVO is the view object returned to the frontend.
type DraftVO struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	Title           string    `json:"title"`
	Subtitle        *string   `json:"subtitle,omitempty"`
	Digest          string    `json:"digest"`
	AuthorName      string    `json:"author_name"`
	CoverURL        *string   `json:"cover_url,omitempty"`
	QualityScore    float64   `json:"quality_score"`
	SimilarityScore float64   `json:"similarity_score"`
	RiskLevel       string    `json:"risk_level"`
	ReviewStatus    string    `json:"review_status"`
	Version         int       `json:"version"`
	Blocks          []BlockVO `json:"blocks"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

// BlockVO is the view object for an article block.
type BlockVO struct {
	ID              string  `json:"id"`
	SortNo          int     `json:"sort_no"`
	BlockType       string  `json:"block_type"`
	Heading         *string `json:"heading,omitempty"`
	TextMD          *string `json:"text_md,omitempty"`
	HTMLFragment    *string `json:"html_fragment,omitempty"`
	AssetURL        *string `json:"asset_url,omitempty"`
	AttributionText *string `json:"attribution_text,omitempty"`
	PromptText      *string `json:"prompt_text,omitempty"`
	Status          string  `json:"status"`
}

// SourceVO is the view object for a source document.
type SourceVO struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	SourceType     string  `json:"source_type"`
	SiteName       string  `json:"site_name"`
	SourceURL      string  `json:"source_url"`
	Author         *string `json:"author,omitempty"`
	PublishedAt    *string `json:"published_at,omitempty"`
	HotScore       float64 `json:"hot_score"`
	RelevanceScore float64 `json:"relevance_score"`
	Summary        *string `json:"summary,omitempty"`
}

// DraftService handles article draft operations.
type DraftService struct {
	draftRepo  postgres.ArticleDraftRepository
	blockRepo  postgres.ArticleBlockRepository
	sourceRepo postgres.SourceDocumentRepository
	taskRepo   postgres.TaskRepository
	assetRepo  postgres.AssetRepository
	storage    adapter.StorageProvider
}

// NewDraftService creates a new DraftService. assetRepo and storage may be nil for
// tests / lightweight contexts; if nil, DraftVO.CoverURL will simply be omitted.
func NewDraftService(
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
	sourceRepo postgres.SourceDocumentRepository,
	taskRepo postgres.TaskRepository,
	assetRepo postgres.AssetRepository,
	storage adapter.StorageProvider,
) *DraftService {
	return &DraftService{
		draftRepo:  draftRepo,
		blockRepo:  blockRepo,
		sourceRepo: sourceRepo,
		taskRepo:   taskRepo,
		assetRepo:  assetRepo,
		storage:    storage,
	}
}

// GetByPublicID retrieves a draft with its blocks by public ID.
func (s *DraftService) GetByPublicID(ctx context.Context, publicID string) (*DraftVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetByPublicID: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetByPublicID: fetch blocks: %w", err)
	}

	// Find the task to get its public_id
	taskPublicID := ""
	if d.TaskID > 0 {
		task, taskErr := s.taskRepo.FindByID(ctx, d.TaskID)
		if taskErr == nil && task != nil {
			taskPublicID = task.PublicID
		}
	}

	return s.toVO(d, blocks, taskPublicID), nil
}

// GetByTaskPublicID retrieves the latest draft for a task by the task's public ID.
func (s *DraftService) GetByTaskPublicID(ctx context.Context, taskPublicID string) (*DraftVO, error) {
	task, err := s.taskRepo.FindByPublicID(ctx, taskPublicID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetByTaskPublicID: %w", err)
	}
	if task == nil {
		return nil, ErrNotFound
	}

	d, err := s.draftRepo.FindLatestByTaskID(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetByTaskPublicID: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetByTaskPublicID: fetch blocks: %w", err)
	}

	return s.toVO(d, blocks, taskPublicID), nil
}

// UpdateDraft updates title, subtitle, digest, or author_name.
func (s *DraftService) UpdateDraft(ctx context.Context, publicID string, title, subtitle, digest *string) (*DraftVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("draftService.UpdateDraft: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	if title != nil {
		d.Title = *title
	}
	if subtitle != nil {
		d.Subtitle = subtitle
	}
	if digest != nil {
		d.Digest = *digest
	}

	if err := s.draftRepo.Update(ctx, d); err != nil {
		return nil, fmt.Errorf("draftService.UpdateDraft: %w", err)
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("draftService.UpdateDraft: fetch blocks: %w", err)
	}

	return s.toVO(d, blocks, ""), nil
}

// UpdateBlock updates a single block's content.
func (s *DraftService) UpdateBlock(ctx context.Context, draftPublicID, blockPublicID string, heading, textMD, htmlFragment *string) (*BlockVO, error) {
	// Verify draft exists
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("draftService.UpdateBlock: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("draftService.UpdateBlock: %w", err)
	}

	var target *draft.ArticleBlock
	for i := range blocks {
		if blocks[i].PublicID == blockPublicID {
			target = &blocks[i]
			break
		}
	}
	if target == nil {
		return nil, ErrNotFound
	}

	if heading != nil {
		target.Heading = heading
	}
	if textMD != nil {
		target.TextMD = textMD
	}
	if htmlFragment != nil {
		target.HTMLFragment = htmlFragment
	}

	if err := s.blockRepo.Update(ctx, target); err != nil {
		return nil, fmt.Errorf("draftService.UpdateBlock: %w", err)
	}

	return s.blockToVO(target), nil
}

// GetTaskSources retrieves source documents for a task.
func (s *DraftService) GetTaskSources(ctx context.Context, taskPublicID string) ([]SourceVO, error) {
	task, err := s.taskRepo.FindByPublicID(ctx, taskPublicID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetTaskSources: %w", err)
	}
	if task == nil {
		return nil, ErrNotFound
	}

	docs, err := s.sourceRepo.FindByTaskID(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("draftService.GetTaskSources: %w", err)
	}

	vos := make([]SourceVO, 0, len(docs))
	for i := range docs {
		vo := SourceVO{
			ID:             docs[i].PublicID,
			Title:          docs[i].Title,
			SourceType:     docs[i].SourceType,
			SiteName:       docs[i].SiteName,
			SourceURL:      docs[i].SourceURL,
			Author:         docs[i].Author,
			HotScore:       docs[i].HotScore,
			RelevanceScore: docs[i].RelevanceScore,
		}
		if docs[i].PublishedAt != nil {
			t := docs[i].PublishedAt.Format("2006-01-02")
			vo.PublishedAt = &t
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *DraftService) toVO(d *draft.ArticleDraft, blocks []draft.ArticleBlock, taskPublicID string) *DraftVO {
	vo := &DraftVO{
		ID:              d.PublicID,
		TaskID:          taskPublicID,
		Title:           d.Title,
		Subtitle:        d.Subtitle,
		Digest:          d.Digest,
		AuthorName:      d.AuthorName,
		QualityScore:    d.QualityScore,
		SimilarityScore: d.SimilarityScore,
		RiskLevel:       d.RiskLevel,
		ReviewStatus:    d.ReviewStatus,
		Version:         d.Version,
		CreatedAt:       d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if url := s.resolveCoverURL(context.Background(), d); url != "" {
		vo.CoverURL = &url
	}

	vo.Blocks = make([]BlockVO, 0, len(blocks))
	for i := range blocks {
		vo.Blocks = append(vo.Blocks, *s.blockToVO(&blocks[i]))
	}

	return vo
}

// resolveCoverURL looks up the cover asset and returns its public storage URL, or
// "" if no cover is linked or any lookup fails (cover is non-critical metadata).
func (s *DraftService) resolveCoverURL(ctx context.Context, d *draft.ArticleDraft) string {
	if d == nil || d.CoverAssetID == nil || s.assetRepo == nil || s.storage == nil {
		return ""
	}
	a, err := s.assetRepo.FindByID(ctx, *d.CoverAssetID)
	if err != nil || a == nil {
		return ""
	}
	url, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey)
	if err != nil {
		return ""
	}
	return url
}

func (s *DraftService) blockToVO(b *draft.ArticleBlock) *BlockVO {
	return &BlockVO{
		ID:        b.PublicID,
		SortNo:    b.SortNo,
		BlockType: b.BlockType,
		Heading:   b.Heading,
		TextMD:    b.TextMD,
		HTMLFragment: b.HTMLFragment,
		PromptText:   b.PromptText,
		Status:       b.Status,
	}
}
