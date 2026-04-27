// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

/** Topic library view object returned by GET /api/v1/reports/topics */
export interface TopicVO {
  public_id: string
  keyword: string
  audience: string
  article_goal: string
  historical_score: number
  recommend_weight: number
  last_used_at: string | null
  created_at: string
  updated_at: string
}

/** Paginated topic list response */
export interface TopicListResponse {
  items: TopicVO[]
  total: number
  page: number
  size: number
}

/** Request body for POST /api/v1/reports/topics */
export interface CreateTopicRequest {
  keyword: string
  audience?: string
  article_goal?: string
}

/** Request body for PATCH /api/v1/reports/topics/:id */
export interface UpdateTopicRequest {
  keyword?: string
  audience?: string
  article_goal?: string
}

/** Request body for POST /api/v1/reports/topics/:id/performance */
export interface PerformanceFeedback {
  read_count: number
  share_count: number
  fans_gained: number
}

/** Response for GET /api/v1/reports/topics/recommendations */
export interface TopicRecommendationsResponse {
  items: TopicVO[]
  count: number
}
