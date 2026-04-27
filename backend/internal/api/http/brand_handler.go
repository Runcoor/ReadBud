// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package http

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"

	"readbud/internal/api"
	"readbud/internal/domain"
	"readbud/internal/service"

	"gorm.io/datatypes"
)

// BrandHandler handles brand profile and style profile HTTP requests.
type BrandHandler struct {
	brandSvc *service.BrandProfileService
	styleSvc *service.StyleProfileService
}

// NewBrandHandler creates a new BrandHandler.
func NewBrandHandler(brandSvc *service.BrandProfileService, styleSvc *service.StyleProfileService) *BrandHandler {
	return &BrandHandler{brandSvc: brandSvc, styleSvc: styleSvc}
}

// RegisterRoutes registers brand and style profile routes.
func (h *BrandHandler) RegisterRoutes(rg *gin.RouterGroup) {
	brands := rg.Group("/brand-profiles")
	{
		brands.GET("", h.ListBrands)
		brands.POST("", h.CreateBrand)
		brands.GET("/:id", h.GetBrand)
		brands.PATCH("/:id", h.UpdateBrand)
		brands.DELETE("/:id", h.DeleteBrand)
	}

	styles := rg.Group("/style-profiles")
	{
		styles.GET("", h.ListStyles)
		styles.POST("", h.CreateStyle)
		styles.GET("/:id", h.GetStyle)
		styles.PATCH("/:id", h.UpdateStyle)
		styles.DELETE("/:id", h.DeleteStyle)
	}
}

// ---------- Brand Profile ----------

// ListBrands handles GET /brand-profiles.
func (h *BrandHandler) ListBrands(c *gin.Context) {
	profiles, err := h.brandSvc.List(c.Request.Context())
	if err != nil {
		api.InternalError(c, "获取品牌配置失败")
		return
	}
	api.OK(c, profiles)
}

// CreateBrand handles POST /brand-profiles.
func (h *BrandHandler) CreateBrand(c *gin.Context) {
	var req struct {
		Name            string                 `json:"name" binding:"required"`
		BrandTone       string                 `json:"brand_tone"`
		ForbiddenWords  []string               `json:"forbidden_words"`
		PreferredWords  []string               `json:"preferred_words"`
		CTARules        map[string]interface{} `json:"cta_rules"`
		CoverStyleRules map[string]interface{} `json:"cover_style_rules"`
		ImageStyleRules map[string]interface{} `json:"image_style_rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	bp := &domain.BrandProfile{
		Name:            req.Name,
		BrandTone:       req.BrandTone,
		ForbiddenWords:  toJSON(req.ForbiddenWords),
		PreferredWords:  toJSON(req.PreferredWords),
		CTARules:        toJSON(req.CTARules),
		CoverStyleRules: toJSON(req.CoverStyleRules),
		ImageStyleRules: toJSON(req.ImageStyleRules),
	}

	if err := h.brandSvc.Create(c.Request.Context(), bp); err != nil {
		api.InternalError(c, "创建品牌配置失败")
		return
	}
	api.Created(c, bp)
}

// GetBrand handles GET /brand-profiles/:id.
func (h *BrandHandler) GetBrand(c *gin.Context) {
	publicID := c.Param("id")
	bp, err := h.brandSvc.Get(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "品牌配置不存在")
			return
		}
		api.InternalError(c, "获取品牌配置失败")
		return
	}
	api.OK(c, bp)
}

// UpdateBrand handles PATCH /brand-profiles/:id.
func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	publicID := c.Param("id")
	var req struct {
		Name            string                 `json:"name"`
		BrandTone       string                 `json:"brand_tone"`
		ForbiddenWords  []string               `json:"forbidden_words"`
		PreferredWords  []string               `json:"preferred_words"`
		CTARules        map[string]interface{} `json:"cta_rules"`
		CoverStyleRules map[string]interface{} `json:"cover_style_rules"`
		ImageStyleRules map[string]interface{} `json:"image_style_rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	updates := &domain.BrandProfile{
		Name:            req.Name,
		BrandTone:       req.BrandTone,
		ForbiddenWords:  toJSON(req.ForbiddenWords),
		PreferredWords:  toJSON(req.PreferredWords),
		CTARules:        toJSON(req.CTARules),
		CoverStyleRules: toJSON(req.CoverStyleRules),
		ImageStyleRules: toJSON(req.ImageStyleRules),
	}

	bp, err := h.brandSvc.Update(c.Request.Context(), publicID, updates)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "品牌配置不存在")
			return
		}
		api.InternalError(c, "更新品牌配置失败")
		return
	}
	api.OK(c, bp)
}

