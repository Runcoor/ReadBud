// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

/** Provider type options */
export type ProviderType = 'llm' | 'image_search' | 'image_gen' | 'search' | 'crawler'

/** Provider type label map */
export const PROVIDER_TYPE_LABELS: Record<ProviderType, string> = {
  llm: '大语言模型',
  image_search: '图片搜索',
  image_gen: '图片生成',
  search: '内容搜索',
  crawler: '网页抓取',
}

/** Provider config view object */
export interface ProviderConfigVO {
  id: string
  provider_type: ProviderType
  provider_name: string
  config_json: Record<string, unknown>
  has_secret: boolean
  status: number
  is_default: boolean
}

/** Provider config request */
export interface ProviderConfigRequest {
  provider_type: ProviderType
  provider_name: string
  config_json: Record<string, unknown>
  secret_json?: string
  status?: number
}

/** Token mode options */
export type TokenMode = 'direct' | 'stable' | 'gateway_v2'

export const TOKEN_MODE_LABELS: Record<TokenMode, string> = {
  direct: '直接获取',
  stable: '稳定令牌',
  gateway_v2: '网关V2模式',
}

/** Delivery mode — how articles reach the WeChat editor. */
export type DeliveryMode = 'api' | 'extension' | 'manual'

export const DELIVERY_MODE_LABELS: Record<DeliveryMode, string> = {
  api: '直连发布',
  extension: '插件填充',
  manual: '手动复制',
}

export const DELIVERY_MODE_HINTS: Record<DeliveryMode, string> = {
  api: '通过 WeChat draft/add API 自动发布。需要已认证的服务号 + AppSecret。',
  extension: '安装浏览器插件后，自动跳转 WeChat 编辑器并填入标题/正文/封面。个人号也可用。',
  manual: '系统准备好内容，由你复制粘贴到 WeChat 编辑器。最保守，永远可用。',
}

/** WeChat account view object */
export interface WechatAccountVO {
  id: string
  name: string
  app_id: string
  token_mode: TokenMode
  delivery_mode: DeliveryMode
  is_default: boolean
  status: number
  remark: string
}

/** WeChat account request */
export interface WechatAccountRequest {
  name: string
  app_id: string
  app_secret?: string
  token_mode: TokenMode
  delivery_mode?: DeliveryMode
  is_default: boolean
  remark: string
}

/** Extension token (browser plugin auth) */
export interface ExtensionTokenVO {
  id: string
  name: string
  token_prefix: string
  last_used_at?: string
  expires_at?: string
  revoked_at?: string
  created_at: string
}

export interface ExtensionTokenIssueResponse {
  token: string
  info: ExtensionTokenVO
}
