<template>
  <div class="overview-report">
    <!-- Header -->
    <header class="mono-header">
      <div class="header-left">
        <el-button text :icon="ArrowLeft" @click="router.push({ name: 'Workbench' })">
          返回工作台
        </el-button>
      </div>
      <div class="header-center">
        <h1 class="page-title">运营总览</h1>
        <p class="page-subtitle">全局数据分析与选题推荐</p>
      </div>
      <div class="header-right">
        <el-button :icon="Download" @click="report.exportReport()">导出报告</el-button>
      </div>
    </header>

    <!-- Loading State -->
    <div v-if="report.loading.value" class="report-loading">
      <el-skeleton :rows="3" animated />
      <div class="skeleton-grid">
        <el-skeleton :rows="2" animated />
        <el-skeleton :rows="2" animated />
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="report.fetchError.value" class="report-error">
      <el-result icon="warning" title="数据加载失败" :sub-title="report.fetchError.value">
        <template #extra>
          <el-button type="primary" @click="report.loadAll('default')">重新加载</el-button>
        </template>
      </el-result>
    </div>

    <!-- Main Content -->
    <template v-else>
      <!-- KPI Cards -->
      <section class="kpi-section">
        <h2 class="section-title">核心指标</h2>
        <div class="kpi-grid">
          <div
            v-for="kpi in report.kpis.value"
            :key="kpi.key"
            class="mono-kpi-card"
          >
            <div class="kpi-label">{{ kpi.label }}</div>
            <div class="kpi-value">
              <span class="kpi-number">{{ formatNumber(kpi.value) }}</span>
              <span v-if="kpi.unit" class="kpi-unit">{{ kpi.unit }}</span>
            </div>
            <div v-if="kpi.trend" class="kpi-trend" :class="'trend-' + kpi.trend">
              <el-icon v-if="kpi.trend === 'up'" :size="12"><Top /></el-icon>
              <el-icon v-else-if="kpi.trend === 'down'" :size="12"><Bottom /></el-icon>
              <span>{{ kpi.trendValue }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Analysis Dimensions -->
      <section class="analysis-section">
        <h2 class="section-title">效果分析</h2>
        <div class="analysis-grid">
          <div
            v-for="dim in report.dimensions.value"
            :key="dim.label"
            class="mono-analysis-card"
          >
            <div class="analysis-card-header">
              <h3 class="analysis-card-title">{{ dim.label }}</h3>
              <span class="analysis-card-desc">{{ dim.description }}</span>
            </div>
            <div class="analysis-chart-container">
              <div class="bar-chart">
                <div
                  v-for="(cat, idx) in dim.categories"
                  :key="cat"
                  class="bar-row"
                >
                  <span class="bar-label">{{ cat }}</span>
                  <div class="bar-track">
                    <div
                      class="bar-fill"
                      :style="{
                        width: dim.values[idx] + '%',
                        opacity: 0.4 + (dim.values[idx] / 100) * 0.6,
                      }"
                    />
                  </div>
                  <span class="bar-value">{{ dim.values[idx] }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Topic Recommendations -->
      <section class="topics-section">
        <div class="section-header">
          <h2 class="section-title">选题推荐</h2>
          <el-button text @click="router.push({ name: 'Workbench' })">
            返回工作台使用 →
          </el-button>
        </div>

        <!-- Empty State -->
        <div v-if="report.recommendations.value.length === 0" class="mono-empty-state">
          <el-icon :size="48" style="color: var(--border-medium)"><DataLine /></el-icon>
          <p class="empty-title">暂无选题数据</p>
          <p class="empty-desc">使用关键词创建内容任务后，系统会自动积累选题推荐</p>
        </div>

        <!-- Topic Cards -->
        <div v-else class="topics-grid">
          <div
            v-for="topic in report.recommendations.value"
            :key="topic.public_id"
            class="mono-topic-card"
          >
            <div class="topic-keyword">{{ topic.keyword }}</div>
            <div class="topic-meta">
              <span v-if="topic.audience" class="topic-audience">
                {{ topic.audience }}
              </span>
              <span v-if="topic.article_goal" class="topic-goal">
                {{ topic.article_goal }}
              </span>
            </div>
            <div class="topic-scores">
              <div class="score-item">
                <span class="score-label">历史评分</span>
                <span class="score-value">{{ topic.historical_score.toFixed(1) }}</span>
              </div>
              <div class="score-item">
                <span class="score-label">推荐权重</span>
                <span class="score-value">{{ topic.recommend_weight.toFixed(1) }}</span>
              </div>
            </div>
            <div v-if="topic.last_used_at" class="topic-used">
              上次使用: {{ formatDateShort(topic.last_used_at) }}
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, Download, Top, Bottom, DataLine } from '@element-plus/icons-vue'
import { useOverviewReport } from '@/composables/useOverviewReport'

const router = useRouter()
const report = useOverviewReport()

