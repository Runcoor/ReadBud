<template>
  <div class="task-progress">
    <!-- Task header info -->
    <div v-if="task" class="progress-header">
      <div class="progress-meta">
        <span class="progress-keyword">{{ task.keyword }}</span>
        <el-tag :type="STATUS_TAG_TYPES[task.status]" size="small" effect="plain">
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
        <el-icon :size="48" color="#D1D5DB"><Promotion /></el-icon>
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

defineEmits<{
  retry: [taskId: string]
}>()

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
  gap: $spacing-lg;
  height: 100%;
}

.progress-header {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.progress-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.progress-keyword {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.progress-bar-wrapper {
  :deep(.el-progress-bar__outer) {
    border-radius: 3px;
  }

  :deep(.el-progress-bar__inner) {
    border-radius: 3px;
    transition: width 0.6s ease;
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
  gap: $spacing-md;
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
  font-size: $font-size-xs;
  font-weight: $font-weight-semibold;
  transition: all $transition-base;
  flex-shrink: 0;
}

.stage-number {
  font-size: $font-size-xs;
}

.stage-line {
  width: 2px;
  flex: 1;
  min-height: 16px;
  transition: background-color $transition-base;
}

// Stage states
.pipeline-stage--pending {
  .stage-dot {
    background-color: $color-bg;
    border: 2px solid $color-border;
    color: $color-text-muted;
  }

  .stage-line {
    background-color: $color-border;
  }

  .stage-label {
    color: $color-text-muted;
  }
}

.pipeline-stage--active {
  .stage-dot {
    background-color: $color-accent;
    border: 2px solid $color-accent;
    color: #fff;
  }

  .stage-line {
    background-color: $color-border;
  }

  .stage-label {
    color: $color-text-primary;
    font-weight: $font-weight-semibold;
  }
}

.pipeline-stage--done {
  .stage-dot {
    background-color: $color-success;
    border: 2px solid $color-success;
    color: #fff;
  }

  .stage-line {
    background-color: $color-success;
  }

  .stage-label {
    color: $color-text-secondary;
  }
}

.stage-content {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding-top: 4px;
  min-height: 28px;
}

.stage-label {
  font-size: $font-size-base;
  transition: color $transition-base;
}

.stage-status {
  font-size: $font-size-xs;
  color: $color-accent;
}

.stage-status--done {
  color: $color-success;
}

.stage-spinning {
  animation: spin 1.2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// Error state
.progress-error {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
  margin-top: $spacing-base;
}

.retry-btn {
  align-self: flex-start;
}

// Empty state
.progress-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  text-align: center;
  padding: $spacing-3xl $spacing-lg;
}

.empty-icon {
  margin-bottom: $spacing-lg;
  opacity: 0.6;
}

.empty-text {
  font-size: $font-size-base;
  color: $color-text-secondary;
  margin-bottom: $spacing-xs;
}

.empty-sub {
  font-size: $font-size-sm;
  color: $color-text-muted;
}
</style>
