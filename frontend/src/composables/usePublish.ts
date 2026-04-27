// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { ref, computed, onMounted, markRaw, type Component } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Position, Clock, Download } from '@element-plus/icons-vue'
import { createPublishJob, cancelPublishJob, getPublishJob } from '@/api/publish'
import { listWechatAccounts } from '@/api/provider'
import type { PublishJobVO } from '@/api/publish'
import type { DraftVO } from '@/types/draft'
import type { WechatAccountVO } from '@/types/provider'

export type PublishMode = 'now' | 'schedule' | 'manual'

export interface ModeOption {
  value: PublishMode
  label: string
  desc: string
  icon: Component
}

export const PUBLISH_MODES: ModeOption[] = [
  { value: 'now', label: '立即发布', desc: '审核通过后即时推送', icon: markRaw(Position) },
  { value: 'schedule', label: '定时发布', desc: '预约指定时间推送', icon: markRaw(Clock) },
  { value: 'manual', label: '手动导出', desc: '导出 HTML 自行发布', icon: markRaw(Download) },
]

const JOB_STATUS_MAP: Record<string, string> = {
  queued: '排队中',
  submitting: '正在提交',
  polling: '等待平台审核',
  awaiting_extension: '等待插件填充',
  awaiting_manual: '等待手动复制',
  success: '发布成功',
  failed: '发布失败',
  cancelled: '已取消',
}

