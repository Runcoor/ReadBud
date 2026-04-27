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
	"readbud/internal/api/middleware"
	"readbud/internal/service"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes registers auth routes on the given router group.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			apiPkg.Unauthorized(c, "用户名或密码错误")
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			apiPkg.Forbidden(c, "账号已被禁用")
			return
		}
		apiPkg.InternalError(c, "登录失败，请稍后重试")
		return
	}

	apiPkg.OK(c, resp)
}

// RefreshToken handles POST /api/v1/auth/refresh.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		apiPkg.Unauthorized(c, "无效的认证信息")
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), claims)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrUserInactive) {
			apiPkg.Unauthorized(c, "账号不可用")
			return
		}
		apiPkg.InternalError(c, "令牌刷新失败")
		return
	}

	apiPkg.OK(c, resp)
}
