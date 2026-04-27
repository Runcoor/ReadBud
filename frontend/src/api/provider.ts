// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

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
