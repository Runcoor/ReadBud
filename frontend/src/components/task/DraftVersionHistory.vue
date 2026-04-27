<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="version-history">
    <div class="version-header">
      <h4 class="version-title">版本历史</h4>
      <el-button
        size="small"
        type="primary"
        plain
        :loading="snapshotting"
        @click="showSnapshotDialog = true"
      >
        保存版本
      </el-button>
    </div>

    <div v-if="loading" class="version-loading">
      <el-skeleton :rows="3" animated />
    </div>

    <div v-else-if="error" class="version-error">
      <el-alert type="error" :title="error" :closable="false" show-icon />
      <el-button size="small" type="primary" plain @click="loadVersions">重试</el-button>
    </div>

    <el-empty
      v-else-if="versions.length === 0"
      description="暂无版本记录"
      :image-size="60"
    />

    <el-timeline v-else class="version-timeline">
      <el-timeline-item
        v-for="version in versions"
        :key="version.id"
        :timestamp="formatTime(version.created_at)"
        placement="top"
        :type="version.id === currentVersionId ? 'primary' : undefined"
      >
        <div
          class="glass-version-card"
          :class="{ 'glass-version-card--current': version.id === currentVersionId }"
        >
          <div class="version-card-header">
            <span class="version-no">v{{ version.version_no }}</span>
            <el-tag
              v-if="version.id === currentVersionId"
              size="small"
              type="success"
            >
              当前版本
            </el-tag>
          </div>
          <p class="version-card-title">{{ version.title }}</p>
          <p v-if="version.change_reason" class="version-card-reason">
            {{ version.change_reason }}
          </p>
          <div class="version-card-actions">
            <el-button
              size="small"
              text
              type="primary"
              @click="handlePreview(version)"
            >
              查看
            </el-button>
            <el-button
              v-if="version.id !== currentVersionId"
              size="small"
              text
              type="warning"
              @click="confirmRollback(version)"
            >
              回滚
            </el-button>
          </div>
        </div>
      </el-timeline-item>
    </el-timeline>

    <!-- Snapshot Dialog -->
    <el-dialog
      v-model="showSnapshotDialog"
      title="保存当前版本"
      width="400"
      class="version-dialog"
    >
      <el-form label-position="top">
        <el-form-item label="变更说明">
          <el-input
            v-model="snapshotReason"
            type="textarea"
            :rows="3"
            placeholder="描述本次修改的内容..."
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSnapshotDialog = false">取消</el-button>
        <el-button type="primary" :loading="snapshotting" @click="handleSnapshot">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- Preview Dialog -->
    <el-dialog
      v-model="showPreviewDialog"
      :title="`版本 v${previewVersion?.version_no ?? ''} 预览`"
      width="680"
      class="version-dialog"
    >
      <div v-if="previewLoading" class="preview-loading">
        <el-skeleton :rows="6" animated />
      </div>
      <div v-else-if="previewDetail" class="preview-content">
        <div class="preview-meta">
          <p><strong>标题：</strong>{{ previewDetail.title }}</p>
          <p><strong>摘要：</strong>{{ previewDetail.digest }}</p>
        </div>
        <el-divider />
        <div class="preview-blocks">
          <div
            v-for="block in previewDetail.blocks"
            :key="block.id"
            class="preview-block"
          >
            <el-tag size="small" class="block-type-tag">{{ block.block_type }}</el-tag>
            <h5 v-if="block.heading" class="block-heading">{{ block.heading }}</h5>
            <p v-if="block.text_md" class="block-text">{{ block.text_md }}</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showPreviewDialog = false">关闭</el-button>
        <el-button
          v-if="previewVersion && previewVersion.id !== currentVersionId"
          type="warning"
          @click="confirmRollback(previewVersion!)"
        >
          回滚到此版本
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listDraftVersions,
  getDraftVersion,
  createSnapshot,
  rollbackVersion,
} from '@/api/version'
import type { DraftVersionVO, DraftVersionDetailVO } from '@/types/version'

const props = defineProps<{
  draftId: string
}>()

const emit = defineEmits<{
  (e: 'rollback'): void
}>()

const loading = ref(false)
const error = ref<string | null>(null)
const versions = ref<DraftVersionVO[]>([])
const currentVersionId = ref<string | null>(null)

const showSnapshotDialog = ref(false)
const snapshotReason = ref('')
const snapshotting = ref(false)

const showPreviewDialog = ref(false)
const previewVersion = ref<DraftVersionVO | null>(null)
const previewDetail = ref<DraftVersionDetailVO | null>(null)
const previewLoading = ref(false)

function formatTime(isoStr: string): string {
  const d = new Date(isoStr)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function loadVersions() {
  if (!props.draftId) return
  loading.value = true
  error.value = null
  try {
    const resp = await listDraftVersions(props.draftId)
    if (resp.code === 0) {
      versions.value = resp.data || []
      if (versions.value.length > 0) {
        currentVersionId.value = versions.value[0].id
      }
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载版本历史失败'
  } finally {
    loading.value = false
  }
}

async function handleSnapshot() {
  snapshotting.value = true
  try {
    const resp = await createSnapshot(props.draftId, {
      change_reason: snapshotReason.value || '手动保存',
    })
    if (resp.code === 0) {
      ElMessage.success('版本已保存')
      showSnapshotDialog.value = false
      snapshotReason.value = ''
      await loadVersions()
    }
  } catch {
    ElMessage.error('保存版本失败')
  } finally {
    snapshotting.value = false
  }
}

async function handlePreview(version: DraftVersionVO) {
  previewVersion.value = version
  showPreviewDialog.value = true
  previewLoading.value = true
  previewDetail.value = null
  try {
    const resp = await getDraftVersion(props.draftId, version.id)
    if (resp.code === 0) {
      previewDetail.value = resp.data
    }
  } catch {
    ElMessage.error('加载版本详情失败')
  } finally {
    previewLoading.value = false
  }
}

async function confirmRollback(version: DraftVersionVO) {
  try {
    await ElMessageBox.confirm(
      `确定要将草稿回滚到版本 v${version.version_no} 吗？此操作会覆盖当前内容。`,
      '确认回滚',
      { type: 'warning', confirmButtonText: '确定回滚', cancelButtonText: '取消' },
    )
    const resp = await rollbackVersion(props.draftId, version.id)
    if (resp.code === 0) {
      ElMessage.success('回滚成功')
      showPreviewDialog.value = false
      await loadVersions()
      emit('rollback')
    }
  } catch (e: unknown) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error('回滚失败')
    }
  }
}

