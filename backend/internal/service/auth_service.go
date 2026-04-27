// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"context"
	"errors"
	"fmt"

	"readbud/internal/api/dto"
	"readbud/internal/domain"
	"readbud/internal/pkg/crypto"
	"readbud/internal/repository/postgres"
)

// Common service errors.
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  postgres.UserRepository
	jwtConfig crypto.JWTConfig
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo postgres.UserRepository, jwtConfig crypto.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// Login validates credentials and returns a JWT token with user info.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != domain.StatusActive {
		return nil, ErrUserInactive
	}

	if err := crypto.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := crypto.GenerateToken(s.jwtConfig, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: generate token: %w", err)
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserVO{
			ID:       user.PublicID,
			Username: user.Username,
			Nickname: user.Nickname,
			Role:     user.Role,
		},
	}, nil
}

// RefreshToken generates a new token from existing valid claims.
func (s *AuthService) RefreshToken(ctx context.Context, claims *crypto.Claims) (*dto.RefreshTokenResponse, error) {
	// Verify user still exists and is active
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Status != domain.StatusActive {
		return nil, ErrUserInactive
	}

	token, err := crypto.RefreshToken(s.jwtConfig, claims)
	if err != nil {
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}

	return &dto.RefreshTokenResponse{Token: token}, nil
}
