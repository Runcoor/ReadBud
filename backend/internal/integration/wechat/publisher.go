package wechat

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// StubWeChatPublisher is a placeholder implementation of adapter.WeChatPublisher
// for development and testing. Will be replaced with real WeChat API calls.
type StubWeChatPublisher struct {
	logger *zap.Logger
}

// NewStubWeChatPublisher creates a new stub WeChat publisher.
func NewStubWeChatPublisher(logger *zap.Logger) *StubWeChatPublisher {
	return &StubWeChatPublisher{logger: logger}
}

// UploadImage uploads an image to WeChat and returns a stub media_id.
func (p *StubWeChatPublisher) UploadImage(ctx context.Context, accessToken string, imageData []byte, filename string) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("StubWeChatPublisher.UploadImage: access token is required")
	}
	if len(imageData) == 0 {
		return "", fmt.Errorf("StubWeChatPublisher.UploadImage: image data is empty")
	}

	mediaID := fmt.Sprintf("stub_media_%s_%d", filename, time.Now().UnixMilli())
	p.logger.Info("stub: uploaded image to WeChat",
		zap.String("filename", filename),
		zap.Int("size_bytes", len(imageData)),
		zap.String("media_id", mediaID),
	)
	return mediaID, nil
}

// UploadContentImage uploads an in-article image and returns a stub WeChat URL.
// In production this calls the WeChat content image upload API which returns
// a URL rather than a media_id (separate from permanent materials).
func (p *StubWeChatPublisher) UploadContentImage(ctx context.Context, accessToken string, imageData []byte, filename string) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("StubWeChatPublisher.UploadContentImage: access token is required")
	}
	if len(imageData) == 0 {
		return "", fmt.Errorf("StubWeChatPublisher.UploadContentImage: image data is empty")
	}
	if len(imageData) > adapter.ContentImageMaxBytes {
		return "", fmt.Errorf("StubWeChatPublisher.UploadContentImage: image exceeds 1MB limit (%d bytes)", len(imageData))
	}

	wechatURL := fmt.Sprintf("https://mmbiz.qpic.cn/stub/%s_%d.png", filename, time.Now().UnixMilli())
	p.logger.Info("stub: uploaded content image to WeChat",
		zap.String("filename", filename),
		zap.Int("size_bytes", len(imageData)),
		zap.String("wechat_url", wechatURL),
	)
	return wechatURL, nil
}

// CreateDraft creates a draft article on WeChat and returns a stub media_id.
func (p *StubWeChatPublisher) CreateDraft(ctx context.Context, accessToken string, article adapter.WeChatArticle) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("StubWeChatPublisher.CreateDraft: access token is required")
	}
	if article.Title == "" {
		return "", fmt.Errorf("StubWeChatPublisher.CreateDraft: article title is required")
	}

	mediaID := fmt.Sprintf("stub_draft_%d", time.Now().UnixMilli())
	p.logger.Info("stub: created WeChat draft",
		zap.String("title", article.Title),
		zap.String("author", article.Author),
		zap.String("media_id", mediaID),
	)
	return mediaID, nil
}

// Publish publishes a draft article and returns a stub result.
func (p *StubWeChatPublisher) Publish(ctx context.Context, accessToken string, mediaID string) (*adapter.WeChatPublishResult, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("StubWeChatPublisher.Publish: access token is required")
	}
	if mediaID == "" {
		return nil, fmt.Errorf("StubWeChatPublisher.Publish: media ID is required")
	}

	result := &adapter.WeChatPublishResult{
		MediaID:   mediaID,
		MsgID:     fmt.Sprintf("stub_msg_%d", time.Now().UnixMilli()),
		PublishID: fmt.Sprintf("stub_pub_%d", time.Now().UnixMilli()),
	}
	p.logger.Info("stub: published WeChat article",
		zap.String("media_id", mediaID),
		zap.String("publish_id", result.PublishID),
	)
	return result, nil
}

// Compile-time check that StubWeChatPublisher satisfies the interface.
var _ adapter.WeChatPublisher = (*StubWeChatPublisher)(nil)