export function usePublish(getDraft: () => DraftVO | null) {
  // --- State ---
  const publishMode = ref<PublishMode>('now')
  const scheduleTime = ref<string>('')
  const selectedAccountId = ref('')
  const accounts = ref<WechatAccountVO[]>([])
  const accountsLoading = ref(false)
  const publishing = ref(false)
  const publishJob = ref<PublishJobVO | null>(null)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // --- Computed: Review ---
  const reviewTagType = computed(() => {
    const draft = getDraft()
    if (!draft) return 'info'
    switch (draft.review_status) {
      case 'pass': return 'success'
      case 'reject': return 'danger'
      default: return 'warning'
    }
  })

  const reviewLabel = computed(() => {
    const draft = getDraft()
    if (!draft) return '待审核'
    switch (draft.review_status) {
      case 'pass': return '审核通过'
      case 'reject': return '审核拒绝'
      default: return '待审核'
    }
  })

  const qualityClass = computed(() => {
    const draft = getDraft()
    if (!draft) return ''
    const score = draft.quality_score
    if (score >= 8) return 'quality--high'
    if (score >= 6) return 'quality--medium'
    return 'quality--low'
  })

  const riskAlertTitle = computed(() => {
    const draft = getDraft()
    if (!draft) return ''
    return draft.risk_level === 'high' ? '高风险内容' : '中风险内容'
  })

  const riskAlertDesc = computed(() => {
    const draft = getDraft()
    if (!draft) return ''
    return draft.risk_level === 'high'
      ? '内容存在高风险项，建议修改后再发布'
      : '内容存在部分风险项，请确认后再发布'
  })

  // --- Computed: Publish ---
  // Returns the delivery mode of the currently selected WeChat account, or
  // 'api' as a safe default (matches existing behavior when no account or no
  // explicit mode is configured).
  const selectedDeliveryMode = computed(() => {
    const acc = accounts.value.find(a => a.id === selectedAccountId.value)
    return acc?.delivery_mode || 'api'
  })

  const isJobInProgress = computed(() => {
    if (!publishJob.value) return false
    return ['queued', 'submitting', 'polling'].includes(publishJob.value.status)
  })

  const isJobAwaitingUser = computed(() => {
    if (!publishJob.value) return false
    return ['awaiting_extension', 'awaiting_manual'].includes(publishJob.value.status)
  })

  const canPublish = computed(() => {
    if (!getDraft()) return false
    if (publishing.value) return false
    if (!selectedAccountId.value && publishMode.value !== 'manual') return false
    if (publishMode.value === 'schedule' && !scheduleTime.value) return false
    if (publishJob.value && isJobInProgress.value) return false
    return true
  })

  const canCancel = computed(() => {
    if (!publishJob.value) return false
    return [
      'queued', 'submitting', 'polling',
      'awaiting_extension', 'awaiting_manual',
    ].includes(publishJob.value.status)
  })

  const publishButtonLabel = computed(() => {
    if (publishing.value) return '发布中...'
    if (selectedDeliveryMode.value === 'extension') {
      return '通过插件发布'
    }
    if (selectedDeliveryMode.value === 'manual') {
      return '准备复制内容'
    }
    switch (publishMode.value) {
      case 'now': return '立即发布'
      case 'schedule': return '确认定时'
      case 'manual': return '导出 HTML'
      default: return '发布'
    }
  })

  const publishButtonIcon = computed(() => {
    switch (publishMode.value) {
      case 'now': return Position
      case 'schedule': return Clock
      case 'manual': return Download
      default: return Position
    }
  })

  // --- Computed: Job Status ---
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
    return JOB_STATUS_MAP[publishJob.value.status] || publishJob.value.status
  })

  const jobModeLabel = computed(() => {
    if (!publishJob.value) return ''
    switch (publishJob.value.publish_mode) {
      case 'now': return '即时'
      case 'schedule': return '定时'
      case 'manual': return '手动'
      default: return publishJob.value.publish_mode
    }
  })

  const jobProgress = computed(() => {
    if (!publishJob.value) return 0
    switch (publishJob.value.status) {
      case 'queued': return 20
      case 'submitting': return 50
      case 'polling': return 80
      case 'success': return 100
      default: return 0
    }
  })

  const jobProgressStatus = computed<'' | 'success' | 'exception'>(() => {
    if (!publishJob.value) return ''
    if (publishJob.value.status === 'success') return 'success'
    if (publishJob.value.status === 'failed') return 'exception'
    return ''
  })

  // --- Computed: Warnings ---
  const publishWarnings = computed(() => {
    const draft = getDraft()
    const warnings: string[] = []
    if (draft?.review_status === 'reject') {
      warnings.push('文章审核未通过，建议修改后再发布')
    }
    if (draft?.risk_level === 'high') {
      warnings.push('内容存在高风险项，发布后可能被平台限制')
    }
    if (accounts.value.length === 0 && !accountsLoading.value) {
      warnings.push('未配置公众号账号，请先前往设置')
    }
    return warnings
  })

  // --- Helpers ---
  function isPastDate(date: Date): boolean {
    return date.getTime() < Date.now() - 86400000
  }

  function maskAppId(appId: string): string {
    if (!appId || appId.length < 8) return appId
    return appId.slice(0, 4) + '****' + appId.slice(-4)
  }

  function formatScheduleTime(isoTime: string): string {
    try {
      const d = new Date(isoTime)
      return d.toLocaleString('zh-CN', {
        month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit',
      })
    } catch {
      return isoTime
    }
  }

  // --- Actions ---
  async function loadAccounts(): Promise<void> {
    accountsLoading.value = true
    try {
      const resp = await listWechatAccounts()
      if (resp.code === 0) {
        accounts.value = resp.data || []
        const defaultAcc = accounts.value.find(a => a.is_default)
        if (defaultAcc) {
          selectedAccountId.value = defaultAcc.id
        } else if (accounts.value.length > 0) {
          selectedAccountId.value = accounts.value[0].id
        }
      }
    } catch {
      // Handled by interceptor
    } finally {
      accountsLoading.value = false
    }
  }

  async function handlePublish(): Promise<void> {
    const draft = getDraft()
    if (!draft) return

    const delivery = selectedDeliveryMode.value
    const modeDesc =
      delivery === 'extension'
        ? '将打开微信公众号编辑器,由浏览器插件自动填充标题、正文、封面。继续吗？'
        : delivery === 'manual'
          ? '将打开微信公众号编辑器,你需要手动粘贴内容。继续吗？'
          : publishMode.value === 'now'
            ? '文章将立即推送至公众号，确认发布？'
            : publishMode.value === 'schedule'
              ? `文章将于 ${formatScheduleTime(scheduleTime.value)} 自动推送，确认设置？`
              : '将导出文章 HTML 内容至剪贴板，确认导出？'

    try {
      await ElMessageBox.confirm(modeDesc, '确认发布', {
        confirmButtonText: '确定', cancelButtonText: '取消', type: 'info',
      })
    } catch {
      return
    }

    publishing.value = true
    try {
      const resp = await createPublishJob({
        draft_id: draft.id,
        wechat_account_id: selectedAccountId.value,
        publish_mode: publishMode.value,
        schedule_at: publishMode.value === 'schedule' ? scheduleTime.value : undefined,
      })
      if (resp.code === 0) {
        publishJob.value = resp.data
        // Extension/manual: open the WeChat editor in a new tab. The plugin
        // (or the user) takes over from there. We DON'T poll — the job stays
        // in awaiting_extension until the plugin reports back via /fulfilled.
        if (resp.data.editor_url && (resp.data.delivery_mode === 'extension' || resp.data.delivery_mode === 'manual')) {
          window.open(resp.data.editor_url, '_blank', 'noopener,noreferrer')
          ElMessage.success(
            resp.data.delivery_mode === 'extension'
              ? '已打开 WeChat 编辑器,插件将自动填充'
              : '已打开 WeChat 编辑器,请手动粘贴内容',
          )
        } else {
          ElMessage.success('发布任务已创建')
          startPolling()
        }
      }
    } catch {
      ElMessage.error('创建发布任务失败，请重试')
    } finally {
      publishing.value = false
    }
  }

  async function handleCancel(): Promise<void> {
    if (!publishJob.value) return

    try {
      await ElMessageBox.confirm('取消后需要重新提交发布，确定取消？', '取消发布', {
        confirmButtonText: '确定取消', cancelButtonText: '返回', type: 'warning',
      })
    } catch {
      return
    }

    try {
      const resp = await cancelPublishJob(publishJob.value.id)
      if (resp.code === 0) {
        publishJob.value = resp.data
        ElMessage.success('发布已取消')
        stopPolling()
      }
    } catch {
      ElMessage.error('取消操作失败')
    }
  }

  function startPolling(): void {
    stopPolling()
    pollTimer = setInterval(async () => {
      if (!publishJob.value || !isJobInProgress.value) {
        stopPolling()
        return
      }
      try {
        const resp = await getPublishJob(publishJob.value.id)
        if (resp.code === 0) {
          publishJob.value = resp.data
          if (!isJobInProgress.value) {
            stopPolling()
            if (publishJob.value?.status === 'success') {
              ElMessage.success('文章发布成功')
            }
          }
        }
      } catch {
        // Silently retry
      }
    }, 3000)
  }

  function stopPolling(): void {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  onMounted(() => {
    loadAccounts()
  })

  return {
    // State
    publishMode,
    scheduleTime,
    selectedAccountId,
    accounts,
    accountsLoading,
    publishing,
    publishJob,
    // Review computed
    reviewTagType,
    reviewLabel,
    qualityClass,
    riskAlertTitle,
    riskAlertDesc,
    // Publish computed
    canPublish,
    canCancel,
    publishButtonLabel,
    publishButtonIcon,
    isJobInProgress,
    isJobAwaitingUser,
    selectedDeliveryMode,
    // Job status computed
    jobStatusTagType,
    jobStatusLabel,
    jobModeLabel,
    jobProgress,
    jobProgressStatus,
    // Warnings
    publishWarnings,
    // Methods
    isPastDate,
    maskAppId,
    formatScheduleTime,
    handlePublish,
    handleCancel,
  }
}
