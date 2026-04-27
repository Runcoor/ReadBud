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

// WechatHandler handles WeChat account HTTP endpoints.
type WechatHandler struct {
	wechatService *service.WechatAccountService
}

// NewWechatHandler creates a new WechatHandler.
func NewWechatHandler(svc *service.WechatAccountService) *WechatHandler {
	return &WechatHandler{wechatService: svc}
}

// RegisterRoutes registers WeChat account routes on the given router group.
func (h *WechatHandler) RegisterRoutes(rg *gin.RouterGroup) {
	wechat := rg.Group("/wechat/accounts")
	{
		wechat.GET("", h.List)
		wechat.POST("", h.Create)
		wechat.PUT("/:id", h.Update)
	}
}

// List handles GET /api/v1/wechat/accounts.
func (h *WechatHandler) List(c *gin.Context) {
	accounts, err := h.wechatService.List(c.Request.Context())
	if err != nil {
		apiPkg.InternalError(c, "获取公众号列表失败")
		return
	}
	apiPkg.OK(c, accounts)
}

// Create handles POST /api/v1/wechat/accounts.
func (h *WechatHandler) Create(c *gin.Context) {
	var req dto.WechatAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.wechatService.Create(c.Request.Context(), req)
	if err != nil {
		apiPkg.InternalError(c, "创建公众号配置失败")
		return
	}
	apiPkg.Created(c, vo)
}

// Update handles PUT /api/v1/wechat/accounts/:id.
func (h *WechatHandler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req dto.WechatAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.wechatService.Update(c.Request.Context(), publicID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "公众号不存在")
			return
		}
		apiPkg.InternalError(c, "更新公众号配置失败")
		return
	}
	apiPkg.OK(c, vo)
}
