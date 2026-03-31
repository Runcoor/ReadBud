package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"readbud/internal/api"
	"readbud/internal/service"
)

// TopicHandler handles topic library HTTP requests.
type TopicHandler struct {
	topicSvc *service.TopicLibraryService
}

// NewTopicHandler creates a new TopicHandler.
func NewTopicHandler(topicSvc *service.TopicLibraryService) *TopicHandler {
	return &TopicHandler{topicSvc: topicSvc}
}

// RegisterRoutes registers topic library routes on the given router group.
func (h *TopicHandler) RegisterRoutes(rg *gin.RouterGroup) {
	topics := rg.Group("/reports/topics")
	{
		topics.GET("", h.ListTopics)
		topics.POST("", h.CreateTopic)
		topics.GET("/recommendations", h.GetRecommendations)
		topics.GET("/search", h.SearchTopics)
		topics.GET("/:id", h.GetTopic)
		topics.PATCH("/:id", h.UpdateTopic)
		topics.DELETE("/:id", h.DeleteTopic)
		topics.POST("/:id/performance", h.UpdatePerformance)
	}
}

// allowedSortFields defines valid sort fields for topic listing.
var allowedSortFields = map[string]bool{
	"recommend_weight": true,
	"created_at":       true,
	"updated_at":       true,
	"keyword":          true,
}

// ListTopics handles GET /api/v1/reports/topics — paginated topic list.
func (h *TopicHandler) ListTopics(c *gin.Context) {
	page := queryInt(c, "page", 1)
	size := queryInt(c, "size", 20)
	sortBy := c.DefaultQuery("sort_by", "recommend_weight")

	// Validate sort field to prevent SQL injection
	if !allowedSortFields[sortBy] {
		sortBy = "recommend_weight"
	}

	// Clamp pagination
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	result, err := h.topicSvc.List(c.Request.Context(), page, size, sortBy)
	if err != nil {
		api.InternalError(c, "获取选题列表失败")
		return
	}

	api.OK(c, result)
}

// CreateTopic handles POST /api/v1/reports/topics — add a new topic.
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	var req service.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	topic, err := h.topicSvc.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrConflict) || isKeywordDuplicate(err) {
			api.Conflict(c, "关键词已存在")
			return
		}
		api.InternalError(c, "创建选题失败")
		return
	}

	api.Created(c, topic)
}

// GetTopic handles GET /api/v1/reports/topics/:id — single topic detail.
func (h *TopicHandler) GetTopic(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "选题 ID 不能为空")
		return
	}

	topic, err := h.topicSvc.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "选题不存在")
			return
		}
		api.InternalError(c, "获取选题失败")
		return
	}

	api.OK(c, topic)
}

// UpdateTopic handles PATCH /api/v1/reports/topics/:id — modify a topic.
func (h *TopicHandler) UpdateTopic(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "选题 ID 不能为空")
		return
	}

	var req service.UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	topic, err := h.topicSvc.Update(c.Request.Context(), publicID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "选题不存在")
			return
		}
		if errors.Is(err, service.ErrConflict) || isKeywordDuplicate(err) {
			api.Conflict(c, "关键词已存在")
			return
		}
		api.InternalError(c, "更新选题失败")
		return
	}

	api.OK(c, topic)
}

// DeleteTopic handles DELETE /api/v1/reports/topics/:id — remove a topic.
func (h *TopicHandler) DeleteTopic(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "选题 ID 不能为空")
		return
	}

	err := h.topicSvc.Delete(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "选题不存在")
			return
		}
		api.InternalError(c, "删除选题失败")
		return
	}

	api.OK(c, gin.H{"deleted": true})
}

// GetRecommendations handles GET /api/v1/reports/topics/recommendations.
func (h *TopicHandler) GetRecommendations(c *gin.Context) {
	limit := queryInt(c, "limit", 10)

	topics, err := h.topicSvc.GetRecommendations(c.Request.Context(), limit)
	if err != nil {
		api.InternalError(c, "获取选题推荐失败")
		return
	}

	api.OK(c, gin.H{
		"items": topics,
		"count": len(topics),
	})
}

// SearchTopics handles GET /api/v1/reports/topics/search?q=xxx.
func (h *TopicHandler) SearchTopics(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		api.BadRequest(c, "搜索关键词不能为空")
		return
	}
	page := queryInt(c, "page", 1)
	size := queryInt(c, "size", 20)

	result, err := h.topicSvc.Search(c.Request.Context(), query, page, size)
	if err != nil {
		api.InternalError(c, "搜索选题失败")
		return
	}

	api.OK(c, result)
}

// UpdatePerformance handles POST /api/v1/reports/topics/:id/performance.
func (h *TopicHandler) UpdatePerformance(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "选题 ID 不能为空")
		return
	}

	var req service.PerformanceFeedback
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	err := h.topicSvc.UpdatePerformance(c.Request.Context(), publicID, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "选题不存在")
			return
		}
		api.InternalError(c, "更新选题表现失败")
		return
	}

	api.OK(c, gin.H{"updated": true})
}

// queryInt parses an integer query parameter with a default fallback.
func queryInt(c *gin.Context, key string, defaultVal int) int {
	str := c.Query(key)
	if str == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal
	}
	return val
}

// isKeywordDuplicate checks if the error indicates a duplicate keyword.
func isKeywordDuplicate(err error) bool {
	return err != nil && (contains(err.Error(), "already exists") || contains(err.Error(), "already in use"))
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

// searchStr is a simple substring search.
func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
