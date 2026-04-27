// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { get, patch, post } from './request'
import type { ApiResponse } from '@/types/api'
import type { DraftVO, BlockVO, SourceVO, CoverVO } from '@/types/draft'

export function getDraft(id: string): Promise<ApiResponse<DraftVO>> {
  return get<ApiResponse<DraftVO>>(`/drafts/${id}`)
}

export function getDraftCover(id: string): Promise<ApiResponse<CoverVO | null>> {
  return get<ApiResponse<CoverVO | null>>(`/drafts/${id}/cover`)
}

export function regenerateDraftCover(id: string): Promise<ApiResponse<CoverVO>> {
  return post<ApiResponse<CoverVO>>(`/drafts/${id}/cover/regenerate`)
}

export function updateDraft(
  id: string,
  data: { title?: string; subtitle?: string; digest?: string },
): Promise<ApiResponse<DraftVO>> {
  return patch<ApiResponse<DraftVO>>(`/drafts/${id}`, data)
}

export function updateBlock(
  draftId: string,
  blockId: string,
  data: { heading?: string; text_md?: string; html_fragment?: string },
): Promise<ApiResponse<BlockVO>> {
  return patch<ApiResponse<BlockVO>>(`/drafts/${draftId}/blocks/${blockId}`, data)
}

export function getTaskSources(taskId: string): Promise<ApiResponse<SourceVO[]>> {
  return get<ApiResponse<SourceVO[]>>(`/tasks/${taskId}/sources`)
}
