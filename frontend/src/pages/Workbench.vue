<template>
  <div class="workbench">
    <header class="workbench-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">内容工作台</span>
      </div>
      <div class="header-actions">
        <el-tooltip content="运营总览" placement="bottom">
          <button class="icon-btn" @click="router.push({ name: 'OverviewReport' })">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3v18h18"/><path d="M7 16l4-8 4 4 4-6"/></svg>
          </button>
        </el-tooltip>
        <el-tooltip content="系统设置" placement="bottom">
          <button class="icon-btn" @click="router.push({ name: 'Settings' })">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </button>
        </el-tooltip>
        <el-tooltip :content="isDarkTheme ? '浅色模式' : '深色模式'" placement="bottom">
          <button class="icon-btn" @click="toggleTheme">
            <svg v-if="!isDarkTheme" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
          </button>
        </el-tooltip>
        <el-tooltip content="退出登录" placement="bottom">
          <button class="icon-btn" @click="handleLogout">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
          </button>
        </el-tooltip>
      </div>
    </header>

    <main class="workbench-main">
      <div class="card-stack">
        <!-- Config card -->
        <div class="stack-card" :class="getCardClass('config')" @click="focusPanel = 'config'">
          <div class="card-inner">
            <div class="panel-header"><span class="panel-title">任务配置</span></div>
            <div class="panel-body">
              <TaskForm
                :disabled="taskStore.isRunning"
                :submitting="taskStore.creating"
                @submit="handleCreateTask"
              />
            </div>
          </div>
        </div>
        <!-- Execute card -->
        <div class="stack-card" :class="getCardClass('execute')" @click="focusPanel = 'execute'">
          <div class="card-inner">
            <div class="panel-header"><span class="panel-title">执行流程</span></div>
            <div class="panel-body">
              <TaskProgress
                :task="taskStore.currentTask"
                @retry="handleRetry"
                @cancel="handleCancel"
              />
            </div>
          </div>
        </div>
        <!-- Preview card -->
        <div class="stack-card" :class="getCardClass('preview')" @click="focusPanel = 'preview'">
          <div class="card-inner">
            <div class="panel-header">
              <span class="panel-title">文章预览</span>
              <el-tag v-if="taskStore.isDone" type="success" size="small">已完成</el-tag>
            </div>
            <div class="panel-body panel-body-right">
              <div v-if="taskStore.isDone && taskStore.currentTask?.result_draft_id" class="right-content">
                <DraftPreview :draft-id="taskStore.currentTask.result_draft_id" />
                <div class="publish-divider" />
                <PublishPanel :draft="currentDraft" />
                <div class="publish-divider" />
                <DistributionPanel :draft-public-id="taskStore.currentTask.result_draft_id" />
              </div>
              <div v-else-if="taskStore.isRunning" class="right-placeholder">
                <el-skeleton :rows="6" animated />
                <p class="placeholder-text">文章生成中，请稍候...</p>
              </div>
              <div v-else class="right-placeholder">
                <div class="empty-preview-icon">
                  <el-icon :size="48" color="#d4d4d4"><Document /></el-icon>
                </div>
                <p class="placeholder-text">任务完成后将在此显示文章预览</p>
                <p class="placeholder-sub">支持逐段编辑、图片替换和一键发布</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTaskStore } from '@/stores/task'
import { useTheme } from '@/composables/useTheme'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document } from '@element-plus/icons-vue'
import TaskForm from '@/components/task/TaskForm.vue'
import TaskProgress from '@/components/task/TaskProgress.vue'
import DraftPreview from '@/components/task/DraftPreview.vue'
import PublishPanel from '@/components/task/PublishPanel.vue'
import DistributionPanel from '@/components/task/DistributionPanel.vue'
import { getDraft } from '@/api/draft'
import type { CreateTaskRequest } from '@/types/task'
import type { DraftVO } from '@/types/draft'

const router = useRouter()
const authStore = useAuthStore()
const taskStore = useTaskStore()
const { theme, toggle: toggleTheme } = useTheme()
const isDarkTheme = computed(() => theme.value === 'dark')

const currentDraft = ref<DraftVO | null>(null)

type FocusPanel = 'config' | 'execute' | 'preview'
const focusPanel = ref<FocusPanel>('config')

// Auto-focus follows task state
watch(() => taskStore.currentTask?.status, (status) => {
  if (!status) { focusPanel.value = 'config'; return }
  if (status === 'pending' || status === 'running') focusPanel.value = 'execute'
  else if (status === 'done') focusPanel.value = 'preview'
})

function getCardClass(panel: FocusPanel) {
  if (focusPanel.value === panel) return 'stack-card--focus'
  const panels: FocusPanel[] = ['config', 'execute', 'preview']
  const focusIdx = panels.indexOf(focusPanel.value)
  const panelIdx = panels.indexOf(panel)
  return panelIdx < focusIdx ? 'stack-card--left' : 'stack-card--right'
}

// Init: recover running task on mount
onMounted(() => { taskStore.init() })

