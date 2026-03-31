<template>
  <div class="publish-panel">
    <!-- Review status -->
    <div v-if="draft" class="review-section">
      <h4 class="section-label">审核状态</h4>
      <div class="review-status">
        <el-tag :type="reviewTagType" size="small" effect="plain">
          {{ reviewLabel }}
        </el-tag>
        <span v-if="draft.quality_score > 0" class="quality-score">
          质量评分: {{ draft.quality_score.toFixed(1) }}
        </span>
      </div>
      <div v-if="draft.risk_level !== 'low'" class="risk-alert">
        <el-alert
          :title="`风险等级: ${riskLabel}`"
          :type="draft.risk_level === 'high' ? 'error' : 'warning'"
          :closable="false"
          show-icon
        />
      </div>
    </div>

    <!-- Publish mode -->
    <div class="mode-section">
      <h4 class="section-label">发布方式</h4>
      <el-radio-group v-model="publishMode" :disabled="publishing" class="mode-group">
        <el-radio value="now">立即发布</el-radio>
        <el-radio value="schedule">定时发布</el-radio>
        <el-radio value="manual">手动导出</el-radio>
      </el-radio-group>

      <div v-if="publishMode === 'schedule'" class="schedule-picker">
        <el-date-picker
          v-model="scheduleTime"
          type="datetime"
          placeholder="选择发布时间"
          :disabled-date="isPastDate"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          style="width: 100%"
        />
      </div>
    </div>

    <!-- Account selector -->
    <div v-if="accounts.length > 1" class="account-section">
      <h4 class="section-label">发布账号</h4>
      <el-select v-model="selectedAccountId" placeholder="选择公众号" style="width: 100%">
        <el-option
          v-for="acc in accounts"
          :key="acc.id"
          :label="acc.name"
          :value="acc.id"
        >
          <span>{{ acc.name }}</span>
          <el-tag v-if="acc.is_default" size="small" type="warning" class="default-tag">默认</el-tag>
        </el-option>
      </el-select>
    </div>

    <!-- Publish status -->
    <div v-if="publishJob" class="status-section">
      <h4 class="section-label">发布状态</h4>
      <div class="publish-status">
        <el-tag :type="jobStatusTagType" effect="plain">{{ jobStatusLabel }}</el-tag>
        <span v-if="publishJob.last_error" class="error-text">{{ publishJob.last_error }}</span>
      </div>
    </div>

    <!-- Action buttons -->
    <div class="action-section">
      <el-button
        type="primary"
        :loading="publishing"
        :disabled="!canPublish"
        class="publish-btn"
        @click="handlePublish"
      >
        {{ publishButtonLabel }}
      </el-button>
      <el-button
        v-if="publishJob && canCancel"
        plain
        @click="handleCancel"
      >
        取消发布
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createPublishJob, cancelPublishJob } from '@/api/publish'
import { listWechatAccounts } from '@/api/provider'
import type { PublishJobVO } from '@/api/publish'
import type { DraftVO } from '@/types/draft'
import type { WechatAccountVO } from '@/types/provider'

interface Props {
  draft: DraftVO | null
}

const props = defineProps<Props>()

const publishMode = ref<'now' | 'schedule' | 'manual'>('now')
const scheduleTime = ref<string>('')
const selectedAccountId = ref('')
const accounts = ref<WechatAccountVO[]>([])
const publishing = ref(false)
const publishJob = ref<PublishJobVO | null>(null)

const reviewTagType = computed(() => {
  if (!props.draft) return 'info'
  switch (props.draft.review_status) {
    case 'pass': return 'success'
    case 'reject': return 'danger'
    default: return 'warning'
  }
})

const reviewLabel = computed(() => {
  if (!props.draft) return '待审核'
  switch (props.draft.review_status) {
    case 'pass': return '审核通过'
    case 'reject': return '审核拒绝'
    default: return '待审核'
  }
})

const riskLabel = computed(() => {
  if (!props.draft) return ''
  switch (props.draft.risk_level) {
    case 'high': return '高风险'
    case 'medium': return '中风险'
    default: return '低风险'
  }
})

const canPublish = computed(() => {
  if (!props.draft) return false
  if (publishing.value) return false
  if (publishMode.value === 'schedule' && !scheduleTime.value) return false
  if (!selectedAccountId.value) return false
  return true
})

