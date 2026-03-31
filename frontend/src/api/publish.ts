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
