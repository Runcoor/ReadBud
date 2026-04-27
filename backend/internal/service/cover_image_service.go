// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain/asset"
	"readbud/internal/domain/draft"
	"readbud/internal/repository/postgres"
	"readbud/internal/service/imageresize"
)

const (
	coverBucket           = "covers"
	coverDownloadCap      = 20 * 1024 * 1024
	coverDownloadTimeout  = 60 * time.Second
	coverDefaultStyleName = "minimal"
	// Final cover dimensions stored and shipped to WeChat (16:9 thumbnail aspect).
	coverWidth  = 1024
	coverHeight = 576
	// Image-gen request dimensions. Most providers (DALL-E 3, gpt-image-1, and
	// the Chinese gateways that follow them) reject sub-1MP sizes, so we ask
	// for a square at the supported minimum and crop down to 16:9 ourselves.
	coverGenWidth  = 1024
	coverGenHeight = 1024
)

// CoverVO is the view object for an article cover image.
type CoverVO struct {
	AssetID       string  `json:"asset_id,omitempty"`
	URL           string  `json:"url"`
	Width         int     `json:"width,omitempty"`
	Height        int     `json:"height,omitempty"`
	IsAIGenerated bool    `json:"is_ai_generated"`
	Prompt        *string `json:"prompt,omitempty"`
}

// CoverImageService generates and manages article cover images.
//
// Flow for a regenerate request:
//  1. Build a style-aware editorial prompt from the draft's title/subtitle/digest/style.
//  2. Call imageGen.Generate to produce raw bytes (base64 / data URL / https URL all handled).
//  3. Persist via storage.Upload + create an Asset row of type cover_image.
//  4. Update draft.CoverAssetID to point at the new asset.
//
// Old covers are intentionally left in storage rather than deleted — they remain
// reachable via their public_id if anything still references them, and a separate
// GC sweep can clean orphans later.
type CoverImageService struct {
	draftRepo postgres.ArticleDraftRepository
	assetRepo postgres.AssetRepository
	imageGen  adapter.ImageGenProvider
	storage   adapter.StorageProvider
	http      *http.Client
	logger    *zap.Logger
}

// NewCoverImageService builds a CoverImageService.
func NewCoverImageService(
	draftRepo postgres.ArticleDraftRepository,
	assetRepo postgres.AssetRepository,
	imageGen adapter.ImageGenProvider,
	storage adapter.StorageProvider,
	logger *zap.Logger,
) *CoverImageService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CoverImageService{
		draftRepo: draftRepo,
		assetRepo: assetRepo,
		imageGen:  imageGen,
		storage:   storage,
		http:      &http.Client{Timeout: coverDownloadTimeout},
		logger:    logger,
	}
}

// ----- Public API -----

// GenerateForDraft generates a fresh cover for the draft and links it. Any previous
// cover (if any) is replaced — the old asset row is left in place, only CoverAssetID
// flips to point at the new one. Safe to call multiple times for retry.
func (s *CoverImageService) GenerateForDraft(ctx context.Context, draftPublicID string) (*CoverVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GenerateForDraft: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}

	prompt := buildCoverPrompt(d)
	gen, err := s.imageGen.Generate(ctx, prompt, adapter.ImageGenOptions{
		Width:  coverGenWidth,
		Height: coverGenHeight,
		Style:  d.StyleUsed,
	})
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GenerateForDraft: image gen: %w", err)
	}

	rawData, err := s.materializeBytes(ctx, gen)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GenerateForDraft: materialize: %w", err)
	}

	// Image-gen returns a square; center-crop to 16:9 to match the WeChat
	// thumbnail aspect. Output is JPEG so persistCoverAsset stores .jpg.
	data, _, err := imageresize.CropToSize(rawData, coverWidth, coverHeight)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GenerateForDraft: crop: %w", err)
	}

	a, publicURL, err := s.persistCoverAsset(ctx, data, prompt)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GenerateForDraft: persist: %w", err)
	}

	d.CoverAssetID = &a.ID
	if err := s.draftRepo.Update(ctx, d); err != nil {
		// The asset is already saved; surface error but the asset can still be reused
		// on a retry (FindBySHA256 path) — we just couldn't link it.
		return nil, fmt.Errorf("coverImageService.GenerateForDraft: update draft: %w", err)
	}

	s.logger.Info("cover image generated",
		zap.Int64("draft_id", d.ID),
		zap.String("draft_public_id", draftPublicID),
		zap.Int64("asset_id", a.ID),
		zap.String("style", d.StyleUsed),
	)

	return &CoverVO{
		AssetID:       a.PublicID,
		URL:           publicURL,
		Width:         gen.Width,
		Height:        gen.Height,
		IsAIGenerated: true,
		Prompt:        a.PromptText,
	}, nil
}

