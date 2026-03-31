package dto

// UpdateDraftRequest represents the PATCH /api/v1/drafts/:id request body.
type UpdateDraftRequest struct {
	Title    *string `json:"title,omitempty"`
	Subtitle *string `json:"subtitle,omitempty"`
	Digest   *string `json:"digest,omitempty"`
}

// UpdateBlockRequest represents the PATCH /api/v1/drafts/:id/blocks/:blockId request body.
type UpdateBlockRequest struct {
	Heading      *string `json:"heading,omitempty"`
	TextMD       *string `json:"text_md,omitempty"`
	HTMLFragment *string `json:"html_fragment,omitempty"`
}

// CreatePublishJobRequest represents the POST /api/v1/publish/jobs request body.
type CreatePublishJobRequest struct {
	DraftID         string  `json:"draft_id" binding:"required"`
	WechatAccountID string  `json:"wechat_account_id" binding:"required"`
	PublishMode     string  `json:"publish_mode" binding:"required,oneof=now schedule manual"`
	ScheduleAt      *string `json:"schedule_at,omitempty"`
}
