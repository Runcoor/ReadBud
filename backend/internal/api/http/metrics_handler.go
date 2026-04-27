// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"readbud/internal/api"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"
)

// MetricsHandler handles metrics-related HTTP requests.
type MetricsHandler struct {
	metricsSvc *service.MetricsService
	wechatRepo postgres.WechatAccountRepository
}

// NewMetricsHandler creates a new MetricsHandler.
func NewMetricsHandler(metricsSvc *service.MetricsService, wechatRepo postgres.WechatAccountRepository) *MetricsHandler {
	return &MetricsHandler{
		metricsSvc: metricsSvc,
		wechatRepo: wechatRepo,
	}
}

// RegisterRoutes registers metrics-related routes on the given router group.
func (h *MetricsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/metrics/sync", h.SyncMetrics)
	rg.GET("/articles/:id/metrics", h.GetArticleMetrics)
	rg.GET("/reports/overview", h.GetOverview)
}

// syncMetricsRequest represents the request body for manual metrics sync.
type syncMetricsRequest struct {
	WechatAccountID string `json:"wechat_account_id" binding:"required"`
}

// SyncMetrics handles POST /api/v1/metrics/sync — manual metrics sync trigger.
func (h *MetricsHandler) SyncMetrics(c *gin.Context) {
	var req syncMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	// Resolve public ID to internal model
	account, err := h.wechatRepo.FindByPublicID(c.Request.Context(), req.WechatAccountID)
	if err != nil {
		api.InternalError(c, "查询公众号账号失败")
		return
	}
	if account == nil {
		api.NotFound(c, "公众号账号不存在")
		return
	}

	result, err := h.metricsSvc.SyncMetrics(c.Request.Context(), account.ID, account.AppID)
	if err != nil {
		api.InternalError(c, "同步指标失败")
		return
	}

	api.OK(c, result)
}

// GetArticleMetrics handles GET /api/v1/articles/:id/metrics — single article metrics.
func (h *MetricsHandler) GetArticleMetrics(c *gin.Context) {
	articleID := c.Param("id")
	if articleID == "" {
		api.BadRequest(c, "文章 ID 不能为空")
		return
	}

	// Parse optional date range (defaults to last 30 days)
	startStr := c.Query("start")
	endStr := c.Query("end")

	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	if startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsed
		}
	}
	if endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsed
		}
	}

	snapshots, err := h.metricsSvc.GetArticleMetrics(c.Request.Context(), articleID, start, end)
	if err != nil {
		api.InternalError(c, "获取文章指标失败")
		return
	}

	api.OK(c, gin.H{
		"article_id": articleID,
		"start":      start.Format("2006-01-02"),
		"end":        end.Format("2006-01-02"),
		"snapshots":  snapshots,
	})
}

// GetOverview handles GET /api/v1/reports/overview — aggregated overview report.
func (h *MetricsHandler) GetOverview(c *gin.Context) {
	accountPublicID := c.Query("wechat_account_id")
	if accountPublicID == "" {
		api.BadRequest(c, "公众号账号 ID 不能为空")
		return
	}

	account, err := h.wechatRepo.FindByPublicID(c.Request.Context(), accountPublicID)
	if err != nil {
		api.InternalError(c, "查询公众号账号失败")
		return
	}
	if account == nil {
		api.NotFound(c, "公众号账号不存在")
		return
	}

	// Get all distinct article IDs for this account
	articleIDs, err := h.metricsSvc.GetDistinctArticleIDs(c.Request.Context(), account.ID)
	if err != nil {
		api.InternalError(c, "获取文章列表失败")
		return
	}

	overview, err := h.metricsSvc.GetOverview(c.Request.Context(), articleIDs)
	if err != nil {
		api.InternalError(c, "获取运营总览失败")
		return
	}

	api.OK(c, overview)
}