// Fetch draft when task completes
watch(
  () => taskStore.currentTask?.result_draft_id,
  async (draftId) => {
    if (draftId) {
      try {
        const res = await getDraft(draftId)
        currentDraft.value = res.data
      } catch {
        currentDraft.value = null
      }
    } else {
      currentDraft.value = null
    }
  },
  { immediate: true },
)

async function handleCreateTask(payload: CreateTaskRequest): Promise<void> {
  try {
    await taskStore.create(payload)
    ElMessage.success('任务已创建，正在执行...')
  } catch {
    // Error handled by request interceptor
  }
}

async function handleRetry(taskId: string): Promise<void> {
  try {
    await taskStore.retry(taskId)
    ElMessage.success('任务已重新提交')
  } catch {
    // Error handled by request interceptor
  }
}

async function handleCancel(taskId: string) {
  try {
    await ElMessageBox.confirm('确定取消当前任务？所有已生成的数据将被删除。', '取消任务', {
      confirmButtonText: '确定取消', cancelButtonText: '继续执行', type: 'warning'
    })
    await taskStore.cancel(taskId)
    ElMessage.success('任务已取消')
  } catch { /* user cancelled dialog */ }
}

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定退出登录？', '退出', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info',
    })
    authStore.logout()
    router.push({ name: 'Login' })
  } catch {
    // User cancelled
  }
}

onUnmounted(() => {
  taskStore.disconnectSSE()
})
</script>

<style lang="scss" scoped>
@use '@/styles/tokens' as *;

.workbench {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--surface-secondary);
}

.workbench-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 32px;
  @include glass-panel-solid;
  border-radius: 0;
  border-left: none;
  border-right: none;
  border-top: none;
  position: sticky;
  top: 0;
  z-index: $z-sticky;
}

.header-brand { display: flex; align-items: center; gap: 12px; }
.header-title { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.header-divider { color: var(--border-medium); }
.header-desc { font-size: 14px; color: var(--text-secondary); }
.header-actions { display: flex; align-items: center; gap: 4px; }

.icon-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  border-radius: $radius-md;
  cursor: pointer;
  transition: all $transition-base;
  &:hover {
    color: var(--text-primary);
    background: var(--surface-tertiary);
  }
}

.workbench-main {
  flex: 1;
  padding: 24px;
  overflow: hidden;
}

.card-stack {
  display: flex;
  align-items: stretch;
  height: calc(100vh - 60px - 48px);
  gap: 12px;
}

.stack-card {
  transition: all 0.45s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  overflow: hidden;

  &--focus {
    width: 50%;
    z-index: 3;
    .card-inner {
      @include glass-panel;
      transform: scale(1);
      opacity: 1;
      box-shadow: var(--shadow-lg);
    }
  }

  &--left, &--right {
    width: 25%;
    z-index: 1;
    .card-inner {
      @include glass-panel;
      transform: scale(0.96);
      opacity: 0.7;
    }
  }
}

.card-inner {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.45s cubic-bezier(0.4, 0, 0.2, 1);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-light);
}

.panel-title { font-size: 15px; font-weight: 600; color: var(--text-primary); }

.panel-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.panel-body-right { display: flex; flex-direction: column; }

.right-content { display: flex; flex-direction: column; gap: 20px; }

.publish-divider { height: 1px; background: var(--border-light); margin: 8px 0; }

.right-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 300px;
  text-align: center;
  padding: 48px 20px;
}

.empty-preview-icon { margin-bottom: 20px; opacity: 0.5; }
.placeholder-text { font-size: 14px; color: var(--text-secondary); margin-bottom: 4px; }
.placeholder-sub { font-size: 13px; color: var(--text-tertiary); }

// El-tag overrides
:deep(.el-tag) { border-radius: 4px; border: none; }
:deep(.el-tag--success) { background: #f0fdf4; color: #16a34a; }
:deep(.el-skeleton) { --el-skeleton-color: var(--surface-tertiary); --el-skeleton-to-color: var(--border-light); }

// Responsive
@media (max-width: $breakpoint-md) {
  .workbench-header { padding: 0 16px; }
  .header-desc, .header-divider { display: none; }
  .card-stack { flex-direction: column; height: auto; gap: 8px; }
  .stack-card { width: 100% !important; }
  .stack-card--focus { min-height: 50vh; }
  .stack-card--focus .card-inner { transform: scale(1); opacity: 1; }
  .stack-card--left .card-inner,
  .stack-card--right .card-inner { transform: scale(0.98); opacity: 0.7; max-height: 120px; overflow: hidden; }
  .workbench-main { padding: 16px; }
  .panel-body { padding: 16px; }
}

@media (max-width: $breakpoint-sm) {
  .workbench-header { height: 52px; padding: 0 12px; }
  .header-title { font-size: 16px; }
  .workbench-main { padding: 8px; }
  .panel-header { padding: 12px 16px; }
  .panel-title { font-size: 14px; }
  .panel-body { padding: 12px; }
}
</style>
