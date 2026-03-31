package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID_GeneratesULID(t *testing.T) {
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id := GetRequestID(c)
		if id == "" {
			t.Error("expected non-empty request ID")
		}
		c.String(http.StatusOK, id)
	})

	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	// Response should have X-Request-ID header
	headerID := w.Header().Get(RequestIDHeader)
	if headerID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	// ULID is 26 characters
	if len(headerID) != 26 {
		t.Errorf("expected ULID length 26, got %d (%s)", len(headerID), headerID)
	}
}

func TestRequestID_PropagatesClientID(t *testing.T) {
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, GetRequestID(c))
	})

	clientID := "my-custom-trace-id-123"
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(RequestIDHeader, clientID)
	r.ServeHTTP(w, c.Request)

	headerID := w.Header().Get(RequestIDHeader)
	if headerID != clientID {
		t.Errorf("expected propagated client ID %q, got %q", clientID, headerID)
	}
}

func TestGetRequestID_NoMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	id := GetRequestID(c)
	if id != "" {
		t.Errorf("expected empty ID without middleware, got %q", id)
	}
}
