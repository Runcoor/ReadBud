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
        <div v-for="kpi in kpis" :key="kpi.key" class="kpi-card">
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
          class="metrics-table"
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
      <div class="avg-section">
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
      borderColor: '#E5E7EB',
      borderWidth: 1,
      textStyle: { color: '#1F2937', fontSize: 12 },
    },
    legend: {
      bottom: 0,
      textStyle: { color: '#6B7280', fontSize: 12 },
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
      axisLine: { lineStyle: { color: '#E5E7EB' } },
      axisLabel: { color: '#9CA3AF', fontSize: 11 },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#F0F0F0', type: 'dashed' } },
      axisLabel: { color: '#9CA3AF', fontSize: 11 },
    },
    series: [
      {
        name: '阅读量',
        type: 'line',
        data: chartReadData.value,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#1B3A5C' },
        itemStyle: { color: '#1B3A5C' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(27, 58, 92, 0.12)' },
            { offset: 1, color: 'rgba(27, 58, 92, 0.01)' },
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
        lineStyle: { width: 2, color: '#5B8DEF' },
        itemStyle: { color: '#5B8DEF' },
      },
      {
        name: '净增粉',
        type: 'line',
        data: chartFansData.value,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#52C41A' },
        itemStyle: { color: '#52C41A' },
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
  gap: $spacing-base;
}

// --- Header ---
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: $spacing-sm;
}

.panel-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0;
}

// --- Loading / Empty ---
.panel-loading {
  padding: $spacing-xl 0;
}

.panel-error {
  padding: $spacing-xl 0;
}

.panel-empty {
  padding: $spacing-2xl 0;
}

.empty-text {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

// --- KPI Grid ---
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: $spacing-md;
}

.kpi-card {
  display: flex;
  flex-direction: column;
  gap: $spacing-xs;
  padding: $spacing-md;
  background-color: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-lg;
  transition: border-color $transition-fast;

  &:hover {
    border-color: rgba($color-accent, 0.3);
  }
}

.kpi-label {
  font-size: $font-size-xs;
  color: $color-text-muted;
  letter-spacing: 0.02em;
}

.kpi-value {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
  line-height: $line-height-tight;
}

.kpi-trend {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: $font-size-xs;
  font-weight: $font-weight-medium;

  &--up {
    color: $color-success;
  }

  &--down {
    color: $color-error;
  }

  &--flat {
    color: $color-text-muted;
  }
}

// --- Chart ---
.chart-section {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.section-title {
  font-size: $font-size-sm;
  font-weight: $font-weight-semibold;
  color: $color-text-secondary;
  margin: 0;
}

.trend-chart {
  width: 100%;
  height: 220px;
}

// --- Table ---
.table-section {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.metrics-table {
  :deep(.el-table__header th) {
    background-color: $color-bg;
    color: $color-text-secondary;
    font-size: $font-size-xs;
    font-weight: $font-weight-medium;
  }

  :deep(.el-table__body td) {
    font-size: $font-size-sm;
    color: $color-text-primary;
  }
}

// --- Averages ---
.avg-section {
  display: flex;
  align-items: center;
  gap: $spacing-xl;
  padding: $spacing-md;
  background-color: $color-bg;
  border-radius: $radius-lg;
}

.avg-item {
  display: flex;
  flex-direction: column;
  gap: $spacing-xs;
}

.avg-divider {
  width: 1px;
  height: 32px;
  background-color: $color-border;
}

.avg-label {
  font-size: $font-size-xs;
  color: $color-text-muted;
}

.avg-value {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $color-primary;
}

// --- Responsive ---
@media (max-width: $breakpoint-md) {
  .kpi-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .trend-chart {
    height: 180px;
  }
}

@media (max-width: $breakpoint-sm) {
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .kpi-value {
    font-size: $font-size-lg;
  }

  .trend-chart {
    height: 160px;
  }

  .avg-section {
    flex-direction: column;
    gap: $spacing-md;
  }

  .avg-divider {
    width: 100%;
    height: 1px;
  }
}
</style>
