<template>
  <div class="task-progress">
    <!-- Task header info -->
    <div v-if="task" class="progress-header">
      <div class="progress-meta">
        <span class="progress-keyword">{{ task.keyword }}</span>
        <el-tag :type="STATUS_TAG_TYPES[task.status]" size="small">
          {{ STATUS_LABELS[task.status] }}
        </el-tag>
      </div>
      <div class="progress-bar-wrapper">
        <el-progress
          :percentage="task.progress"
          :status="progressStatus"
          :stroke-width="6"
          :show-text="true"
          :format="formatProgress"
        />
      </div>
    </div>

    <div class="flow-divider" :class="{ 'flow-divider--active': task && (task.status === 'running' || task.status === 'pending') }"></div>

    <!-- Pipeline stages -->
    <div v-if="task" class="pipeline">
      <div
        v-for="(stage, index) in TASK_STAGES"
        :key="stage.key"
        class="pipeline-stage"
        :class="stageClass(stage.key, index)"
      >
        <div class="stage-indicator">
          <div class="stage-dot">
            <el-icon v-if="stageState(stage.key, index) === 'done'" :size="14">
              <Check />
            </el-icon>
            <el-icon v-else-if="stageState(stage.key, index) === 'active'" :size="14" class="stage-spinning">
              <Loading />
            </el-icon>
            <span v-else class="stage-number">{{ index + 1 }}</span>
          </div>
          <div v-if="index < TASK_STAGES.length - 1" class="stage-line" />
        </div>
        <div class="stage-content">
          <span class="stage-label">{{ stage.label }}</span>
          <span v-if="stageState(stage.key, index) === 'active'" class="stage-status">执行中...</span>
          <span v-else-if="stageState(stage.key, index) === 'done'" class="stage-status stage-status--done">完成</span>
        </div>
      </div>
    </div>

    <div v-if="task && (task.status === 'running' || task.status === 'pending')" class="cancel-row">
      <button class="cancel-btn" @click="$emit('cancel', task.id)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        取消任务
      </button>
    </div>

    <!-- Error message -->
    <div v-if="task?.status === 'failed' && task.error_message" class="progress-error">
      <el-alert
        :title="task.error_message"
        type="error"
        :closable="false"
        show-icon
      />
      <el-button
        type="primary"
        plain
        size="small"
        class="retry-btn"
        @click="$emit('retry', task.id)"
      >
        重新执行
      </el-button>
    </div>

    <!-- Empty state when no task -->
    <div v-if="!task" class="progress-empty">
      <div class="empty-icon">
        <el-icon :size="48" color="var(--border-medium)"><Promotion /></el-icon>
      </div>
      <p class="empty-text">配置任务参数后点击"开始生成"</p>
      <p class="empty-sub">系统将自动完成内容采集、分析、撰写全流程</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Check, Loading, Promotion } from '@element-plus/icons-vue'
import type { TaskVO, TaskStage } from '@/types/task'
import { TASK_STAGES, STATUS_LABELS, STATUS_TAG_TYPES } from '@/types/task'

interface Props {
  task: TaskVO | null
}

const props = defineProps<Props>()

const emit = defineEmits<{ retry: [taskId: string]; cancel: [taskId: string] }>()

type StageState = 'pending' | 'active' | 'done'

/** Get the index of the current active stage */
function currentStageIndex(): number {
  if (!props.task?.current_stage) return -1
  return TASK_STAGES.findIndex(s => s.key === props.task?.current_stage)
}

/** Determine the visual state of a stage */
function stageState(key: TaskStage, index: number): StageState {
  if (!props.task) return 'pending'

  if (props.task.status === 'done') return 'done'
  if (props.task.status === 'failed') {
    const activeIdx = currentStageIndex()
    if (index < activeIdx) return 'done'
    if (index === activeIdx) return 'active'
    return 'pending'
  }

  const activeIdx = currentStageIndex()
  if (activeIdx < 0) return 'pending'

  if (index < activeIdx) return 'done'
  if (index === activeIdx && (props.task.status === 'running')) return 'active'
  return 'pending'
}

/** CSS class for a pipeline stage */
function stageClass(key: TaskStage, index: number): string {
  const state = stageState(key, index)
  return `pipeline-stage--${state}`
}

/** Progress bar status */
const progressStatus = computed(() => {
  if (!props.task) return undefined
  if (props.task.status === 'done') return 'success' as const
  if (props.task.status === 'failed') return 'exception' as const
  return undefined
})

function formatProgress(percentage: number): string {
  return `${percentage}%`
}
</script>

<style lang="scss" scoped>
.task-progress {
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
}

.progress-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.progress-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.progress-keyword {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

// Flow divider with animation
.flow-divider {
  height: 2px;
  background: var(--border-light);
  position: relative;
  overflow: hidden;

  &--active::after {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, transparent, var(--text-primary), transparent);
    animation: flow 1.8s ease-in-out infinite;
  }
}

