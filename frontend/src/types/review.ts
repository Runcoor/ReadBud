// Review rule type definitions

export type RuleType = 'keyword_blacklist' | 'pattern_match' | 'content_policy'
export type RiskLevel = 'low' | 'medium' | 'high'

export interface ReviewRuleVO {
  id: string
  rule_type: RuleType
  rule_content: string
  risk_level: RiskLevel
  is_enabled: number
  created_at: string
  updated_at: string
}

export interface CreateReviewRuleRequest {
  rule_type: RuleType
  rule_content: string
  risk_level: RiskLevel
  is_enabled?: number
}

export interface UpdateReviewRuleRequest {
  rule_type?: RuleType
  rule_content?: string
  risk_level?: RiskLevel
}

export interface RuleViolation {
  rule_id: string
  rule_type: string
  risk_level: string
  detail: string
}

export const RULE_TYPE_LABELS: Record<RuleType, string> = {
  keyword_blacklist: '关键词黑名单',
  pattern_match: '正则匹配',
  content_policy: '内容策略',
}

export const RISK_LEVEL_LABELS: Record<RiskLevel, string> = {
  low: '低风险',
  medium: '中风险',
  high: '高风险',
}