// GetForDraft returns the current cover for the draft, or nil if none is linked.
func (s *CoverImageService) GetForDraft(ctx context.Context, draftPublicID string) (*CoverVO, error) {
	d, err := s.draftRepo.FindByPublicID(ctx, draftPublicID)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GetForDraft: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	if d.CoverAssetID == nil {
		return nil, nil
	}
	a, err := s.assetRepo.FindByID(ctx, *d.CoverAssetID)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GetForDraft: load asset: %w", err)
	}
	if a == nil {
		return nil, nil
	}
	url, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("coverImageService.GetForDraft: storage url: %w", err)
	}
	width := 0
	height := 0
	if a.Width != nil {
		width = *a.Width
	}
	if a.Height != nil {
		height = *a.Height
	}
	return &CoverVO{
		AssetID:       a.PublicID,
		URL:           url,
		Width:         width,
		Height:        height,
		IsAIGenerated: a.IsAIGenerated == 1,
		Prompt:        a.PromptText,
	}, nil
}

// EnsureCover returns the existing cover URL or, if missing, generates one and returns it.
// Used by the publish pipeline to guarantee a thumb is available before calling WeChat.
func (s *CoverImageService) EnsureCover(ctx context.Context, d *draft.ArticleDraft) (string, error) {
	if d == nil {
		return "", errors.New("coverImageService.EnsureCover: nil draft")
	}
	if d.CoverAssetID != nil {
		a, err := s.assetRepo.FindByID(ctx, *d.CoverAssetID)
		if err != nil {
			return "", fmt.Errorf("coverImageService.EnsureCover: load asset: %w", err)
		}
		if a != nil {
			url, err := s.storage.GetURL(ctx, a.Bucket, a.ObjectKey)
			if err == nil && url != "" {
				return url, nil
			}
		}
	}
	vo, err := s.GenerateForDraft(ctx, d.PublicID)
	if err != nil {
		return "", err
	}
	return vo.URL, nil
}

// ----- Internals -----

func (s *CoverImageService) materializeBytes(ctx context.Context, gen *adapter.GeneratedImage) ([]byte, error) {
	if gen == nil {
		return nil, errors.New("nil gen result")
	}
	if gen.Base64 != "" {
		return decodeBase64Payload(gen.Base64)
	}
	if gen.URL == "" {
		return nil, errors.New("gen result has neither URL nor base64 payload")
	}
	if strings.HasPrefix(gen.URL, "data:") {
		return decodeBase64Payload(gen.URL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gen.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, coverDownloadCap))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}

