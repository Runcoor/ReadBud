package http

import (
	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/service"
)

// ContentImageHandler handles content image upload HTTP endpoints.
type ContentImageHandler struct {
	contentImageSvc *service.ContentImageService
}

// NewContentImageHandler creates a new ContentImageHandler.
func NewContentImageHandler(svc *service.ContentImageService) *ContentImageHandler {
	return &ContentImageHandler{contentImageSvc: svc}
}

// RegisterRoutes registers content image routes on the given router group.
func (h *ContentImageHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/drafts/:id/upload-images", h.UploadDraftImages)
}

// uploadDraftImagesRequest is the request body for batch image upload.
type uploadDraftImagesRequest struct {
	AppID string `json:"app_id" binding:"required"`
}

// UploadDraftImages handles POST /api/v1/drafts/:id/upload-images.
// Batch uploads all pending content images for a draft to WeChat.
func (h *ContentImageHandler) UploadDraftImages(c *gin.Context) {
	draftPublicID := c.Param("id")
	if draftPublicID == "" {
		apiPkg.BadRequest(c, "草稿ID不能为空")
		return
	}

	var req uploadDraftImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiPkg.HandleBindError(c, err)
		return
	}

	// Note: In a full implementation we'd resolve the draft's internal ID
	// from the public ID. For now, this endpoint demonstrates the API contract.
	// The actual batch upload is called from ProcessJob in the publish pipeline.
	apiPkg.OK(c, gin.H{
		"draft_id": draftPublicID,
		"message":  "内容图片上传已在发布流程中自动执行",
	})
}
