// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/service"
)

// DraftVersionHandler handles draft version and citation HTTP endpoints.
type DraftVersionHandler struct {
	versionService  *service.DraftVersionService
	citationService *service.CitationService
}

// NewDraftVersionHandler creates a new DraftVersionHandler.
func NewDraftVersionHandler(versionSvc *service.DraftVersionService, citationSvc *service.CitationService) *DraftVersionHandler {
	return &DraftVersionHandler{
		versionService:  versionSvc,
		citationService: citationSvc,
	}
}

// RegisterRoutes registers draft version and citation routes on the given router group.
func (h *DraftVersionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	drafts := rg.Group("/drafts")
	{
		drafts.GET("/:id/versions", h.ListVersions)
		drafts.GET("/:id/versions/:versionId", h.GetVersion)
		drafts.POST("/:id/versions/snapshot", h.CreateSnapshot)
		drafts.POST("/:id/versions/:versionId/rollback", h.Rollback)
		drafts.GET("/:id/citations", h.GetDraftCitations)
		drafts.GET("/:id/blocks/:blockId/citations", h.GetBlockCitations)
		drafts.POST("/:id/citations", h.AddCitation)
	}
}

// ListVersions handles GET /api/v1/drafts/:id/versions.
func (h *DraftVersionHandler) ListVersions(c *gin.Context) {
	draftID := c.Param("id")

	vos, err := h.versionService.ListVersions(c.Request.Context(), draftID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "获取版本列表失败")
		return
	}

	apiPkg.OK(c, vos)
}

// GetVersion handles GET /api/v1/drafts/:id/versions/:versionId.
func (h *DraftVersionHandler) GetVersion(c *gin.Context) {
	versionID := c.Param("versionId")

	vo, err := h.versionService.GetVersion(c.Request.Context(), versionID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "版本不存在")
			return
		}
		apiPkg.InternalError(c, "获取版本详情失败")
		return
	}

	apiPkg.OK(c, vo)
}

// createSnapshotRequest is the request body for creating a snapshot.
type createSnapshotRequest struct {
	ChangeReason string `json:"change_reason" binding:"required,max=255"`
	OperatorID   *int64 `json:"operator_id,omitempty"`
}

// CreateSnapshot handles POST /api/v1/drafts/:id/versions/snapshot.
func (h *DraftVersionHandler) CreateSnapshot(c *gin.Context) {
	draftPublicID := c.Param("id")

	var req createSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	// Resolve draft to get internal ID.
	err := h.versionService.CreateSnapshotByPublicID(c.Request.Context(), draftPublicID, req.OperatorID, req.ChangeReason)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "创建版本快照失败")
		return
	}

	apiPkg.OK(c, nil)
}

// rollbackRequest is the request body for rollback (currently empty, IDs come from URL).
type rollbackRequest struct {
	OperatorID *int64 `json:"operator_id,omitempty"`
}

// Rollback handles POST /api/v1/drafts/:id/versions/:versionId/rollback.
func (h *DraftVersionHandler) Rollback(c *gin.Context) {
	draftID := c.Param("id")
	versionID := c.Param("versionId")

	var req rollbackRequest
	// Body is optional; ignore bind errors for empty body.
	_ = c.ShouldBindJSON(&req)

	err := h.versionService.Rollback(c.Request.Context(), draftID, versionID, req.OperatorID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿或版本不存在")
			return
		}
		apiPkg.InternalError(c, "回滚版本失败")
		return
	}

	apiPkg.OK(c, nil)
}

// GetDraftCitations handles GET /api/v1/drafts/:id/citations.
func (h *DraftVersionHandler) GetDraftCitations(c *gin.Context) {
	draftID := c.Param("id")

	vos, err := h.citationService.GetDraftCitations(c.Request.Context(), draftID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "获取引用列表失败")
		return
	}

	apiPkg.OK(c, vos)
}

// GetBlockCitations handles GET /api/v1/drafts/:id/blocks/:blockId/citations.
func (h *DraftVersionHandler) GetBlockCitations(c *gin.Context) {
	draftID := c.Param("id")
	blockID := c.Param("blockId")

	vos, err := h.citationService.GetBlockCitations(c.Request.Context(), draftID, blockID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿或内容块不存在")
			return
		}
		apiPkg.InternalError(c, "获取内容块引用失败")
		return
	}

	apiPkg.OK(c, vos)
}

// AddCitation handles POST /api/v1/drafts/:id/citations.
func (h *DraftVersionHandler) AddCitation(c *gin.Context) {
	draftID := c.Param("id")

	var req service.AddCitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	err := h.citationService.AddCitation(c.Request.Context(), draftID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿、内容块或来源不存在")
			return
		}
		apiPkg.InternalError(c, "添加引用失败")
		return
	}

	apiPkg.OK(c, nil)
}
