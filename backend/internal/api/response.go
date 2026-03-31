package api

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

// CtxKeyRequestID is the context key used by the RequestID middleware.
// Duplicated here to avoid import cycle between api and middleware packages.
const ctxKeyRequestID = "request_id"

// Response is the standard API response envelope.
// Matches spec: { "code": 0, "message": "ok", "data": {}, "request_id": "" }
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id"`
}

// getRequestID returns the request ID from the Gin context (set by RequestID middleware).
// Falls back to generating a new ULID if the middleware hasn't been applied.
func getRequestID(c *gin.Context) string {
	if val, exists := c.Get(ctxKeyRequestID); exists {
		if id, ok := val.(string); ok && id != "" {
			return id
		}
	}
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// OK sends a successful response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      0,
		Message:   "ok",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 response with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:      0,
		Message:   "created",
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// NoContent sends a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response.
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

// ErrorWithDetails sends an error response with structured field-level details.
func ErrorWithDetails(c *gin.Context, httpStatus int, code int, message string, details interface{}) {
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: getRequestID(c),
	})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, 400, message)
}

// ValidationError sends a 422 error response with field-level validation details.
func ValidationError(c *gin.Context, message string, details []FieldError) {
	ErrorWithDetails(c, http.StatusUnprocessableEntity, 422, message, details)
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, 401, message)
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, 403, message)
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, 404, message)
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, 409, message)
}

// TooLarge sends a 413 error response.
func TooLarge(c *gin.Context, message string) {
	Error(c, http.StatusRequestEntityTooLarge, 413, message)
}

// TooManyRequests sends a 429 error response.
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, 429, message)
}

// InternalError sends a 500 error response.
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, 500, message)
}

// ServiceUnavailable sends a 503 error response.
func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, 503, message)
}
