<template>
  <div class="workbench">
    <header class="workbench-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">内容工作台</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'OverviewReport' })">运营总览</el-button>
        <el-button text @click="router.push({ name: 'Settings' })">设置</el-button>
        <span v-if="authStore.user" class="header-user">{{ authStore.user.nickname }}</span>
        <el-button text @click="handleLogout">退出</el-button>
      </div>
    </header>

    <main class="workbench-main">
      <aside class="panel panel-left">
        <div class="panel-header">
          <h2 class="panel-title">任务配置</h2>
        </div>
        <div class="panel-body">
          <TaskForm
            :disabled="taskStore.isRunning"
            :submitting="taskStore.creating"
            @submit="handleCreateTask"
          />
        </div>
      </aside>

      <section class="panel panel-center">
        <div class="panel-header">
          <h2 class="panel-title">执行流程</h2>
        </div>
        <div class="panel-body">
          <TaskProgress
            :task="taskStore.currentTask"
            @retry="handleRetry"
          />
        </div>
      </section>

      <aside class="panel panel-right">
        <div class="panel-header">
          <div class="panel-header-row">
            <h2 class="panel-title">文章预览</h2>
            <el-tag v-if="taskStore.isDone" type="success" size="small" effect="plain">已完成</el-tag>
          </div>
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
              <el-icon :size="48" color="#D1D5DB"><Document /></el-icon>
            </div>
            <p class="placeholder-text">任务完成后将在此显示文章预览</p>
            <p class="placeholder-sub">支持逐段编辑、图片替换和一键发布</p>
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTaskStore } from '@/stores/task'
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

const currentDraft = ref<DraftVO | null>(null)

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
.workbench {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $color-bg;
}

.workbench-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 $spacing-xl;
  background-color: $color-card-bg;
  border-bottom: 1px solid $color-border;
  box-shadow: $shadow-card;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.header-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-bold;
  color: $color-primary;
}

.header-divider {
  color: $color-border;
}

.header-desc {
  font-size: $font-size-base;
  color: $color-text-secondary;
}

.header-user {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  margin-right: $spacing-sm;
}

.workbench-main {
  display: grid;
  grid-template-columns: 320px 1fr 420px;
  gap: $spacing-base;
  flex: 1;
  padding: $spacing-base;
  overflow: hidden;
}

.panel {
  display: flex;
  flex-direction: column;
  background-color: $color-card-bg;
  border-radius: $radius-lg;
  border: 1px solid $color-border;
  overflow: hidden;
}

.panel-header {
  padding: $spacing-base $spacing-lg;
  border-bottom: 1px solid $color-divider;
}

.panel-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.panel-body {
  flex: 1;
  padding: $spacing-lg;
  overflow-y: auto;
}

.panel-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-body-right {
  display: flex;
  flex-direction: column;
}

.right-content {
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

.publish-divider {
  height: 1px;
  background-color: $color-divider;
  margin: $spacing-sm 0;
}

.right-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 300px;
  text-align: center;
  padding: $spacing-2xl $spacing-lg;
}

.empty-preview-icon {
  margin-bottom: $spacing-lg;
  opacity: 0.5;
}

.placeholder-text {
  font-size: $font-size-base;
  color: $color-text-secondary;
  margin-bottom: $spacing-xs;
}

.placeholder-sub {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

// Responsive: Condensed two-column at 1024px
@media (max-width: $breakpoint-lg) {
  .workbench-main {
    grid-template-columns: 280px 1fr 360px;
  }
}

// Responsive: Stacked layout at 1024px
@media (max-width: $breakpoint-md) {
  .workbench-header {
    padding: 0 $spacing-base;
  }

  .header-desc {
    display: none;
  }

  .header-divider {
    display: none;
  }

  .workbench-main {
    grid-template-columns: 1fr;
    overflow-y: auto;
    padding: $spacing-sm;
    gap: $spacing-sm;
  }

  .panel {
    max-height: none;
  }

  .panel-body {
    padding: $spacing-base;
  }

  .right-placeholder {
    min-height: 200px;
    padding: $spacing-xl $spacing-base;
  }
}

// Responsive: Compact at 768px
@media (max-width: $breakpoint-sm) {
  .workbench-header {
    height: 48px;
    padding: 0 $spacing-sm;
  }

  .header-title {
    font-size: $font-size-md;
  }

  .header-actions {
    gap: 0;

    :deep(.el-button) {
      padding: $spacing-xs;
      font-size: $font-size-sm;
    }
  }

  .header-user {
    display: none;
  }

  .workbench-main {
    padding: $spacing-xs;
    gap: $spacing-xs;
  }

  .panel-header {
    padding: $spacing-sm $spacing-base;
  }

  .panel-title {
    font-size: $font-size-base;
  }

  .panel-body {
    padding: $spacing-sm;
  }
}
</style>
