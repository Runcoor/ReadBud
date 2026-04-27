<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="workbench">
    <!-- 48px topbar -->
    <AppTopBar :crumb="topbarCrumb" user-initial="Y">
      <template #right>
        <!-- AUTOSAVED badge -->
        <div class="autosave-badge">
          <StatusDot kind="sprout" :size="6" />
          <span>AUTOSAVED · {{ autosaveAgo }}</span>
        </div>

        <div class="topbar-rail">
          <el-tooltip content="历史记录" placement="bottom">
            <button class="rail-icon" @click="showHistory = !showHistory">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            </button>
          </el-tooltip>
          <el-tooltip content="文章历史" placement="bottom">
            <button class="rail-icon" @click="router.push({ name: 'ArticleHistory' })">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
            </button>
          </el-tooltip>
          <el-tooltip content="品牌档案" placement="bottom">
            <button class="rail-icon" @click="router.push({ name: 'BrandProfiles' })">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
            </button>
          </el-tooltip>
          <el-tooltip content="运营总览" placement="bottom">
            <button class="rail-icon" @click="router.push({ name: 'OverviewReport' })">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 3v18h18"/><path d="M7 16l4-8 4 4 4-6"/></svg>
            </button>
          </el-tooltip>
          <el-tooltip content="系统设置" placement="bottom">
            <button class="rail-icon" @click="router.push({ name: 'Settings' })">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            </button>
          </el-tooltip>
          <el-tooltip :content="isDarkTheme ? '浅色模式' : '深色模式'" placement="bottom">
            <button class="rail-icon" @click="toggleTheme">
              <svg v-if="!isDarkTheme" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
            </button>
          </el-tooltip>
          <el-tooltip content="退出登录" placement="bottom">
            <button class="rail-icon" @click="handleLogout">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
            </button>
          </el-tooltip>
        </div>
      </template>
    </AppTopBar>

    <!-- Main 3-column layout -->
    <div class="workspace">
      <!-- COL 1: 320px config -->
      <aside class="col-config">
        <SectionLabel
          title="任务配置"
          code="01 · CONFIG"
          hint="描述清楚你想要什么，AI 会沿着主线完成全部 10 步流程"
        />
        <div class="config-form">
          <TaskForm
            :disabled="taskStore.isRunning"
            :submitting="taskStore.creating"
            @submit="handleCreateTask"
          />
        </div>
      </aside>

      <!-- COL 2: 280px pipeline -->
      <aside class="col-pipeline">
        <SectionLabel
          title="执行流程"
          code="02 · PIPELINE"
          :hint="pipelineHint"
        />

        <!-- Progress card -->
        <div class="progress-card">
          <div class="progress-card__head">
            <span class="progress-card__title">{{ progressTitle }}</span>
            <span class="progress-card__status">
              <StatusDot :kind="progressDotKind" :size="6" />
              <span>{{ progressStatusLabel }}</span>
            </span>
          </div>
          <div class="progress-card__bar">
            <div class="progress-card__fill" :style="{ width: progressPercent + '%' }" />
          </div>
          <div class="progress-card__meta">
            <span class="mono">{{ doneCount }}/{{ TASK_STAGES.length }}</span>
            <span class="mono">{{ elapsedLabel }}</span>
          </div>
          <div v-if="canCancel" class="progress-card__action">
            <button class="cancel-link" @click="handleCancel(taskStore.currentTask!.id)">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              取消任务
            </button>
          </div>
          <div v-if="taskStore.isFailed" class="progress-card__action">
            <button class="retry-link" @click="handleRetry(taskStore.currentTask!.id)">重试</button>
          </div>
        </div>

        <!-- Vertical timeline (shared component) -->
        <PipelineTimeline :task="taskStore.currentTask" />
      </aside>

      <!-- COL 3: preview -->
      <section class="col-preview">
        <div class="preview-header">
          <SectionLabel
            title="文章预览"
            code="03 · PREVIEW"
            hint="点击任一段落可在右侧打开编辑器进行调整"
          />
          <div class="preview-header__right">
            <PillTabs
              v-model="previewMode"
              :options="previewModeOptions"
              :compact="true"
            />
            <span class="preview-header__divider" />
            <button class="export-chip" @click="handleExport">
              导出
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="7" y1="17" x2="17" y2="7"/><polyline points="7 7 17 7 17 17"/></svg>
            </button>
          </div>
        </div>

        <div class="preview-body" :class="`preview-body--${previewMode}`">
          <!-- Has draft -->
          <template v-if="hasDraft">
            <div class="phone-frame" :class="`phone-frame--${previewMode}`">
              <span v-if="previewMode === 'mobile'" class="phone-frame__notch" />
              <div class="phone-frame__screen">
                <DraftPreview
                  ref="draftPreviewRef"
                  :draft-id="taskStore.currentTask!.result_draft_id!"
                  :mode="previewMode"
                />
              </div>
            </div>

            <div class="preview-extras">
              <div class="extras-divider" />
              <PublishPanel :draft="currentDraft" />
              <div class="extras-divider" />
              <DistributionPanel :draft-public-id="taskStore.currentTask!.result_draft_id!" />
            </div>
          </template>

          <!-- Loading state -->
          <div v-else-if="taskStore.isRunning" class="preview-empty">
            <span class="preview-empty__caption mono">GENERATING · 文章生成中，请稍候</span>
          </div>

          <!-- Empty state -->
          <div v-else class="preview-empty">
            <span class="preview-empty__caption mono">等待生成 · 任务完成后此处显示文章预览</span>
          </div>
        </div>
      </section>
    </div>

    <!-- History drawer -->
    <teleport to="body">
      <div v-if="showHistory" class="history-overlay" @click.self="showHistory = false">
        <aside class="history-drawer">
          <header class="drawer-header">
            <div class="drawer-header__title">
              <span>历史记录</span>
              <MonoChip kind="default">{{ historyTasks.length.toString().padStart(2, '0') }}</MonoChip>
            </div>
            <button class="rail-icon" @click="showHistory = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </header>

          <div class="drawer-search">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input v-model="historySearch" type="text" placeholder="搜索关键词" />
          </div>

          <div class="drawer-filters">
            <button
              v-for="opt in historyFilterOptions"
              :key="opt.value"
              class="filter-pill"
              :class="{ 'is-active': historyFilter === opt.value }"
              @click="historyFilter = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>

          <div class="drawer-body">
            <div v-if="historyLoading" class="drawer-state">加载中...</div>
            <div v-else-if="filteredHistory.length === 0" class="drawer-state">暂无历史记录</div>
            <div v-else class="history-list">
              <button
                v-for="task in filteredHistory"
                :key="task.id"
                class="history-item"
                @click="openHistoryPreview(task)"
              >
                <div class="history-item__main">
                  <span class="history-item__keyword">{{ task.keyword }}</span>
                  <span class="history-item__date mono">{{ formatDate(task.created_at) }}</span>
                </div>
                <div class="history-item__status">
                  <StatusDot :kind="task.status === 'done' ? 'sprout' : 'danger'" :size="6" />
                  <span>{{ task.status === 'done' ? '已完成' : '失败' }}</span>
                </div>
              </button>
            </div>
          </div>

          <footer class="drawer-footer mono">
            共 {{ historyTasks.length }} 条 · {{ historyDoneCount }} 已完成 · {{ historyFailedCount }} 失败
          </footer>
        </aside>
      </div>
    </teleport>

    <!-- Full-screen preview modal -->
    <teleport to="body">
      <div v-if="previewTask" class="fullscreen-preview-overlay">
        <div class="fullscreen-preview">
          <header class="preview-toolbar">
            <div class="preview-toolbar__left">
              <BrandLogo :size="18" :show-label="true" />
              <span class="preview-toolbar__sep" />
              <span class="preview-toolbar__crumb">历史记录 / {{ previewTask.keyword }}</span>
            </div>
            <div class="preview-toolbar__right">
              <span class="preview-toolbar__status">
                <StatusDot :kind="previewTask.status === 'done' ? 'sprout' : 'danger'" :size="6" />
                <span>{{ previewTask.status === 'done' ? '已完成' : '失败' }}</span>
              </span>
              <button class="rail-icon" @click="closePreview">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
          </header>
          <div class="preview-body-fs">
            <div class="preview-article">
              <DraftPreview v-if="previewTask.result_draft_id" :draft-id="previewTask.result_draft_id" />
              <div v-else class="drawer-state">该任务没有生成稿件</div>
            </div>
            <aside v-if="previewTask.result_draft_id && previewTask.status === 'done'" class="preview-sidebar">
              <PublishPanel :draft="previewDraft" />
              <div class="extras-divider" />
              <DistributionPanel :draft-public-id="previewTask.result_draft_id" />
            </aside>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTaskStore } from '@/stores/task'
