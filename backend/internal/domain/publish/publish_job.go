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
const (
	JobStatusQueued     = "queued"
	JobStatusSubmitting = "submitting"
	JobStatusPolling    = "polling"
	JobStatusSuccess    = "success"
	JobStatusFailed     = "failed"
	JobStatusCancelled  = "cancelled"
)
