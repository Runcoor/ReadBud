package dto

import "time"

// CreateTaskRequest is the DTO for POST /api/v1/tasks.
type CreateTaskRequest struct {
	Keyword         string     `json:"keyword" binding:"required,min=1,max=255"`
	Audience        string     `json:"audience" binding:"omitempty,max=255"`
	Tone            string     `json:"tone" binding:"omitempty,max=64"`
	TargetWords     int        `json:"target_words" binding:"omitempty,min=500,max=20000"`
	ImageMode       string     `json:"image_mode" binding:"required,oneof=auto search_only generate_only"`
	ChartMode       int16      `json:"chart_mode" binding:"omitempty,min=0,max=2"`
	PublishMode     string     `json:"publish_mode" binding:"required,oneof=manual now schedule"`
	PublishAt       *time.Time `json:"publish_at,omitempty"`
	WechatAccountID *string    `json:"wechat_account_id,omitempty" binding:"omitempty,min=1,max=64"`
	ArticleStyle    string     `json:"article_style" binding:"omitempty,oneof=minimal magazine listicle narrative faq casual"`
	VisualEnhance   *bool      `json:"visual_enhance"`
	BrandProfileID  *string    `json:"brand_profile_id,omitempty"`
}

// TaskVO is the view object for task display.
type TaskVO struct {
	ID              string     `json:"id"`
	TaskNo          string     `json:"task_no"`
	Keyword         string     `json:"keyword"`
	Audience        string     `json:"audience"`
	Tone            string     `json:"tone"`
	TargetWords     int        `json:"target_words"`
	ImageMode       string     `json:"image_mode"`
	ChartMode       int16      `json:"chart_mode"`
	PublishMode     string     `json:"publish_mode"`
	PublishAt       *time.Time `json:"publish_at,omitempty"`
	Status          string     `json:"status"`
	Progress        int        `json:"progress"`
	CurrentStage    string     `json:"current_stage"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	ResultDraftID   *string    `json:"result_draft_id,omitempty"`
	ArticleStyle    string     `json:"article_style"`
	VisualEnhance   bool       `json:"visual_enhance"`
	BrandProfileID  *string    `json:"brand_profile_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TaskListResponse wraps a paginated list of tasks.
type TaskListResponse struct {
	Items    []TaskVO `json:"items"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}
