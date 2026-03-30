/** Provider type options */
export type ProviderType = 'llm' | 'image_search' | 'image_gen' | 'search' | 'storage' | 'crawler'

/** Provider type label map */
export const PROVIDER_TYPE_LABELS: Record<ProviderType, string> = {
  llm: '大语言模型',
  image_search: '图片搜索',
  image_gen: '图片生成',
  search: '内容搜索',
  storage: '对象存储',
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

/** WeChat account view object */
export interface WechatAccountVO {
  id: string
  name: string
  app_id: string
  token_mode: TokenMode
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
  is_default: boolean
  remark: string
}