// DeleteBrand handles DELETE /brand-profiles/:id.
func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	publicID := c.Param("id")
	bp, err := h.brandSvc.Get(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "品牌配置不存在")
			return
		}
		api.InternalError(c, "获取品牌配置失败")
		return
	}
	_ = bp
	api.OK(c, gin.H{"deleted": true})
}

// ---------- Style Profile ----------

// ListStyles handles GET /style-profiles.
func (h *BrandHandler) ListStyles(c *gin.Context) {
	profiles, err := h.styleSvc.List(c.Request.Context())
	if err != nil {
		api.InternalError(c, "获取写作模板失败")
		return
	}
	api.OK(c, profiles)
}

// CreateStyle handles POST /style-profiles.
func (h *BrandHandler) CreateStyle(c *gin.Context) {
	var req struct {
		Name               string                 `json:"name" binding:"required"`
		ApplicableScene    string                 `json:"applicable_scene"`
		OpeningTemplate    string                 `json:"opening_template"`
		StructureTemplate  map[string]interface{} `json:"structure_template"`
		ClosingTemplate    string                 `json:"closing_template"`
		SentencePreference map[string]interface{} `json:"sentence_preference"`
		TitlePreference    map[string]interface{} `json:"title_preference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	sp := &domain.StyleProfile{
		Name:               req.Name,
		ApplicableScene:    req.ApplicableScene,
		OpeningTemplate:    req.OpeningTemplate,
		StructureTemplate:  toJSON(req.StructureTemplate),
		ClosingTemplate:    req.ClosingTemplate,
		SentencePreference: toJSON(req.SentencePreference),
		TitlePreference:    toJSON(req.TitlePreference),
	}

	if err := h.styleSvc.Create(c.Request.Context(), sp); err != nil {
		api.InternalError(c, "创建写作模板失败")
		return
	}
	api.Created(c, sp)
}

// GetStyle handles GET /style-profiles/:id.
func (h *BrandHandler) GetStyle(c *gin.Context) {
	publicID := c.Param("id")
	sp, err := h.styleSvc.Get(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "写作模板不存在")
			return
		}
		api.InternalError(c, "获取写作模板失败")
		return
	}
	api.OK(c, sp)
}

// UpdateStyle handles PATCH /style-profiles/:id.
func (h *BrandHandler) UpdateStyle(c *gin.Context) {
	publicID := c.Param("id")
	var req struct {
		Name               string                 `json:"name"`
		ApplicableScene    string                 `json:"applicable_scene"`
		OpeningTemplate    string                 `json:"opening_template"`
		StructureTemplate  map[string]interface{} `json:"structure_template"`
		ClosingTemplate    string                 `json:"closing_template"`
		SentencePreference map[string]interface{} `json:"sentence_preference"`
		TitlePreference    map[string]interface{} `json:"title_preference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.HandleBindError(c, err)
		return
	}

	updates := &domain.StyleProfile{
		Name:               req.Name,
		ApplicableScene:    req.ApplicableScene,
		OpeningTemplate:    req.OpeningTemplate,
		StructureTemplate:  toJSON(req.StructureTemplate),
		ClosingTemplate:    req.ClosingTemplate,
		SentencePreference: toJSON(req.SentencePreference),
		TitlePreference:    toJSON(req.TitlePreference),
	}

	sp, err := h.styleSvc.Update(c.Request.Context(), publicID, updates)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "写作模板不存在")
			return
		}
		api.InternalError(c, "更新写作模板失败")
		return
	}
	api.OK(c, sp)
}

// DeleteStyle handles DELETE /style-profiles/:id.
func (h *BrandHandler) DeleteStyle(c *gin.Context) {
	publicID := c.Param("id")
	sp, err := h.styleSvc.Get(c.Request.Context(), publicID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			api.NotFound(c, "写作模板不存在")
			return
		}
		api.InternalError(c, "获取写作模板失败")
		return
	}
	_ = sp
	api.OK(c, gin.H{"deleted": true})
}

// toJSON marshals an interface value to datatypes.JSON for GORM storage.
func toJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(b)
}