func decodeBase64Payload(raw string) ([]byte, error) {
	if i := strings.Index(raw, ","); i >= 0 && strings.Contains(raw[:i], "base64") {
		raw = raw[i+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return decoded, nil
}

func (s *CoverImageService) persistCoverAsset(ctx context.Context, data []byte, prompt string) (*asset.Asset, string, error) {
	if len(data) == 0 {
		return nil, "", errors.New("empty image bytes")
	}
	mime := http.DetectContentType(data)
	ext := "png"
	switch mime {
	case "image/png":
		ext = "png"
	case "image/jpeg":
		ext = "jpg"
	case "image/webp":
		ext = "webp"
	default:
		return nil, "", fmt.Errorf("unsupported mime %q", mime)
	}

	now := time.Now().UTC()
	id := ulid.MustNew(ulid.Timestamp(now), crand.Reader).String()
	key := fmt.Sprintf("%04d%02d/%s.%s", now.Year(), now.Month(), id, ext)

	publicURL, err := s.storage.Upload(ctx, coverBucket, key, data, mime)
	if err != nil {
		return nil, "", fmt.Errorf("upload: %w", err)
	}

	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	sz := int64(len(data))
	promptCopy := prompt
	w := coverWidth
	h := coverHeight

	a := &asset.Asset{
		AssetType:          asset.AssetTypeCoverImage,
		SourceKind:         asset.SourceKindGenerated,
		MimeType:           mime,
		StorageProvider:    "local",
		Bucket:             coverBucket,
		ObjectKey:          key,
		Width:              &w,
		Height:             &h,
		SizeBytes:          &sz,
		SHA256:             sha,
		PromptText:         &promptCopy,
		IsAIGenerated:      1,
		WechatUploadStatus: asset.WechatUploadPending,
	}
	if err := s.assetRepo.Create(ctx, a); err != nil {
		_ = s.storage.Delete(context.Background(), coverBucket, key)
		return nil, "", fmt.Errorf("create asset: %w", err)
	}
	_ = bytes.TrimSpace // keep bytes import alive across future tweaks
	return a, publicURL, nil
}

// ----- Prompt building -----

// buildCoverPrompt produces a high-quality editorial prompt tuned for clean, sophisticated
// AI-generated covers. It bakes in:
//   - the article's title/subtitle/digest as semantic anchor (kept in Chinese — modern
//     diffusion models handle Chinese semantics, and translating loses nuance)
//   - a style-specific aesthetic palette matching the chosen preset (minimal/magazine/stitch)
//   - hard constraints (no text, no faces, abstract conceptual focus) that diffusion models
//     respect best when stated as positive design directives rather than negative prompts
func buildCoverPrompt(d *draft.ArticleDraft) string {
	style := strings.ToLower(strings.TrimSpace(d.StyleUsed))
	if style == "" {
		style = coverDefaultStyleName
	}
	aesthetic := stylePromptAesthetic(style)

	var sb strings.Builder
	sb.WriteString("Editorial magazine cover image for a Chinese WeChat article.\n\n")

	sb.WriteString("ARTICLE\n")
	sb.WriteString("Title: ")
	sb.WriteString(strings.TrimSpace(d.Title))
	sb.WriteString("\n")
	if d.Subtitle != nil && strings.TrimSpace(*d.Subtitle) != "" {
		sb.WriteString("Subtitle: ")
		sb.WriteString(strings.TrimSpace(*d.Subtitle))
		sb.WriteString("\n")
	}
	if strings.TrimSpace(d.Digest) != "" {
		sb.WriteString("Summary: ")
		sb.WriteString(strings.TrimSpace(d.Digest))
		sb.WriteString("\n")
	}

	sb.WriteString("\nDESIGN BRIEF\n")
	sb.WriteString("Render an abstract, conceptual, metaphorical visualization that captures the article's essence. ")
	sb.WriteString("Treat it like a premium magazine cover (Wallpaper, Kinfolk, Monocle) — never a literal illustration.\n")

	sb.WriteString("\nAESTHETIC\n")
	sb.WriteString(aesthetic)
	sb.WriteString("\n")

	sb.WriteString("\nCOMPOSITION\n")
	sb.WriteString("- 16:9 aspect ratio, optimized for small-thumbnail readability (must remain legible at 200×114 px)\n")
	sb.WriteString("- Single strong focal subject, generous negative space, calm and uncluttered\n")
	sb.WriteString("- Sophisticated lighting with depth and atmosphere, photographic clarity or refined illustration\n")
	sb.WriteString("- Editorial-grade craft — looks intentional, considered, restrained\n")

	sb.WriteString("\nSTRICT REQUIREMENTS\n")
	sb.WriteString("- Absolutely no text, no letters, no Chinese characters, no numbers, no captions, no logos, no watermarks, no signatures\n")
	sb.WriteString("- No human faces, no recognizable people, no crowds, no dense scenes\n")
	sb.WriteString("- No cartoon, no childish style, no clipart, no stock-photo cliché\n")
	sb.WriteString("- Abstract conceptual visualization only — objects, materials, light, geometry, atmosphere\n")
	sb.WriteString("- Modern, sophisticated, professional, calm, refined\n")

	return sb.String()
}

// stylePromptAesthetic returns the per-preset aesthetic instructions used by buildCoverPrompt.
// Palettes are chosen to match the corresponding StyleConfig in pipeline/compiler.go so the
// cover and the article body feel like they belong to the same publication.
func stylePromptAesthetic(style string) string {
	switch style {
	case "magazine":
		return "Vintage editorial print aesthetic. " +
			"Warm muted palette: paper cream (#F2EFE8), deep ink black, dusty red accent (#E63946) used very sparingly (max 5% of image), soft dusty rose secondary. " +
			"Subtle paper texture, slight halftone or grain, evocative of mid-century magazine print. " +
			"Composition reminiscent of Wallpaper magazine or Kinfolk — refined, atmospheric, editorial."
	case "stitch":
		return "Warm hand-crafted vintage aesthetic. " +
			"Cream and amber palette: warm cream base (#FCFAF5), terracotta / sienna accent (#D2691E used as a focal warm tone), rich warm shadows (#1A1815). " +
			"Soft natural lighting, subtle paper or textile texture, organic shapes, hand-finished feel. " +
			"Evocative of crafted leather goods, artisan workshops, warm autumn light."
	default: // "minimal" or unknown
		return "Minimalist Swiss-design aesthetic. " +
			"Strict monochromatic palette: pure ink black (#111111), bright clean white (#FFFFFF), with a single muted highlight color — warm yellow (#FFE94D) used very sparingly (max 5% of image). " +
			"Strong geometric forms, generous white space, high contrast, sans-serif sensibility. " +
			"Subtle gradients allowed; no decorative ornament. Calm, modern, considered."
	}
}
