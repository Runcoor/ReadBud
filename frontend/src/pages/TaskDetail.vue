<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="task-detail">
    <AppTopBar :crumb="`任务 · ${task?.task_no || ''}`" user-initial="Y">
      <template #right>
        <button class="rail-btn" @click="router.push({ name: 'Workbench' })">← 返回工作台</button>
      </template>
    </AppTopBar>

    <main v-if="loading" class="detail-main">
      <div class="card">
        <el-skeleton :rows="10" animated />
      </div>
    </main>

    <main v-else-if="task" class="detail-main">
      <!-- Hero -->
      <section class="td-hero">
        <h1 class="td-hero__title">{{ task.keyword }}</h1>
        <MonoChip class="td-hero__code">TASK · {{ task.task_no }}</MonoChip>
        <p class="td-hero__sub">
          创建于 {{ formatDateTime(task.created_at) }} · 当前状态：
          <span class="td-hero__status">
            <StatusDot :kind="statusDotKind" :size="6" />
            <span>{{ statusLabel }}</span>
          </span>
        </p>
      </section>

      <!-- Info card -->
      <section class="card info-card">
        <SectionLabel title="任务信息" code="INFO" />
        <div class="info-fields">
          <div class="info-field">
            <span class="field-label">目标读者</span>
            <span class="field-value">{{ task.audience || '—' }}</span>
          </div>
          <div class="info-field">
            <span class="field-label">文章风格</span>
            <span class="field-value">{{ task.tone || '—' }}</span>
          </div>
          <div class="info-field">
            <span class="field-label">目标字数</span>
            <span class="field-value mono">{{ task.target_words || '—' }}</span>
          </div>
          <div class="info-field">
            <span class="field-label">配图模式</span>
            <span class="field-value">{{ imageModeLabel }}</span>
          </div>
          <div class="info-field">
            <span class="field-label">发布方式</span>
            <span class="field-value">{{ publishModeLabel }}</span>
          </div>
        </div>

        <!-- Custom progress bar -->
        <div class="td-progress">
          <div class="td-progress__track">
            <div
              class="td-progress__fill"
              :class="{ 'is-failed': task.status === 'failed' }"
              :style="{ width: progressWidth + '%' }"
            />
          </div>
          <div class="td-progress-meta mono">
            {{ progressWidth }}% · {{ progressTextLabel }}
          </div>
        </div>

        <!-- Failure block -->
        <div v-if="task.status === 'failed'" class="td-error">
          <div class="td-error__msg">
            <StatusDot kind="danger" :size="6" />
            <span>{{ task.error_message || '任务执行失败' }}</span>
          </div>
          <div class="td-error__actions">
            <el-button type="primary" plain size="small" @click="handleRetry">重新执行</el-button>
          </div>
        </div>
      </section>

      <!-- Pipeline -->
      <section class="card">
        <SectionLabel title="执行流程" code="PIPELINE" />
        <PipelineTimeline :task="task" />
      </section>

      <!-- Two-column layout: sources + draft -->
      <div class="detail-columns">
        <section class="card column-card">
          <SectionLabel title="来源文章" code="SOURCES" :hint="`${sources.length} 篇`" />
          <div v-if="sourcesLoading" class="column-card__body">
            <el-skeleton :rows="4" animated />
          </div>
          <div v-else-if="sources.length === 0" class="column-card__body empty-state">
            <span class="empty-state__caption mono">EMPTY · 暂无来源文章</span>
          </div>
          <div v-else class="column-card__body source-list">
            <div
              v-for="src in sources"
              :key="src.id"
              class="source-item"
            >
              <div class="source-title-row">
                <MonoChip>{{ sourceTypeLabel(src.source_type) }}</MonoChip>
                <a :href="src.source_url" target="_blank" rel="noopener" class="source-title">
                  {{ src.title }}
                </a>
              </div>
              <div class="source-meta">
                <span v-if="src.site_name">{{ src.site_name }}</span>
                <span v-if="src.author">· {{ src.author }}</span>
                <span v-if="src.published_at" class="mono">· {{ src.published_at }}</span>
              </div>
              <div class="source-scores">
                <div class="score-bar">
                  <span class="score-label">热度</span>
                  <div class="score-track">
                    <div
                      class="score-fill score-fill--hot"
                      :style="{ width: Math.min(src.hot_score, 100) + '%' }"
                    />
                  </div>
                  <span class="score-value mono">{{ src.hot_score.toFixed(1) }}</span>
                </div>
                <div class="score-bar">
                  <span class="score-label">相关</span>
                  <div class="score-track">
                    <div
                      class="score-fill score-fill--rel"
                      :style="{ width: Math.min(src.relevance_score, 100) + '%' }"
                    />
                  </div>
                  <span class="score-value mono">{{ src.relevance_score.toFixed(1) }}</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="card column-card">
          <SectionLabel title="文章预览" code="DRAFT" />
          <div class="column-card__body">
            <DraftPreview :draft-id="task.result_draft_id || null" />
          </div>
        </section>
      </div>

      <!-- Distribution / publish history -->
      <section v-if="task.result_draft_id" class="card">
        <SectionLabel title="分发管理" code="PUBLISH HISTORY" />
        <DistributionPanel :draft-public-id="task.result_draft_id" />
      </section>
    </main>

    <main v-else class="detail-main">
      <div class="card empty-card">
        <span class="empty-state__caption mono">NOT FOUND · 任务不存在或已删除</span>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getTask, retryTask } from '@/api/task'
