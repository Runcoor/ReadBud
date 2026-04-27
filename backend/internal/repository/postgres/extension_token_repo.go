// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// ExtensionTokenRepository defines data access for browser-extension tokens.
type ExtensionTokenRepository interface {
	Create(ctx context.Context, t *domain.ExtensionToken) error
	FindByHash(ctx context.Context, hash string) (*domain.ExtensionToken, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.ExtensionToken, error)
	ListByUser(ctx context.Context, userID int64) ([]domain.ExtensionToken, error)
	Revoke(ctx context.Context, id int64, at time.Time) error
	UpdateLastUsed(ctx context.Context, id int64, at time.Time) error
}

type extensionTokenRepo struct {
	db *gorm.DB
}

// NewExtensionTokenRepository builds a PostgreSQL-backed repository.
func NewExtensionTokenRepository(db *gorm.DB) ExtensionTokenRepository {
	return &extensionTokenRepo{db: db}
}

func (r *extensionTokenRepo) Create(ctx context.Context, t *domain.ExtensionToken) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("extensionTokenRepo.Create: %w", err)
	}
	return nil
}

func (r *extensionTokenRepo) FindByHash(ctx context.Context, hash string) (*domain.ExtensionToken, error) {
	var t domain.ExtensionToken
	if err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL", hash).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("extensionTokenRepo.FindByHash: %w", err)
	}
	return &t, nil
}

func (r *extensionTokenRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.ExtensionToken, error) {
	var t domain.ExtensionToken
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("extensionTokenRepo.FindByPublicID: %w", err)
	}
	return &t, nil
}

func (r *extensionTokenRepo) ListByUser(ctx context.Context, userID int64) ([]domain.ExtensionToken, error) {
	var ts []domain.ExtensionToken
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&ts).Error; err != nil {
		return nil, fmt.Errorf("extensionTokenRepo.ListByUser: %w", err)
	}
	return ts, nil
}

func (r *extensionTokenRepo) Revoke(ctx context.Context, id int64, at time.Time) error {
	if err := r.db.WithContext(ctx).Model(&domain.ExtensionToken{}).
		Where("id = ?", id).
		Update("revoked_at", at).Error; err != nil {
		return fmt.Errorf("extensionTokenRepo.Revoke: %w", err)
	}
	return nil
}

func (r *extensionTokenRepo) UpdateLastUsed(ctx context.Context, id int64, at time.Time) error {
	if err := r.db.WithContext(ctx).Model(&domain.ExtensionToken{}).
		Where("id = ?", id).
		Update("last_used_at", at).Error; err != nil {
		return fmt.Errorf("extensionTokenRepo.UpdateLastUsed: %w", err)
	}
	return nil
}
