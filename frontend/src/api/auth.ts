import { post } from './request'
import type { ApiResponse, LoginRequest, LoginResponse } from '@/types/api'

/** POST /api/v1/auth/login */
export function login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
  return post<ApiResponse<LoginResponse>>('/auth/login', data)
}

/** POST /api/v1/auth/refresh */
export function refreshToken(): Promise<ApiResponse<{ token: string }>> {
  return post<ApiResponse<{ token: string }>>('/auth/refresh')
}