import { getTaskSources } from '@/api/draft'
import DraftPreview from '@/components/task/DraftPreview.vue'
import DistributionPanel from '@/components/task/DistributionPanel.vue'
import AppTopBar from '@/components/common/AppTopBar.vue'
import StatusDot from '@/components/common/StatusDot.vue'
import SectionLabel from '@/components/common/SectionLabel.vue'
import MonoChip from '@/components/common/MonoChip.vue'
import PipelineTimeline from '@/components/common/PipelineTimeline.vue'
import type { TaskVO } from '@/types/task'
import type { SourceVO } from '@/types/draft'
import { IMAGE_MODE_LABELS, PUBLISH_MODE_LABELS, STATUS_LABELS } from '@/types/task'

const route = useRoute()
const router = useRouter()

const task = ref<TaskVO | null>(null)
const loading = ref(true)
const sources = ref<SourceVO[]>([])
const sourcesLoading = ref(false)

const statusLabel = computed(() => {
  if (!task.value) return ''
  return STATUS_LABELS[task.value.status] || task.value.status
})

const statusDotKind = computed<'sprout' | 'danger' | 'warn' | 'mute'>(() => {
  const s = task.value?.status
  if (s === 'done') return 'sprout'
  if (s === 'failed') return 'danger'
  if (s === 'running' || s === 'pending') return 'warn'
  return 'mute'
})

const imageModeLabel = computed(() => {
  if (!task.value) return '—'
  return IMAGE_MODE_LABELS[task.value.image_mode] || task.value.image_mode
})

const publishModeLabel = computed(() => {
  if (!task.value) return '—'
  return PUBLISH_MODE_LABELS[task.value.publish_mode] || task.value.publish_mode
})

const progressStatus = computed(() => {
  if (!task.value) return undefined
  if (task.value.status === 'done') return 'success' as const
  if (task.value.status === 'failed') return 'exception' as const
  return undefined
})

const progressWidth = computed(() => {
  const t = task.value
  if (!t) return 0
  if (t.status === 'done') return 100
  return Math.max(0, Math.min(100, t.progress || 0))
})

const progressTextLabel = computed(() => {
  const t = task.value
  if (!t) return '—'
  if (t.status === 'done') return '已完成'
  if (t.status === 'failed') return '执行失败'
  if (t.status === 'cancelled') return '已取消'
  if (t.status === 'pending') return '排队中'
  if (t.status === 'running') return '执行中'
  return statusLabel.value
})

function formatDateTime(s: string | undefined | null): string {
  if (!s) return '—'
  const d = new Date(s)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function sourceTypeLabel(t: string): string {
  const map: Record<string, string> = { web: '网页', news: '新闻', wechat: '公众号', blog: '博客' }
  return map[t] || t
}

async function fetchTask(): Promise<void> {
  loading.value = true
  try {
    const id = route.params.id as string
    const res = await getTask(id)
    task.value = res.data
  } catch {
    task.value = null
  } finally {
    loading.value = false
  }
}

async function fetchSources(): Promise<void> {
  if (!task.value) return
  sourcesLoading.value = true
  try {
    const res = await getTaskSources(task.value.id)
    sources.value = res.data || []
  } catch {
    sources.value = []
  } finally {
    sourcesLoading.value = false
  }
}

async function handleRetry(): Promise<void> {
  if (!task.value) return
  try {
    const res = await retryTask(task.value.id)
    task.value = res.data
    ElMessage.success('任务已重新提交')
  } catch {
    // Handled by interceptor
  }
}

// `progressStatus` retained as a computed so existing callers that import the
// page structure continue to type-check; reference it once to keep linters quiet.
void progressStatus

onMounted(async () => {
  await fetchTask()
  if (task.value) {
    fetchSources()
  }
})
</script>

<style scoped lang="scss">
@use '@/styles/tokens' as *;

.task-detail {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--brand-paper);
  font-family: var(--font-sans);
  color: var(--text-primary);
}

.mono {
  font-family: var(--font-mono);
  letter-spacing: 0.04em;
}

