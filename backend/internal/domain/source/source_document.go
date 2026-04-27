// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package source

import (
	"time"

	"gorm.io/datatypes"

	"readbud/internal/domain"
)

// SourceDocument represents the source_documents table per spec Section 11.5.
type SourceDocument struct {
	domain.BaseModel
	TaskID         int64          `gorm:"type:bigint;not null;index:idx_source_task_score,priority:1" json:"task_id"`
	SourceType     string         `gorm:"type:varchar(32);not null" json:"source_type"`
	SiteName       string         `gorm:"type:varchar(128)" json:"site_name"`
	SourceURL      string         `gorm:"type:text;not null" json:"source_url"`
	Title          string         `gorm:"type:varchar(512);not null" json:"title"`
	Author         *string        `gorm:"type:varchar(128)" json:"author,omitempty"`
	PublishedAt    *time.Time     `gorm:"type:timestamptz" json:"published_at,omitempty"`
	CrawledAt      time.Time      `gorm:"type:timestamptz;not null" json:"crawled_at"`
	RawHTML        string         `gorm:"type:text" json:"-"`
	PlainText      string         `gorm:"type:text" json:"-"`
	SummaryJSON    datatypes.JSON `gorm:"type:jsonb" json:"summary_json,omitempty"`
	HotScore       float64        `gorm:"type:decimal(10,2);not null;default:0;index:idx_source_task_score,priority:2,sort:desc" json:"hot_score"`
	RelevanceScore float64        `gorm:"type:decimal(10,2);not null;default:0" json:"relevance_score"`
	DataPointsJSON datatypes.JSON `gorm:"type:jsonb" json:"data_points_json,omitempty"`
	ImageCluesJSON datatypes.JSON `gorm:"type:jsonb" json:"image_clues_json,omitempty"`
	LicenseNote    *string        `gorm:"type:varchar(255)" json:"license_note,omitempty"`
}

// TableName overrides the default table name.
func (SourceDocument) TableName() string {
	return "source_documents"
}

// Source type constants.
const (
	SourceTypeWeb    = "web"
	SourceTypeNews   = "news"
	SourceTypeWechat = "wechat"
	SourceTypeBlog   = "blog"
)
