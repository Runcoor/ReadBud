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
  created_at: string
}

export function createPublishJob(data: CreatePublishJobRequest): Promise<ApiResponse<PublishJobVO>> {
  return post<ApiResponse<PublishJobVO>>('/publish/jobs', data)
}

export function getPublishJob(id: string): Promise<ApiResponse<PublishJobVO>> {
  return get<ApiResponse<PublishJobVO>>(`/publish/jobs/${id}`)
}

export function cancelPublishJob(id: string): Promise<ApiResponse<PublishJobVO>> {
  return post<ApiResponse<PublishJobVO>>(`/publish/jobs/${id}/cancel`)
}
