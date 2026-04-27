// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Review rule API client

import { get, post, del, patch } from './request'
import type { ApiResponse } from '@/types/api'
import type {
  ReviewRuleVO,
  CreateReviewRuleRequest,
  UpdateReviewRuleRequest,
  RuleViolation,
} from '@/types/review'

export function listReviewRules(): Promise<ApiResponse<ReviewRuleVO[]>> {
  return get('/review-rules')
}

export function getReviewRule(id: string): Promise<ApiResponse<ReviewRuleVO>> {
  return get(`/review-rules/${id}`)
}

export function createReviewRule(data: CreateReviewRuleRequest): Promise<ApiResponse<ReviewRuleVO>> {
  return post('/review-rules', data)
}

export function updateReviewRule(id: string, data: UpdateReviewRuleRequest): Promise<ApiResponse<ReviewRuleVO>> {
  return patch(`/review-rules/${id}`, data)
}

export function deleteReviewRule(id: string): Promise<ApiResponse<null>> {
  return del(`/review-rules/${id}`)
}

export function toggleReviewRule(id: string, isEnabled: number): Promise<ApiResponse<ReviewRuleVO>> {
  return post(`/review-rules/${id}/toggle`, { is_enabled: isEnabled })
}

export function evaluateContent(content: string): Promise<ApiResponse<RuleViolation[]>> {
  return post('/review-rules/evaluate', { content })
}
