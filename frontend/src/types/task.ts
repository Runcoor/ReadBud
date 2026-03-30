// Task-related type definitions

export type TaskStatus =
  | 'pending'
  | 'collecting'
  | 'analyzing'
  | 'writing'
  | 'asseting'
  | 'review_ready'
  | 'publishing'
  | 'published'
  | 'failed'

export type ImageMode = 'auto' | 'search_only' | 'generate_only'

export type PublishMode = 'manual' | 'now' | 'schedule'

export interface CreateTaskRequest {
  keyword: string
  audience: string
  tone: string
  target_words: number
  image_mode: ImageMode
  chart_mode: boolean
  publish_mode: PublishMode
  publish_at?: string
  wechat_account_id: string
}

export interface TaskVO {
  id: string
  task_no: string
  keyword: string
  audience: string
  tone: string
  target_words: number
  image_mode: ImageMode
  chart_mode: boolean
  publish_mode: PublishMode
  publish_at?: string
  status: TaskStatus
  progress: number
  current_stage: string
  error_message?: string
  result_draft_id?: string
  created_at: string
  updated_at: string
}

export const TASK_STAGES = [
  { key: 'collecting', label: '采集' },
  { key: 'dedup', label: '去重' },
  { key: 'analyzing', label: '爆文分析' },
  { key: 'outlining', label: '文章提纲' },
  { key: 'writing', label: '正文生成' },
  { key: 'image_matching', label: '图片匹配' },
  { key: 'chart_gen', label: '图表生成' },
  { key: 'html_compile', label: 'HTML编译' },
  { key: 'review', label: '审核检查' },
  { key: 'publishing', label: '发布' },
] as const
