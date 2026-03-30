package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/service"
)

// ProviderHandler handles provider config HTTP endpoints.
type ProviderHandler struct {
	providerService *service.ProviderConfigService
}

// NewProviderHandler creates a new ProviderHandler.
func NewProviderHandler(svc *service.ProviderConfigService) *ProviderHandler {
	return &ProviderHandler{providerService: svc}
}

// RegisterRoutes registers provider routes on the given router group.
func (h *ProviderHandler) RegisterRoutes(rg *gin.RouterGroup) {
	providers := rg.Group("/providers")
	{
		providers.GET("", h.List)
		providers.POST("", h.Create)
		providers.PUT("/:id", h.Update)
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
		apiPkg.BadRequest(c, "请求参数不正确")
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
		apiPkg.BadRequest(c, "请求参数不正确")
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
