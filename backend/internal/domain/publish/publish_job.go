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

// PublishJob represents the publish_jobs table per spec Section 11.9.
type PublishJob struct {
	domain.BaseModel
	DraftID              int64          `gorm:"type:bigint;not null" json:"draft_id"`
	WechatAccountID      int64          `gorm:"type:bigint;not null" json:"wechat_account_id"`
	PublishMode          string         `gorm:"type:varchar(32);not null" json:"publish_mode"`
	ScheduleAt           *time.Time     `gorm:"type:timestamptz" json:"schedule_at,omitempty"`
	Status               string         `gorm:"type:varchar(32);not null;default:'queued';index:idx_job_status_schedule,priority:1" json:"status"`
	RetryCount           int            `gorm:"type:int;not null;default:0" json:"retry_count"`
	LastError            *string        `gorm:"type:text" json:"last_error,omitempty"`
	ProviderRequestJSON  datatypes.JSON `gorm:"type:jsonb" json:"-"`
	ProviderResponseJSON datatypes.JSON `gorm:"type:jsonb" json:"-"`
}

// TableName overrides the default table name.
func (PublishJob) TableName() string {
	return "publish_jobs"
}

// Publish job status constants.
//
// Status flow by delivery mode:
//
//	api:       queued -> submitting -> polling -> success | failed | cancelled
//	extension: awaiting_extension -> success (when extension reports back) | cancelled
//	manual:    awaiting_manual    -> success (user clicks "已发布")        | cancelled
const (
	JobStatusQueued            = "queued"
	JobStatusSubmitting        = "submitting"
	JobStatusPolling           = "polling"
	JobStatusSuccess           = "success"
	JobStatusFailed            = "failed"
	JobStatusCancelled         = "cancelled"
	JobStatusAwaitingExtension = "awaiting_extension"
	JobStatusAwaitingManual    = "awaiting_manual"
)
