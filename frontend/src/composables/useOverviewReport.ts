// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { getReportsOverview } from '@/api/metrics'
import { listTopics, getTopicRecommendations } from '@/api/topic'
import type { MetricsOverviewVO, KpiItem } from '@/types/metrics'
import type { TopicVO } from '@/types/topic'

/** Analysis dimension item for chart display */
export interface AnalysisDimension {
  label: string
  description: string
  categories: string[]
  values: number[]
  color: string
}

export function useOverviewReport() {
  // --- State ---
  const loading = ref(false)
  const fetchError = ref<string | null>(null)
  const overview = ref<MetricsOverviewVO | null>(null)
  const topics = ref<TopicVO[]>([])
  const recommendations = ref<TopicVO[]>([])
  const wechatAccountId = ref<string>('')
  const hasData = ref(false)

  // --- Computed: 5 KPIs ---
  const kpis = computed<KpiItem[]>(() => {
    const data = overview.value
    if (!data) {
      return defaultKpis()
    }

    return [
      {
        key: 'total_reads',
        label: '阅读量',
        value: data.total_reads,
        unit: '',
        trend: undefined,
        trendValue: undefined,
      },
      {
        key: 'total_articles',
        label: '文章数',
        value: data.total_articles,
        unit: '篇',
        trend: undefined,
        trendValue: undefined,
      },
      {
        key: 'total_shares',
        label: '分享次数',
        value: data.total_shares,
        unit: '',
        trend: undefined,
        trendValue: undefined,
      },
      {
        key: 'avg_reads',
        label: '篇均阅读',
        value: data.total_articles > 0
          ? Math.round(data.total_reads / data.total_articles)
          : 0,
        unit: '',
        trend: undefined,
        trendValue: undefined,
      },
      {
        key: 'total_fans',
        label: '净增粉',
        value: data.total_fans_added,
        unit: '',
        trend: undefined,
        trendValue: undefined,
      },
    ]
  })

  // --- Computed: 4 Analysis Dimensions (mock until real analytics API) ---
  const dimensions = computed<AnalysisDimension[]>(() => {
    return [
      {
        label: '标题效果分析',
        description: '哪类标题更强',
        categories: ['疑问式', '数字式', '利益式', '故事式', '热点式'],
        values: generateDimensionValues(5),
        color: '#1B3A5C',
      },
      {
        label: '开头效果分析',
        description: '哪类开头更强',
        categories: ['痛点切入', '数据引用', '场景描述', '提问式', '故事式'],
        values: generateDimensionValues(5),
        color: '#5B8DEF',
      },
      {
        label: '配图效果分析',
        description: '哪类图片更强',
        categories: ['数据图表', '实景摄影', '插画设计', '对比图', '信息图'],
        values: generateDimensionValues(5),
        color: '#52C41A',
      },
      {
        label: 'CTA 效果分析',
        description: '哪类 CTA 更强',
        categories: ['关注引导', '转发引导', '评论互动', '链接跳转', '二维码'],
        values: generateDimensionValues(5),
        color: '#FAAD14',
      },
    ]
  })

  // --- Computed: Top keywords for chart ---
  const topKeywords = computed(() => {
    return recommendations.value.slice(0, 8).map((t) => ({
      keyword: t.keyword,
      score: t.historical_score,
      weight: t.recommend_weight,
    }))
  })

  // --- Actions ---
  async function fetchOverview(): Promise<void> {
    if (!wechatAccountId.value) return

    loading.value = true
    try {
      const resp = await getReportsOverview(wechatAccountId.value)
      if (resp.code === 0) {
        overview.value = resp.data
        hasData.value = true
      }
    } catch {
      // Handled by interceptor
    } finally {
      loading.value = false
    }
  }

  async function fetchTopics(): Promise<void> {
    try {
      const [listResp, recResp] = await Promise.all([
        listTopics(1, 50, 'historical_score'),
        getTopicRecommendations(10),
      ])
      if (listResp.code === 0) {
        topics.value = listResp.data.items
      }
      if (recResp.code === 0) {
        recommendations.value = recResp.data.items
      }
    } catch {
      // Handled by interceptor
    }
  }

  async function loadAll(accountId?: string): Promise<void> {
    if (accountId) {
      wechatAccountId.value = accountId
    }
    fetchError.value = null
    try {
      await Promise.all([fetchOverview(), fetchTopics()])
    } catch (e: unknown) {
      fetchError.value = e instanceof Error ? e.message : '加载运营数据失败'
    }
  }

  function exportReport(): void {
    ElMessage.info('导出功能开发中...')
  }

  // --- Helpers ---
  function defaultKpis(): KpiItem[] {
    const labels = ['阅读量', '文章数', '分享次数', '篇均阅读', '净增粉']
    const keys = ['total_reads', 'total_articles', 'total_shares', 'avg_reads', 'total_fans']
    return labels.map((label, i) => ({
      key: keys[i],
      label,
      value: 0,
      unit: i === 1 ? '篇' : '',
    }))
  }

  /** Generate plausible mock dimension values (for display demo). */
  function generateDimensionValues(count: number): number[] {
    const base = [72, 65, 58, 45, 38]
    return base.slice(0, count)
  }

  return {
    // State
    loading,
    fetchError,
    overview,
    topics,
    recommendations,
    wechatAccountId,
    hasData,
    // Computed
    kpis,
    dimensions,
    topKeywords,
    // Actions
    fetchOverview,
    fetchTopics,
    loadAll,
    exportReport,
  }
}
