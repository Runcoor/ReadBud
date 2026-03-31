import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getArticleMetrics } from '@/api/metrics'
import type { MetricsSnapshotVO, KpiItem } from '@/types/metrics'
import { DATE_RANGE_PRESETS } from '@/types/metrics'

export function useMetrics(getArticleId: () => string | null) {
  // --- State ---
  const snapshots = ref<MetricsSnapshotVO[]>([])
  const loading = ref(false)
  const dateRange = ref<string>('30d')
  const hasData = ref(false)

  // --- Computed: Date range ---
  const dateRangeOptions = computed(() =>
    Object.entries(DATE_RANGE_PRESETS).map(([key, preset]) => ({
      value: key,
      label: preset.label,
    })),
  )

  const dateParams = computed(() => {
    const preset = DATE_RANGE_PRESETS[dateRange.value]
    const end = new Date()
    const start = new Date()
    start.setDate(start.getDate() - (preset?.days ?? 30))
    return {
      start: formatDate(start),
      end: formatDate(end),
    }
  })

  // --- Computed: Latest KPIs ---
  const latestSnapshot = computed<MetricsSnapshotVO | null>(() => {
    if (snapshots.value.length === 0) return null
    return snapshots.value[snapshots.value.length - 1]
  })

  const kpis = computed<KpiItem[]>(() => {
    const latest = latestSnapshot.value
    const prev = snapshots.value.length >= 2 ? snapshots.value[snapshots.value.length - 2] : null

    return [
      buildKpi('reads', '阅读量', latest?.read_count, prev?.read_count, ''),
      buildKpi('read_users', '阅读人数', latest?.read_user_count, prev?.read_user_count, ''),
      buildKpi('shares', '分享次数', latest?.share_count, prev?.share_count, ''),
      buildKpi('share_users', '分享人数', latest?.share_user_count, prev?.share_user_count, ''),
      buildKpi('fans', '净增粉', latest?.net_fans_count, prev?.net_fans_count, ''),
    ]
  })

  // --- Computed: Chart data ---
  const chartDates = computed(() =>
    snapshots.value.map((s) => formatShortDate(s.metric_date)),
  )

  const chartReadData = computed(() =>
    snapshots.value.map((s) => s.read_count ?? 0),
  )

  const chartShareData = computed(() =>
    snapshots.value.map((s) => s.share_count ?? 0),
  )

  const chartFansData = computed(() =>
    snapshots.value.map((s) => s.net_fans_count ?? 0),
  )

  // --- Computed: Average calculations ---
  const avgReads = computed(() => {
    if (snapshots.value.length === 0) return 0
    const total = snapshots.value.reduce((sum, s) => sum + (s.read_count ?? 0), 0)
    return Math.round(total / snapshots.value.length)
  })

  const avgShares = computed(() => {
    if (snapshots.value.length === 0) return 0
    const total = snapshots.value.reduce((sum, s) => sum + (s.share_count ?? 0), 0)
    return Math.round(total / snapshots.value.length)
  })

  // --- Actions ---
  async function fetchMetrics(): Promise<void> {
    const articleId = getArticleId()
    if (!articleId) return

    loading.value = true
    try {
      const resp = await getArticleMetrics(articleId, dateParams.value.start, dateParams.value.end)
      if (resp.code === 0) {
        snapshots.value = resp.data.snapshots ?? []
        hasData.value = snapshots.value.length > 0
      }
    } catch {
      // Handled by interceptor
    } finally {
      loading.value = false
    }
  }

  // --- Helpers ---
  function formatDate(d: Date): string {
    return d.toISOString().slice(0, 10)
  }

  function formatShortDate(dateStr: string): string {
    const d = new Date(dateStr)
    return `${d.getMonth() + 1}/${d.getDate()}`
  }

  function buildKpi(
    key: string,
    label: string,
    current: number | null | undefined,
    previous: number | null | undefined,
    unit: string,
  ): KpiItem {
    const val = current ?? 0
    const prevVal = previous ?? 0
    const diff = val - prevVal

    let trend: 'up' | 'down' | 'flat' = 'flat'
    if (diff > 0) trend = 'up'
    else if (diff < 0) trend = 'down'

    return {
      key,
      label,
      value: val,
      unit,
      trend: previous !== null && previous !== undefined ? trend : undefined,
      trendValue: previous !== null && previous !== undefined ? Math.abs(diff) : undefined,
    }
  }

  // Watch date range changes
  watch(dateRange, () => {
    fetchMetrics()
  })

  return {
    // State
    snapshots,
    loading,
    dateRange,
    hasData,
    // Computed
    dateRangeOptions,
    latestSnapshot,
    kpis,
    chartDates,
    chartReadData,
    chartShareData,
    chartFansData,
    avgReads,
    avgShares,
    // Actions
    fetchMetrics,
  }
}
