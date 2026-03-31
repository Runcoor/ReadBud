package service

import (
	"context"
	"encoding/json"
	"fmt"

	"readbud/internal/domain"
	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
)

// DraftVersionVO is the view object for a draft version list item.
type DraftVersionVO struct {
	ID           string `json:"id"`
	VersionNo    int    `json:"version_no"`
	Title        string `json:"title"`
	Digest       string `json:"digest"`
	OperatorID   *int64 `json:"operator_id,omitempty"`
	ChangeReason string `json:"change_reason"`
	CreatedAt    string `json:"created_at"`
}

// DraftVersionDetailVO is the view object for a full draft version with blocks.
type DraftVersionDetailVO struct {
	DraftVersionVO
	Blocks []BlockVO `json:"blocks"`
}

// DraftVersionService handles draft version operations.
type DraftVersionService struct {
	versionRepo  postgres.DraftVersionRepository
	draftRepo    postgres.ArticleDraftRepository
	blockRepo    postgres.ArticleBlockRepository
	citationRepo postgres.ContentCitationRepository
	maxVersions  int
}

// NewDraftVersionService creates a new DraftVersionService.
func NewDraftVersionService(
	versionRepo postgres.DraftVersionRepository,
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
	citationRepo postgres.ContentCitationRepository,
) *DraftVersionService {
	return &DraftVersionService{
		versionRepo:  versionRepo,
		draftRepo:    draftRepo,
		blockRepo:    blockRepo,
		citationRepo: citationRepo,
		maxVersions:  20,
	}
}

// CreateSnapshot snapshots the current draft state as a new version.
func (s *DraftVersionService) CreateSnapshot(ctx context.Context, draftID int64, operatorID *int64, changeReason string) error {
	d, err := s.draftRepo.FindByID(ctx, draftID)
	if err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshot: %w", err)
	}
	if d == nil {
		return ErrNotFound
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshot: fetch blocks: %w", err)
	}

	blocksJSON, err := json.Marshal(blocks)
	if err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshot: marshal blocks: %w", err)
	}

	// Determine the next version number.
	latest, err := s.versionRepo.GetLatestVersion(ctx, d.ID)
	if err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshot: get latest: %w", err)
	}
	nextNo := 1
	if latest != nil {
		nextNo = latest.VersionNo + 1
	}

	v := &domain.DraftVersion{
		DraftID:      d.ID,
		VersionNo:    nextNo,
		Title:        d.Title,
		Digest:       d.Digest,
		BlocksJSON:   blocksJSON,
		OperatorID:   operatorID,
		ChangeReason: changeReason,
	}

	if err := s.versionRepo.Create(ctx, v); err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshot: %w", err)
	}

	// Prune old versions if exceeding max.
	count, err := s.versionRepo.CountByDraftID(ctx, d.ID)
	if err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshot: count: %w", err)
	}
	if count > int64(s.maxVersions) {
		if err := s.versionRepo.DeleteOldVersions(ctx, d.ID, s.maxVersions); err != nil {
			return fmt.Errorf("draftVersionService.CreateSnapshot: prune: %w", err)
		}
	}

	return nil
}

