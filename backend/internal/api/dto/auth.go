// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package dto

// LoginRequest is the DTO for POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// LoginResponse is the VO returned after successful login.
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserVO  `json:"user"`
}

// UserVO is the view object for user information (never includes password).
type UserVO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// RefreshTokenResponse is the VO for token refresh.
type RefreshTokenResponse struct {
	Token string `json:"token"`
}