import { useTheme } from '@/composables/useTheme'
import { ElMessage, ElMessageBox } from 'element-plus'
import TaskForm from '@/components/task/TaskForm.vue'
import DraftPreview from '@/components/task/DraftPreview.vue'
import PublishPanel from '@/components/task/PublishPanel.vue'
import DistributionPanel from '@/components/task/DistributionPanel.vue'
import AppTopBar from '@/components/common/AppTopBar.vue'
import BrandLogo from '@/components/common/BrandLogo.vue'
import SectionLabel from '@/components/common/SectionLabel.vue'
import PillTabs from '@/components/common/PillTabs.vue'
import StatusDot from '@/components/common/StatusDot.vue'
import MonoChip from '@/components/common/MonoChip.vue'
import PipelineTimeline from '@/components/common/PipelineTimeline.vue'
import { getDraft } from '@/api/draft'
import { listTasks } from '@/api/task'
import type { CreateTaskRequest, TaskVO } from '@/types/task'
import type { DraftVO } from '@/types/draft'

// === Pipeline definition ===
const TASK_STAGES = [
  { key: 'keyword_expand', label: '关键词扩展' },
  { key: 'source_search', label: '素材搜集' },
  { key: 'content_crawl', label: '内容采集' },
  { key: 'hot_score', label: '热度评分' },
  { key: 'article_write', label: '文案撰写' },
  { key: 'image_match', label: '图片匹配' },
  { key: 'chart_gen', label: '图表生成' },
  { key: 'html_compile', label: 'HTML 编译' },
  { key: 'self_check', label: '自检校对' },
  { key: 'publish', label: '发布' },
] as const