const canCancel = computed(() => {
  if (!publishJob.value) return false
  return ['queued', 'submitting', 'polling'].includes(publishJob.value.status)
})

const publishButtonLabel = computed(() => {
  if (publishing.value) return '发布中...'
  switch (publishMode.value) {
    case 'now': return '立即发布'
    case 'schedule': return '确认定时发布'
    case 'manual': return '导出 HTML'
    default: return '发布'
  }
})

const jobStatusTagType = computed(() => {
  if (!publishJob.value) return 'info'
  switch (publishJob.value.status) {
    case 'success': return 'success'
    case 'failed': return 'danger'
    case 'cancelled': return 'info'
    default: return 'warning'
  }
})

const jobStatusLabel = computed(() => {
  if (!publishJob.value) return ''
  const map: Record<string, string> = {
    queued: '排队中',
    submitting: '提交中',
    polling: '等待审核',
    success: '发布成功',
    failed: '发布失败',
    cancelled: '已取消',
  }
  return map[publishJob.value.status] || publishJob.value.status
})

function isPastDate(date: Date): boolean {
  return date.getTime() < Date.now() - 86400000
}

async function loadAccounts(): Promise<void> {
  try {
    const resp = await listWechatAccounts()
    if (resp.code === 0) {
      accounts.value = resp.data || []
      // Auto-select default account
      const defaultAcc = accounts.value.find(a => a.is_default)
      if (defaultAcc) {
        selectedAccountId.value = defaultAcc.public_id
      } else if (accounts.value.length > 0) {
        selectedAccountId.value = accounts.value[0].public_id
      }
    }
  } catch {
    // Handled by interceptor
  }
}

async function handlePublish(): Promise<void> {
  if (!props.draft) return

  try {
    await ElMessageBox.confirm(
      publishMode.value === 'now'
        ? '确定立即发布此文章？'
        : publishMode.value === 'schedule'
          ? `确定定时发布？发布时间: ${scheduleTime.value}`
          : '确定导出 HTML？',
      '确认发布',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' },
    )
  } catch {
    return
  }

  publishing.value = true
  try {
    const resp = await createPublishJob({
      draft_id: props.draft.id,
      wechat_account_id: selectedAccountId.value,
      publish_mode: publishMode.value,
      schedule_at: publishMode.value === 'schedule' ? scheduleTime.value : undefined,
    })
    if (resp.code === 0) {
      publishJob.value = resp.data
      ElMessage.success('发布任务已创建')
    }
  } catch {
    ElMessage.error('创建发布任务失败')
  } finally {
    publishing.value = false
  }
}

async function handleCancel(): Promise<void> {
  if (!publishJob.value) return

  try {
    await ElMessageBox.confirm('确定取消发布？', '取消发布', {
      confirmButtonText: '确定',
      cancelButtonText: '返回',
      type: 'warning',
    })
  } catch {
    return
  }

  try {
    const resp = await cancelPublishJob(publishJob.value.id)
    if (resp.code === 0) {
      publishJob.value = resp.data
      ElMessage.success('发布已取消')
    }
  } catch {
    ElMessage.error('取消失败')
  }
}

onMounted(() => {
  loadAccounts()
})
</script>

<style lang="scss" scoped>
.publish-panel {
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

.section-label {
  font-size: $font-size-sm;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin-bottom: $spacing-sm;
}

.review-section {
  .review-status {
    display: flex;
    align-items: center;
    gap: $spacing-md;
  }

  .quality-score {
    font-size: $font-size-xs;
    color: $color-text-secondary;
  }

  .risk-alert {
    margin-top: $spacing-sm;
  }
}

.mode-section {
  .mode-group {
    display: flex;
    flex-direction: column;
    gap: $spacing-xs;
  }

  .schedule-picker {
    margin-top: $spacing-sm;
  }
}

.account-section {
  .default-tag {
    margin-left: $spacing-sm;
  }
}

.status-section {
  .publish-status {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
  }

  .error-text {
    font-size: $font-size-xs;
    color: $color-error;
  }
}

.action-section {
  display: flex;
  gap: $spacing-sm;
  padding-top: $spacing-md;
  border-top: 1px solid $color-divider;

  .publish-btn {
    flex: 1;
  }
}
</style>
