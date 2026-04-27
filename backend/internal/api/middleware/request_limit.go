// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"readbud/internal/api"
)

const (
	// DefaultMaxBodySize is the default maximum request body size (2 MB).
	DefaultMaxBodySize = 2 << 20
	// LargeMaxBodySize is the limit for file upload endpoints (10 MB).
	LargeMaxBodySize = 10 << 20
)

// RequestSizeLimit limits the request body size to prevent abuse.
// maxBytes specifies the maximum allowed body size in bytes.
func RequestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			api.TooLarge(c, "请求体过大，请减少数据量")
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
