// Metrics types for article performance tracking

export interface MetricsSnapshotVO {
  article_id: string
  metric_date: string
  read_count: number | null
  read_user_count: number | null
  share_count: number | null
  share_user_count: number | null
  add_fans_count: number | null
  cancel_fans_count: number | null
  net_fans_count: number | null
}

export interface ArticleMetricsResponse {
  article_id: string
  start: string
  end: string
  snapshots: MetricsSnapshotVO[]
}

export interface MetricsOverviewVO {
  total_articles: number
  total_reads: number
  total_shares: number
  total_fans_added: number
}

export interface MetricsSyncResult {
  articles_synced: number
  fans_synced: boolean
  errors?: string[]
}

// KPI display configuration
export interface KpiItem {
  key: string
  label: string
  value: number
  unit: string
  trend?: 'up' | 'down' | 'flat'
  trendValue?: number
}

// Date range presets for metrics queries
export const DATE_RANGE_PRESETS: Record<string, { label: string; days: number }> = {
  '7d': { label: '近 7 天', days: 7 },
  '14d': { label: '近 14 天', days: 14 },
  '30d': { label: '近 30 天', days: 30 },
}
