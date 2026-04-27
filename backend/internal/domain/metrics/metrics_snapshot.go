// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package metrics

import (
	"time"

	"gorm.io/datatypes"

	"readbud/internal/domain"
)

// MetricsSnapshot represents the metrics_snapshots table per spec Section 11.11.
// This table is partitioned by month on metric_date in PostgreSQL.
type MetricsSnapshot struct {
	domain.BaseModel
	WechatAccountID int64          `gorm:"type:bigint;not null" json:"wechat_account_id"`
	ArticleID       string         `gorm:"type:varchar(128);not null;index:idx_metrics_article_date,priority:1" json:"article_id"`
	MetricDate      time.Time      `gorm:"type:date;not null;index:idx_metrics_article_date,priority:2" json:"metric_date"`
	ReadCount       *int           `gorm:"type:int" json:"read_count,omitempty"`
	ReadUserCount   *int           `gorm:"type:int" json:"read_user_count,omitempty"`
	ShareCount      *int           `gorm:"type:int" json:"share_count,omitempty"`
	ShareUserCount  *int           `gorm:"type:int" json:"share_user_count,omitempty"`
	AddFansCount    *int           `gorm:"type:int" json:"add_fans_count,omitempty"`
	CancelFansCount *int           `gorm:"type:int" json:"cancel_fans_count,omitempty"`
	NetFansCount    *int           `gorm:"type:int" json:"net_fans_count,omitempty"`
	RawJSON         datatypes.JSON `gorm:"type:jsonb" json:"raw_json,omitempty"`
}

// TableName overrides the default table name.
func (MetricsSnapshot) TableName() string {
	return "metrics_snapshots"
}
