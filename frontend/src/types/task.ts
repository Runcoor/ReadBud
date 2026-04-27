// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Task-related type definitions

export type TaskStatus =
  | 'pending'
  | 'running'
  | 'done'
  | 'failed'
  | 'cancelled'

export type TaskStage =
  | 'keyword_expand'
  | 'source_search'
  | 'content_crawl'
  | 'hot_score'
  | 'article_write'
  | 'image_match'
  | 'chart_gen'
  | 'html_compile'
  | 'review'
  | 'publish'

export type ImageMode = 'auto' | 'search_only' | 'generate_only'

export type PublishMode = 'manual' | 'now' | 'schedule'

export type ArticleStyle = 'minimal' | 'magazine' | 'stitch'

export interface CreateTaskRequest {
  keyword: string
  audience?: string
  tone?: string
  target_words?: number
  image_mode: ImageMode
  chart_mode?: number
  publish_mode: PublishMode
  publish_at?: string
  wechat_account_id?: string
  article_style?: ArticleStyle
  visual_enhance?: boolean
  brand_profile_id?: string
}

export interface TaskVO {
  id: string
  task_no: string
  keyword: string
  audience: string
  tone: string
  target_words: number
  image_mode: ImageMode
  chart_mode: number
  publish_mode: PublishMode
  publish_at?: string
  article_style: string
  visual_enhance: boolean
  brand_profile_id?: string
  status: TaskStatus
  progress: number
  current_stage: string
  error_message?: string
  result_draft_id?: string
  created_at: string
  updated_at: string
}

export interface TaskListResponse {
  items: TaskVO[]
  total: number
  page: number
  page_size: number
}

/** Pipeline stage definitions for UI rendering */
export interface StageDefinition {
  key: TaskStage
  label: string
  icon: string
}

export const TASK_STAGES: StageDefinition[] = [
  { key: 'keyword_expand', label: '关键词扩展', icon: 'Search' },
  { key: 'source_search', label: '素材搜索', icon: 'Collection' },
  { key: 'content_crawl', label: '内容采集', icon: 'Download' },
  { key: 'hot_score', label: '热度评分', icon: 'TrendCharts' },
  { key: 'article_write', label: '文章撰写', icon: 'Edit' },
  { key: 'image_match', label: '图片匹配', icon: 'Picture' },
  { key: 'chart_gen', label: '图表生成', icon: 'DataLine' },
  { key: 'html_compile', label: 'HTML编译', icon: 'Document' },
  { key: 'review', label: '审核检查', icon: 'CircleCheck' },
  { key: 'publish', label: '发布', icon: 'Promotion' },
] as const

/** Status display label map */
export const STATUS_LABELS: Record<TaskStatus, string> = {
  pending: '排队中',
  running: '执行中',
  done: '已完成',
  failed: '失败',
  cancelled: '已取消',
}

/** Status tag type for Element Plus */
export const STATUS_TAG_TYPES: Record<TaskStatus, '' | 'success' | 'warning' | 'danger' | 'info'> = {
  pending: 'info',
  running: '',
  done: 'success',
  failed: 'danger',
  cancelled: 'warning',
}

/** Image mode labels */
export const IMAGE_MODE_LABELS: Record<ImageMode, string> = {
  auto: '自动（搜索优先）',
  search_only: '仅搜索',
  generate_only: '仅生成',
}

/** Publish mode labels */
export const PUBLISH_MODE_LABELS: Record<PublishMode, string> = {
  manual: '手动发布',
  now: '立即发布',
  schedule: '定时发布',
}

export const ARTICLE_STYLE_LABELS: Record<ArticleStyle, string> = {
  minimal: '极简专业 · 黑白底 + 荧光黄高亮',
  magazine: '杂志编辑 · 米色纸张 + 报刊红',
  stitch: '暖橙手账 · 米色底 + 暖橙强调',
}

/** Detail blurb shown beneath each style option. */
export const ARTICLE_STYLE_DETAILS: Record<ArticleStyle, string> = {
  minimal: '衬线标题 + 等宽编号，技术、AI、产品、知识深度内容首选。',
  magazine: 'Bodoni 大字 + 报头报尾，品牌故事、人物专访、深度观点。',
  stitch: '居中标题 + 装饰短横，教程、生活方式、轻量科普。',
}