// ListVersions lists version history for a draft by its public ID.
func (s *DraftVersionService) ListVersions(ctx context.Context, draftPublicID string) ([]DraftVersionVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("draftVersionService.ListVersions: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	versions, err := s.versionRepo.ListByDraftID(ctx, d.ID, 0)
	if err != nil {
		return nil, fmt.Errorf("draftVersionService.ListVersions: %w", err)
	}

	vos := make([]DraftVersionVO, 0, len(versions))
	for i := range versions {
		vos = append(vos, s.toVersionVO(&versions[i]))
	}
	return vos, nil
}

// GetVersion gets a full version detail including blocks.
func (s *DraftVersionService) GetVersion(ctx context.Context, versionPublicID string) (*DraftVersionDetailVO, error) {
	v, err := s.versionRepo.FindByPublicID(ctx, versionPublicID)
	if err != nil {
		return nil, fmt.Errorf("draftVersionService.GetVersion: %w", err)
	}
	if v == nil {
		return nil, ErrNotFound
	}

	var blocks []draft.ArticleBlock
	if len(v.BlocksJSON) > 0 {
		if err := json.Unmarshal(v.BlocksJSON, &blocks); err != nil {
			return nil, fmt.Errorf("draftVersionService.GetVersion: unmarshal blocks: %w", err)
		}
	}

	blockVOs := make([]BlockVO, 0, len(blocks))
	for i := range blocks {
		blockVOs = append(blockVOs, BlockVO{
			ID:           blocks[i].PublicID,
			SortNo:       blocks[i].SortNo,
			BlockType:    blocks[i].BlockType,
			Heading:      blocks[i].Heading,
			TextMD:       blocks[i].TextMD,
			HTMLFragment: blocks[i].HTMLFragment,
			PromptText:   blocks[i].PromptText,
			Status:       blocks[i].Status,
		})
	}

	return &DraftVersionDetailVO{
		DraftVersionVO: s.toVersionVO(v),
		Blocks:         blockVOs,
	}, nil
}

// Rollback restores a draft to a previous version's state.
func (s *DraftVersionService) Rollback(ctx context.Context, draftPublicID, versionPublicID string, operatorID *int64) error {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return fmt.Errorf("draftVersionService.Rollback: %w", err)
	}
	if d == nil {
		return ErrNotFound
	}

	v, err := s.versionRepo.FindByPublicID(ctx, versionPublicID)
	if err != nil {
		return fmt.Errorf("draftVersionService.Rollback: find version: %w", err)
	}
	if v == nil {
		return ErrNotFound
	}
	if v.DraftID != d.ID {
		return ErrNotFound
	}

	// Snapshot current state before rollback.
	if err := s.CreateSnapshot(ctx, d.ID, operatorID, fmt.Sprintf("auto-snapshot before rollback to v%d", v.VersionNo)); err != nil {
		return fmt.Errorf("draftVersionService.Rollback: pre-snapshot: %w", err)
	}

	// Restore draft metadata.
	d.Title = v.Title
	d.Digest = v.Digest
	if err := s.draftRepo.Update(ctx, d); err != nil {
		return fmt.Errorf("draftVersionService.Rollback: update draft: %w", err)
	}

	// Replace blocks.
	if err := s.blockRepo.DeleteByDraftID(ctx, d.ID); err != nil {
		return fmt.Errorf("draftVersionService.Rollback: delete blocks: %w", err)
	}

	var blocks []draft.ArticleBlock
	if len(v.BlocksJSON) > 0 {
		if err := json.Unmarshal(v.BlocksJSON, &blocks); err != nil {
			return fmt.Errorf("draftVersionService.Rollback: unmarshal blocks: %w", err)
		}
	}

	if len(blocks) > 0 {
		// Reset IDs so GORM creates new records.
		for i := range blocks {
			blocks[i].ID = 0
			blocks[i].PublicID = ""
			blocks[i].DraftID = d.ID
		}
		if err := s.blockRepo.CreateBatch(ctx, blocks); err != nil {
			return fmt.Errorf("draftVersionService.Rollback: create blocks: %w", err)
		}
	}

	return nil
}

// CreateSnapshotByPublicID resolves a draft by public ID and creates a snapshot.
func (s *DraftVersionService) CreateSnapshotByPublicID(ctx context.Context, draftPublicID string, operatorID *int64, changeReason string) error {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return fmt.Errorf("draftVersionService.CreateSnapshotByPublicID: %w", err)
	}
	if d == nil {
		return ErrNotFound
	}
	return s.CreateSnapshot(ctx, d.ID, operatorID, changeReason)
}

func (s *DraftVersionService) toVersionVO(v *domain.DraftVersion) DraftVersionVO {
	return DraftVersionVO{
		ID:           v.PublicID,
		VersionNo:    v.VersionNo,
		Title:        v.Title,
		Digest:       v.Digest,
		OperatorID:   v.OperatorID,
		ChangeReason: v.ChangeReason,
		CreatedAt:    v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
