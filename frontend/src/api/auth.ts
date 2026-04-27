// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { post } from './request'
import type { ApiResponse, LoginRequest, LoginResponse } from '@/types/api'

/** POST /api/v1/auth/login */
export function login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
  return post<ApiResponse<LoginResponse>>('/auth/login', data)
}

/** POST /api/v1/auth/refresh */
export function refreshToken(): Promise<ApiResponse<{ token: string }>> {
  return post<ApiResponse<{ token: string }>>('/auth/refresh')
}
