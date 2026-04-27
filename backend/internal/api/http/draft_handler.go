// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/service"
)

// DraftHandler handles draft and source HTTP endpoints.
type DraftHandler struct {
	draftService   *service.DraftService
	coverService   *service.CoverImageService
	packageService *service.WechatPackageService
}

// NewDraftHandler creates a new DraftHandler. coverService and packageService may be
// nil — the corresponding endpoints will then return a 500 with an explanatory message.
func NewDraftHandler(
	svc *service.DraftService,
	coverSvc *service.CoverImageService,
	pkgSvc *service.WechatPackageService,
) *DraftHandler {
	return &DraftHandler{
		draftService:   svc,
		coverService:   coverSvc,
		packageService: pkgSvc,
	}
}

// RegisterRoutes registers draft routes on the given router group.
func (h *DraftHandler) RegisterRoutes(rg *gin.RouterGroup) {
	drafts := rg.Group("/drafts")
	{
		drafts.GET("/:id", h.Get)
		drafts.PATCH("/:id", h.Update)
		drafts.PATCH("/:id/blocks/:blockId", h.UpdateBlock)
		drafts.GET("/:id/cover", h.GetCover)
		drafts.POST("/:id/cover/regenerate", h.RegenerateCover)
	}
}

// RegisterPackageRoute mounts the wechat-package endpoint on a router group that
// uses the combined-auth middleware (accepts both webapp JWT and extension token).
// Kept separate from RegisterRoutes so the auth wiring stays explicit at the
// composition root.
func (h *DraftHandler) RegisterPackageRoute(rg *gin.RouterGroup) {
	rg.GET("/drafts/:id/wechat-package", h.GetWechatPackage)
}

// GetWechatPackage handles GET /api/v1/drafts/:id/wechat-package.
// Returns the bundle the browser extension uses to auto-fill the WeChat editor.
func (h *DraftHandler) GetWechatPackage(c *gin.Context) {
	if h.packageService == nil {
		apiPkg.InternalError(c, "插件分发服务未启用")
		return
	}
	publicID := c.Param("id")
	pkg, err := h.packageService.GetForDraft(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, err.Error())
		return
	}
	apiPkg.OK(c, pkg)
}

// GetCover handles GET /api/v1/drafts/:id/cover.
func (h *DraftHandler) GetCover(c *gin.Context) {
	if h.coverService == nil {
		apiPkg.InternalError(c, "封面服务未启用")
		return
	}
	publicID := c.Param("id")
	vo, err := h.coverService.GetForDraft(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "获取封面失败")
		return
	}
	if vo == nil {
		apiPkg.OK(c, nil)
		return
	}
	apiPkg.OK(c, vo)
}

// RegenerateCover handles POST /api/v1/drafts/:id/cover/regenerate.
func (h *DraftHandler) RegenerateCover(c *gin.Context) {
	if h.coverService == nil {
		apiPkg.InternalError(c, "封面服务未启用")
		return
	}
	publicID := c.Param("id")
	vo, err := h.coverService.GenerateForDraft(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "封面生成失败: "+err.Error())
		return
	}
	apiPkg.OK(c, vo)
}

// Get handles GET /api/v1/drafts/:id.
func (h *DraftHandler) Get(c *gin.Context) {
	publicID := c.Param("id")

	vo, err := h.draftService.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "获取草稿详情失败")
		return
	}

	apiPkg.OK(c, vo)
}

// Update handles PATCH /api/v1/drafts/:id.
func (h *DraftHandler) Update(c *gin.Context) {
	publicID := c.Param("id")

	var req dto.UpdateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.draftService.UpdateDraft(c.Request.Context(), publicID, req.Title, req.Subtitle, req.Digest)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "草稿不存在")
			return
		}
		apiPkg.InternalError(c, "更新草稿失败")
		return
	}

	apiPkg.OK(c, vo)
}

// UpdateBlock handles PATCH /api/v1/drafts/:id/blocks/:blockId.
func (h *DraftHandler) UpdateBlock(c *gin.Context) {
	draftID := c.Param("id")
	blockID := c.Param("blockId")

	var req dto.UpdateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.draftService.UpdateBlock(c.Request.Context(), draftID, blockID, req.Heading, req.TextMD, req.HTMLFragment)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "内容块不存在")
			return
		}
		apiPkg.InternalError(c, "更新内容块失败")
		return
	}

	apiPkg.OK(c, vo)
}

// SourceHandler handles source-related endpoints on tasks.
type SourceHandler struct {
	draftService *service.DraftService
}

// NewSourceHandler creates a new SourceHandler.
func NewSourceHandler(svc *service.DraftService) *SourceHandler {
	return &SourceHandler{draftService: svc}
}

// GetTaskSources handles GET /api/v1/tasks/:id/sources.
func (h *SourceHandler) GetTaskSources(c *gin.Context) {
	taskID := c.Param("id")

	sources, err := h.draftService.GetTaskSources(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "任务不存在")
			return
		}
		apiPkg.InternalError(c, "获取来源文章失败")
		return
	}

	apiPkg.OK(c, sources)
}