watch(() => props.draftId, () => {
  loadVersions()
})

onMounted(() => {
  loadVersions()
})
</script>

<style lang="scss" scoped>
.version-history {
  padding: 16px;
}

.version-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.version-title {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.version-loading,
.version-error {
  padding: 20px 0;
}

.version-error {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-start;
}

// Glass timeline overrides
.version-timeline {
  padding-left: 4px;

  :deep(.el-timeline-item__tail) {
    border-left: 2px solid rgba(255, 255, 255, 0.1);
  }

  :deep(.el-timeline-item__node) {
    background: rgba(255, 255, 255, 0.15);
    border: none;
  }

  :deep(.el-timeline-item__node--primary) {
    background: #6366f1;
    box-shadow: 0 0 8px rgba(99, 102, 241, 0.3);
  }

  :deep(.el-timeline-item__timestamp) {
    color: rgba(255, 255, 255, 0.3) !important;
    font-size: 12px;
  }
}

// Glass version card
.glass-version-card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 14px;
  transition: all 0.15s ease;

  &:hover {
    border-color: rgba(99, 102, 241, 0.3);
    background: rgba(255, 255, 255, 0.06);
  }

  &--current {
    border-color: rgba(99, 102, 241, 0.4);
    background: rgba(99, 102, 241, 0.06);
  }
}

.version-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.version-no {
  font-size: 13px;
  font-weight: 600;
  color: #818cf8;
}

.version-card-title {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0 0 4px 0;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-card-reason {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.35);
  margin: 0 0 10px 0;
}

.version-card-actions {
  display: flex;
  gap: 6px;
}

// Glass tag overrides
:deep(.el-tag) {
  background: rgba(255, 255, 255, 0.08) !important;
  border-color: rgba(255, 255, 255, 0.15) !important;
  color: rgba(255, 255, 255, 0.7) !important;

:deep(.el-tag.el-tag--success ) {
    background: rgba(16, 185, 129, 0.15) !important;
    border-color: rgba(16, 185, 129, 0.3) !important;
    color: #10b981 !important;
  }
}

// Glass button overrides
:deep(.el-button--primary) {
  background: linear-gradient(135deg, #6366f1, #818cf8) !important;
  border: 1px solid rgba(99, 102, 241, 0.5) !important;
  color: #fff !important;
  border-radius: 10px !important;
  &:hover {
    box-shadow: 0 0 20px rgba(99, 102, 241, 0.3) !important;
  }
}

:deep(.el-button--warning) {
  color: #f59e0b !important;
}

// Glass skeleton
:deep(.el-skeleton) {
  --el-skeleton-color: rgba(255, 255, 255, 0.06);
  --el-skeleton-to-color: rgba(255, 255, 255, 0.12);
}

// Glass empty
:deep(.el-empty__description p) {
  color: rgba(255, 255, 255, 0.4) !important;
}

// Glass alert
:deep(.el-alert) {
  background: rgba(239, 68, 68, 0.1) !important;
  border: 1px solid rgba(239, 68, 68, 0.2) !important;
  border-radius: 10px;
  .el-alert__title {
    color: #ef4444 !important;
  }
}

// Glass dialog
:deep(.el-dialog) {
  background: rgba(15, 15, 30, 0.95) !important;
  backdrop-filter: blur(40px);
  -webkit-backdrop-filter: blur(40px);
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 20px !important;
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.5) !important;

  .el-dialog__header {
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    padding: 20px 24px;
  }

  .el-dialog__title {
    color: #fff !important;
    font-weight: 600;
  }

  .el-dialog__body {
    padding: 24px;
  }

  .el-dialog__footer {
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    padding: 16px 24px;
  }
}

// Glass form
:deep(.el-form-item__label) {
  color: rgba(255, 255, 255, 0.6) !important;
}

:deep(.el-textarea__inner) {
  background: rgba(255, 255, 255, 0.06) !important;
  border: 1px solid rgba(255, 255, 255, 0.12) !important;
  box-shadow: none !important;
  border-radius: 10px !important;
  color: #fff !important;
  &::placeholder {
    color: rgba(255, 255, 255, 0.3) !important;
  }
}

:deep(.el-divider) {
  border-color: rgba(255, 255, 255, 0.08);
}

// Preview content
.preview-content {
  max-height: 500px;
  overflow-y: auto;
}

.preview-meta {
  p {
    margin: 0 0 6px 0;
    font-size: 14px;
    color: rgba(255, 255, 255, 0.8);
    line-height: 1.6;

    strong {
      color: rgba(255, 255, 255, 0.5);
    }
  }
}

.preview-blocks {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview-block {
  padding: 14px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.03);
}

.block-type-tag {
  margin-bottom: 6px;
}

.block-heading {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  margin: 6px 0;
}

.block-text {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
}

.preview-loading {
  padding: 20px;
}

// Responsive
@media (max-width: 768px) {
  .version-history {
    padding: 10px;
  }

  .version-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
}
</style>
