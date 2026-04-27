<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
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
      <div class="mono-material-card">
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
      <div class="mono-material-card">
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
      <div class="mono-material-card">
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
      <div class="mono-material-card">
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
      <div class="mono-material-card">
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
  gap: 16px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: #0a0a0a;
  margin: 0;
}

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
}

:deep(.el-button--danger) { color: #ef4444 !important; }

:deep(.el-button) {
  color: #525252 !important;
  &:hover { color: #0a0a0a !important; }
}

.panel-loading { padding: 24px 0; }

:deep(.el-skeleton) {
  --el-skeleton-color: #f5f5f5;
  --el-skeleton-to-color: #e8e8e8;
}

.panel-generating {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 48px 0;
}

.generating-icon {
  color: #525252;
  animation: spin 1.2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.generating-text {
  font-size: 14px;
  font-weight: 500;
  color: #0a0a0a;
  margin: 0;
}

.generating-hint {
  font-size: 13px;
  color: #d4d4d4;
  margin: 0;
}

.panel-empty { padding: 32px 0; }

:deep(.el-empty__description p) { color: #525252 !important; }

.empty-text {
  font-size: 13px;
  color: #525252;
  margin: 0 0 4px;
}

.empty-hint {
  font-size: 12px;
  color: #d4d4d4;
  margin: 0;
}

.panel-error { padding: 24px 0; }

.panel-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mono-material-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px;
  background: var(--surface-card);
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  transition: all 0.15s ease;

  &:hover {
    border-color: #0a0a0a;
  }
}

.material-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.material-label {
  font-size: 11px;
  font-weight: 600;
  color: #525252;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.material-text {
  font-size: 13px;
  color: #1a1a1a;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;

  &--highlight {
    font-weight: 500;
    color: #0a0a0a;
    padding: 10px 14px;
    background: #f5f5f5;
    border-radius: 8px;
    border-left: 3px solid #0a0a0a;
  }

  &--suggestion {
    color: #525252;
    font-size: 12px;
    line-height: 1.6;
  }
}

.panel-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px solid #e8e8e8;
}

.footer-time {
  font-size: 12px;
  color: #d4d4d4;
}

@media (max-width: 768px) {
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .mono-material-card {
    padding: 10px;
  }
}
</style>
