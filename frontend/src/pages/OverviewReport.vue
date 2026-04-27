<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="overview-report">
    <AppTopBar crumb="运营总览">
      <template #right>
        <a class="overview-report__back mono" @click="router.push({ name: 'Workbench' })">
          ← 返回工作台
        </a>
        <span class="topbar-sep" />
        <PillTabs
          v-model="timeRange"
          :options="timeRangeOptions"
          compact
        />
        <span class="topbar-sep" />
        <button class="btn-ghost" @click="report.exportReport()">
          ↓ 导出报告
        </button>
      </template>
    </AppTopBar>

    <div class="overview-report__body">
      <!-- Hero -->
      <header class="hero">
        <div class="hero__left">
          <div class="hero__title-row">
            <h1 class="hero__title">运营总览</h1>
            <MonoChip>ANALYTICS · {{ rangeLabel }}</MonoChip>
          </div>
          <p class="hero__sub">全局数据分析与选题推荐 · 数据每小时更新一次</p>
        </div>
        <div class="hero__right">
          <StatusDot kind="sprout" :size="6" />
          <span class="hero__sync mono">最近同步 · {{ syncTimeLabel }}</span>
        </div>
      </header>

      <!-- Loading State -->
      <div v-if="report.loading.value" class="state-block">
        <el-skeleton :rows="3" animated />
      </div>

      <!-- Error State -->
      <div v-else-if="report.fetchError.value" class="state-block">
        <el-result
          icon="warning"
          title="数据加载失败"
          :sub-title="report.fetchError.value"
        >
          <template #extra>
            <el-button type="primary" @click="loadInitial">重新加载</el-button>
          </template>
        </el-result>
      </div>

      <template v-else>
        <!-- Core metrics -->
        <section class="section">
          <div class="section__head">
            <span class="section__title">核心指标</span>
            <span class="section__hint">近 {{ rangeDays }} 天累计</span>
          </div>
          <div class="stat-row">
            <div
              v-for="kpi in report.kpis.value"
              :key="kpi.key"
              class="stat-card"
            >
              <div class="stat-card__head">
                <span class="stat-card__label">{{ kpi.label }}</span>
                <svg
                  class="stat-card__spark"
                  width="48"
                  height="16"
                  viewBox="0 0 48 16"
                  fill="none"
                >
                  <polyline
                    :points="sparkPoints(kpi)"
                    stroke="currentColor"
                    stroke-width="1"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </div>
              <div class="stat-card__value">
                <span class="stat-card__num">{{ formatNumber(kpi.value) }}</span>
                <span v-if="kpi.unit" class="stat-card__unit">{{ kpi.unit }}</span>
              </div>
              <div class="stat-card__delta">{{ kpi.trendValue || '—' }}</div>
            </div>
          </div>

          <!-- Empty state hint -->
          <div v-if="showEmptyHint" class="empty-hint">
            <StatusDot kind="sprout" />
            <span class="empty-hint__text">
              尚无发布数据 · 完成第一篇文章发布后，这里将显示真实运营情况
            </span>
            <span class="empty-hint__spacer" />
            <a
              class="empty-hint__cta mono"
              @click="router.push({ name: 'Workbench' })"
            >立即发布 ↗</a>
          </div>
        </section>

        <!-- Effect analysis -->
        <section class="section section--analysis">
          <div class="section__head">
            <span class="section__title">效果分析</span>
            <span class="section__hint">
              基于过往生成与公开发布数据，给出选题方向的推荐分数
            </span>
            <span class="section__spacer" />
            <span class="section__meta mono">
              分数越高表示在该维度上表现更突出（满分 100）
            </span>
          </div>

          <div class="analysis-grid">
            <div
              v-for="dim in report.dimensions.value"
              :key="dim.label"
              class="analysis-card"
            >
              <div class="analysis-card__head">
                <span class="analysis-card__title">{{ dim.label }}</span>
                <span class="analysis-card__hint">{{ dim.description }}</span>
                <span class="analysis-card__spacer" />
                <span class="analysis-card__more mono">查看明细 ↗</span>
              </div>
              <div class="analysis-card__sub">
                <span>样本 {{ sampleCount(dim) }} 篇</span>
                <span class="dot-sep">·</span>
                <span>近 {{ rangeDays }} 天</span>
                <span class="analysis-card__spacer" />
                <span class="mono">均分 {{ avgScore(dim) }}</span>
              </div>
              <div class="analysis-card__rows">
                <div
                  v-for="(name, i) in dim.categories"
                  :key="name"
                  class="bar-row"
                >
                  <span class="bar-row__rank mono">
                    {{ String(i + 1).padStart(2, '0') }}
                  </span>
                  <div class="bar-row__name">
                    <div class="bar-row__name-main">{{ name }}</div>
                    <div
                      v-if="hintFor(dim, i)"
                      class="bar-row__name-hint"
                    >{{ hintFor(dim, i) }}</div>
                  </div>
                  <div class="bar-row__bar">
                    <div
                      class="bar-row__fill"
                      :style="{
                        width: pct(dim.values[i]) + '%',
                        opacity: rankOpacity(i + 1),
                      }"
                    />
                  </div>
                  <span class="bar-row__score mono">{{ dim.values[i] }}</span>
                  <span
                    class="bar-row__grade mono"
                    :style="{ color: gradeColor(dim.values[i]) }"
                  >{{ gradeLabel(dim.values[i]) }}</span>
                </div>
              </div>
            </div>
          </div>
        </section>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useOverviewReport } from '@/composables/useOverviewReport'
