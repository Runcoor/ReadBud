// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package dto

// UpdateDraftRequest represents the PATCH /api/v1/drafts/:id request body.
type UpdateDraftRequest struct {
	Title    *string `json:"title,omitempty" binding:"omitempty,max=200"`
	Subtitle *string `json:"subtitle,omitempty" binding:"omitempty,max=200"`
	Digest   *string `json:"digest,omitempty" binding:"omitempty,max=500"`
}

// UpdateBlockRequest represents the PATCH /api/v1/drafts/:id/blocks/:blockId request body.
type UpdateBlockRequest struct {
	Heading      *string `json:"heading,omitempty" binding:"omitempty,max=200"`
	TextMD       *string `json:"text_md,omitempty" binding:"omitempty,max=50000"`
	HTMLFragment *string `json:"html_fragment,omitempty" binding:"omitempty,max=100000"`
}

// CreatePublishJobRequest represents the POST /api/v1/publish/jobs request body.
type CreatePublishJobRequest struct {
	DraftID         string  `json:"draft_id" binding:"required,min=1,max=64"`
	WechatAccountID string  `json:"wechat_account_id" binding:"required,min=1,max=64"`
	PublishMode     string  `json:"publish_mode" binding:"required,oneof=now schedule manual"`
	ScheduleAt      *string `json:"schedule_at,omitempty"`
}
