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

export type ArticleStyle = 'minimal' | 'magazine' | 'listicle' | 'narrative' | 'faq' | 'casual'

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
  minimal: '极简专业',
  magazine: '杂志编辑',
  listicle: '清单干货',
  narrative: '叙事故事',
  faq: '问答拆解',
  casual: '轻社交',
}