onMounted(() => {
  // Load with a default account placeholder — real app would select account
  report.loadAll('default')
})

function formatNumber(val: number): string {
  if (val >= 10000) {
    return (val / 10000).toFixed(1) + 'w'
  }
  return val.toLocaleString()
}

function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}
</script>

<style lang="scss" scoped>
@use '@/styles/tokens' as *;

.overview-report {
  min-height: 100vh;
  padding: 0 48px 48px;
  margin: 0 auto;
  background: var(--surface-secondary);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', sans-serif;
}

.mono-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 0;
  margin-bottom: 32px;
  border-bottom: 1px solid var(--border-light);
}

.header-center {
  text-align: center;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  line-height: 1.2;
  color: var(--text-primary);
}

.page-subtitle {
  font-size: 13px;
  color: var(--border-medium);
  margin: 4px 0 0;
}

:deep(.el-button) {
  color: var(--text-secondary) !important;
  &:hover {
    color: var(--text-primary) !important;
    background: var(--surface-tertiary) !important;
  }
}

:deep(.el-button--primary) {
  background: var(--text-primary) !important;
  border-color: var(--text-primary) !important;
  color: var(--text-inverse) !important;
  border-radius: 8px !important;
  &:hover { background: var(--text-primary) !important; border-color: var(--text-primary) !important; opacity: 0.85; }
  &:active { transform: scale(0.98); }
}

.report-error {
  display: flex;
  justify-content: center;
  padding: 80px 0;

  :deep(.el-result__title) {
    color: var(--text-primary) !important;
  }
  :deep(.el-result__subtitle) {
    color: var(--text-secondary) !important;
  }
}

.report-loading {
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 32px 0;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

:deep(.el-skeleton) {
  --el-skeleton-color: var(--surface-tertiary);
  --el-skeleton-to-color: var(--border-light);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 20px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;

  .section-title {
    margin-bottom: 0;
  }
}

// KPI Cards
.kpi-section {
  margin-bottom: 40px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 20px;
}

.mono-kpi-card,
.mono-analysis-card,
.mono-topic-card {
  @include glass-panel;
  padding: 24px;
  transition: all 150ms ease;
  &:hover {
    box-shadow: var(--shadow-lg);
  }
}

.mono-kpi-card {
  padding: 24px 20px;
}

.kpi-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 10px;
}

.kpi-value {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.kpi-number {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.kpi-unit {
  font-size: 13px;
  color: var(--border-medium);
}

.kpi-trend {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-top: 8px;
  font-size: 12px;

  &.trend-up {
    color: #22c55e;
  }

  &.trend-down {
    color: #ef4444;
  }

  &.trend-flat {
    color: var(--border-medium);
  }
}

// Analysis Cards
.analysis-section {
  margin-bottom: 40px;
}

.analysis-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.analysis-card-header {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 20px;
}

.analysis-card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.analysis-card-desc {
  font-size: 12px;
  color: var(--border-medium);
}

// Bar Chart
.bar-chart {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bar-label {
  flex: 0 0 72px;
  font-size: 13px;
  color: var(--text-secondary);
  text-align: right;
}

.bar-track {
  flex: 1;
  height: 20px;
  background: var(--surface-tertiary);
  border-radius: 6px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 6px;
  background-color: var(--text-primary);
  transition: width 0.6s ease;
}

.bar-value {
  flex: 0 0 32px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  text-align: right;
}

// Topics Section
.topics-section {
  margin-bottom: 32px;
}

.mono-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: 12px;
}

.empty-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-secondary);
  margin: 20px 0 6px;
}

.empty-desc {
  font-size: 13px;
  color: var(--border-medium);
  margin: 0;
}

.topics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.mono-topic-card {
  padding: 20px;
  cursor: default;

  &:hover {
    border-color: var(--text-primary);
  }
}

.topic-keyword {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 10px;
}

.topic-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 14px;
}

.topic-audience,
.topic-goal {
  display: inline-block;
  padding: 2px 10px;
  background: var(--surface-tertiary);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.topic-scores {
  display: flex;
  gap: 20px;
  margin-bottom: 8px;
}

.score-item {
  display: flex;
  flex-direction: column;
}

.score-label {
  font-size: 11px;
  color: var(--border-medium);
}

.score-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.topic-used {
  font-size: 11px;
  color: var(--border-medium);
}

@media (max-width: 1024px) {
  .overview-report {
    padding: 0 24px 32px;
  }

  .kpi-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .analysis-grid {
    grid-template-columns: 1fr;
  }

  .topics-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .overview-report {
    padding: 0 16px 24px;
  }

  .mono-header {
    flex-direction: column;
    gap: 12px;
    padding: 16px 0;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .kpi-number {
    font-size: 22px;
  }

  .topics-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .bar-label {
    flex: 0 0 56px;
    font-size: 12px;
  }
}
</style>
