import { get, patch } from './request'
import type { ApiResponse } from '@/types/api'
import type { DraftVO, BlockVO, SourceVO } from '@/types/draft'

export function getDraft(id: string): Promise<ApiResponse<DraftVO>> {
  return get<ApiResponse<DraftVO>>(`/drafts/${id}`)
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
