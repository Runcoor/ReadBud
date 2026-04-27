// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/repository/postgres"
)

// WechatPackageVO is the bundle the browser extension fetches before auto-filling
// the WeChat editor. Everything the plugin needs to populate the form is here:
//
//   - Text fields (title/author/digest/source_url) go straight into <input> elements.
//   - content_html is the styled article body (already image-substituted, ready to paste
//     into the WeChat rich-text editor as HTML).
//   - The cover comes both as a URL (for preview) AND as an inline base64 data URL
//     (so the extension can construct a File object to drop into the cover-image
//     <input type="file"> on mp.weixin.qq.com without dealing with CORS).
//
// The base64 payload is intentionally inlined rather than served as a separate
// endpoint — keeps the extension's network surface minimal (one fetch covers
// everything) and dodges any third-party-cookie / CORS friction the plugin
// might otherwise hit when downloading from a different origin.
type WechatPackageVO struct {
	DraftID        string  `json:"draft_id"`
	Title          string  `json:"title"`
	Subtitle       *string `json:"subtitle,omitempty"`
	Author         string  `json:"author"`
	Digest         string  `json:"digest"`
	SourceURL      *string `json:"source_url,omitempty"`
	ContentHTML    string  `json:"content_html"`
	CoverURL       string  `json:"cover_url,omitempty"`
	CoverBase64    string  `json:"cover_base64,omitempty"`
	CoverMimeType  string  `json:"cover_mime_type,omitempty"`
	CoverFilename  string  `json:"cover_filename,omitempty"`
	GeneratedAt    string  `json:"generated_at,omitempty"`
}

// WechatPackageService assembles the data the browser extension needs.
//
// It deliberately does NOT auto-generate a missing cover here — the Settings
// page surfaces that as a separate "封面缺失,先去生成一张" affordance so the
// user retains control. If a cover has been generated previously it is included
// with full bytes (base64) so the extension can re-upload it programmatically.
type WechatPackageService struct {
	draftRepo postgres.ArticleDraftRepository
	assetRepo postgres.AssetRepository
	storage   adapter.StorageProvider
	logger    *zap.Logger
}

// NewWechatPackageService builds the service.
func NewWechatPackageService(
	draftRepo postgres.ArticleDraftRepository,
	assetRepo postgres.AssetRepository,
	storage adapter.StorageProvider,
	logger *zap.Logger,
) *WechatPackageService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &WechatPackageService{
		draftRepo: draftRepo,
		assetRepo: assetRepo,
		storage:   storage,
		logger:    logger,
	}
}

// GetForDraft assembles a package for the given draft public ID.
func (s *WechatPackageService) GetForDraft(ctx context.Context, draftPublicID string) (*WechatPackageVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("wechatPackageService.GetForDraft: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	if d.CompiledHTML == "" {
		return nil, errors.New("草稿未生成排版 HTML, 请先在编辑页完成排版")
	}

	vo := &WechatPackageVO{
		DraftID:     d.PublicID,
		Title:       d.Title,
		Subtitle:    d.Subtitle,
		Author:      d.AuthorName,
		Digest:      d.Digest,
		SourceURL:   d.ContentSourceURL,
		ContentHTML: d.CompiledHTML,
		GeneratedAt: d.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}

	if d.CoverAssetID != nil && s.assetRepo != nil && s.storage != nil {
		a, err := s.assetRepo.FindByID(ctx, *d.CoverAssetID)
		if err == nil && a != nil {
			if url, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey); err == nil {
				vo.CoverURL = url
			}
			if data, err := s.storage.Download(ctx, a.Bucket, a.ObjectKey); err == nil && len(data) > 0 {
				vo.CoverBase64 = base64.StdEncoding.EncodeToString(data)
				vo.CoverMimeType = a.MimeType
				vo.CoverFilename = coverFilenameForMime(a.MimeType, d.PublicID)
			} else if err != nil {
				s.logger.Warn("download cover bytes failed",
					zap.String("draft_public_id", draftPublicID),
					zap.Int64("asset_id", a.ID),
					zap.Error(err),
				)
			}
		}
	}

	return vo, nil
}

func coverFilenameForMime(mime, draftID string) string {
	switch mime {
	case "image/png":
		return draftID + ".png"
	case "image/jpeg":
		return draftID + ".jpg"
	case "image/webp":
		return draftID + ".webp"
	default:
		return draftID + ".bin"
	}
}
