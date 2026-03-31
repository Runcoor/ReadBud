package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// testRequest is a sample DTO for validation tests.
type testRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleBindError_ValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"","password":""}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req testRequest
	err := c.ShouldBindJSON(&req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	handled := HandleBindError(c, err)
	if !handled {
		t.Fatal("expected HandleBindError to return true")
	}

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	var resp Response
	if decErr := json.Unmarshal(w.Body.Bytes(), &resp); decErr != nil {
		t.Fatalf("failed to decode response: %v", decErr)
	}

	if resp.Code != 422 {
		t.Errorf("expected code 422, got %d", resp.Code)
	}

	if resp.Message != "输入参数验证失败" {
		t.Errorf("expected Chinese validation message, got %q", resp.Message)
	}

	if resp.Details == nil {
		t.Fatal("expected non-nil details")
	}

	if resp.RequestID == "" {
		t.Error("expected non-empty request_id")
	}
}

func TestHandleBindError_MalformedJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid json`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req testRequest
	err := c.ShouldBindJSON(&req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	handled := HandleBindError(c, err)
	if !handled {
		t.Fatal("expected HandleBindError to return true")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandleBindError_NilError(t *testing.T) {
	handled := HandleBindError(nil, nil)
	if handled {
		t.Error("expected HandleBindError to return false for nil error")
	}
}

func TestTranslateFieldError_ChineseLabels(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req testRequest
	err := c.ShouldBindJSON(&req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleBindError(c, err)

	body := w.Body.String()
	// Should contain Chinese field names
	if !strings.Contains(body, "用户名") && !strings.Contains(body, "密码") {
		t.Errorf("expected Chinese field labels in response, got: %s", body)
	}
}
