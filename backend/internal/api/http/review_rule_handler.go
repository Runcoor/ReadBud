package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	"readbud/internal/api"
	"readbud/internal/api/dto"
	"readbud/internal/domain"
	"readbud/internal/service"
)

// ReviewRuleHandler handles review rule HTTP requests.
type ReviewRuleHandler struct {
	ruleSvc *service.ReviewRuleService
}

// NewReviewRuleHandler creates a new ReviewRuleHandler.
func NewReviewRuleHandler(ruleSvc *service.ReviewRuleService) *ReviewRuleHandler {
	return &ReviewRuleHandler{ruleSvc: ruleSvc}
}

// RegisterRoutes registers review rule routes on the given router group.
func (h *ReviewRuleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rules := rg.Group("/review-rules")
	{
		rules.GET("", h.ListRules)
		rules.POST("", h.CreateRule)
		rules.POST("/evaluate", h.EvaluateContent)
		rules.GET("/:id", h.GetRule)
		rules.PUT("/:id", h.UpdateRule)
		rules.DELETE("/:id", h.DeleteRule)
		rules.POST("/:id/toggle", h.ToggleRule)
	}
}

// ListRules handles GET /api/v1/review-rules — list all rules.
func (h *ReviewRuleHandler) ListRules(c *gin.Context) {
	rules, err := h.ruleSvc.List(c.Request.Context())
	if err != nil {
		api.InternalError(c, "获取审核规则列表失败")
		return
	}

	api.OK(c, rules)
}

// CreateRule handles POST /api/v1/review-rules — create a new rule.
func (h *ReviewRuleHandler) CreateRule(c *gin.Context) {
	var req dto.CreateReviewRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	rule := &domain.ReviewRule{
		RuleType:    req.RuleType,
		RuleContent: req.RuleContent,
		RiskLevel:   req.RiskLevel,
		IsEnabled:   1,
	}
	if req.IsEnabled != nil {
		rule.IsEnabled = *req.IsEnabled
	}

	if err := h.ruleSvc.Create(c.Request.Context(), rule); err != nil {
		api.InternalError(c, "创建审核规则失败")
		return
	}

	api.Created(c, rule)
}

// GetRule handles GET /api/v1/review-rules/:id — get a single rule.
func (h *ReviewRuleHandler) GetRule(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "规则 ID 不能为空")
		return
	}

	rule, err := h.ruleSvc.Get(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "审核规则不存在")
			return
		}
		api.InternalError(c, "获取审核规则失败")
		return
	}

	api.OK(c, rule)
}

// UpdateRule handles PUT /api/v1/review-rules/:id — update a rule.
func (h *ReviewRuleHandler) UpdateRule(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "规则 ID 不能为空")
		return
	}

	var req dto.UpdateReviewRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	rule, err := h.ruleSvc.Update(c.Request.Context(), publicID, req.RuleType, req.RuleContent, req.RiskLevel)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "审核规则不存在")
			return
		}
		api.InternalError(c, "更新审核规则失败")
		return
	}

	api.OK(c, rule)
}

// DeleteRule handles DELETE /api/v1/review-rules/:id — delete a rule.
func (h *ReviewRuleHandler) DeleteRule(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "规则 ID 不能为空")
		return
	}

	err := h.ruleSvc.Delete(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "审核规则不存在")
			return
		}
		api.InternalError(c, "删除审核规则失败")
		return
	}

	api.OK(c, gin.H{"deleted": true})
}

// ToggleRule handles POST /api/v1/review-rules/:id/toggle — toggle enabled/disabled.
func (h *ReviewRuleHandler) ToggleRule(c *gin.Context) {
	publicID := c.Param("id")
	if publicID == "" {
		api.BadRequest(c, "规则 ID 不能为空")
		return
	}

	var req dto.ToggleReviewRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	rule, err := h.ruleSvc.Toggle(c.Request.Context(), publicID, req.IsEnabled == 1)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "审核规则不存在")
			return
		}
		api.InternalError(c, "切换审核规则状态失败")
		return
	}

	api.OK(c, rule)
}

// EvaluateContent handles POST /api/v1/review-rules/evaluate — evaluate content against all rules.
func (h *ReviewRuleHandler) EvaluateContent(c *gin.Context) {
	var req dto.EvaluateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	violations, err := h.ruleSvc.EvaluateContent(c.Request.Context(), req.Content)
	if err != nil {
		api.InternalError(c, "内容审核评估失败")
		return
	}

	api.OK(c, gin.H{
		"violations": violations,
		"count":      len(violations),
	})
}
