// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// WeChat publishing endpoints.
const (
	endpointUploadContentImg = "https://api.weixin.qq.com/cgi-bin/media/uploadimg"
	endpointAddMaterial      = "https://api.weixin.qq.com/cgi-bin/material/add_material"
	endpointDraftAdd         = "https://api.weixin.qq.com/cgi-bin/draft/add"
	endpointFreepublish      = "https://api.weixin.qq.com/cgi-bin/freepublish/submit"
	endpointFreepublishGet   = "https://api.weixin.qq.com/cgi-bin/freepublish/get"

	defaultPublisherTimeout = 30 * time.Second
)

// RealWeChatPublisher is the production implementation of adapter.WeChatPublisher.
// It calls real WeChat OA endpoints; the caller is responsible for providing a valid
// access_token (typically from RealTokenProvider).
type RealWeChatPublisher struct {
	http   *http.Client
	logger *zap.Logger
}

// NewRealWeChatPublisher builds a production WeChatPublisher.
// httpClient may be nil; a default 30s-timeout client will be used.
func NewRealWeChatPublisher(httpClient *http.Client, logger *zap.Logger) *RealWeChatPublisher {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultPublisherTimeout}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RealWeChatPublisher{http: httpClient, logger: logger}
}

// ----- UploadContentImage -----

