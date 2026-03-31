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
          class="version-card"
          :class="{ 'version-card--current': version.id === currentVersionId }"
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
  padding: $spacing-base;
}

.version-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-lg;
}

.version-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0;
}

.version-loading,
.version-error {
  padding: $spacing-lg 0;
}

.version-error {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  align-items: flex-start;
}

.version-timeline {
  padding-left: $spacing-xs;
}

.version-card {
  background: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-base;
  padding: $spacing-md;
  transition: border-color $transition-fast;

  &:hover {
    border-color: $color-accent;
  }

  &--current {
    border-color: $color-accent;
    background: rgba($color-accent, 0.03);
  }
}

.version-card-header {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  margin-bottom: $spacing-xs;
}

.version-no {
  font-size: $font-size-sm;
  font-weight: $font-weight-semibold;
  color: $color-primary;
}

.version-card-title {
  font-size: $font-size-base;
  color: $color-text-primary;
  margin: 0 0 $spacing-xs 0;
  line-height: $line-height-normal;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-card-reason {
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin: 0 0 $spacing-sm 0;
}

.version-card-actions {
  display: flex;
  gap: $spacing-xs;
}

.preview-content {
  max-height: 500px;
  overflow-y: auto;
}

.preview-meta {
  p {
    margin: 0 0 $spacing-xs 0;
    font-size: $font-size-base;
    color: $color-text-primary;
    line-height: $line-height-relaxed;
  }
}

.preview-blocks {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.preview-block {
  padding: $spacing-md;
  border: 1px solid $color-border;
  border-radius: $radius-sm;
  background: $color-bg;
}

.block-type-tag {
  margin-bottom: $spacing-xs;
}

.block-heading {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: $spacing-xs 0;
}

.block-text {
  font-size: $font-size-base;
  color: $color-text-secondary;
  line-height: $line-height-relaxed;
  margin: 0;
  white-space: pre-wrap;
}

.preview-loading {
  padding: $spacing-lg;
}

// Responsive
@media (max-width: $breakpoint-sm) {
  .version-history {
    padding: $spacing-sm;
  }

  .version-header {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-sm;
  }
}
</style>
