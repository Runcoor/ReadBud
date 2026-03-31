package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"
)

// PublishHandler handles publish job HTTP endpoints.
type PublishHandler struct {
	publishService *service.PublishService
	draftRepo      postgres.ArticleDraftRepository
	wechatRepo     postgres.WechatAccountRepository
}

// NewPublishHandler creates a new PublishHandler.
func NewPublishHandler(
	svc *service.PublishService,
	draftRepo postgres.ArticleDraftRepository,
	wechatRepo postgres.WechatAccountRepository,
) *PublishHandler {
	return &PublishHandler{
		publishService: svc,
		draftRepo:      draftRepo,
		wechatRepo:     wechatRepo,
	}
}

// RegisterRoutes registers publish routes on the given router group.
func (h *PublishHandler) RegisterRoutes(rg *gin.RouterGroup) {
	publish := rg.Group("/publish/jobs")
	{
		publish.POST("", h.CreateJob)
		publish.GET("/:id", h.GetJob)
		publish.POST("/:id/cancel", h.CancelJob)
		publish.POST("/:id/retry", h.RetryJob)
	}
}

// CreateJob handles POST /api/v1/publish/jobs.
func (h *PublishHandler) CreateJob(c *gin.Context) {
	var req dto.CreatePublishJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	// Resolve draft internal ID from public ID
	d, err := h.draftRepo.FindByPublicID(c.Request.Context(), req.DraftID)
	if err != nil || d == nil {
		apiPkg.NotFound(c, "草稿不存在")
		return
	}

	// Resolve wechat account internal ID from public ID
	wa, err := h.wechatRepo.FindByPublicID(c.Request.Context(), req.WechatAccountID)
	if err != nil || wa == nil {
		apiPkg.NotFound(c, "公众号账号不存在")
		return
	}

	job, err := h.publishService.CreateJob(c.Request.Context(), d.ID, wa.ID, req.PublishMode)
	if err != nil {
		apiPkg.InternalError(c, "创建发布任务失败")
		return
	}

	apiPkg.Created(c, gin.H{
		"id":           job.PublicID,
		"draft_id":     req.DraftID,
		"publish_mode": job.PublishMode,
		"status":       job.Status,
		"created_at":   job.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// GetJob handles GET /api/v1/publish/jobs/:id.
func (h *PublishHandler) GetJob(c *gin.Context) {
	publicID := c.Param("id")

	job, err := h.publishService.GetJob(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "发布任务不存在")
			return
		}
		apiPkg.InternalError(c, "获取发布任务失败")
		return
	}

	resp := gin.H{
		"id":           job.PublicID,
		"publish_mode": job.PublishMode,
		"status":       job.Status,
		"retry_count":  job.RetryCount,
		"last_error":   job.LastError,
		"created_at":   job.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Include article_url from publish record if job succeeded
	if job.Status == "success" {
		articleURL, _ := h.publishService.GetArticleURL(c.Request.Context(), job.ID)
		if articleURL != "" {
			resp["article_url"] = articleURL
		}
	}

	apiPkg.OK(c, resp)
}

// RetryJob handles POST /api/v1/publish/jobs/:id/retry.
func (h *PublishHandler) RetryJob(c *gin.Context) {
	publicID := c.Param("id")

	job, err := h.publishService.RetryJob(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "发布任务不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidState) {
			apiPkg.BadRequest(c, "当前任务状态不支持重试")
			return
		}
		apiPkg.InternalError(c, "重试发布任务失败")
		return
	}

	apiPkg.OK(c, gin.H{
		"id":     job.PublicID,
		"status": job.Status,
	})
}

// CancelJob handles POST /api/v1/publish/jobs/:id/cancel.
func (h *PublishHandler) CancelJob(c *gin.Context) {
	publicID := c.Param("id")

	job, err := h.publishService.CancelJob(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "发布任务不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidState) {
			apiPkg.BadRequest(c, "当前任务状态不支持取消")
			return
		}
		apiPkg.InternalError(c, "取消发布任务失败")
		return
	}

	apiPkg.OK(c, gin.H{
		"id":     job.PublicID,
		"status": job.Status,
	})
}