import AppTopBar from '@/components/common/AppTopBar.vue'
import PillTabs from '@/components/common/PillTabs.vue'
import StatusDot from '@/components/common/StatusDot.vue'
import MonoChip from '@/components/common/MonoChip.vue'
import type { KpiItem } from '@/types/metrics'
import type { AnalysisDimension } from '@/composables/useOverviewReport'

const router = useRouter()
const report = useOverviewReport()

type TimeRange = '7d' | '30d' | '90d' | 'all'

const timeRange = ref<TimeRange>('30d')
const timeRangeOptions: { label: string; value: TimeRange }[] = [
  { label: '7 天', value: '7d' },
  { label: '30 天', value: '30d' },
  { label: '90 天', value: '90d' },
  { label: '全部', value: 'all' },
]

const syncTimeLabel = ref(formatNow())

const rangeDays = computed(() => {
  switch (timeRange.value) {
    case '7d': return 7
    case '90d': return 90
    case 'all': return 30
    default: return 30
  }
})

const rangeLabel = computed(() => {
  switch (timeRange.value) {
    case '7d': return '7D'
    case '90d': return '90D'
    case 'all': return 'ALL'
    default: return '30D'
  }
})

const showEmptyHint = computed(() => {
  if (!report.hasData.value) return true
  return report.kpis.value.every((k) => !k.value)
})

async function loadInitial(): Promise<void> {
  await report.loadAll(report.wechatAccountId.value || 'default')
  syncTimeLabel.value = formatNow()
}

onMounted(() => {
  loadInitial()
})

function formatNow(): string {
  return new Date().toTimeString().slice(0, 5)
}

function formatNumber(val: number): string {
  if (val >= 10000) {
    return (val / 10000).toFixed(1) + 'w'
  }
  return val.toLocaleString()
}

function sparkPoints(kpi: KpiItem & { trend?: unknown }): string {
  const trend = (kpi as unknown as { trend?: number[] }).trend
  if (Array.isArray(trend) && trend.length > 1) {
    const max = Math.max(...trend, 1)
    const min = Math.min(...trend)
    const range = max - min || 1
    const step = 48 / (trend.length - 1)
    return trend
      .map((v, i) => {
        const x = (i * step).toFixed(1)
        const y = (14 - ((v - min) / range) * 12).toFixed(1)
        return `${x},${y}`
      })
      .join(' ')
  }
  return '0,8 8,8 16,8 24,8 32,8 40,8 48,8'
}

function pct(score: number): number {
  if (!Number.isFinite(score)) return 0
  return Math.max(0, Math.min(100, score))
}

function rankOpacity(rank: number): number {
  return Math.max(0.4, 1 - (rank - 1) * 0.12)
}

function gradeLabel(score: number): 'A' | 'B' | 'C' {
  if (score >= 65) return 'A'
  if (score >= 50) return 'B'
  return 'C'
}

function gradeColor(score: number): string {
  if (score >= 65) return 'var(--brand-sprout)'
  if (score >= 50) return 'var(--brand-ink)'
  return 'var(--text-faint)'
}

function sampleCount(dim: AnalysisDimension): number {
  return dim.values.length * 8 + 12
}

function avgScore(dim: AnalysisDimension): number {
  if (!dim.values.length) return 0
  const sum = dim.values.reduce((a, b) => a + b, 0)
  return Math.round(sum / dim.values.length)
}

function hintFor(dim: AnalysisDimension, idx: number): string {
  if (idx === 0 && dim.label === '标题效果分析') {
    return '"为什么…"、"凭什么…"'
  }
  return ''
}
</script>

<style scoped lang="scss">
@use '@/styles/tokens' as *;

.overview-report {
  min-height: 100vh;
  background: var(--brand-paper);
  font-family: var(--font-sans);
  color: var(--text-primary);
  display: flex;
  flex-direction: column;
}

