<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="metrics-panel">
    <!-- Header -->
    <div class="panel-header">
      <h3 class="panel-title">数据复盘</h3>
      <el-radio-group v-model="dateRange" size="small">
        <el-radio-button
          v-for="opt in dateRangeOptions"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="panel-loading">
      <el-skeleton :rows="4" animated />
    </div>

    <!-- Error State -->
    <div v-else-if="fetchError" class="panel-error">
      <el-result icon="warning" title="数据加载失败" :sub-title="fetchError">
        <template #extra>
          <el-button size="small" type="primary" @click="fetchMetrics">重试</el-button>
        </template>
      </el-result>
    </div>

    <!-- Empty State -->
    <div v-else-if="!hasData" class="panel-empty">
      <el-empty description="暂无数据" :image-size="64">
        <template #description>
          <p class="empty-text">文章发布后数据将在此展示</p>
        </template>
      </el-empty>
    </div>

    <!-- Data Content -->
    <template v-else>
      <!-- KPI Cards -->
      <div class="kpi-grid">
        <div v-for="kpi in kpis" :key="kpi.key" class="mono-kpi-card">
          <span class="kpi-label">{{ kpi.label }}</span>
          <span class="kpi-value">{{ formatNumber(kpi.value) }}</span>
          <span
            v-if="kpi.trend && kpi.trendValue !== undefined"
            class="kpi-trend"
            :class="`kpi-trend--${kpi.trend}`"
          >
            <el-icon :size="12">
              <Top v-if="kpi.trend === 'up'" />
              <Bottom v-else-if="kpi.trend === 'down'" />
              <Minus v-else />
            </el-icon>
            {{ kpi.trendValue }}
          </span>
        </div>
      </div>

      <!-- Trend Chart -->
      <div class="chart-section">
        <h4 class="section-title">趋势变化</h4>
        <div ref="chartRef" class="trend-chart" />
      </div>

      <!-- Daily Breakdown Table -->
      <div class="table-section">
        <h4 class="section-title">每日明细</h4>
        <el-table
          :data="tableData"
          stripe
          size="small"
          class="mono-table"
          :max-height="280"
        >
          <el-table-column prop="date" label="日期" width="100" />
          <el-table-column prop="reads" label="阅读" align="right" width="80" />
          <el-table-column prop="readUsers" label="阅读人数" align="right" width="90" />
          <el-table-column prop="shares" label="分享" align="right" width="80" />
          <el-table-column prop="fans" label="净增粉" align="right" width="80" />
        </el-table>
      </div>

      <!-- Averages Summary -->
      <div class="mono-avg-section">
        <div class="avg-item">
          <span class="avg-label">日均阅读</span>
          <span class="avg-value">{{ formatNumber(avgReads) }}</span>
        </div>
        <div class="avg-divider" />
        <div class="avg-item">
          <span class="avg-label">日均分享</span>
          <span class="avg-value">{{ formatNumber(avgShares) }}</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { Top, Bottom, Minus } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { useMetrics } from '@/composables/useMetrics'

interface Props {
  articleId: string | null
}

const props = defineProps<Props>()

const {
  snapshots,
  loading,
  fetchError,
  dateRange,
  hasData,
  dateRangeOptions,
  kpis,
  chartDates,
  chartReadData,
  chartShareData,
  chartFansData,
  avgReads,
  avgShares,
  fetchMetrics,
} = useMetrics(() => props.articleId)

// --- Chart ---
const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

function initChart(): void {
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)
  updateChart()
}

function updateChart(): void {
  if (!chartInstance) return

  chartInstance.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#fff',
      borderColor: '#e8e8e8',
      borderWidth: 1,
      textStyle: { color: '#0a0a0a', fontSize: 12 },
    },
    legend: {
      bottom: 0,
      textStyle: { color: '#525252', fontSize: 12 },
      itemWidth: 12,
      itemHeight: 3,
    },
    grid: {
      top: 12,
      left: 8,
      right: 8,
      bottom: 36,
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: chartDates.value,
      axisLine: { lineStyle: { color: '#e8e8e8' } },
      axisLabel: { color: '#d4d4d4', fontSize: 11 },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#f5f5f5', type: 'dashed' } },
      axisLabel: { color: '#d4d4d4', fontSize: 11 },
    },
    series: [
      {
        name: '阅读量',
        type: 'line',
        data: chartReadData.value,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#0a0a0a' },
        itemStyle: { color: '#0a0a0a' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(10, 10, 10, 0.1)' },
            { offset: 1, color: 'rgba(10, 10, 10, 0.01)' },
          ]),
        },
      },
      {
        name: '分享',
        type: 'line',
        data: chartShareData.value,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#525252' },
        itemStyle: { color: '#525252' },
      },
      {
        name: '净增粉',
        type: 'line',
        data: chartFansData.value,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#22c55e' },
        itemStyle: { color: '#22c55e' },
      },
    ],
  })
}

