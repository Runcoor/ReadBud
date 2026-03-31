package service

import (
	"context"
	"fmt"

	"readbud/internal/domain"
	"readbud/internal/repository/postgres"
)

// CitationVO is the view object for a content citation.
type CitationVO struct {
	ID               string `json:"id"`
	BlockID          string `json:"block_id"`
	SourceDocumentID string `json:"source_document_id"`
	CitationType     string `json:"citation_type"`
	CitationText     string `json:"citation_text"`
	SourceLink       string `json:"source_link"`
	SourceNote       string `json:"source_note"`
}

// AddCitationRequest is the request to add a citation.
type AddCitationRequest struct {
	BlockPublicID          string `json:"block_id" binding:"required"`
	SourceDocumentPublicID string `json:"source_document_id" binding:"required"`
	CitationType           string `json:"citation_type" binding:"required"`
	CitationText           string `json:"citation_text"`
	SourceLink             string `json:"source_link"`
	SourceNote             string `json:"source_note"`
}

// CitationService handles content citation operations.
type CitationService struct {
	citationRepo postgres.ContentCitationRepository
	draftRepo    postgres.ArticleDraftRepository
	blockRepo    postgres.ArticleBlockRepository
	sourceRepo   postgres.SourceDocumentRepository
}

// NewCitationService creates a new CitationService.
func NewCitationService(
	citationRepo postgres.ContentCitationRepository,
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
	sourceRepo postgres.SourceDocumentRepository,
) *CitationService {
	return &CitationService{
		citationRepo: citationRepo,
		draftRepo:    draftRepo,
		blockRepo:    blockRepo,
		sourceRepo:   sourceRepo,
	}
}

// GetDraftCitations retrieves all citations for a draft.
func (s *CitationService) GetDraftCitations(ctx context.Context, draftPublicID string) ([]CitationVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("citationService.GetDraftCitations: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	citations, err := s.citationRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("citationService.GetDraftCitations: %w", err)
	}

	// Build lookup maps for blocks and source docs.
	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("citationService.GetDraftCitations: fetch blocks: %w", err)
	}
	blockMap := make(map[int64]string, len(blocks))
	for i := range blocks {
		blockMap[blocks[i].ID] = blocks[i].PublicID
	}

	return s.toCitationVOs(citations, blockMap), nil
}

// GetBlockCitations retrieves citations for a specific block.
func (s *CitationService) GetBlockCitations(ctx context.Context, draftPublicID, blockPublicID string) ([]CitationVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("citationService.GetBlockCitations: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("citationService.GetBlockCitations: fetch blocks: %w", err)
	}

	var blockID int64
	blockMap := make(map[int64]string, len(blocks))
	for i := range blocks {
		blockMap[blocks[i].ID] = blocks[i].PublicID
		if blocks[i].PublicID == blockPublicID {
			blockID = blocks[i].ID
		}
	}
	if blockID == 0 {
		return nil, ErrNotFound
	}

	citations, err := s.citationRepo.FindByBlockID(ctx, blockID)
	if err != nil {
		return nil, fmt.Errorf("citationService.GetBlockCitations: %w", err)
	}

	return s.toCitationVOs(citations, blockMap), nil
}

// AddCitation adds a citation to a draft block.
func (s *CitationService) AddCitation(ctx context.Context, draftPublicID string, req AddCitationRequest) error {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return fmt.Errorf("citationService.AddCitation: %w", err)
	}
	if d == nil {
		return ErrNotFound
	}

	// Resolve block internal ID.
	blocks, err := s.blockRepo.FindByDraftID(ctx, d.ID)
	if err != nil {
		return fmt.Errorf("citationService.AddCitation: fetch blocks: %w", err)
	}
	var blockID int64
	for i := range blocks {
		if blocks[i].PublicID == req.BlockPublicID {
			blockID = blocks[i].ID
			break
		}
	}
	if blockID == 0 {
		return ErrNotFound
	}

	// Resolve source document internal ID by matching public ID within the draft's task sources.
	sources, err := s.sourceRepo.FindByTaskID(ctx, d.TaskID)
	if err != nil {
		return fmt.Errorf("citationService.AddCitation: fetch sources: %w", err)
	}
	var sourceDocID int64
	for i := range sources {
		if sources[i].PublicID == req.SourceDocumentPublicID {
			sourceDocID = sources[i].ID
			break
		}
	}
	if sourceDocID == 0 {
		return ErrNotFound
	}

	c := &domain.ContentCitation{
		DraftID:          d.ID,
		BlockID:          blockID,
		SourceDocumentID: sourceDocID,
		CitationType:     req.CitationType,
		CitationText:     req.CitationText,
		SourceLink:       req.SourceLink,
		SourceNote:       req.SourceNote,
	}

	if err := s.citationRepo.Create(ctx, c); err != nil {
		return fmt.Errorf("citationService.AddCitation: %w", err)
	}

	return nil
}

func (s *CitationService) toCitationVOs(citations []domain.ContentCitation, blockMap map[int64]string) []CitationVO {
	vos := make([]CitationVO, 0, len(citations))
	for i := range citations {
		vos = append(vos, CitationVO{
			ID:               citations[i].PublicID,
			BlockID:          blockMap[citations[i].BlockID],
			SourceDocumentID: fmt.Sprintf("%d", citations[i].SourceDocumentID),
			CitationType:     citations[i].CitationType,
			CitationText:     citations[i].CitationText,
			SourceLink:       citations[i].SourceLink,
			SourceNote:       citations[i].SourceNote,
		})
	}
	return vos
}
