// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package publish

import (
	"time"

	"gorm.io/datatypes"

	"readbud/internal/domain"
)

// PublishRecord represents the publish_records table per spec Section 11.10.
type PublishRecord struct {
	domain.BaseModel
	PublishJobID    int64          `gorm:"type:bigint;not null" json:"publish_job_id"`
	DraftID         int64          `gorm:"type:bigint;not null" json:"draft_id"`
	WechatAccountID int64          `gorm:"type:bigint;not null" json:"wechat_account_id"`
	DraftMediaID    *string        `gorm:"type:varchar(128)" json:"draft_media_id,omitempty"`
	PublishID       *string        `gorm:"type:varchar(128)" json:"publish_id,omitempty"`
	ArticleID       *string        `gorm:"type:varchar(128)" json:"article_id,omitempty"`
	ArticleURL      *string        `gorm:"type:text" json:"article_url,omitempty"`
	PublishedAt     *time.Time     `gorm:"type:timestamptz" json:"published_at,omitempty"`
	PublishStatus   string         `gorm:"type:varchar(32);not null" json:"publish_status"`
	ExtraJSON       datatypes.JSON `gorm:"type:jsonb" json:"extra_json,omitempty"`
}

// TableName overrides the default table name.
func (PublishRecord) TableName() string {
	return "publish_records"
}
