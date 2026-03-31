import { get, post } from './request'
import type { ApiResponse } from '@/types/api'
import type { ArticleMetricsResponse, MetricsSyncResult, MetricsOverviewVO } from '@/types/metrics'

/** GET /api/v1/articles/:id/metrics */
export function getArticleMetrics(
  articleId: string,
  start?: string,
  end?: string,
): Promise<ApiResponse<ArticleMetricsResponse>> {
  const params: Record<string, string> = {}
  if (start) params.start = start
  if (end) params.end = end
  return get<ApiResponse<ArticleMetricsResponse>>(`/articles/${articleId}/metrics`, { params })
}

/** POST /api/v1/metrics/sync */
export function syncMetrics(wechatAccountId: string): Promise<ApiResponse<MetricsSyncResult>> {
  return post<ApiResponse<MetricsSyncResult>>('/metrics/sync', {
    wechat_account_id: wechatAccountId,
  })
}

/** GET /api/v1/reports/overview */
export function getReportsOverview(wechatAccountId: string): Promise<ApiResponse<MetricsOverviewVO>> {
  return get<ApiResponse<MetricsOverviewVO>>('/reports/overview', {
    params: { wechat_account_id: wechatAccountId },
  })
}
