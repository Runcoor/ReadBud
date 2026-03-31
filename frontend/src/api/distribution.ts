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