const router = useRouter()
const authStore = useAuthStore()
const taskStore = useTaskStore()
const { theme, toggle: toggleTheme } = useTheme()
const isDarkTheme = computed(() => theme.value === 'dark')

const currentDraft = ref<DraftVO | null>(null)

const showHistory = ref(false)
const historyTasks = ref<TaskVO[]>([])
const historyLoading = ref(false)
const historySearch = ref('')
const historyFilter = ref<'all' | 'done' | 'failed'>('all')
const previewTask = ref<TaskVO | null>(null)
const previewDraft = ref<DraftVO | null>(null)

// Preview mode + ref for export
type PreviewMode = 'mobile' | 'desktop' | 'markdown'
const previewMode = ref<PreviewMode>('mobile')
const previewModeOptions = [
  { label: '移动', value: 'mobile' as const },
  { label: '桌面', value: 'desktop' as const },
  { label: 'MD', value: 'markdown' as const },
]
const draftPreviewRef = ref<{ getMarkdown: () => string; getTitle: () => string } | null>(null)

// === Tick for elapsed/autosave ===
const now = ref(Date.now())
let tickerId: number | undefined
onMounted(() => {
  tickerId = window.setInterval(() => { now.value = Date.now() }, 1000)
})
onUnmounted(() => {
  if (tickerId) window.clearInterval(tickerId)
})

const lastSaveTs = ref(Date.now())
const autosaveAgo = computed(() => {
  const diff = Math.max(0, Math.floor((now.value - lastSaveTs.value) / 1000))
  if (diff < 60) return `${diff}s ago`
  const m = Math.floor(diff / 60)
  return `${m}m ago`
})

