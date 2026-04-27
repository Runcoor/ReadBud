// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { del, get, post } from './request'
import type { ApiResponse } from '@/types/api'
import type {
  ExtensionTokenVO,
  ExtensionTokenIssueResponse,
} from '@/types/provider'

/** GET /api/v1/extension-tokens */
export function listExtensionTokens(): Promise<ApiResponse<ExtensionTokenVO[]>> {
  return get<ApiResponse<ExtensionTokenVO[]>>('/extension-tokens')
}

/** POST /api/v1/extension-tokens */
export function issueExtensionToken(payload: {
  name?: string
  ttl_hours?: number
}): Promise<ApiResponse<ExtensionTokenIssueResponse>> {
  return post<ApiResponse<ExtensionTokenIssueResponse>>('/extension-tokens', payload)
}

/** DELETE /api/v1/extension-tokens/:id */
export function revokeExtensionToken(id: string): Promise<unknown> {
  return del(`/extension-tokens/${id}`)
}