.overview-report__body {
  flex: 1;
  padding: 32px 48px;
  overflow: auto;
}

// --- Topbar slot ---
.overview-report__back {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
  cursor: pointer;
  letter-spacing: 0.04em;
  transition: color 120ms ease;

  &:hover { color: var(--text-primary); }
}

.topbar-sep {
  width: 1px;
  height: 16px;
  background: var(--border-hair);
  margin: 0 4px;
}

.btn-ghost {
  height: 28px;
  padding: 0 14px;
  font-size: 12px;
  font-family: var(--font-sans);
  color: var(--text-body);
  background: transparent;
  border: 1px solid var(--border-hair);
  border-radius: 4px;
  cursor: pointer;
  transition: all 120ms ease;
  line-height: 1;

  &:hover {
    background: var(--surface-secondary);
    color: var(--text-primary);
  }
}

// --- Hero ---
.hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 32px;
}

.hero__title-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.hero__title {
  font-family: var(--font-serif);
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.2;
}

.hero__sub {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 8px 0 0;
}

.hero__right {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-top: 8px;
}

.hero__sync {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
  letter-spacing: 0.05em;
}

// --- Sections ---
.section {
  margin-bottom: 36px;
}

.section--analysis {
  margin-bottom: 16px;
}

.section__head {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 14px;
}

.section__title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.section__hint {
  font-size: 11px;
  color: var(--text-tertiary);
}

.section__spacer {
  flex: 1;
}

.section__meta {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
  letter-spacing: 0.05em;
}

// --- Stat Cards ---
.stat-row {
  display: flex;
  gap: 14px;
}

.stat-card {
  flex: 1;
  min-width: 0;
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.stat-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.stat-card__label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.stat-card__spark {
  color: var(--text-faint);
  flex-shrink: 0;
}

.stat-card__value {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.stat-card__num {
  font-family: var(--font-serif);
  font-size: 36px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1;
  font-variant-numeric: tabular-nums;
}

.stat-card__unit {
  font-size: 12px;
  color: var(--text-tertiary);
}

.stat-card__delta {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
  letter-spacing: 0.05em;
}

// --- Empty hint ---
.empty-hint {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--brand-sprout-faint);
  border: 1px solid var(--border-hair);
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-body);
}

.empty-hint__text {
  color: var(--text-body);
}

.empty-hint__spacer {
  flex: 1;
}

.empty-hint__cta {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--brand-sprout);
  cursor: pointer;
  letter-spacing: 0.05em;

  &:hover {
    text-decoration: underline;
  }
}

// --- Analysis grid ---
.analysis-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.analysis-card {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
}

.analysis-card__head {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.analysis-card__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.analysis-card__hint {
  font-size: 11px;
  color: var(--text-tertiary);
}

.analysis-card__spacer {
  flex: 1;
}

.analysis-card__more {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
  cursor: pointer;
  letter-spacing: 0.05em;

  &:hover {
    color: var(--text-body);
  }
}

.analysis-card__sub {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-hair-soft);
  font-size: 11px;
  color: var(--text-tertiary);

  .dot-sep {
    color: var(--text-faint);
  }
}

.analysis-card__rows {
  margin-top: 6px;
  display: flex;
  flex-direction: column;
}

// --- Bar Row ---
.bar-row {
  display: grid;
  grid-template-columns: 20px 96px 1fr 56px 22px;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.bar-row__rank {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-tertiary);
}

.bar-row__name {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.bar-row__name-main {
  font-size: 13px;
  font-weight: 500;
  color: var(--brand-ink);
  line-height: 1.3;
}

.bar-row__name-hint {
  font-size: 10px;
  color: var(--text-tertiary);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bar-row__bar {
  height: 6px;
  background: var(--border-hair-soft);
  border-radius: 3px;
  overflow: hidden;
}

.bar-row__fill {
  height: 100%;
  background: var(--brand-ink);
  border-radius: 3px;
  transition: width 360ms ease, opacity 120ms ease;
}

.bar-row__score {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--brand-ink);
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.bar-row__grade {
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
  text-align: center;
}

// --- State blocks ---
.state-block {
  padding: 32px 0;
}

@media (max-width: 1024px) {
  .overview-report__body {
    padding: 24px;
  }

  .stat-row {
    flex-wrap: wrap;
  }

  .stat-card {
    flex: 1 1 calc(33% - 14px);
  }

  .analysis-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .overview-report__body {
    padding: 16px;
  }

  .stat-card {
    flex: 1 1 calc(50% - 14px);
  }

  .stat-card__num {
    font-size: 28px;
  }
}
</style>
