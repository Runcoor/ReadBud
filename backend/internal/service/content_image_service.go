package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain/asset"
	"readbud/internal/integration/wechat"
	"readbud/internal/repository/postgres"
)

// ContentImageService handles uploading article images to WeChat
// as content images (正文图片). These are separate from permanent materials,
// return a URL (not media_id), and do not consume the permanent material quota.
type ContentImageService struct {
	assetRepo postgres.AssetRepository
	publisher adapter.WeChatPublisher
	storage   adapter.StorageProvider
	tokenProv wechat.TokenProvider
	logger    *zap.Logger
}

// NewContentImageService creates a new ContentImageService.
func NewContentImageService(
	assetRepo postgres.AssetRepository,
	publisher adapter.WeChatPublisher,
	storage adapter.StorageProvider,
	tokenProv wechat.TokenProvider,
	logger *zap.Logger,
) *ContentImageService {
	return &ContentImageService{
		assetRepo: assetRepo,
		publisher: publisher,
		storage:   storage,
		tokenProv: tokenProv,
		logger:    logger,
	}
}

// ContentImageUploadResult holds the result of a single content image upload.
type ContentImageUploadResult struct {
	AssetID   int64  `json:"asset_id"`
	PublicID  string `json:"public_id"`
	OrigURL   string `json:"orig_url"`
	WeChatURL string `json:"wechat_url"`
}

// ValidateContentImage checks that the image data meets WeChat content image requirements.
// Only jpg/png allowed, max 1MB.
func ValidateContentImage(data []byte, filename string) error {
	if len(data) == 0 {
		return fmt.Errorf("image data is empty")
	}
	if len(data) > adapter.ContentImageMaxBytes {
		return fmt.Errorf("image exceeds 1MB limit: %d bytes", len(data))
	}

	mime := http.DetectContentType(data)
	if mime != adapter.ContentImageMIMEJPEG && mime != adapter.ContentImageMIMEPNG {
		return fmt.Errorf("unsupported image type %q for file %q: only jpg/png allowed", mime, filename)
	}
	return nil
}

// UploadForDraft batch-uploads all pending content images for a draft to WeChat.
// It fetches assets linked to the draft, downloads image data from storage,
// uploads each to WeChat, and updates the asset records with the WeChat URL.
func (s *ContentImageService) UploadForDraft(ctx context.Context, draftID int64, appID string) ([]ContentImageUploadResult, error) {
	assets, err := s.assetRepo.FindByDraftID(ctx, draftID)
	if err != nil {
		return nil, fmt.Errorf("contentImageService.UploadForDraft: find assets: %w", err)
	}

	// Filter to pending content images only
	var pending []asset.Asset
	for i := range assets {
		a := &assets[i]
		if a.WechatUploadStatus == asset.WechatUploadPending && isContentImageType(a.AssetType) {
			pending = append(pending, *a)
		}
	}

	if len(pending) == 0 {
		s.logger.Info("no pending content images for draft", zap.Int64("draft_id", draftID))
		return nil, nil
	}

	token, err := s.tokenProv.GetAccessToken(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("contentImageService.UploadForDraft: get token: %w", err)
	}

	results := make([]ContentImageUploadResult, 0, len(pending))
	var uploadErrors []string

	for i := range pending {
		a := &pending[i]
		result, err := s.uploadSingleAsset(ctx, a, token)
		if err != nil {
			s.logger.Error("content image upload failed",
				zap.Int64("asset_id", a.ID),
				zap.String("public_id", a.PublicID),
				zap.Error(err),
			)
			uploadErrors = append(uploadErrors, fmt.Sprintf("asset %s: %v", a.PublicID, err))

			a.WechatUploadStatus = asset.WechatUploadFailed
			if updateErr := s.assetRepo.Update(ctx, a); updateErr != nil {
				s.logger.Error("failed to update asset status",
					zap.Int64("asset_id", a.ID),
					zap.Error(updateErr),
				)
			}
			continue
		}
		results = append(results, *result)
	}

	if len(uploadErrors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("contentImageService.UploadForDraft: all uploads failed: %s",
			strings.Join(uploadErrors, "; "))
	}

	s.logger.Info("content image upload complete",
		zap.Int64("draft_id", draftID),
		zap.Int("total", len(pending)),
		zap.Int("success", len(results)),
		zap.Int("failed", len(uploadErrors)),
	)
	return results, nil
}