// === Topbar crumb ===
const topbarCrumb = computed(() => {
  const kw = taskStore.currentTask?.keyword
  return kw ? `新任务 / ${kw}` : '新任务'
})

// === Pipeline state derivation ===
const currentStageIndex = computed(() => {
  const stage = taskStore.currentTask?.current_stage
  if (!stage) return -1
  return TASK_STAGES.findIndex((s) => s.key === stage)
})

function stageState(idx: number): 'done' | 'active' | 'pending' {
  const task = taskStore.currentTask
  if (!task) return 'pending'
  if (task.status === 'done') return 'done'
  const cur = currentStageIndex.value
  if (cur < 0) return 'pending'
  if (idx < cur) return 'done'
  if (idx === cur && task.status === 'running') return 'active'
  return 'pending'
}

function stageMeta(idx: number): string {
  const state = stageState(idx)
  if (state === 'done') return '完成'
  if (state === 'active') return '执行中'
  return '—'
}

const doneCount = computed(() => {
  let c = 0
  for (let i = 0; i < TASK_STAGES.length; i++) {
    if (stageState(i) === 'done') c++
  }
  return c
})

const progressPercent = computed(() => {
  const t = taskStore.currentTask
  if (!t) return 0
  if (t.status === 'done') return 100
  return Math.max(0, Math.min(100, t.progress || 0))
})

const progressTitle = computed(() => {
  return taskStore.currentTask?.keyword || '尚未开始任务'
})

const progressStatusLabel = computed(() => {
  const s = taskStore.currentTask?.status
  if (!s) return '空闲'
  if (s === 'done') return '已完成'
  if (s === 'running') return '执行中'
  if (s === 'pending') return '排队中'
  if (s === 'failed') return '失败'
  if (s === 'cancelled') return '已取消'
  return s
})

const progressDotKind = computed<'sprout' | 'warn' | 'danger' | 'mute'>(() => {
  const s = taskStore.currentTask?.status
  if (s === 'done') return 'sprout'
  if (s === 'failed' || s === 'cancelled') return 'danger'
  if (s === 'running' || s === 'pending') return 'warn'
  return 'mute'
})

const pipelineHint = computed(() => {
  const t = taskStore.currentTask
  if (!t) return '等待启动 · 共 10 步'
  if (t.status === 'done') return '10 步全部完成'
  if (t.status === 'failed') return '执行失败'
  if (t.status === 'cancelled') return '已取消'
  return `${doneCount.value}/${TASK_STAGES.length} 进行中`
})

