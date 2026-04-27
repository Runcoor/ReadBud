// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/service"
)

// ExtensionAuth is a gin middleware that authenticates a request using either
// a regular JWT (Authorization: Bearer eyJ…) OR a long-lived extension token
// (Authorization: Bearer rbex_…). It is intended for routes consumed by both
// the web frontend AND the browser extension — namely the wechat-package
// endpoint that the extension uses to fetch article data.
//
// Resolution order:
//  1. If `c.Get(CtxKeyUserID)` already contains an int64 (set by an upstream
//     JWTAuth middleware), the request is already authenticated; pass through.
//  2. Otherwise, parse the bearer token. Tokens prefixed with `rbex_` are
//     resolved via ExtensionTokenService; everything else falls through with
//     401 (the JWT path should already have run by then).
func ExtensionAuth(svc *service.ExtensionTokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, already := GetUserID(c); already {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apiPkg.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			apiPkg.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}
		tok := strings.TrimSpace(parts[1])
		if !strings.HasPrefix(tok, "rbex_") {
			apiPkg.Unauthorized(c, "插件令牌无效")
			c.Abort()
			return
		}

		uid, err := svc.Authenticate(c.Request.Context(), tok)
		if err != nil {
			apiPkg.Unauthorized(c, "插件令牌无效或已过期")
			c.Abort()
			return
		}
		c.Set(CtxKeyUserID, uid)
		c.Next()
	}
}

// CombinedAuth tries the extension token first when the Authorization header
// is shaped like one (rbex_…), and falls back to the supplied JWT middleware
// otherwise. Install on routes that should accept BOTH webapp users and the
// browser extension — namely the wechat-package endpoint.
func CombinedAuth(extSvc *service.ExtensionTokenService, jwtMiddleware gin.HandlerFunc) gin.HandlerFunc {
	extMW := ExtensionAuth(extSvc)
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Authorization"), "rbex_") {
			extMW(c)
			return
		}
		jwtMiddleware(c)
	}
}
