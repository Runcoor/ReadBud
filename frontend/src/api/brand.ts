// Brand profile and style profile API client

import { get, post, patch, del } from './request'
import type { ApiResponse } from '@/types/api'
import type {
  BrandProfileVO,
  CreateBrandProfileRequest,
  UpdateBrandProfileRequest,
  StyleProfileVO,
  CreateStyleProfileRequest,
  UpdateStyleProfileRequest,
} from '@/types/brand'

// Brand profiles
export function listBrandProfiles(): Promise<ApiResponse<BrandProfileVO[]>> {
  return get('/brand-profiles')
}

export function getBrandProfile(id: string): Promise<ApiResponse<BrandProfileVO>> {
  return get(`/brand-profiles/${id}`)
}

export function createBrandProfile(data: CreateBrandProfileRequest): Promise<ApiResponse<BrandProfileVO>> {
  return post('/brand-profiles', data)
}

export function updateBrandProfile(id: string, data: UpdateBrandProfileRequest): Promise<ApiResponse<BrandProfileVO>> {
  return patch(`/brand-profiles/${id}`, data)
}

export function deleteBrandProfile(id: string): Promise<ApiResponse<null>> {
  return del(`/brand-profiles/${id}`)
}

// Style profiles
export function listStyleProfiles(): Promise<ApiResponse<StyleProfileVO[]>> {
  return get('/style-profiles')
}

export function getStyleProfile(id: string): Promise<ApiResponse<StyleProfileVO>> {
  return get(`/style-profiles/${id}`)
}

export function createStyleProfile(data: CreateStyleProfileRequest): Promise<ApiResponse<StyleProfileVO>> {
  return post('/style-profiles', data)
}

export function updateStyleProfile(id: string, data: UpdateStyleProfileRequest): Promise<ApiResponse<StyleProfileVO>> {
  return patch(`/style-profiles/${id}`, data)
}

export function deleteStyleProfile(id: string): Promise<ApiResponse<null>> {
  return del(`/style-profiles/${id}`)
}
