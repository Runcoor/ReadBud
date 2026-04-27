// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { post, get } from './request'
import type { ApiResponse } from '@/types/api'

export interface CreatePublishJobRequest {
  draft_id: string
  wechat_account_id: string
  publish_mode: 'now' | 'schedule' | 'manual'
  schedule_at?: string
}

export interface PublishJobVO {
  id: string
  draft_id: string
  publish_mode: string
  status: string
  retry_count: number
  last_error?: string
  article_url?: string
  /** How the article reaches WeChat. Set by the backend when the job is created
   *  based on the bound WeChat account's delivery_mode setting. */
  delivery_mode?: 'api' | 'extension' | 'manual'
  /** Deeplink to open the WeChat editor (extension/manual modes only). */
  editor_url?: string
  created_at: string
}

/** POST /api/v1/publish/jobs */
export function createPublishJob(data: CreatePublishJobRequest): Promise<ApiResponse<PublishJobVO>> {
  return post<ApiResponse<PublishJobVO>>('/publish/jobs', data)
}

/** GET /api/v1/publish/jobs/:id */
export function getPublishJob(id: string): Promise<ApiResponse<PublishJobVO>> {
  return get<ApiResponse<PublishJobVO>>(`/publish/jobs/${id}`)
}

/** POST /api/v1/publish/jobs/:id/cancel */
export function cancelPublishJob(id: string): Promise<ApiResponse<PublishJobVO>> {
  return post<ApiResponse<PublishJobVO>>(`/publish/jobs/${id}/cancel`)
}

/** POST /api/v1/publish/jobs/:id/retry */
export function retryPublishJob(id: string): Promise<ApiResponse<PublishJobVO>> {
  return post<ApiResponse<PublishJobVO>>(`/publish/jobs/${id}/retry`)
}

/** POST /api/v1/publish/jobs/:id/fulfilled —
 *  Used by the extension flow (or manual "已发布") to flip awaiting_extension/manual
 *  to success and optionally record the article URL. */
export function markPublishJobFulfilled(
  id: string,
  articleURL?: string,
): Promise<ApiResponse<PublishJobVO>> {
  return post<ApiResponse<PublishJobVO>>(`/publish/jobs/${id}/fulfilled`, articleURL ? { article_url: articleURL } : {})
}
