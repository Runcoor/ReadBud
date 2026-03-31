package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/service"
)

// TaskHandler handles task HTTP endpoints.
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: svc}
}

// RegisterRoutes registers task routes on the given router group.
func (h *TaskHandler) RegisterRoutes(rg *gin.RouterGroup) {
	tasks := rg.Group("/tasks")
	{
		tasks.POST("", h.Create)
		tasks.GET("", h.List)
		tasks.GET("/:id", h.Get)
		tasks.POST("/:id/retry", h.Retry)
		tasks.POST("/:id/cancel", h.CancelTask)
	}
}

// Create handles POST /api/v1/tasks.
func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	vo, err := h.taskService.Create(c.Request.Context(), req)
	if err != nil {
		apiPkg.InternalError(c, "创建任务失败")
		return
	}

	apiPkg.Created(c, vo)
}

// Get handles GET /api/v1/tasks/:id.
func (h *TaskHandler) Get(c *gin.Context) {
	publicID := c.Param("id")

	vo, err := h.taskService.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "任务不存在")
			return
		}
		apiPkg.InternalError(c, "获取任务详情失败")
		return
	}

	apiPkg.OK(c, vo)
}

// List handles GET /api/v1/tasks.
func (h *TaskHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Clamp pagination values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	status := c.Query("status")

	resp, err := h.taskService.ListRecent(c.Request.Context(), page, pageSize, status)
	if err != nil {
		apiPkg.InternalError(c, "获取任务列表失败")
		return
	}

	apiPkg.OK(c, resp)
}

// CancelTask handles POST /api/v1/tasks/:id/cancel.
func (h *TaskHandler) CancelTask(c *gin.Context) {
	publicID := c.Param("id")

	if err := h.taskService.CancelTask(c.Request.Context(), publicID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "任务不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidState) {
			apiPkg.BadRequest(c, "当前任务状态不支持取消")
			return
		}
		apiPkg.InternalError(c, "取消任务失败")
		return
	}

	apiPkg.OK(c, nil)
}

// Retry handles POST /api/v1/tasks/:id/retry.
func (h *TaskHandler) Retry(c *gin.Context) {
	publicID := c.Param("id")

	vo, err := h.taskService.Retry(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			apiPkg.NotFound(c, "任务不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidState) {
			apiPkg.BadRequest(c, "当前任务状态不支持重试")
			return
		}
		apiPkg.InternalError(c, "重试任务失败")
		return
	}

	apiPkg.OK(c, vo)
}
