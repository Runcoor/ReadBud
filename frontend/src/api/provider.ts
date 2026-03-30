import { get, post } from './request'
import type { ApiResponse } from '@/types/api'
import type { ProviderConfigVO, ProviderConfigRequest, WechatAccountVO, WechatAccountRequest } from '@/types/provider'

/** GET /api/v1/providers */
export function listProviders(): Promise<ApiResponse<ProviderConfigVO[]>> {
  return get<ApiResponse<ProviderConfigVO[]>>('/providers')
}

/** POST /api/v1/providers */
export function createProvider(data: ProviderConfigRequest): Promise<ApiResponse<ProviderConfigVO>> {
  return post<ApiResponse<ProviderConfigVO>>('/providers', data)
}

/** GET /api/v1/wechat/accounts */
export function listWechatAccounts(): Promise<ApiResponse<WechatAccountVO[]>> {
  return get<ApiResponse<WechatAccountVO[]>>('/wechat/accounts')
}

/** POST /api/v1/wechat/accounts */
export function createWechatAccount(data: WechatAccountRequest): Promise<ApiResponse<WechatAccountVO>> {
  return post<ApiResponse<WechatAccountVO>>('/wechat/accounts', data)
}
