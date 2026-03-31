package task

import (
	"time"

	"readbud/internal/domain"
)

// ContentTask represents the content_tasks table per spec Section 11.4.
type ContentTask struct {
	domain.BaseModel
	TaskNo          string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"task_no"`
	Keyword         string     `gorm:"type:varchar(255);not null" json:"keyword"`
	Audience        string     `gorm:"type:varchar(255)" json:"audience"`
	Tone            string     `gorm:"type:varchar(64)" json:"tone"`
	TargetWords     int        `gorm:"type:int;not null;default:2000" json:"target_words"`
	ImageMode       string     `gorm:"type:varchar(32);not null;default:'auto'" json:"image_mode"`
	ChartMode       int16      `gorm:"type:smallint;not null;default:1" json:"chart_mode"`
	PublishMode     string     `gorm:"type:varchar(32);not null;default:'manual'" json:"publish_mode"`
	PublishAt       *time.Time `gorm:"type:timestamptz" json:"publish_at,omitempty"`
	WechatAccountID *int64     `gorm:"type:bigint;index" json:"wechat_account_id,omitempty"`
	Status          string     `gorm:"type:varchar(32);not null;default:'pending';index:idx_task_status_created,priority:1" json:"status"`
	Progress        int        `gorm:"type:int;not null;default:0" json:"progress"`
	CurrentStage    string     `gorm:"type:varchar(64)" json:"current_stage"`
	ErrorMessage    *string    `gorm:"type:text" json:"error_message,omitempty"`
	ResultDraftID   *int64     `gorm:"type:bigint" json:"result_draft_id,omitempty"`
}

// TableName overrides the default table name.
func (ContentTask) TableName() string {
	return "content_tasks"
}

// Task status constants.
const (
	StatusPending     = "pending"
	StatusCollecting  = "collecting"
	StatusAnalyzing   = "analyzing"
	StatusWriting     = "writing"
	StatusAsseting    = "asseting"
	StatusReviewReady = "review_ready"
	StatusPublishing  = "publishing"
	StatusPublished   = "published"
	StatusFailed      = "failed"
)

// Task status: running is used during active pipeline execution.
const StatusRunning = "running"

// Task status: done marks a completed task.
const StatusDone = "done"

// Task status: cancelled marks a user-cancelled task.
const StatusCancelled = "cancelled"

// Task stage constants (pipeline stages).
const (
	StageKeywordExpand = "keyword_expand"
	StageSourceSearch  = "source_search"
	StageContentCrawl  = "content_crawl"
	StageHotScore      = "hot_score"
	StageArticleWrite  = "article_write"
	StageImageMatch    = "image_match"
	StageChartGen      = "chart_gen"
	StageHTMLCompile   = "html_compile"
	StageReview        = "review"
	StagePublish       = "publish"
)

// Image mode constants.
const (
	ImageModeAuto         = "auto"
	ImageModeSearchOnly   = "search_only"
	ImageModeGenerateOnly = "generate_only"
)

// Publish mode constants.
const (
	PublishModeManual   = "manual"
	PublishModeNow      = "now"
	PublishModeSchedule = "schedule"
)
