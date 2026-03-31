package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	"readbud/internal/api"
	"readbud/internal/service"
)

// DistributionHandler handles distribution package HTTP requests.
type DistributionHandler struct {
	distSvc *service.DistributionService
}

// NewDistributionHandler creates a new DistributionHandler.
func NewDistributionHandler(distSvc *service.DistributionService) *DistributionHandler {
	return &DistributionHandler{distSvc: distSvc}
}

// RegisterRoutes registers distribution package routes on the given router group.
func (h *DistributionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	dist := rg.Group("/distributions")
	{
		dist.POST("/generate", h.Generate)
		dist.GET("/by-draft/:draftId", h.GetByDraft)
		dist.GET("/:id", h.GetByID)
		dist.DELETE("/:id", h.Delete)
	}
}

// Generate handles POST /api/v1/distributions/generate — generate distribution materials.
func (h *DistributionHandler) Generate(c *gin.Context) {
	var req service.GenerateDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	result, err := h.distSvc.Generate(c.Request.Context(), req.DraftPublicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "草稿不存在")
			return
		}
		api.InternalError(c, "生成分发素材包失败")
		return
	}

	api.Created(c, result)
}

// GetByDraft handles GET /api/v1/distributions/by-draft/:draftId — get by draft public ID.
func (h *DistributionHandler) GetByDraft(c *gin.Context) {
	draftID := c.Param("draftId")
	if draftID == "" {
		api.BadRequest(c, "草稿 ID 不能为空")
		return
	}

	result, err := h.distSvc.GetByDraftPublicID(c.Request.Context(), draftID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "分发素材包不存在")
			return
		}
		api.InternalError(c, "获取分发素材包失败")
		return
	}

	api.OK(c, result)
}

// GetByID handles GET /api/v1/distributions/:id — get by package public ID.
func (h *DistributionHandler) GetByID(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "素材包 ID 不能为空")
		return
	}

	result, err := h.distSvc.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "分发素材包不存在")
			return
		}
		api.InternalError(c, "获取分发素材包失败")
		return
	}

	api.OK(c, result)
}

// Delete handles DELETE /api/v1/distributions/:id — remove distribution package.
func (h *DistributionHandler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "素材包 ID 不能为空")
		return
	}

	err := h.distSvc.Delete(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "分发素材包不存在")
			return
		}
		api.InternalError(c, "删除分发素材包失败")
		return
	}

	api.OK(c, gin.H{"deleted": true})
}
