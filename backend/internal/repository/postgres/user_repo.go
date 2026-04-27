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

// UserRepository defines the interface for user data access.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepo.FindByUsername: %w", err)
	}
	return &user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepo.FindByID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) FindByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepo.FindByPublicID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("userRepo.Create: %w", err)
	}
	return nil
}
