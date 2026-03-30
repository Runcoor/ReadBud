package draft

import (
	"gorm.io/datatypes"

	"readbud/internal/domain"
)

// ArticleBlock represents the article_blocks table per spec Section 11.7.
type ArticleBlock struct {
	domain.BaseModel
	DraftID       int64          `gorm:"type:bigint;not null;index:idx_block_draft_sort,priority:1" json:"draft_id"`
	SortNo        int            `gorm:"type:int;not null;index:idx_block_draft_sort,priority:2" json:"sort_no"`
	BlockType     string         `gorm:"type:varchar(32);not null" json:"block_type"`
	Heading       *string        `gorm:"type:varchar(255)" json:"heading,omitempty"`
	TextMD        *string        `gorm:"type:text" json:"text_md,omitempty"`
	HTMLFragment  *string        `gorm:"type:text" json:"html_fragment,omitempty"`
	AssetID       *int64         `gorm:"type:bigint" json:"asset_id,omitempty"`
	SourceRefsJSON datatypes.JSON `gorm:"type:jsonb" json:"source_refs_json,omitempty"`
	PromptText    *string        `gorm:"type:text" json:"prompt_text,omitempty"`
	Status        string         `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
}

// TableName overrides the default table name.
func (ArticleBlock) TableName() string {
	return "article_blocks"
}

// Block type constants.
const (
	BlockTypeLead      = "lead"
	BlockTypeSection   = "section"
	BlockTypeImage     = "image"
	BlockTypeChart     = "chart"
	BlockTypeQuote     = "quote"
	BlockTypeChecklist = "checklist"
	BlockTypeCTA       = "cta"
	BlockTypeTitle     = "title"
	BlockTypeSubtitle  = "subtitle"
	BlockTypeSummary   = "summary"
)