function handleResize(): void {
  chartInstance?.resize()
}

// --- Table data ---
const tableData = computed(() =>
  snapshots.value.map((s) => ({
    date: formatTableDate(s.metric_date),
    reads: s.read_count ?? '-',
    readUsers: s.read_user_count ?? '-',
    shares: s.share_count ?? '-',
    fans: s.net_fans_count ?? '-',
  })),
)

// --- Helpers ---
function formatNumber(n: number): string {
  if (n >= 10000) {
    return (n / 10000).toFixed(1) + '万'
  }
  return n.toLocaleString('zh-CN')
}

function formatTableDate(dateStr: string): string {
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

// --- Lifecycle ---
onMounted(() => {
  fetchMetrics()
})

onBeforeUnmount(() => {
  chartInstance?.dispose()
  window.removeEventListener('resize', handleResize)
})

// Watch for data changes to update chart
watch([hasData, chartDates, chartReadData], async () => {
  if (hasData.value) {
    await nextTick()
    if (!chartInstance) {
      initChart()
      window.addEventListener('resize', handleResize)
    } else {
      updateChart()
    }
  }
})

// Re-fetch when articleId changes
watch(() => props.articleId, () => {
  fetchMetrics()
})
</script>

<style lang="scss" scoped>
.metrics-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: #0a0a0a;
  margin: 0;
}

// Radio buttons
:deep(.el-radio-button__inner) {
  background: var(--surface-card) !important;
  border-color: #e8e8e8 !important;
  color: #525252 !important;
  font-size: 12px;

  &:hover {
    color: #0a0a0a !important;
  }
}

:deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  box-shadow: -1px 0 0 0 #0a0a0a !important;
}

// Loading / Empty
.panel-loading { padding: 24px 0; }
.panel-error { padding: 24px 0; }
.panel-empty { padding: 32px 0; }

.empty-text {
  font-size: 13px;
  color: #d4d4d4;
}

:deep(.el-skeleton) { --el-skeleton-color: #f5f5f5; --el-skeleton-to-color: #e8e8e8; }
:deep(.el-empty__description p) { color: #525252 !important; }
:deep(.el-result__title) { color: #0a0a0a !important; }
:deep(.el-result__subtitle) { color: #525252 !important; }

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
}

// KPI Grid
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
}

.mono-kpi-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px;
  background: var(--surface-card);
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  transition: all 0.15s ease;

  &:hover {
    border-color: #0a0a0a;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }
}

.kpi-label {
  font-size: 11px;
  color: #d4d4d4;
  letter-spacing: 0.02em;
}

.kpi-value {
  font-size: 20px;
  font-weight: 700;
  color: #0a0a0a;
  line-height: 1.2;
}

.kpi-trend {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  font-weight: 500;

  &--up { color: #22c55e; }
  &--down { color: #ef4444; }
  &--flat { color: #d4d4d4; }
}

// Chart
.chart-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #525252;
  margin: 0;
}

.trend-chart {
  width: 100%;
  height: 220px;
}

// Table
.table-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

:deep(.el-table) {
  --el-table-border-color: #e8e8e8;
  --el-table-header-bg-color: #fafafa;
  --el-table-row-hover-bg-color: #f5f5f5;
  th { font-weight: 600 !important; color: #525252 !important; font-size: 12px; }
  td { font-size: 13px; }
}

// Averages
.mono-avg-section {
  display: flex;
  align-items: center;
  gap: 28px;
  padding: 14px;
  background: var(--surface-card);
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.avg-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.avg-divider {
  width: 1px;
  height: 32px;
  background-color: #e8e8e8;
}

.avg-label {
  font-size: 11px;
  color: #d4d4d4;
}

.avg-value {
  font-size: 18px;
  font-weight: 600;
  color: #0a0a0a;
}

@media (max-width: 1024px) {
  .kpi-grid { grid-template-columns: repeat(3, 1fr); }
  .trend-chart { height: 180px; }
}

@media (max-width: 768px) {
  .panel-header { flex-direction: column; align-items: flex-start; }
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .kpi-value { font-size: 17px; }
  .trend-chart { height: 160px; }
  .mono-avg-section { flex-direction: column; gap: 12px; }
  .avg-divider { width: 100%; height: 1px; }
}
</style>
