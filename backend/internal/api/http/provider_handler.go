package http

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	apiPkg "readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/integration"
	"readbud/internal/service"
)

// ProviderHandler handles provider config HTTP endpoints.
type ProviderHandler struct {
	providerService *service.ProviderConfigService
	factory         *integration.ProviderFactory
	logger          *zap.Logger
}

// NewProviderHandler creates a new ProviderHandler.
func NewProviderHandler(svc *service.ProviderConfigService, factory *integration.ProviderFactory, logger *zap.Logger) *ProviderHandler {
	return &ProviderHandler{
		providerService: svc,
		factory:         factory,
		logger:          logger,
	}
}

// RegisterRoutes registers provider routes on the given router group.
func (h *ProviderHandler) RegisterRoutes(rg *gin.RouterGroup) {
	providers := rg.Group("/providers")
	{
		providers.GET("", h.List)
		providers.POST("", h.Create)
		providers.PUT("/:id", h.Update)
		providers.PATCH("/:id", h.Update)
		providers.DELETE("/:id", h.Delete)
		providers.POST("/:id/test", h.TestConnection)
		providers.POST("/:id/set-default", h.SetDefault)
	}
}

// List handles GET /api/v1/providers.
func (h *ProviderHandler) List(c *gin.Context) {
	configs, err := h.providerService.List(c.Request.Context())
	if err != nil {
		apiPkg.InternalError(c, "获取配置列表失败")
		return
	}
	apiPkg.OK(c, configs)
}

// Create handles POST /api/v1/providers.
func (h *ProviderHandler) Create(c *gin.Context) {
	var req dto.ProviderConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.providerService.Create(c.Request.Context(), req)
	if err != nil {
		apiPkg.InternalError(c, "创建配置失败")
		return
	}
	apiPkg.Created(c, vo)
}

// Update handles PUT /api/v1/providers/:id.
func (h *ProviderHandler) Update(c *gin.Context) {
	publicID := c.Param("id")
	var req dto.ProviderConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.providerService.Update(c.Request.Context(), publicID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "配置不存在")
			return
		}
		apiPkg.InternalError(c, "更新配置失败")
		return
	}
	apiPkg.OK(c, vo)
}

// Delete handles DELETE /api/v1/providers/:id.
func (h *ProviderHandler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	err := h.providerService.Delete(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "配置不存在")
			return
		}
		apiPkg.InternalError(c, "删除配置失败")
		return
	}
	apiPkg.OK(c, nil)
}

// SetDefault handles POST /api/v1/providers/:id/set-default.
func (h *ProviderHandler) SetDefault(c *gin.Context) {
	publicID := c.Param("id")
	vo, err := h.providerService.SetDefault(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "配置不存在")
			return
		}
		apiPkg.InternalError(c, "设置默认失败")
		return
	}
	apiPkg.OK(c, vo)
}

// TestConnection handles POST /api/v1/providers/:id/test.
// It loads the provider config, decrypts the secret, and makes a minimal test API call.
func (h *ProviderHandler) TestConnection(c *gin.Context) {
	publicID := c.Param("id")
	ctx := c.Request.Context()

	cfg, err := h.providerService.FindByPublicID(ctx, publicID)
	if err != nil {
		apiPkg.InternalError(c, "查找配置失败")
		return
	}
	if cfg == nil {
		apiPkg.NotFound(c, "配置不存在")
		return
	}

	if h.factory == nil {
		apiPkg.InternalError(c, "provider factory not available")
		return
	}

	// Parse config to determine format
	var configData struct {
		Format    string `json:"format"`
		APIFormat string `json:"api_format"`
		BaseURL   string `json:"base_url"`
		Model     string `json:"model"`
	}
	if err := json.Unmarshal(cfg.ConfigJSON, &configData); err != nil {
		apiPkg.InternalError(c, "解析配置失败")
		return
	}
	// Use format or api_format (legacy)
	effectiveFormat := configData.Format
	if effectiveFormat == "" {
		effectiveFormat = configData.APIFormat
	}

	// Decrypt secret
	secret, err := h.providerService.DecryptSecret(ctx, cfg)
	if err != nil {
		apiPkg.InternalError(c, "解密密钥失败")
		return
	}

	var apiKey string
	if secret != "" {
		// Try JSON format first: {"api_key": "sk-xxx"}
		var secretData struct {
			APIKey string `json:"api_key"`
		}
		if err := json.Unmarshal([]byte(secret), &secretData); err == nil && secretData.APIKey != "" {
			apiKey = secretData.APIKey
		} else {
			// Fallback: treat the whole string as the API key (legacy format)
			apiKey = secret
		}
	}

	testErr := integration.TestProviderConnection(ctx, cfg.ProviderType, cfg.ProviderName, effectiveFormat, configData.BaseURL, apiKey, configData.Model, h.logger)
	if testErr != nil {
		h.logger.Warn("provider test connection failed",
			zap.String("provider_id", publicID),
			zap.Error(testErr),
		)
		apiPkg.OK(c, gin.H{
			"success": false,
			"error":   testErr.Error(),
		})
		return
	}

	apiPkg.OK(c, gin.H{
		"success": true,
		"message": "连接测试成功",
	})
}
