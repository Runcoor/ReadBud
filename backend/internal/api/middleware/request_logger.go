// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"readbud/internal/pkg/logger"
)

// sensitiveHeaders lists headers whose values must not appear in logs.
var sensitiveHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"set-cookie":    true,
	"x-api-key":     true,
}

// sensitiveParams lists query/body params that must be masked in logs.
var sensitiveParams = map[string]bool{
	"password":   true,
	"secret":     true,
	"app_secret": true,
	"secret_json": true,
	"token":      true,
}

// RequestLogger logs every HTTP request with structured Zap fields.
// Fields: request_id, method, path, status, latency_ms, client_ip, user_id, user_agent, content_length.
// Sensitive headers are redacted. 4xx/5xx responses are logged at warn/error level.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := sanitizeQuery(c.Request.URL.RawQuery)

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID := GetRequestID(c)

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", rawQuery),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int64("content_length", c.Request.ContentLength),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Add user info if authenticated
		if userID, ok := GetUserID(c); ok {
			fields = append(fields, zap.Int64("user_id", userID))
		}
		if username, exists := c.Get(CtxKeyUsername); exists {
			if name, ok := username.(string); ok {
				fields = append(fields, zap.String("username", name))
			}
		}

		// Add error info if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		l := logger.L
		if l == nil {
			return
		}

		switch {
		case status >= 500:
			l.Error("request completed",
				fields...,
			)
		case status >= 400:
			l.Warn("request completed",
				fields...,
			)
		default:
			l.Info("request completed",
				fields...,
			)
		}
	}
}

// sanitizeQuery removes sensitive parameter values from the query string.
func sanitizeQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	parts := strings.Split(rawQuery, "&")
	sanitized := make([]string, 0, len(parts))
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		key := strings.ToLower(kv[0])
		if sensitiveParams[key] && len(kv) == 2 {
			sanitized = append(sanitized, kv[0]+"=[REDACTED]")
		} else {
			sanitized = append(sanitized, part)
		}
	}
	return strings.Join(sanitized, "&")
}

// SanitizeHeaders returns a map of headers with sensitive values redacted.
// Exported for use in audit trail logging.
func SanitizeHeaders(headers map[string][]string) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if sensitiveHeaders[strings.ToLower(key)] {
			result[key] = "[REDACTED]"
		} else if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}
