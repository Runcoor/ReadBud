// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

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
