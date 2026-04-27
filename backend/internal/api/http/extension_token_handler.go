// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/api/middleware"
	"readbud/internal/service"
)

// ExtensionTokenHandler exposes endpoints for managing browser-extension tokens.
type ExtensionTokenHandler struct {
	svc *service.ExtensionTokenService
}

// NewExtensionTokenHandler builds the handler.
func NewExtensionTokenHandler(svc *service.ExtensionTokenService) *ExtensionTokenHandler {
	return &ExtensionTokenHandler{svc: svc}
}

// RegisterRoutes wires the user-facing token-management routes.
func (h *ExtensionTokenHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/extension-tokens")
	{
		g.GET("", h.List)
		g.POST("", h.Issue)
		g.DELETE("/:id", h.Revoke)
	}
}

// List handles GET /api/v1/extension-tokens.
func (h *ExtensionTokenHandler) List(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		apiPkg.Unauthorized(c, "未登录")
		return
	}
	vos, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		apiPkg.InternalError(c, "获取插件令牌失败")
		return
	}
	apiPkg.OK(c, vos)
}

// Issue handles POST /api/v1/extension-tokens.
// The plaintext token is returned exactly once in the response body.
func (h *ExtensionTokenHandler) Issue(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		apiPkg.Unauthorized(c, "未登录")
		return
	}
	var req dto.ExtensionTokenIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// allow empty body — defaults are fine
		req = dto.ExtensionTokenIssueRequest{}
	}
	var ttl time.Duration
	if req.TTLHours > 0 {
		ttl = time.Duration(req.TTLHours) * time.Hour
	}
	plaintext, vo, err := h.svc.Issue(c.Request.Context(), uid, strings.TrimSpace(req.Name), ttl)
	if err != nil {
		apiPkg.InternalError(c, "签发插件令牌失败")
		return
	}
	apiPkg.OK(c, dto.ExtensionTokenIssueResponse{Token: plaintext, Info: *vo})
}

// Revoke handles DELETE /api/v1/extension-tokens/:id.
func (h *ExtensionTokenHandler) Revoke(c *gin.Context) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		apiPkg.Unauthorized(c, "未登录")
		return
	}
	publicID := c.Param("id")
	if err := h.svc.Revoke(c.Request.Context(), uid, publicID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "令牌不存在")
			return
		}
		apiPkg.InternalError(c, "撤销插件令牌失败")
		return
	}
	c.Status(http.StatusNoContent)
}
