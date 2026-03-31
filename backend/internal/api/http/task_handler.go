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
	}
}

// Create handles POST /api/v1/tasks.
func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.BadRequest(c, "请输入有效的任务参数")
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

	resp, err := h.taskService.ListRecent(c.Request.Context(), page, pageSize)
	if err != nil {
		apiPkg.InternalError(c, "获取任务列表失败")
		return
	}

	apiPkg.OK(c, resp)
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
		apiPkg.BadRequest(c, err.Error())
		return
	}

	apiPkg.OK(c, vo)
}