type uploadContentImgResp struct {
	URL     string `json:"url"`
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// UploadContentImage uploads an image used inside article body via /media/uploadimg.
// Returns an mmbiz.qpic.cn URL safe to embed in WeChat HTML. Max 1MB, jpg/png only.
func (p *RealWeChatPublisher) UploadContentImage(ctx context.Context, accessToken string, imageData []byte, filename string) (string, error) {
	if accessToken == "" {
		return "", errors.New("UploadContentImage: access token is required")
	}
	if len(imageData) == 0 {
		return "", errors.New("UploadContentImage: image data is empty")
	}
	if len(imageData) > adapter.ContentImageMaxBytes {
		return "", fmt.Errorf("UploadContentImage: image exceeds 1MB limit (%d bytes)", len(imageData))
	}

	endpoint := endpointUploadContentImg + "?access_token=" + url.QueryEscape(accessToken)
	body, contentType, err := buildMediaMultipart("media", filename, imageData)
	if err != nil {
		return "", fmt.Errorf("UploadContentImage: build multipart: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return "", fmt.Errorf("UploadContentImage: build request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	var resp uploadContentImgResp
	if err := p.doJSON(req, &resp); err != nil {
		return "", fmt.Errorf("UploadContentImage: %w", err)
	}
	if resp.Errcode != 0 {
		return "", &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "uploadimg"}
	}
	if resp.URL == "" {
		return "", errors.New("UploadContentImage: empty url in response")
	}
	p.logger.Info("uploaded content image to WeChat",
		zap.String("filename", filename),
		zap.Int("bytes", len(imageData)),
		zap.String("url", resp.URL),
	)
	return resp.URL, nil
}

// ----- UploadImage (permanent material, used for thumbnails) -----

type addMaterialResp struct {
	MediaID string `json:"media_id"`
	URL     string `json:"url"`
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// UploadImage uploads an image as permanent material via /material/add_material?type=image
// and returns the media_id. This consumes the permanent-material quota and is suitable
// for cover/thumbnail images referenced by draft.thumb_media_id.
func (p *RealWeChatPublisher) UploadImage(ctx context.Context, accessToken string, imageData []byte, filename string) (string, error) {
	if accessToken == "" {
		return "", errors.New("UploadImage: access token is required")
	}
	if len(imageData) == 0 {
		return "", errors.New("UploadImage: image data is empty")
	}
	mediaID, _, err := p.addMaterial(ctx, accessToken, imageData, filename)
	return mediaID, err
}

func (p *RealWeChatPublisher) addMaterial(ctx context.Context, accessToken string, imageData []byte, filename string) (string, string, error) {
	endpoint := endpointAddMaterial +
		"?access_token=" + url.QueryEscape(accessToken) +
		"&type=image"
	body, contentType, err := buildMediaMultipart("media", filename, imageData)
	if err != nil {
		return "", "", fmt.Errorf("addMaterial: build multipart: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return "", "", fmt.Errorf("addMaterial: build request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	var resp addMaterialResp
	if err := p.doJSON(req, &resp); err != nil {
		return "", "", fmt.Errorf("addMaterial: %w", err)
	}
	if resp.Errcode != 0 {
		return "", "", &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "add_material"}
	}
	if resp.MediaID == "" {
		return "", "", errors.New("addMaterial: empty media_id in response")
	}
	p.logger.Info("uploaded permanent image material to WeChat",
		zap.String("filename", filename),
		zap.Int("bytes", len(imageData)),
		zap.String("media_id", resp.MediaID),
	)
	return resp.MediaID, resp.URL, nil
}

// ----- CreateDraft -----

type draftAddArticle struct {
	Title            string `json:"title"`
	Author           string `json:"author,omitempty"`
	Digest           string `json:"digest,omitempty"`
	Content          string `json:"content"`
	ContentSourceURL string `json:"content_source_url,omitempty"`
	ThumbMediaID     string `json:"thumb_media_id"`
	NeedOpenComment  int    `json:"need_open_comment,omitempty"`
	OnlyFansCanComment int  `json:"only_fans_can_comment,omitempty"`
}

type draftAddReq struct {
	Articles []draftAddArticle `json:"articles"`
}

type draftAddResp struct {
	MediaID string `json:"media_id"`
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// CreateDraft creates a draft article via /draft/add. If article.ThumbURL is set, it is
// downloaded and uploaded as permanent material first to obtain thumb_media_id. WeChat
// requires thumb_media_id for all draft articles, so callers must supply ThumbURL.
func (p *RealWeChatPublisher) CreateDraft(ctx context.Context, accessToken string, article adapter.WeChatArticle) (string, error) {
	if accessToken == "" {
		return "", errors.New("CreateDraft: access token is required")
	}
	if article.Title == "" {
		return "", errors.New("CreateDraft: article title is required")
	}
	if article.Content == "" {
		return "", errors.New("CreateDraft: article content is required")
	}

	thumbMediaID, err := p.resolveThumbMediaID(ctx, accessToken, article.ThumbURL)
	if err != nil {
		return "", fmt.Errorf("CreateDraft: resolve thumb: %w", err)
	}
	if thumbMediaID == "" {
		return "", errors.New("CreateDraft: thumb_media_id required (set article.ThumbURL to a fetchable image)")
	}

	body, err := json.Marshal(draftAddReq{
		Articles: []draftAddArticle{{
			Title:            article.Title,
			Author:           article.Author,
			Digest:           article.Digest,
			Content:          article.Content,
			ContentSourceURL: article.SourceURL,
			ThumbMediaID:     thumbMediaID,
		}},
	})
	if err != nil {
		return "", fmt.Errorf("CreateDraft: marshal request: %w", err)
	}

	endpoint := endpointDraftAdd + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("CreateDraft: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var resp draftAddResp
	if err := p.doJSON(req, &resp); err != nil {
		return "", fmt.Errorf("CreateDraft: %w", err)
	}
	if resp.Errcode != 0 {
		return "", &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "draft/add"}
	}
	if resp.MediaID == "" {
		return "", errors.New("CreateDraft: empty media_id in response")
	}
	p.logger.Info("created WeChat draft",
		zap.String("title", article.Title),
		zap.String("media_id", resp.MediaID),
	)
	return resp.MediaID, nil
}

// resolveThumbMediaID downloads thumbURL and uploads it as permanent material,
// returning the resulting media_id. Returns "" if thumbURL is empty.
func (p *RealWeChatPublisher) resolveThumbMediaID(ctx context.Context, accessToken, thumbURL string) (string, error) {
	if strings.TrimSpace(thumbURL) == "" {
		return "", nil
	}
	imageData, filename, err := p.fetchImage(ctx, thumbURL)
	if err != nil {
		return "", fmt.Errorf("fetch thumb: %w", err)
	}
	mediaID, _, err := p.addMaterial(ctx, accessToken, imageData, filename)
	if err != nil {
		return "", err
	}
	return mediaID, nil
}

// fetchImage downloads an image from the given URL. Returns bytes and a derived filename.
func (p *RealWeChatPublisher) fetchImage(ctx context.Context, imgURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build request: %w", err)
	}
	res, err := p.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("http: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, "", fmt.Errorf("http %d for %s", res.StatusCode, imgURL)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}
	filename := filenameFromURL(imgURL, res.Header.Get("Content-Type"))
	return data, filename, nil
}

func filenameFromURL(rawURL, contentType string) string {
	ext := ".jpg"
	switch {
	case strings.Contains(contentType, "png"):
		ext = ".png"
	case strings.Contains(contentType, "gif"):
		ext = ".gif"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		ext = ".jpg"
	}
	if u, err := url.Parse(rawURL); err == nil {
		base := u.Path
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		if strings.Contains(base, ".") {
			return base
		}
		if base != "" {
			return base + ext
		}
	}
	return fmt.Sprintf("thumb_%d%s", time.Now().UnixNano(), ext)
}

// ----- Publish -----

type freepublishReq struct {
	MediaID string `json:"media_id"`
}

type freepublishResp struct {
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
	PublishID  string `json:"publish_id"`
	MsgDataID  string `json:"msg_data_id"`
}

// Publish submits a draft for free publication via /freepublish/submit. WeChat treats
// publication as asynchronous — the returned PublishID identifies the job; final article
// URLs become available via /freepublish/get afterwards.
func (p *RealWeChatPublisher) Publish(ctx context.Context, accessToken string, mediaID string) (*adapter.WeChatPublishResult, error) {
	if accessToken == "" {
		return nil, errors.New("Publish: access token is required")
	}
	if mediaID == "" {
		return nil, errors.New("Publish: media_id is required")
	}
	body, err := json.Marshal(freepublishReq{MediaID: mediaID})
	if err != nil {
		return nil, fmt.Errorf("Publish: marshal: %w", err)
	}
	endpoint := endpointFreepublish + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Publish: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var resp freepublishResp
	if err := p.doJSON(req, &resp); err != nil {
		return nil, fmt.Errorf("Publish: %w", err)
	}
	if resp.Errcode != 0 {
		return nil, &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "freepublish/submit"}
	}
	p.logger.Info("submitted WeChat free-publish",
		zap.String("media_id", mediaID),
		zap.String("publish_id", resp.PublishID),
	)
	return &adapter.WeChatPublishResult{
		MediaID:   mediaID,
		MsgID:     resp.MsgDataID,
		PublishID: resp.PublishID,
	}, nil
}

// ----- Helpers -----

// buildMediaMultipart builds a multipart/form-data body with a single "media" file part.
// WeChat's media upload endpoints look at the Content-Disposition filename to determine
// extension, so we set that explicitly with a content-type derived from extension.
func buildMediaMultipart(fieldname, filename string, data []byte) (io.Reader, string, error) {
	if filename == "" {
		filename = fmt.Sprintf("upload_%d.jpg", time.Now().UnixNano())
	}
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, fieldname, filename))
	hdr.Set("Content-Type", contentTypeFromFilename(filename))

	part, err := w.CreatePart(hdr)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf, w.FormDataContentType(), nil
}

func contentTypeFromFilename(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// doJSON executes an HTTP request and decodes the JSON body into v.
func (p *RealWeChatPublisher) doJSON(req *http.Request, v any) error {
	res, err := p.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("decode json: %w (body=%s)", err, strings.TrimSpace(string(raw)))
	}
	return nil
}

// ----- Stub (kept for tests / fallback) -----

// StubWeChatPublisher is a placeholder implementation of adapter.WeChatPublisher
// for development and testing. Do NOT use in production.
type StubWeChatPublisher struct {
	logger *zap.Logger
}

// NewStubWeChatPublisher creates a new stub WeChat publisher.
func NewStubWeChatPublisher(logger *zap.Logger) *StubWeChatPublisher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StubWeChatPublisher{logger: logger}
}

func (p *StubWeChatPublisher) UploadImage(_ context.Context, accessToken string, imageData []byte, filename string) (string, error) {
	if accessToken == "" {
		return "", errors.New("StubWeChatPublisher.UploadImage: access token is required")
	}
	if len(imageData) == 0 {
		return "", errors.New("StubWeChatPublisher.UploadImage: image data is empty")
	}
	mediaID := fmt.Sprintf("stub_media_%s_%d", filename, time.Now().UnixMilli())
	p.logger.Info("stub: uploaded image", zap.String("filename", filename), zap.String("media_id", mediaID))
	return mediaID, nil
}

func (p *StubWeChatPublisher) UploadContentImage(_ context.Context, accessToken string, imageData []byte, filename string) (string, error) {
	if accessToken == "" {
		return "", errors.New("StubWeChatPublisher.UploadContentImage: access token is required")
	}
	if len(imageData) == 0 {
		return "", errors.New("StubWeChatPublisher.UploadContentImage: image data is empty")
	}
	if len(imageData) > adapter.ContentImageMaxBytes {
		return "", fmt.Errorf("StubWeChatPublisher.UploadContentImage: image exceeds 1MB limit (%d bytes)", len(imageData))
	}
	wechatURL := fmt.Sprintf("https://mmbiz.qpic.cn/stub/%s_%d.png", filename, time.Now().UnixMilli())
	p.logger.Info("stub: uploaded content image", zap.String("filename", filename), zap.String("url", wechatURL))
	return wechatURL, nil
}

func (p *StubWeChatPublisher) CreateDraft(_ context.Context, accessToken string, article adapter.WeChatArticle) (string, error) {
	if accessToken == "" {
		return "", errors.New("StubWeChatPublisher.CreateDraft: access token is required")
	}
	if article.Title == "" {
		return "", errors.New("StubWeChatPublisher.CreateDraft: article title is required")
	}
	mediaID := fmt.Sprintf("stub_draft_%d", time.Now().UnixMilli())
	p.logger.Info("stub: created draft", zap.String("title", article.Title), zap.String("media_id", mediaID))
	return mediaID, nil
}

func (p *StubWeChatPublisher) Publish(_ context.Context, accessToken string, mediaID string) (*adapter.WeChatPublishResult, error) {
	if accessToken == "" {
		return nil, errors.New("StubWeChatPublisher.Publish: access token is required")
	}
	if mediaID == "" {
		return nil, errors.New("StubWeChatPublisher.Publish: media ID is required")
	}
	return &adapter.WeChatPublishResult{
		MediaID:   mediaID,
		MsgID:     fmt.Sprintf("stub_msg_%d", time.Now().UnixMilli()),
		PublishID: fmt.Sprintf("stub_pub_%d", time.Now().UnixMilli()),
	}, nil
}

// Compile-time interface checks.
var (
	_ adapter.WeChatPublisher = (*RealWeChatPublisher)(nil)
	_ adapter.WeChatPublisher = (*StubWeChatPublisher)(nil)
)