const elapsedLabel = computed(() => {
  const t = taskStore.currentTask
  if (!t?.created_at) return '00:00'
  const start = new Date(t.created_at).getTime()
  const diff = Math.max(0, Math.floor((now.value - start) / 1000))
  const m = Math.floor(diff / 60)
  const s = diff % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

const canCancel = computed(() => {
  const s = taskStore.currentTask?.status
  return s === 'running' || s === 'pending'
})

// === Draft loading ===
const hasDraft = computed(() => {
  return !!(taskStore.isDone && taskStore.currentTask?.result_draft_id)
})

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

watch(() => previewTask.value?.result_draft_id, async (draftId) => {
  if (draftId) {
    try {
      const res = await getDraft(draftId)
      previewDraft.value = res.data
    } catch {
      previewDraft.value = null
    }
  } else {
    previewDraft.value = null
  }
}, { immediate: true })

function closePreview() {
  previewTask.value = null
  previewDraft.value = null
}

// === History ===
const historyFilterOptions = [
  { label: '全部', value: 'all' as const },
  { label: '已完成', value: 'done' as const },
  { label: '失败', value: 'failed' as const },
]

const historyDoneCount = computed(() => historyTasks.value.filter(t => t.status === 'done').length)
const historyFailedCount = computed(() => historyTasks.value.filter(t => t.status === 'failed').length)

const filteredHistory = computed(() => {
  let list = historyTasks.value
  if (historyFilter.value !== 'all') {
    list = list.filter(t => t.status === historyFilter.value)
  }
  const q = historySearch.value.trim().toLowerCase()
  if (q) {
    list = list.filter(t => t.keyword.toLowerCase().includes(q))
  }
  return list
})

watch(showHistory, async (open) => {
  if (open) {
    historyLoading.value = true
    try {
      const res = await listTasks(1, 50)
      historyTasks.value = res.data.items.filter(t => t.status === 'done' || t.status === 'failed')
    } catch { /* ignore */ }
    finally { historyLoading.value = false }
  }
})

function openHistoryPreview(task: TaskVO) {
  previewTask.value = task
  showHistory.value = false
}

function formatDate(dateStr: string) {
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

// === Lifecycle ===
onMounted(() => { taskStore.init() })
onUnmounted(() => { taskStore.disconnectSSE() })

// === Action handlers ===
async function handleCreateTask(payload: CreateTaskRequest): Promise<void> {
  try {
    await taskStore.create(payload)
    lastSaveTs.value = Date.now()
    ElMessage.success('任务已创建，正在执行...')
  } catch {
    // handled by interceptor
  }
}

async function handleRetry(taskId: string): Promise<void> {
  try {
    await taskStore.retry(taskId)
    ElMessage.success('任务已重新提交')
  } catch { /* ignore */ }
}

async function handleCancel(taskId: string) {
  try {
    await ElMessageBox.confirm('确定取消当前任务？所有已生成的数据将被删除。', '取消任务', {
      confirmButtonText: '确定取消', cancelButtonText: '继续执行', type: 'warning',
    })
    await taskStore.cancel(taskId)
    ElMessage.success('任务已取消')
  } catch { /* user dismissed */ }
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
  } catch { /* user cancelled */ }
}

function handleExport() {
  const md = draftPreviewRef.value?.getMarkdown() || ''
  if (!md.trim()) {
    ElMessage.warning('暂无可导出的草稿内容')
    return
  }
  const rawTitle = draftPreviewRef.value?.getTitle() || 'draft'
  const safeTitle = rawTitle.replace(/[\\/:*?"<>|]+/g, '_').slice(0, 80) || 'draft'
  const blob = new Blob([md], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${safeTitle}.md`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success(`已导出 ${safeTitle}.md`)
}
</script>

<style scoped lang="scss">
@use '@/styles/tokens' as *;

.workbench {
  display: flex;
  flex-direction: column;
  // Bind to viewport so .workspace (flex:1, min-height:0) caps each column
  // and per-column overflow-y:auto activates instead of letting the document
  // scroll as a whole.
  height: 100vh;
  overflow: hidden;
  background: var(--brand-paper);
  font-family: var(--font-sans);
  color: var(--text-primary);
}

.mono {
  font-family: var(--font-mono);
  letter-spacing: 0.04em;
}

// === Topbar autosave + rail ===
.autosave-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-body);
  background: var(--brand-sprout-soft);
  padding: 4px 8px;
  border-radius: 3px;
  letter-spacing: 0.04em;
}

.topbar-rail {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.rail-icon {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  background: transparent;
  border: 1px solid var(--border-hair);
  color: var(--text-body);
  cursor: pointer;
  transition: all 120ms ease;
  border-radius: 0;

  &:hover {
    color: var(--text-primary);
    border-color: var(--border-medium);
    background: var(--brand-paper-warm);
  }
}

// === Main 3-column layout ===
.workspace {
  flex: 1;
  display: flex;
  align-items: stretch;
  min-height: 0;
}

.col-config {
  width: 320px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-hair);
  background: var(--surface-card);
  padding: 20px 22px;
  overflow-y: auto;
}

.col-pipeline {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-hair);
  background: var(--brand-paper);
  padding: 20px 22px;
  overflow-y: auto;
}

.col-preview {
  flex: 1;
  min-width: 0;
  background: var(--brand-paper);
  padding: 20px 32px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

// === Config form re-skin (light overrides) ===
.config-form {
  :deep(.el-form-item__label) {
    font-size: 11px;
    color: var(--text-tertiary);
    font-family: var(--font-mono);
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 0 0 6px 0;
    line-height: 1.4;
  }

  :deep(.form-section-title) {
    display: none; // hide subsection titles, SectionLabel handles top-level
  }

  :deep(.form-section) {
    margin-bottom: 16px;
  }

  :deep(.el-input__wrapper),
  :deep(.el-select .el-input__wrapper),
  :deep(.el-textarea__inner) {
    box-shadow: none !important;
    border: 1px solid var(--border-hair);
    border-radius: 3px;
    background: var(--surface-card);
    transition: border-color 120ms ease;

    &:hover {
      border-color: var(--border-medium);
    }

    &.is-focus, &:focus, &:focus-within {
      border-color: var(--brand-ink);
    }
  }

  :deep(.el-button--primary) {
    width: 100%;
    height: 40px;
    background: var(--brand-ink);
    border-color: var(--brand-ink);
    color: var(--text-inverse);
    border-radius: 3px;
    font-weight: 500;
    letter-spacing: 0.18em;
    font-size: 13px;
  }
}

// === Progress card ===
.progress-card {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  padding: 14px 14px 12px;
  margin-bottom: 18px;
  border-radius: 3px;

  &__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
    gap: 8px;
  }

  &__title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-body);
    flex-shrink: 0;
  }

  &__bar {
    height: 3px;
    background: var(--brand-paper-warm);
    border-radius: 1px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  &__fill {
    height: 100%;
    background: var(--brand-ink);
    transition: width 240ms ease;
  }

  &__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 11px;
    color: var(--text-tertiary);
  }

  &__action {
    margin-top: 10px;
    border-top: 1px solid var(--border-hair-soft);
    padding-top: 8px;
  }
}

.cancel-link, .retry-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  font-size: 11px;
  font-family: var(--font-mono);
  letter-spacing: 0.06em;
  cursor: pointer;
  padding: 0;

  &:hover {
    color: var(--brand-danger);
  }
}

.retry-link:hover {
  color: var(--brand-ink);
}

// === Preview header ===
.preview-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 4px;

  > :deep(.section-label) {
    flex: 1;
    margin-bottom: 0;
  }

  &__right {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  &__divider {
    width: 1px;
    height: 16px;
    background: var(--border-hair);
  }
}

.export-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--brand-sprout-soft);
  border: none;
  color: var(--brand-sprout);
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  padding: 5px 9px;
  border-radius: 3px;
  cursor: pointer;
  transition: opacity 120ms ease;

  &:hover { opacity: 0.85; }
}

// === Preview body ===
.preview-body {
  flex: 1;
  margin-top: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.phone-frame {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 32px;
  padding: 20px;
  position: relative;
  box-shadow: var(--shadow-md);
  width: 360px;
  max-width: 100%;
  flex-shrink: 0;

  &--desktop {
    width: 100%;
    max-width: 880px;
    border-radius: 6px;
    padding: 32px 40px;
  }

  &--markdown {
    width: 100%;
    max-width: 720px;
    border-radius: 6px;
    padding: 24px 28px;
    font-family: var(--font-mono);
  }

  &__notch {
    position: absolute;
    top: 8px;
    left: 50%;
    transform: translateX(-50%);
    width: 80px;
    height: 5px;
    background: var(--border-hair);
    border-radius: 3px;
    z-index: 2;
  }

  &__screen {
    margin-top: 14px;
    max-height: 720px;
    overflow-y: auto;
  }

  &--desktop &__screen,
  &--markdown &__screen {
    margin-top: 0;
    max-height: none;
  }
}

.preview-extras {
  width: 100%;
  max-width: 720px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.extras-divider {
  height: 1px;
  background: var(--border-hair);
  width: 100%;
}

.preview-empty {
  border: 1px dashed var(--border-hair);
  background: var(--surface-card);
  width: 100%;
  max-width: 480px;
  min-height: 320px;
  display: grid;
  place-items: center;
  border-radius: 4px;
  padding: 40px 20px;

  &__caption {
    font-size: 11px;
    color: var(--text-tertiary);
    letter-spacing: 0.06em;
    text-transform: uppercase;
    text-align: center;
  }
}

// === History drawer ===
.history-overlay {
  position: fixed;
  inset: 0;
  z-index: $z-modal;
  background: rgba(10, 10, 10, 0.32);
}

.history-drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 380px;
  max-width: 92vw;
  background: var(--surface-card);
  border-left: 1px solid var(--border-hair);
  display: flex;
  flex-direction: column;
  z-index: $z-modal;
  font-family: var(--font-sans);
  box-shadow: var(--shadow-drawer, -24px 0 48px -16px rgba(0,0,0,0.16));
  animation: drawer-slide 240ms ease;
}

@keyframes drawer-slide {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-hair);
  height: 48px;
  flex-shrink: 0;

  &__title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }
}

.drawer-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin: 14px 18px 0;
  border: 1px solid var(--border-hair);
  border-radius: 3px;
  color: var(--text-tertiary);

  input {
    flex: 1;
    border: none;
    outline: none;
    background: transparent;
    font-size: 12px;
    color: var(--text-primary);
    font-family: var(--font-sans);

    &::placeholder {
      color: var(--text-tertiary);
    }
  }
}

.drawer-filters {
  display: flex;
  gap: 6px;
  padding: 12px 18px 14px;
  border-bottom: 1px solid var(--border-hair);
}

.filter-pill {
  background: transparent;
  border: 1px solid var(--border-hair);
  color: var(--text-body);
  font-size: 11px;
  padding: 4px 10px;
  border-radius: 3px;
  cursor: pointer;
  font-family: var(--font-sans);
  transition: all 120ms ease;

  &:hover {
    border-color: var(--border-medium);
  }

  &.is-active {
    background: var(--brand-ink);
    color: var(--text-inverse);
    border-color: var(--brand-ink);
  }
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px 12px;
}

.drawer-state {
  padding: 40px 16px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 12px;
}

.history-list {
  display: flex;
  flex-direction: column;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 8px;
  border: none;
  border-bottom: 1px solid var(--border-hair-soft);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: background 120ms ease;

  &:hover {
    background: var(--brand-paper-warm);
  }

  &__main {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
    flex: 1;
  }

  &__keyword {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__date {
    font-size: 10px;
    color: var(--text-tertiary);
  }

  &__status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-body);
    flex-shrink: 0;
  }
}

.drawer-footer {
  padding: 12px 18px;
  border-top: 1px solid var(--border-hair);
  font-size: 10px;
  color: var(--text-tertiary);
  letter-spacing: 0.06em;
  text-align: center;
  flex-shrink: 0;
}

// === Fullscreen preview ===
.fullscreen-preview-overlay {
  position: fixed;
  inset: 0;
  z-index: $z-modal;
  background: var(--brand-paper);
  animation: fade-in 0.2s ease;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.fullscreen-preview {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--brand-paper);
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 48px;
  border-bottom: 1px solid var(--border-hair);
  background: var(--surface-card);
  flex-shrink: 0;

  &__left {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  &__sep {
    width: 1px;
    height: 16px;
    background: var(--border-hair);
  }

  &__crumb {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  &__right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-body);
  }
}

.preview-body-fs {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.preview-article {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  justify-content: center;
}

.preview-sidebar {
  width: 380px;
  flex-shrink: 0;
  border-left: 1px solid var(--border-hair);
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: var(--surface-card);
}

// === Responsive ===
@media (max-width: $breakpoint-md) {
  .workspace {
    flex-direction: column;
  }

  .col-config,
  .col-pipeline {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--border-hair);
  }

  .col-preview {
    padding: 20px 16px;
  }

  .preview-body-fs {
    flex-direction: column;
  }

  .preview-sidebar {
    width: 100%;
    border-left: none;
    border-top: 1px solid var(--border-hair);
    max-height: 40vh;
  }
}
</style>
