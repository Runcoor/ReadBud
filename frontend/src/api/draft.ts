import { get } from './request'
import type { ApiResponse } from '@/types/api'
import type { DraftVO, SourceVO } from '@/types/draft'

export function getDraft(id: string): Promise<ApiResponse<DraftVO>> {
  return get<ApiResponse<DraftVO>>(`/drafts/${id}`)
}

export function getTaskSources(taskId: string): Promise<ApiResponse<SourceVO[]>> {
  return get<ApiResponse<SourceVO[]>>(`/tasks/${taskId}/sources`)
}
