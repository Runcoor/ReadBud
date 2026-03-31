<template>
  <div class="publish-status-card" :class="statusClass">
    <!-- Header Row: status + mode badge -->
    <div class="status-header">
      <div class="status-indicator">
        <span class="status-dot" />
        <span class="status-text">{{ statusLabel }}</span>
      </div>
      <el-tag :type="modeTagType" size="small" effect="plain">{{ modeLabel }}</el-tag>
    </div>

    <!-- Progress bar for in-progress states -->
    <el-progress
      v-if="isInProgress"
      :percentage="progressPct"
      :status="progressStatus"
      :stroke-width="4"
      :show-text="false"
      class="status-progress"
    />

    <!-- Step indicator: queued → submitting → polling → done -->
    <div v-if="isInProgress" class="step-indicator">
      <div
        v-for="(step, idx) in steps"
        :key="step.key"
        class="step-item"
        :class="{ 'step-item--active': stepIndex >= idx, 'step-item--current': stepIndex === idx }"
      >
        <span class="step-dot" />
        <span class="step-label">{{ step.label }}</span>
      </div>
    </div>

    <!-- Success: article URL link -->
    <div v-if="job.status === 'success'" class="success-section">
      <el-icon :size="32" class="success-icon"><CircleCheckFilled /></el-icon>
      <p class="success-text">文章已成功发布至公众号</p>
      <a
        v-if="job.article_url"
        :href="job.article_url"
        target="_blank"
        rel="noopener noreferrer"
        class="article-link"
      >
        <el-icon :size="14"><Link /></el-icon>
        <span>查看文章</span>
        <el-icon :size="12" class="link-external"><TopRight /></el-icon>
      </a>
      <span v-else class="article-link-pending">文章链接获取中...</span>
    </div>

    <!-- Failed: error message + retry -->
    <div v-if="job.status === 'failed'" class="failed-section">
      <div class="error-message">
        <el-icon :size="14" class="error-icon"><CircleCloseFilled /></el-icon>
        <span>{{ job.last_error || '发布失败，请重试' }}</span>
      </div>
      <el-button
        type="primary"
        plain
        size="small"
        :loading="retrying"
        class="retry-btn"
        @click="handleRetry"
      >
        <el-icon v-if="!retrying"><RefreshRight /></el-icon>
        重新发布
      </el-button>
    </div>

    <!-- Cancelled -->
    <div v-if="job.status === 'cancelled'" class="cancelled-section">
      <p class="cancelled-text">发布已取消，可重新配置后再次发布</p>
    </div>

    <!-- Footer: meta info -->
    <div class="status-footer">
      <span class="meta-item">
        <el-icon :size="12"><Clock /></el-icon>
        {{ formatTime(job.created_at) }}
      </span>
      <span v-if="job.retry_count > 0" class="meta-item meta-item--warn">
        重试 {{ job.retry_count }} 次
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  CircleCheckFilled, CircleCloseFilled, Link, TopRight,
  RefreshRight, Clock,
} from '@element-plus/icons-vue'
import { retryPublishJob } from '@/api/publish'
import type { PublishJobVO } from '@/api/publish'

interface Props {
  job: PublishJobVO
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'retry', job: PublishJobVO): void
  (e: 'update', job: PublishJobVO): void
}>()

const retrying = ref(false)

const steps = [
  { key: 'queued', label: '排队' },
  { key: 'submitting', label: '提交' },
  { key: 'polling', label: '审核' },
]

const STATUS_LABELS: Record<string, string> = {
  queued: '排队中',
  submitting: '正在提交',
  polling: '等待平台审核',
  success: '发布成功',
  failed: '发布失败',
  cancelled: '已取消',
}

const statusLabel = computed(() => STATUS_LABELS[props.job.status] || props.job.status)

const statusClass = computed(() => `publish-status-card--${props.job.status}`)

const isInProgress = computed(() =>
  ['queued', 'submitting', 'polling'].includes(props.job.status),
)

const stepIndex = computed(() => {
  const map: Record<string, number> = { queued: 0, submitting: 1, polling: 2 }
  return map[props.job.status] ?? -1
})

const progressPct = computed(() => {
  const map: Record<string, number> = { queued: 20, submitting: 50, polling: 80, success: 100 }
  return map[props.job.status] ?? 0
})

const progressStatus = computed<'' | 'success' | 'exception'>(() => {
  if (props.job.status === 'success') return 'success'
  if (props.job.status === 'failed') return 'exception'
  return ''
})

