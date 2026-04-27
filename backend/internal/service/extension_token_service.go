// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"readbud/internal/api/dto"
	"readbud/internal/domain"
	"readbud/internal/repository/postgres"
)

// extensionTokenPrefix is prepended to every issued plaintext token. Mostly a
// human aid — when a user shows you a token starting with `rbex_…` you can
// immediately tell what it is and which system to look in.
const extensionTokenPrefix = "rbex_"

// ExtensionTokenService manages long-lived browser-extension credentials.
type ExtensionTokenService struct {
	repo postgres.ExtensionTokenRepository
}

// NewExtensionTokenService constructs the service.
func NewExtensionTokenService(repo postgres.ExtensionTokenRepository) *ExtensionTokenService {
	return &ExtensionTokenService{repo: repo}
}

// Issue creates a new extension token for the user and returns the plaintext
// value alongside the VO. The plaintext is shown to the user EXACTLY ONCE — the
// DB only ever holds the sha256 hash, so a leaked DB cannot be replayed.
func (s *ExtensionTokenService) Issue(ctx context.Context, userID int64, name string, ttl time.Duration) (string, *dto.ExtensionTokenVO, error) {
	if name == "" {
		name = "默认插件"
	}

	plaintext, err := generateExtensionTokenSecret()
	if err != nil {
		return "", nil, fmt.Errorf("extensionTokenService.Issue: secret: %w", err)
	}
	hash := hashExtensionToken(plaintext)
	prefix := plaintext[:min(12, len(plaintext))]

	t := &domain.ExtensionToken{
		UserID:      userID,
		Name:        name,
		TokenHash:   hash,
		TokenPrefix: prefix,
	}
	if ttl > 0 {
		exp := time.Now().UTC().Add(ttl)
		t.ExpiresAt = &exp
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return "", nil, fmt.Errorf("extensionTokenService.Issue: %w", err)
	}

	vo := toExtensionTokenVO(*t)
	return plaintext, &vo, nil
}

// List returns all extension tokens belonging to a user (revoked included so the
// user can see their full history; the UI greys revoked entries out).
func (s *ExtensionTokenService) List(ctx context.Context, userID int64) ([]dto.ExtensionTokenVO, error) {
	ts, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("extensionTokenService.List: %w", err)
	}
	vos := make([]dto.ExtensionTokenVO, 0, len(ts))
	for _, t := range ts {
		vos = append(vos, toExtensionTokenVO(t))
	}
	return vos, nil
}

// Revoke marks a token as revoked. Idempotent — calling on an already-revoked
// token is a no-op success.
func (s *ExtensionTokenService) Revoke(ctx context.Context, userID int64, publicID string) error {
	t, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("extensionTokenService.Revoke: %w", err)
	}
	if t == nil {
		return ErrNotFound
	}
	if t.UserID != userID {
		return ErrNotFound // hide existence from non-owners
	}
	if t.RevokedAt != nil {
		return nil
	}
	now := time.Now().UTC()
	if err := s.repo.Revoke(ctx, t.ID, now); err != nil {
		return fmt.Errorf("extensionTokenService.Revoke: %w", err)
	}
	return nil
}

// Authenticate validates a plaintext token; returns the user ID on success.
// Used by the extension auth middleware.
func (s *ExtensionTokenService) Authenticate(ctx context.Context, plaintext string) (int64, error) {
	if plaintext == "" {
		return 0, errors.New("empty token")
	}
	hash := hashExtensionToken(plaintext)
	t, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return 0, fmt.Errorf("extensionTokenService.Authenticate: %w", err)
	}
	if t == nil {
		return 0, errors.New("token not found")
	}
	if t.RevokedAt != nil {
		return 0, errors.New("token revoked")
	}
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now().UTC()) {
		return 0, errors.New("token expired")
	}

	// Best-effort last-used update — don't block auth on a write hiccup.
	_ = s.repo.UpdateLastUsed(ctx, t.ID, time.Now().UTC())
	return t.UserID, nil
}

// ----- helpers -----

func generateExtensionTokenSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := crand.Read(buf); err != nil {
		return "", err
	}
	return extensionTokenPrefix + hex.EncodeToString(buf), nil
}

func hashExtensionToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func toExtensionTokenVO(t domain.ExtensionToken) dto.ExtensionTokenVO {
	vo := dto.ExtensionTokenVO{
		ID:          t.PublicID,
		Name:        t.Name,
		TokenPrefix: t.TokenPrefix,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
	}
	if t.LastUsedAt != nil {
		s := t.LastUsedAt.Format(time.RFC3339)
		vo.LastUsedAt = &s
	}
	if t.ExpiresAt != nil {
		s := t.ExpiresAt.Format(time.RFC3339)
		vo.ExpiresAt = &s
	}
	if t.RevokedAt != nil {
		s := t.RevokedAt.Format(time.RFC3339)
		vo.RevokedAt = &s
	}
	return vo
}
