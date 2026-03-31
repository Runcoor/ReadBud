<template>
  <div class="task-detail">
    <header class="detail-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">任务详情</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Workbench' })">返回工作台</el-button>
      </div>
    </header>

    <main v-if="loading" class="detail-main">
      <el-skeleton :rows="10" animated />
    </main>

    <main v-else-if="task" class="detail-main">
      <!-- Task info card -->
      <section class="info-card">
        <div class="info-header">
          <div class="info-title-row">
            <h2 class="info-keyword">{{ task.keyword }}</h2>
            <el-tag :type="statusTagType" effect="plain">{{ statusLabel }}</el-tag>
          </div>
          <p class="info-meta">
            任务编号: {{ task.task_no }} · 创建于 {{ formatDateTime(task.created_at) }}
          </p>
        </div>
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
            <span class="field-value">{{ task.target_words || '—' }}</span>
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

        <!-- Progress bar -->
        <div class="info-progress">
          <el-progress
            :percentage="task.progress"
            :status="progressStatus"
            :stroke-width="8"
          />
        </div>

        <!-- Retry button for failed tasks -->
        <div v-if="task.status === 'failed'" class="info-error">
          <el-alert :title="task.error_message || '任务执行失败'" type="error" :closable="false" show-icon />
          <el-button type="primary" plain size="small" @click="handleRetry">重新执行</el-button>
        </div>
      </section>

      <!-- Two-column layout: sources + preview -->
      <div class="detail-columns">
        <!-- Source articles -->
        <section class="column-card">
          <div class="card-header">
            <h3 class="card-title">来源文章</h3>
            <el-tag size="small" type="info" effect="plain">{{ sources.length }} 篇</el-tag>
          </div>
          <div v-if="sourcesLoading" class="card-body">
            <el-skeleton :rows="4" animated />
          </div>
          <div v-else-if="sources.length === 0" class="card-body">
            <el-empty description="暂无来源文章" :image-size="60" />
          </div>
          <div v-else class="card-body source-list">
            <div
              v-for="src in sources"
              :key="src.id"
              class="source-item"
            >
              <div class="source-title-row">
                <el-tag :type="sourceTypeTag(src.source_type)" size="small" effect="plain">
                  {{ sourceTypeLabel(src.source_type) }}
                </el-tag>
                <a :href="src.source_url" target="_blank" rel="noopener" class="source-title">
                  {{ src.title }}
                </a>
              </div>
              <div class="source-meta">
                <span v-if="src.site_name">{{ src.site_name }}</span>
                <span v-if="src.author">· {{ src.author }}</span>
                <span v-if="src.published_at">· {{ src.published_at }}</span>
              </div>
              <div class="source-scores">
                <div class="score-bar">
                  <span class="score-label">热度</span>
                  <el-progress
                    :percentage="Math.min(src.hot_score, 100)"
                    :stroke-width="4"
                    :show-text="false"
                    color="#5B8DEF"
                  />
                  <span class="score-value">{{ src.hot_score.toFixed(1) }}</span>
                </div>
                <div class="score-bar">
                  <span class="score-label">相关</span>
                  <el-progress
                    :percentage="Math.min(src.relevance_score, 100)"
                    :stroke-width="4"
                    :show-text="false"
                    color="#52C41A"
                  />
                  <span class="score-value">{{ src.relevance_score.toFixed(1) }}</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Draft preview -->
        <section class="column-card">
          <div class="card-header">
            <h3 class="card-title">文章预览</h3>
          </div>
          <div class="card-body">
            <DraftPreview :draft-id="task.result_draft_id || null" />
          </div>
        </section>
      </div>
    </main>

    <main v-else class="detail-main">
      <el-empty description="任务不存在或已删除" />
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
import type { TaskVO } from '@/types/task'
import type { SourceVO } from '@/types/draft'
import { IMAGE_MODE_LABELS, PUBLISH_MODE_LABELS, STATUS_LABELS, STATUS_TAG_TYPES } from '@/types/task'

