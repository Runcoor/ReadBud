<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="publish-status-card" :class="statusClass">
    <!-- Header Row: status + mode badge -->
    <div class="status-header">
      <div class="status-indicator">
        <span class="status-dot" />
        <span class="status-text">{{ statusLabel }}</span>
      </div>
      <el-tag :type="modeTagType" size="small">{{ modeLabel }}</el-tag>
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

    <!-- Step indicator: queued -> submitting -> polling -> done -->
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

    <!-- Awaiting extension/manual: user needs to confirm publish completed -->
    <div v-if="isAwaitingUser" class="awaiting-section">
      <p class="awaiting-text">
        {{ job.status === 'awaiting_extension'
          ? '已打开 WeChat 编辑器。完成填充后,在编辑器里点「群发」。发完后回到这里点下面的按钮。'
          : '已打开 WeChat 编辑器。请手动粘贴标题/正文/封面后,在编辑器里点「群发」。完成后点下面的按钮。' }}
      </p>
      <div class="awaiting-actions">
        <el-button
          type="primary"
          :loading="markingFulfilled"
          @click="handleMarkFulfilled"
        >已发布,标记完成</el-button>
        <el-input
          v-model="manualArticleURL"
          placeholder="可选: 粘贴文章链接"
          size="small"
          clearable
        />
      </div>
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
import { retryPublishJob, markPublishJobFulfilled } from '@/api/publish'
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
const markingFulfilled = ref(false)
const manualArticleURL = ref('')

const isAwaitingUser = computed(() =>
  ['awaiting_extension', 'awaiting_manual'].includes(props.job.status),
)

const steps = [
  { key: 'queued', label: '排队' },
  { key: 'submitting', label: '提交' },
  { key: 'polling', label: '审核' },
]

const STATUS_LABELS: Record<string, string> = {
  queued: '排队中',
  submitting: '正在提交',
  polling: '等待平台审核',
  awaiting_extension: '等待插件填充',
  awaiting_manual: '等待手动复制',
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

async function handleMarkFulfilled(): Promise<void> {
  markingFulfilled.value = true
  try {
    const resp = await markPublishJobFulfilled(props.job.id, manualArticleURL.value || undefined)
    if (resp.code === 0) {
      ElMessage.success('已标记为发布完成')
      emit('update', resp.data)
    }
  } catch {
    ElMessage.error('标记失败,请稍后重试')
  } finally {
    markingFulfilled.value = false
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
  padding: 16px;
  border-radius: 8px;
  background: var(--surface-card);
  border: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: border-color 0.15s ease;

  &--success {
    border-color: #bbf7d0;
    background: #f0fdf4;
  }

  &--failed {
    border-color: #fecaca;
    background: #fef2f2;
  }

  &--cancelled {
    border-color: #e8e8e8;
    background: #fafafa;
    opacity: 0.75;
  }
}

.status-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #eab308;
  flex-shrink: 0;

  .publish-status-card--success & { background-color: #22c55e; }
  .publish-status-card--failed & { background-color: #ef4444; }
  .publish-status-card--cancelled & { background-color: #d4d4d4; }

  .publish-status-card--queued &,
  .publish-status-card--submitting &,
  .publish-status-card--polling & {
    animation: statusPulse 1.5s ease-in-out infinite;
  }

  .publish-status-card--awaiting_extension &,
  .publish-status-card--awaiting_manual & {
    background-color: #6366f1;
    animation: statusPulse 2s ease-in-out infinite;
  }
}

.status-text {
  font-size: 14px;
  font-weight: 600;
  color: #0a0a0a;
}

// Progress
.status-progress {
  margin-top: -4px;

  :deep(.el-progress-bar__outer) {
    background-color: #e8e8e8 !important;
  }

  :deep(.el-progress-bar__inner) {
    background: #525252;
  }
}

// Step Indicator
.step-indicator {
  display: flex;
  gap: 20px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.step-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #e8e8e8;
  transition: all 0.15s ease;

  .step-item--active & { background-color: #0a0a0a; }
  .step-item--current & {
    background-color: #0a0a0a;
    box-shadow: 0 0 0 3px rgba(10, 10, 10, 0.15);
  }
}

.step-label {
  font-size: 12px;
  color: #d4d4d4;
  transition: color 0.15s ease;

  .step-item--active & { color: #525252; }
  .step-item--current & { color: #0a0a0a; font-weight: 500; }
}

// Success
.success-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 14px 0;
}

.success-icon { color: #22c55e; }

.success-text {
  font-size: 13px;
  color: #525252;
}

.article-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  background: var(--surface-card);
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  font-size: 13px;
  color: #0a0a0a;
  text-decoration: none;
  transition: all 0.15s ease;

  &:hover {
    border-color: #0a0a0a;
  }
}

.link-external { opacity: 0.5; }

.article-link-pending {
  font-size: 12px;
  color: #d4d4d4;
}

// Failed
.failed-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.error-message {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 13px;
  color: #ef4444;
  line-height: 1.4;
  word-break: break-all;
}

.error-icon { flex-shrink: 0; margin-top: 2px; }

.retry-btn { align-self: flex-start; }

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
}

// Awaiting (extension / manual)
.awaiting-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 6px 0;
}

.awaiting-text {
  font-size: 13px;
  line-height: 1.6;
  color: #525252;
  margin: 0;
}

.awaiting-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

// Cancelled
.cancelled-section { padding: 4px 0; }

.cancelled-text {
  font-size: 13px;
  color: #d4d4d4;
}

// Footer
.status-footer {
  display: flex;
  align-items: center;
  gap: 14px;
  padding-top: 10px;
  border-top: 1px solid #e8e8e8;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #d4d4d4;

  &--warn { color: #eab308; }
}

// Tag
:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: #f5f5f5 !important; color: #525252 !important; }
:deep(.el-tag--warning) { background: #fefce8 !important; color: #ca8a04 !important; }

@keyframes statusPulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.85); }
}

@media (max-width: 1024px) {
  .step-indicator { gap: 14px; }
  .success-section { padding: 8px 0; }
}

@media (max-width: 768px) {
  .publish-status-card { padding: 12px; }
  .step-indicator { flex-wrap: wrap; gap: 10px; }
}
</style>