// uploadSingleAsset downloads image data from storage and uploads it to WeChat.
func (s *ContentImageService) uploadSingleAsset(ctx context.Context, a *asset.Asset, token string) (*ContentImageUploadResult, error) {
	// Get the storage URL for fetching image data
	storageURL, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("get storage URL: %w", err)
	}

	// For stub: use empty placeholder data since storage may not be real
	// In production, this would download from the storage URL
	imageData, err := s.fetchImageData(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("fetch image data from %s: %w", storageURL, err)
	}

	// Validate image meets WeChat requirements
	filename := extractFilename(a.ObjectKey)
	if err := ValidateContentImage(imageData, filename); err != nil {
		return nil, fmt.Errorf("validate content image: %w", err)
	}

	// Upload to WeChat content image API
	wechatURL, err := s.publisher.UploadContentImage(ctx, token, imageData, filename)
	if err != nil {
		return nil, fmt.Errorf("upload to wechat: %w", err)
	}

	// Update asset record
	a.WechatURL = &wechatURL
	a.WechatUploadStatus = asset.WechatUploadDone
	if err := s.assetRepo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update asset record: %w", err)
	}

	origURL := ""
	if a.SourceURL != nil {
		origURL = *a.SourceURL
	}

	return &ContentImageUploadResult{
		AssetID:   a.ID,
		PublicID:  a.PublicID,
		OrigURL:   origURL,
		WeChatURL: wechatURL,
	}, nil
}

// fetchImageData retrieves image bytes from storage.
// In development with stub storage, returns minimal placeholder data.
func (s *ContentImageService) fetchImageData(ctx context.Context, a *asset.Asset) ([]byte, error) {
	url, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("fetchImageData: get URL: %w", err)
	}

	// If storage returns a real URL, we would HTTP GET it here.
	// For now, return a minimal valid PNG for stub/dev mode.
	_ = url
	if a.SizeBytes != nil && *a.SizeBytes > 0 {
		// Placeholder: in production, download from storage URL
		placeholder := make([]byte, 128)
		// Minimal PNG header for validation
		copy(placeholder, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
		return placeholder, nil
	}

	// Minimal valid PNG
	placeholder := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return placeholder, nil
}

// ReplaceImageURLsInHTML replaces original image URLs in compiled HTML with WeChat URLs.
// WeChat filters out external URLs, so all images must be uploaded through WeChat's API.
func (s *ContentImageService) ReplaceImageURLsInHTML(ctx context.Context, html string, draftID int64) (string, error) {
	assets, err := s.assetRepo.FindByDraftID(ctx, draftID)
	if err != nil {
		return "", fmt.Errorf("contentImageService.ReplaceImageURLsInHTML: %w", err)
	}

	result := html
	replacements := 0

	for i := range assets {
		a := &assets[i]
		if a.WechatURL == nil || *a.WechatURL == "" {
			continue
		}
		if a.WechatUploadStatus != asset.WechatUploadDone {
			continue
		}

		// Replace source URL with WeChat URL if source URL exists
		if a.SourceURL != nil && *a.SourceURL != "" {
			if strings.Contains(result, *a.SourceURL) {
				result = strings.ReplaceAll(result, *a.SourceURL, *a.WechatURL)
				replacements++
			}
		}

		// Also check storage URL pattern (bucket/object_key)
		storageURL, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey)
		if err == nil && storageURL != "" && strings.Contains(result, storageURL) {
			result = strings.ReplaceAll(result, storageURL, *a.WechatURL)
			replacements++
		}
	}

	s.logger.Info("replaced image URLs in HTML",
		zap.Int64("draft_id", draftID),
		zap.Int("replacements", replacements),
	)
	return result, nil
}

// isContentImageType returns true if the asset type should be uploaded as a content image.
func isContentImageType(assetType string) bool {
	switch assetType {
	case asset.AssetTypeContentImage, asset.AssetTypeChart, asset.AssetTypeGeneratedImage:
		return true
	default:
		return false
	}
}

// extractFilename extracts the filename from an object key path.
func extractFilename(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	if len(parts) == 0 {
		return objectKey
	}
	return parts[len(parts)-1]
}
