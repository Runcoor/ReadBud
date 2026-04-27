// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
)

// WechatAccountRepository defines the interface for WeChat account data access.
type WechatAccountRepository interface {
	List(ctx context.Context) ([]domain.WechatAccount, error)
	FindByID(ctx context.Context, id int64) (*domain.WechatAccount, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.WechatAccount, error)
	FindByAppID(ctx context.Context, appID string) (*domain.WechatAccount, error)
	FindDefault(ctx context.Context) (*domain.WechatAccount, error)
	Create(ctx context.Context, acct *domain.WechatAccount) error
	Update(ctx context.Context, acct *domain.WechatAccount) error
	ClearDefault(ctx context.Context) error
}

type wechatAccountRepo struct {
	db *gorm.DB
}

// NewWechatAccountRepository creates a new PostgreSQL-backed WeChat account repository.
func NewWechatAccountRepository(db *gorm.DB) WechatAccountRepository {
	return &wechatAccountRepo{db: db}
}

func (r *wechatAccountRepo) List(ctx context.Context) ([]domain.WechatAccount, error) {
	var accounts []domain.WechatAccount
	if err := r.db.WithContext(ctx).Order("is_default DESC, name").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("wechatAccountRepo.List: %w", err)
	}
	return accounts, nil
}

func (r *wechatAccountRepo) FindByID(ctx context.Context, id int64) (*domain.WechatAccount, error) {
	var acct domain.WechatAccount
	if err := r.db.WithContext(ctx).First(&acct, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("wechatAccountRepo.FindByID: %w", err)
	}
	return &acct, nil
}

func (r *wechatAccountRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.WechatAccount, error) {
	var acct domain.WechatAccount
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&acct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("wechatAccountRepo.FindByPublicID: %w", err)
	}
	return &acct, nil
}

func (r *wechatAccountRepo) FindByAppID(ctx context.Context, appID string) (*domain.WechatAccount, error) {
	var acct domain.WechatAccount
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).First(&acct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("wechatAccountRepo.FindByAppID: %w", err)
	}
	return &acct, nil
}

func (r *wechatAccountRepo) FindDefault(ctx context.Context) (*domain.WechatAccount, error) {
	var acct domain.WechatAccount
	if err := r.db.WithContext(ctx).Where("is_default = 1 AND status = ?", domain.StatusActive).First(&acct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("wechatAccountRepo.FindDefault: %w", err)
	}
	return &acct, nil
}

func (r *wechatAccountRepo) Create(ctx context.Context, acct *domain.WechatAccount) error {
	if err := r.db.WithContext(ctx).Create(acct).Error; err != nil {
		return fmt.Errorf("wechatAccountRepo.Create: %w", err)
	}
	return nil
}

func (r *wechatAccountRepo) Update(ctx context.Context, acct *domain.WechatAccount) error {
	if err := r.db.WithContext(ctx).Save(acct).Error; err != nil {
		return fmt.Errorf("wechatAccountRepo.Update: %w", err)
	}
	return nil
}

func (r *wechatAccountRepo) ClearDefault(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Model(&domain.WechatAccount{}).
		Where("is_default = 1").
		Update("is_default", 0).Error; err != nil {
		return fmt.Errorf("wechatAccountRepo.ClearDefault: %w", err)
	}
	return nil
}
