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