@keyframes flow {
  0% { left: -100%; }
  100% { left: 100%; }
}

// Progress bar
.progress-bar-wrapper {
  :deep(.el-progress-bar__outer) {
    background-color: var(--border-light) !important;
    border-radius: 4px;
  }

  :deep(.el-progress-bar__inner) {
    border-radius: 4px;
    transition: width 0.6s ease;
    background: var(--text-secondary);
  }

  :deep(.el-progress__text) {
    color: var(--text-secondary) !important;
  }
}

// Pipeline stages
.pipeline {
  display: flex;
  flex-direction: column;
  gap: 0;
  flex: 1;
  overflow-y: auto;
}

.pipeline-stage {
  display: flex;
  gap: 14px;
  min-height: 44px;
}

.stage-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 28px;
  flex-shrink: 0;
}

.stage-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.stage-number {
  font-size: 11px;
}

.stage-line {
  width: 2px;
  flex: 1;
  min-height: 16px;
  transition: background-color 0.2s ease;
}

// Stage states
.pipeline-stage--pending {
  .stage-dot {
    background: var(--surface-tertiary);
    border: 2px solid var(--border-light);
    color: var(--border-medium);
  }

  .stage-line {
    background-color: var(--border-light);
  }

  .stage-label {
    color: var(--border-medium);
  }
}

.pipeline-stage--active {
  .stage-dot {
    background: var(--text-primary);
    border: 2px solid var(--text-primary);
    color: var(--text-inverse);
    animation: pulse 1.5s ease-in-out infinite;
  }

  .stage-line {
    background-color: var(--border-light);
  }

  .stage-label {
    color: var(--text-primary);
    font-weight: 600;
  }
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(10, 10, 10, 0.3); }
  50% { box-shadow: 0 0 0 6px rgba(10, 10, 10, 0); }
}

.pipeline-stage--done {
  .stage-dot {
    background: #f0fdf4;
    border: 2px solid var(--status-success);
    color: var(--status-success);
    animation: check-in 0.3s ease-out;
  }

  .stage-line {
    background-color: var(--status-success);
  }

  .stage-label {
    color: var(--text-secondary);
  }
}

@keyframes check-in {
  0% { transform: scale(0.5); opacity: 0; }
  100% { transform: scale(1); opacity: 1; }
}

.stage-content {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 4px;
  min-height: 28px;
}

.stage-label {
  font-size: 14px;
  transition: color 0.2s ease;
}

.stage-status {
  font-size: 12px;
  color: var(--text-secondary);
}

.stage-status--done {
  color: var(--status-success);
}

.stage-spinning {
  animation: spin 1.2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// Cancel button
.cancel-row {
  display: flex;
  justify-content: center;
  padding-top: 16px;
}

.cancel-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: transparent;
  border: 1px solid var(--border-light);
  border-radius: 8px;
  color: var(--text-tertiary);
  font-size: 13px;
  cursor: pointer;
  transition: all 150ms ease;

  &:hover {
    color: var(--status-danger);
    border-color: var(--status-danger);
    background: rgba(239, 68, 68, 0.05);
  }
}

// Error state
.progress-error {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 12px;
}

:deep(.el-alert) {
  border-radius: 8px;
  &.el-alert--error {
    background: #fef2f2 !important;
    border: 1px solid #fecaca !important;
    .el-alert__title { color: #dc2626 !important; }
  }
}

.retry-btn {
  align-self: flex-start;
}

:deep(.el-button--primary) {
  background: var(--text-primary) !important;
  border-color: var(--text-primary) !important;
  color: var(--text-inverse) !important;
  border-radius: 8px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
  &:active { transform: scale(0.98); }
}

// Tag overrides
:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: var(--surface-tertiary) !important; color: var(--text-secondary) !important; }
:deep(.el-tag--success) { background: #f0fdf4 !important; color: #16a34a !important; }
:deep(.el-tag--danger) { background: #fef2f2 !important; color: #dc2626 !important; }
:deep(.el-tag--warning) { background: #fefce8 !important; color: #ca8a04 !important; }

// Empty state
.progress-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  text-align: center;
  padding: 48px 20px;
}

.empty-icon {
  margin-bottom: 20px;
  opacity: 0.6;
}

.empty-text {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.empty-sub {
  font-size: 13px;
  color: var(--text-tertiary);
}

@media (max-width: 768px) {
  .progress-keyword {
    font-size: 14px;
  }

  .pipeline-stage {
    gap: 10px;
    min-height: 36px;
  }

  .stage-indicator {
    width: 24px;
  }

  .stage-dot {
    width: 24px;
    height: 24px;

    .el-icon {
      font-size: 12px !important;
    }
  }

  .stage-label {
    font-size: 13px;
  }

  .stage-status {
    font-size: 10px;
  }

  .progress-empty {
    padding: 24px 12px;
  }

  .empty-text {
    font-size: 13px;
  }
}
</style>