const route = useRoute()
const router = useRouter()

const task = ref<TaskVO | null>(null)
const loading = ref(true)
const sources = ref<SourceVO[]>([])
const sourcesLoading = ref(false)

const statusTagType = computed(() => {
  if (!task.value) return 'info'
  return STATUS_TAG_TYPES[task.value.status] || 'info'
})

const statusLabel = computed(() => {
  if (!task.value) return ''
  return STATUS_LABELS[task.value.status] || task.value.status
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

function formatDateTime(s: string): string {
  const d = new Date(s)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function sourceTypeLabel(t: string): string {
  const map: Record<string, string> = { web: '网页', news: '新闻', wechat: '公众号', blog: '博客' }
  return map[t] || t
}

function sourceTypeTag(t: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const map: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    web: 'info', news: '', wechat: 'success', blog: 'warning',
  }
  return map[t] || 'info'
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

onMounted(async () => {
  await fetchTask()
  if (task.value) {
    fetchSources()
  }
})
</script>

<style lang="scss" scoped>
.task-detail {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $color-bg;
}

.detail-header {
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

.detail-main {
  max-width: 1280px;
  width: 100%;
  margin: 0 auto;
  padding: $spacing-xl;
  display: flex;
  flex-direction: column;
  gap: $spacing-xl;
}

// Info card
.info-card {
  background: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-lg;
  padding: $spacing-xl;
}

.info-header {
  margin-bottom: $spacing-lg;
}

.info-title-row {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  margin-bottom: $spacing-xs;
}

.info-keyword {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.info-meta {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.info-fields {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: $spacing-md $spacing-xl;
  margin-bottom: $spacing-lg;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.field-label {
  font-size: $font-size-xs;
  color: $color-text-muted;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.field-value {
  font-size: $font-size-base;
  color: $color-text-primary;
}

.info-progress {
  margin-bottom: $spacing-md;
}

.info-error {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
  margin-top: $spacing-md;

  .el-button {
    align-self: flex-start;
  }
}

// Two-column layout
.detail-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-xl;
}

.column-card {
  background: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-lg;
  display: flex;
  flex-direction: column;
  max-height: 70vh;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-base $spacing-lg;
  border-bottom: 1px solid $color-divider;
}

.card-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.card-body {
  flex: 1;
  padding: $spacing-lg;
  overflow-y: auto;
}

// Source list
.source-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-base;
}

.source-item {
  padding: $spacing-md;
  border: 1px solid $color-divider;
  border-radius: $radius-base;
  transition: border-color $transition-base;

  &:hover {
    border-color: $color-accent;
  }
}

.source-title-row {
  display: flex;
  align-items: flex-start;
  gap: $spacing-sm;
  margin-bottom: $spacing-xs;
}

.source-title {
  font-size: $font-size-base;
  color: $color-text-primary;
  text-decoration: none;
  line-height: $line-height-tight;

  &:hover {
    color: $color-accent;
  }
}

.source-meta {
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin-bottom: $spacing-sm;
}

.source-scores {
  display: flex;
  gap: $spacing-lg;
}

.score-bar {
  flex: 1;
  display: flex;
  align-items: center;
  gap: $spacing-xs;

  .el-progress {
    flex: 1;
  }
}

.score-label {
  font-size: $font-size-xs;
  color: $color-text-muted;
  width: 28px;
  flex-shrink: 0;
}

.score-value {
  font-size: $font-size-xs;
  color: $color-text-secondary;
  width: 28px;
  text-align: right;
  flex-shrink: 0;
}

@media (max-width: $breakpoint-md) {
  .detail-columns {
    grid-template-columns: 1fr;
  }

  .info-fields {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: $breakpoint-sm) {
  .detail-main {
    padding: $spacing-base;
  }

  .info-fields {
    grid-template-columns: 1fr;
  }
}
</style>
