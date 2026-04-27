// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { get, post, del } from './request'
import type { ApiResponse } from '@/types/api'
import type { DistributionVO, GenerateDistributionRequest } from '@/types/distribution'

export function generateDistribution(
  data: GenerateDistributionRequest,
): Promise<ApiResponse<DistributionVO>> {
  return post<ApiResponse<DistributionVO>>('/distributions/generate', data)
}

export function getDistributionByDraft(
  draftPublicId: string,
): Promise<ApiResponse<DistributionVO>> {
  return get<ApiResponse<DistributionVO>>(`/distributions/by-draft/${draftPublicId}`)
}

export function getDistribution(
  publicId: string,
): Promise<ApiResponse<DistributionVO>> {
  return get<ApiResponse<DistributionVO>>(`/distributions/${publicId}`)
}

export function deleteDistribution(
  publicId: string,
): Promise<ApiResponse<{ deleted: boolean }>> {
  return del<ApiResponse<{ deleted: boolean }>>(`/distributions/${publicId}`)
}
