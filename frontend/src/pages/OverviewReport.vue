<template>
  <div class="overview-report">
    <!-- Header -->
    <header class="report-header">
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
            class="kpi-card"
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
            class="analysis-card"
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
                        backgroundColor: dim.color,
                        opacity: 0.6 + (dim.values[idx] / 100) * 0.4,
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
          <el-button text type="primary" @click="router.push({ name: 'Workbench' })">
            返回工作台使用 →
          </el-button>
        </div>

        <!-- Empty State -->
        <div v-if="report.recommendations.value.length === 0" class="empty-state">
          <el-icon :size="48" color="#D1D5DB"><DataLine /></el-icon>
          <p class="empty-title">暂无选题数据</p>
          <p class="empty-desc">使用关键词创建内容任务后，系统会自动积累选题推荐</p>
        </div>

        <!-- Topic Cards -->
        <div v-else class="topics-grid">
          <div
            v-for="topic in report.recommendations.value"
            :key="topic.public_id"
            class="topic-card"
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
@use '@/styles/_tokens.scss' as *;

.overview-report {
  min-height: 100vh;
  background: $color-bg;
  padding: $spacing-2xl $spacing-3xl;
  max-width: 1440px;
  margin: 0 auto;
}

// --- Header ---
.report-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-2xl;
  padding-bottom: $spacing-lg;
  border-bottom: 1px solid $color-border;
}

.header-center {
  text-align: center;
}

.page-title {
  font-size: $font-size-2xl;
  font-weight: $font-weight-semibold;
  color: $color-primary;
  margin: 0;
  line-height: $line-height-tight;
}

.page-subtitle {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin: $spacing-xs 0 0;
}

// --- Error ---
.report-error {
  display: flex;
  justify-content: center;
  padding: $spacing-4xl 0;
}

// --- Loading ---
.report-loading {
  display: flex;
  flex-direction: column;
  gap: $spacing-xl;
  padding: $spacing-2xl 0;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-xl;
}

// --- Section ---
.section-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0 0 $spacing-lg;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-lg;

  .section-title {
    margin-bottom: 0;
  }
}

// --- KPI Cards ---
.kpi-section {
  margin-bottom: $spacing-3xl;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: $spacing-lg;
}

.kpi-card {
  background: $color-card-bg;
  border-radius: $radius-lg;
  padding: $spacing-xl $spacing-lg;
  box-shadow: $shadow-card;
  transition: box-shadow $transition-base;

  &:hover {
    box-shadow: $shadow-card-hover;
  }
}

.kpi-label {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin-bottom: $spacing-sm;
}

.kpi-value {
  display: flex;
  align-items: baseline;
  gap: $spacing-xs;
}

.kpi-number {
  font-size: $font-size-3xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
  line-height: $line-height-tight;
}

.kpi-unit {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.kpi-trend {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-top: $spacing-sm;
  font-size: $font-size-xs;

  &.trend-up {
    color: $color-success;
  }

  &.trend-down {
    color: $color-error;
  }

  &.trend-flat {
    color: $color-text-muted;
  }
}

// --- Analysis Cards ---
.analysis-section {
  margin-bottom: $spacing-3xl;
}

.analysis-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-lg;
}

.analysis-card {
  background: $color-card-bg;
  border-radius: $radius-lg;
  padding: $spacing-xl;
  box-shadow: $shadow-card;
}

.analysis-card-header {
  display: flex;
  align-items: baseline;
  gap: $spacing-sm;
  margin-bottom: $spacing-lg;
}

.analysis-card-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0;
}

.analysis-card-desc {
  font-size: $font-size-xs;
  color: $color-text-muted;
}

// --- Bar Chart (inline) ---
.bar-chart {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.bar-label {
  flex: 0 0 72px;
  font-size: $font-size-sm;
  color: $color-text-secondary;
  text-align: right;
}

.bar-track {
  flex: 1;
  height: 20px;
  background: $color-bg;
  border-radius: $radius-sm;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: $radius-sm;
  transition: width $transition-slow;
}

.bar-value {
  flex: 0 0 32px;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  text-align: right;
}

// --- Topics Section ---
.topics-section {
  margin-bottom: $spacing-2xl;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-4xl $spacing-xl;
  background: $color-card-bg;
  border-radius: $radius-lg;
  box-shadow: $shadow-card;
}

.empty-title {
  font-size: $font-size-md;
  font-weight: $font-weight-medium;
  color: $color-text-secondary;
  margin: $spacing-lg 0 $spacing-xs;
}

.empty-desc {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin: 0;
}

.topics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: $spacing-lg;
}

.topic-card {
  background: $color-card-bg;
  border-radius: $radius-lg;
  padding: $spacing-lg;
  box-shadow: $shadow-card;
  transition: box-shadow $transition-base;
  cursor: default;

  &:hover {
    box-shadow: $shadow-card-hover;
  }
}

.topic-keyword {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-primary;
  margin-bottom: $spacing-sm;
}

.topic-meta {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-xs;
  margin-bottom: $spacing-md;
}

.topic-audience,
.topic-goal {
  display: inline-block;
  padding: 2px $spacing-sm;
  background: $color-bg;
  border-radius: $radius-sm;
  font-size: $font-size-xs;
  color: $color-text-secondary;
}

.topic-scores {
  display: flex;
  gap: $spacing-lg;
  margin-bottom: $spacing-sm;
}

.score-item {
  display: flex;
  flex-direction: column;
}

.score-label {
  font-size: $font-size-xs;
  color: $color-text-muted;
}

.score-value {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.topic-used {
  font-size: $font-size-xs;
  color: $color-text-muted;
}

// --- Responsive: 1024px ---
@media (max-width: $breakpoint-md) {
  .overview-report {
    padding: $spacing-xl $spacing-lg;
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

// --- Responsive: 768px ---
@media (max-width: $breakpoint-sm) {
  .overview-report {
    padding: $spacing-lg $spacing-base;
  }

  .report-header {
    flex-direction: column;
    gap: $spacing-md;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .kpi-number {
    font-size: $font-size-xl;
  }

  .topics-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .bar-label {
    flex: 0 0 56px;
    font-size: $font-size-xs;
  }
}
</style>
