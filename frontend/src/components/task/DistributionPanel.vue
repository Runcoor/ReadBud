<template>
  <div class="distribution-panel">
    <!-- Header -->
    <div class="panel-header">
      <h3 class="panel-title">分发素材包</h3>
      <el-button
        type="primary"
        size="small"
        :loading="generating"
        :disabled="!draftPublicId"
        @click="handleGenerate"
      >
        {{ distribution ? '重新生成' : '生成素材包' }}
      </el-button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="panel-loading">
      <el-skeleton :rows="6" animated />
    </div>

    <!-- Generating -->
    <div v-else-if="generating" class="panel-generating">
      <el-icon class="generating-icon" :size="24">
        <Loading />
      </el-icon>
      <p class="generating-text">正在生成分发素材包...</p>
      <p class="generating-hint">基于文章内容生成社群文案、朋友圈文案等分发素材</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="!distribution" class="panel-empty">
      <el-empty :image-size="64">
        <template #description>
          <p class="empty-text">暂无分发素材包</p>
          <p class="empty-hint">点击上方按钮，自动生成适合各渠道的分发文案</p>
        </template>
      </el-empty>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="panel-error">
      <el-result icon="warning" :sub-title="error">
        <template #extra>
          <el-button size="small" @click="handleGenerate">重试</el-button>
        </template>
      </el-result>
    </div>

    <!-- Content -->
    <div v-else class="panel-content">
      <!-- Community Copy -->
      <div class="material-card">
        <div class="material-header">
          <span class="material-label">社群文案</span>
          <el-tooltip content="复制" placement="top">
            <el-button
              text
              size="small"
              :icon="CopyDocument"
              @click="copyText(distribution.community_copy, '社群文案')"
            />
          </el-tooltip>
        </div>
        <p class="material-text">{{ distribution.community_copy }}</p>
      </div>

      <!-- Moments Copy -->
      <div class="material-card">
        <div class="material-header">
          <span class="material-label">朋友圈文案</span>
          <el-tooltip content="复制" placement="top">
            <el-button
              text
              size="small"
              :icon="CopyDocument"
              @click="copyText(distribution.moments_copy, '朋友圈文案')"
            />
          </el-tooltip>
        </div>
        <p class="material-text">{{ distribution.moments_copy }}</p>
      </div>

      <!-- Summary Card -->
      <div class="material-card">
        <div class="material-header">
          <span class="material-label">摘要卡片</span>
          <el-tooltip content="复制" placement="top">
            <el-button
              text
              size="small"
              :icon="CopyDocument"
              @click="copyText(distribution.summary_card, '摘要卡片')"
            />
          </el-tooltip>
        </div>
        <p class="material-text material-text--highlight">{{ distribution.summary_card }}</p>
      </div>

      <!-- Comment Guide -->
      <div class="material-card">
        <div class="material-header">
          <span class="material-label">评论区引导语</span>
          <el-tooltip content="复制" placement="top">
            <el-button
              text
              size="small"
              :icon="CopyDocument"
              @click="copyText(distribution.comment_guide, '评论区引导语')"
            />
          </el-tooltip>
        </div>
        <p class="material-text">{{ distribution.comment_guide }}</p>
      </div>

      <!-- Next Topic Suggestion -->
      <div class="material-card">
        <div class="material-header">
          <span class="material-label">下篇选题建议</span>
          <el-tooltip content="复制" placement="top">
            <el-button
              text
              size="small"
              :icon="CopyDocument"
              @click="copyText(distribution.next_topic_suggestion, '选题建议')"
            />
          </el-tooltip>
        </div>
        <p class="material-text material-text--suggestion">{{ distribution.next_topic_suggestion }}</p>
      </div>

      <!-- Footer info -->
      <div class="panel-footer">
        <span class="footer-time">生成于 {{ formatTime(distribution.updated_at) }}</span>
        <el-button
          text
          size="small"
          type="danger"
          @click="handleDelete"
        >
          删除
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch, onMounted } from 'vue'
import { CopyDocument, Loading } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useDistribution } from '@/composables/useDistribution'

interface Props {
  draftPublicId: string | null
}

const props = defineProps<Props>()

const {
  distribution,
  loading,
  generating,
  error,
  loadByDraft,
  generate,
  remove,
} = useDistribution()

async function handleGenerate() {
  if (!props.draftPublicId) return
  await generate(props.draftPublicId)
}

async function handleDelete() {
  if (!distribution.value) return

  try {
    await ElMessageBox.confirm(
      '确定删除分发素材包？删除后需要重新生成。',
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
    await remove(distribution.value.public_id)
    ElMessage.success('分发素材包已删除')
  } catch {
    // User cancelled
  }
}

function copyText(text: string, label: string) {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success(`${label}已复制到剪贴板`)
  }).catch(() => {
    ElMessage.warning('复制失败，请手动复制')
  })
}

function formatTime(dateStr: string): string {
  const d = new Date(dateStr)
  const month = d.getMonth() + 1
  const day = d.getDate()
  const hour = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  return `${month}月${day}日 ${hour}:${min}`
}

onMounted(() => {
  if (props.draftPublicId) {
    loadByDraft(props.draftPublicId)
  }
})

watch(() => props.draftPublicId, (newId) => {
  if (newId) {
    loadByDraft(newId)
  }
})
</script>

<style lang="scss" scoped>
.distribution-panel {
  display: flex;
  flex-direction: column;
  gap: $spacing-base;
}

// --- Header ---
.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0;
}

// --- Loading / Empty / Generating ---
.panel-loading {
  padding: $spacing-xl 0;
}

.panel-generating {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-3xl 0;
}

.generating-icon {
  color: $color-accent;
  animation: spin 1.2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.generating-text {
  font-size: $font-size-base;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  margin: 0;
}

.generating-hint {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin: 0;
}

.panel-empty {
  padding: $spacing-2xl 0;
}

.empty-text {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  margin: 0 0 $spacing-xs;
}

.empty-hint {
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin: 0;
}

.panel-error {
  padding: $spacing-xl 0;
}

// --- Content ---
.panel-content {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

// --- Material Card ---
.material-card {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  padding: $spacing-md;
  background-color: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-lg;
  transition: border-color $transition-fast;

  &:hover {
    border-color: rgba($color-accent, 0.3);
  }
}

.material-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.material-label {
  font-size: $font-size-xs;
  font-weight: $font-weight-semibold;
  color: $color-primary;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.material-text {
  font-size: $font-size-sm;
  color: $color-text-primary;
  line-height: $line-height-relaxed;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;

  &--highlight {
    font-weight: $font-weight-medium;
    color: $color-primary;
    padding: $spacing-sm $spacing-md;
    background-color: rgba($color-primary, 0.04);
    border-radius: $radius-sm;
    border-left: 3px solid $color-primary;
  }

  &--suggestion {
    color: $color-text-secondary;
    font-size: $font-size-xs;
    line-height: $line-height-relaxed;
  }
}

// --- Footer ---
.panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: $spacing-sm;
  border-top: 1px solid $color-divider;
}

.footer-time {
  font-size: $font-size-xs;
  color: $color-text-muted;
}

// --- Responsive ---
@media (max-width: $breakpoint-sm) {
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-sm;
  }

  .material-card {
    padding: $spacing-sm;
  }
}
</style>