// === Topbar action ===
.rail-btn {
  background: transparent;
  border: none;
  padding: 0 4px;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: color 120ms ease;

  &:hover {
    color: var(--text-primary);
  }
}

// === Main ===
.detail-main {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 24px 48px 64px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

// === Hero ===
.td-hero {
  padding: 32px 0 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;

  &__title {
    font-family: var(--font-serif);
    font-size: 28px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    line-height: 1.25;
  }

  &__code {
    align-self: flex-start;
  }

  &__sub {
    font-size: 13px;
    color: var(--text-tertiary);
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  &__status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--text-body);
  }
}

// === Card primitive ===
.card {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  padding: 24px;
  box-shadow: none;
}

.empty-card {
  display: grid;
  place-items: center;
  min-height: 240px;
}

// === Info card ===
.info-card {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 24px;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.field-value {
  font-size: 13px;
  color: var(--text-primary);
}

// === Custom progress ===
.td-progress {
  display: flex;
  flex-direction: column;
  gap: 6px;

  &__track {
    height: 3px;
    background: var(--border-hair-soft);
    border-radius: 1px;
    overflow: hidden;
  }

  &__fill {
    height: 100%;
    background: var(--brand-ink);
    transition: width 240ms ease;

    &.is-failed {
      background: var(--brand-danger);
    }
  }
}

.td-progress-meta {
  font-size: 11px;
  color: var(--text-tertiary);
}

// === Error block ===
.td-error {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--brand-danger-soft);
  background: var(--brand-danger-soft);
  border-radius: 4px;

  &__msg {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--brand-danger);
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.btn-ghost {
  background: transparent;
  border: 1px solid var(--border-hair);
  color: var(--text-body);
  font-size: 12px;
  padding: 6px 12px;
  border-radius: 3px;
  cursor: pointer;
  font-family: var(--font-sans);
  transition: all 120ms ease;

  &:hover {
    color: var(--text-primary);
    border-color: var(--border-medium);
  }

  &--danger {
    color: var(--brand-danger);
    border-color: var(--brand-danger-soft);

    &:hover {
      background: var(--brand-danger-soft);
      color: var(--brand-danger);
    }
  }
}

// === Two-column ===
.detail-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.column-card {
  display: flex;
  flex-direction: column;
  max-height: 70vh;
  padding: 20px 22px;

  &__body {
    flex: 1;
    overflow-y: auto;
    margin-top: 4px;
  }
}

.empty-state {
  display: grid;
  place-items: center;
  min-height: 160px;

  &__caption {
    font-size: 11px;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }
}

// === Source list ===
.source-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.source-item {
  padding: 14px;
  border: 1px solid var(--border-hair);
  border-radius: 4px;
  background: var(--surface-card);
  transition: border-color 120ms ease;

  &:hover {
    border-color: var(--border-medium);
  }
}

.source-title-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 6px;
}

.source-title {
  font-size: 13px;
  color: var(--text-primary);
  text-decoration: none;
  line-height: 1.45;

  &:hover {
    color: var(--brand-ink);
    text-decoration: underline;
  }
}

.source-meta {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-bottom: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.source-scores {
  display: flex;
  gap: 16px;
}

.score-bar {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.score-label {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  width: 28px;
  flex-shrink: 0;
}

.score-track {
  flex: 1;
  height: 3px;
  background: var(--border-hair-soft);
  border-radius: 1px;
  overflow: hidden;
}

.score-fill {
  height: 100%;
  transition: width 240ms ease;

  &--hot {
    background: var(--brand-warn);
  }

  &--rel {
    background: var(--brand-ink);
  }
}

.score-value {
  font-size: 11px;
  color: var(--text-body);
  width: 32px;
  text-align: right;
  flex-shrink: 0;
}

// === Element Plus light overrides ===
:deep(.el-skeleton) {
  --el-skeleton-color: var(--brand-paper-warm);
  --el-skeleton-to-color: var(--border-hair);
}

:deep(.el-button--primary) {
  background: var(--brand-ink) !important;
  border-color: var(--brand-ink) !important;
  color: var(--text-inverse) !important;
  border-radius: 3px !important;

  &:hover { opacity: 0.85; }
  &:active { transform: scale(0.98); }
}

:deep(.el-button--primary.is-plain) {
  background: transparent !important;
  border: 1px solid var(--brand-ink) !important;
  color: var(--brand-ink) !important;

  &:hover {
    background: var(--brand-ink) !important;
    color: var(--text-inverse) !important;
  }
}

@media (max-width: $breakpoint-md) {
  .detail-columns {
    grid-template-columns: 1fr;
  }

  .info-fields {
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  }
}

@media (max-width: $breakpoint-sm) {
  .detail-main {
    padding: 16px 16px 48px;
  }

  .td-hero {
    padding: 16px 0 8px;

    &__title {
      font-size: 22px;
    }
  }
}
</style>