const modeLabel = computed(() => {
  const map: Record<string, string> = { now: '即时发布', schedule: '定时发布', manual: '手动导出' }
  return map[props.job.publish_mode] || props.job.publish_mode
})

const modeTagType = computed(() => {
  const map: Record<string, string> = { now: '', schedule: 'warning', manual: 'info' }
  return (map[props.job.publish_mode] || 'info') as 'warning' | 'info' | ''
})

function formatTime(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleString('zh-CN', {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  } catch {
    return iso
  }
}

async function handleRetry(): Promise<void> {
  retrying.value = true
  try {
    const resp = await retryPublishJob(props.job.id)
    if (resp.code === 0) {
      ElMessage.success('已重新提交发布')
      emit('retry', resp.data)
      emit('update', resp.data)
    }
  } catch {
    ElMessage.error('重试失败，请稍后再试')
  } finally {
    retrying.value = false
  }
}
</script>

<style lang="scss" scoped>
.publish-status-card {
  padding: $spacing-base;
  border-radius: $radius-lg;
  background-color: $color-bg;
  border: 1px solid $color-border;
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
  transition: border-color $transition-base;

  &--success {
    border-color: rgba($color-success, 0.3);
    background-color: rgba($color-success, 0.03);
  }

  &--failed {
    border-color: rgba($color-error, 0.3);
    background-color: rgba($color-error, 0.03);
  }

  &--cancelled {
    border-color: $color-border;
    background-color: $color-bg;
    opacity: 0.75;
  }
}

// --- Header ---
.status-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: $color-warning;
  flex-shrink: 0;

  .publish-status-card--success & { background-color: $color-success; }
  .publish-status-card--failed & { background-color: $color-error; }
  .publish-status-card--cancelled & { background-color: $color-metal; }

  .publish-status-card--queued &,
  .publish-status-card--submitting &,
  .publish-status-card--polling & {
    animation: statusPulse 1.5s ease-in-out infinite;
  }
}

.status-text {
  font-size: $font-size-base;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

// --- Progress ---
.status-progress {
  margin-top: -$spacing-xs;
}

// --- Step Indicator ---
.step-indicator {
  display: flex;
  gap: $spacing-lg;
}

.step-item {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
}

.step-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: $color-border;
  transition: all $transition-fast;

  .step-item--active & {
    background-color: $color-accent;
  }

  .step-item--current & {
    background-color: $color-accent;
    box-shadow: 0 0 0 3px rgba($color-accent, 0.2);
  }
}

.step-label {
  font-size: $font-size-xs;
  color: $color-text-muted;
  transition: color $transition-fast;

  .step-item--active & { color: $color-text-secondary; }
  .step-item--current & { color: $color-accent; font-weight: $font-weight-medium; }
}

// --- Success ---
.success-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-md 0;
}

.success-icon { color: $color-success; }

.success-text {
  font-size: $font-size-sm;
  color: $color-text-secondary;
}

.article-link {
  display: inline-flex;
  align-items: center;
  gap: $spacing-xs;
  padding: $spacing-xs $spacing-md;
  background-color: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-base;
  font-size: $font-size-sm;
  color: $color-accent;
  text-decoration: none;
  transition: all $transition-fast;

  &:hover {
    border-color: $color-accent;
    background-color: rgba($color-accent, 0.04);
  }
}

.link-external { opacity: 0.5; }

.article-link-pending {
  font-size: $font-size-xs;
  color: $color-text-muted;
}

// --- Failed ---
.failed-section {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.error-message {
  display: flex;
  align-items: flex-start;
  gap: $spacing-xs;
  font-size: $font-size-sm;
  color: $color-error;
  line-height: $line-height-normal;
  word-break: break-all;
}

.error-icon { flex-shrink: 0; margin-top: 2px; }

.retry-btn {
  align-self: flex-start;
}

// --- Cancelled ---
.cancelled-section {
  padding: $spacing-xs 0;
}

.cancelled-text {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

// --- Footer ---
.status-footer {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding-top: $spacing-sm;
  border-top: 1px solid $color-divider;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: $spacing-xs;
  font-size: $font-size-xs;
  color: $color-text-muted;

  &--warn { color: $color-warning; }
}

// --- Animation ---
@keyframes statusPulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.85); }
}

// --- Responsive ---
@media (max-width: $breakpoint-md) {
  .step-indicator { gap: $spacing-md; }
  .success-section { padding: $spacing-sm 0; }
}

@media (max-width: $breakpoint-sm) {
  .publish-status-card { padding: $spacing-md; }
  .step-indicator { flex-wrap: wrap; gap: $spacing-sm; }
}
</style>
