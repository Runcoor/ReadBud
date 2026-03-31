import { get, post, patch, del } from './request'
import type { ApiResponse } from '@/types/api'
import type {
  TopicVO,
  TopicListResponse,
  CreateTopicRequest,
  UpdateTopicRequest,
  PerformanceFeedback,
  TopicRecommendationsResponse,
} from '@/types/topic'

/** GET /api/v1/reports/topics — paginated topic list */
export function listTopics(
  page = 1,
  size = 20,
  sortBy = 'recommend_weight',
): Promise<ApiResponse<TopicListResponse>> {
  return get<ApiResponse<TopicListResponse>>('/reports/topics', {
    params: { page, size, sort_by: sortBy },
  })
}

/** POST /api/v1/reports/topics — create a new topic */
export function createTopic(
  data: CreateTopicRequest,
): Promise<ApiResponse<TopicVO>> {
  return post<ApiResponse<TopicVO>>('/reports/topics', data)
}

/** GET /api/v1/reports/topics/:id — single topic */
export function getTopic(publicId: string): Promise<ApiResponse<TopicVO>> {
  return get<ApiResponse<TopicVO>>(`/reports/topics/${publicId}`)
}

/** PATCH /api/v1/reports/topics/:id — update topic */
export function updateTopic(
  publicId: string,
  data: UpdateTopicRequest,
): Promise<ApiResponse<TopicVO>> {
  return patch<ApiResponse<TopicVO>>(`/reports/topics/${publicId}`, data)
}

/** DELETE /api/v1/reports/topics/:id — remove topic */
export function deleteTopic(
  publicId: string,
): Promise<ApiResponse<{ deleted: boolean }>> {
  return del<ApiResponse<{ deleted: boolean }>>(`/reports/topics/${publicId}`)
}

/** GET /api/v1/reports/topics/recommendations — top recommended topics */
export function getTopicRecommendations(
  limit = 10,
): Promise<ApiResponse<TopicRecommendationsResponse>> {
  return get<ApiResponse<TopicRecommendationsResponse>>(
    '/reports/topics/recommendations',
    { params: { limit } },
  )
}

/** GET /api/v1/reports/topics/search — search topics by keyword */
export function searchTopics(
  query: string,
  page = 1,
  size = 20,
): Promise<ApiResponse<TopicListResponse>> {
  return get<ApiResponse<TopicListResponse>>('/reports/topics/search', {
    params: { q: query, page, size },
  })
}

/** POST /api/v1/reports/topics/:id/performance — update performance score */
export function updateTopicPerformance(
  publicId: string,
  feedback: PerformanceFeedback,
): Promise<ApiResponse<{ updated: boolean }>> {
  return post<ApiResponse<{ updated: boolean }>>(
    `/reports/topics/${publicId}/performance`,
    feedback,
  )
}
