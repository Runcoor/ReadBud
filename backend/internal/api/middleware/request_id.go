package middleware

import (
	"crypto/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

// CtxKeyRequestID is the context key for the unique request ID.
const CtxKeyRequestID = "request_id"

// RequestIDHeader is the HTTP header name for request ID propagation.
const RequestIDHeader = "X-Request-ID"

// RequestID generates a unique ULID for each request and injects it into
// the Gin context. If the client provides an X-Request-ID header, it is
// used as-is (for distributed tracing). The response always echoes the ID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
		}
		c.Set(CtxKeyRequestID, reqID)
		c.Header(RequestIDHeader, reqID)
		c.Next()
	}
}

// GetRequestID extracts the request ID from the Gin context.
func GetRequestID(c *gin.Context) string {
	val, exists := c.Get(CtxKeyRequestID)
	if !exists {
		return ""
	}
	id, ok := val.(string)
	if !ok {
		return ""
	}
	return id
}
