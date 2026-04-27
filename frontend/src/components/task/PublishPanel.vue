<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="publish-panel">
    <!-- Review Status -->
    <section class="publish-section">
      <div class="section-header">
        <el-icon :size="14" class="section-icon"><CircleCheck /></el-icon>
        <h4 class="section-label">审核状态</h4>
      </div>
      <div v-if="draft" class="review-row">
        <el-tag :type="reviewTagType" size="small">{{ reviewLabel }}</el-tag>
        <span v-if="draft.quality_score > 0" class="quality-badge">
          <span class="quality-label">质量</span>
          <span class="quality-value" :class="qualityClass">{{ draft.quality_score.toFixed(1) }}</span>
        </span>
      </div>
      <el-alert
        v-if="draft?.risk_level && draft.risk_level !== 'low'"
        :title="riskAlertTitle"
        :description="riskAlertDesc"
        :type="draft.risk_level === 'high' ? 'error' : 'warning'"
        :closable="false"
        show-icon
        class="risk-alert"
      />
      <div v-if="!draft" class="section-empty">
        <span class="section-empty-text">暂无审核信息</span>
      </div>
    </section>

    <!-- Publish Mode -->
    <section class="publish-section">
      <div class="section-header">
        <el-icon :size="14" class="section-icon"><Promotion /></el-icon>
        <h4 class="section-label">发布方式</h4>
      </div>
      <div class="mode-cards">
        <div
          v-for="mode in PUBLISH_MODES"
          :key="mode.value"
          class="mode-card"
          :class="{ 'mode-card--active': publishMode === mode.value, 'mode-card--disabled': publishing }"
          @click="!publishing && (publishMode = mode.value)"
        >
          <el-icon :size="18" class="mode-card-icon"><component :is="mode.icon" /></el-icon>
          <div class="mode-card-content">
            <span class="mode-card-title">{{ mode.label }}</span>
            <span class="mode-card-desc">{{ mode.desc }}</span>
          </div>
        </div>
      </div>
      <transition name="slide-fade">
        <div v-if="publishMode === 'schedule'" class="schedule-section">
          <label class="field-label">发布时间</label>
          <el-date-picker
            v-model="scheduleTime"
            type="datetime"
            placeholder="选择发布时间"
            :disabled-date="isPastDate"
            :disabled="publishing"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            class="schedule-picker"
          />
          <p v-if="scheduleTime" class="schedule-hint">
            将于 {{ formatScheduleTime(scheduleTime) }} 自动发布
          </p>
        </div>
      </transition>
    </section>

    <!-- Account Selector -->
    <section class="publish-section">
      <div class="section-header">
        <el-icon :size="14" class="section-icon"><User /></el-icon>
        <h4 class="section-label">发布账号</h4>
      </div>
      <div v-if="accountsLoading" class="section-loading">
        <el-skeleton :rows="1" animated />
      </div>
      <div v-else-if="accounts.length === 0" class="section-empty">
        <span class="section-empty-text">未配置公众号账号</span>
        <el-button type="primary" link size="small" @click="$router.push({ name: 'Settings' })">
          前往配置
        </el-button>
      </div>
      <div v-else class="account-list">
        <div
          v-for="acc in accounts"
          :key="acc.id"
          class="account-item"
          :class="{ 'account-item--selected': selectedAccountId === acc.id }"
          @click="selectedAccountId = acc.id"
        >
          <div class="account-avatar">{{ acc.name.charAt(0) }}</div>
          <div class="account-info">
            <span class="account-name">{{ acc.name }}</span>
            <span class="account-appid">{{ maskAppId(acc.app_id) }}</span>
          </div>
          <el-tag v-if="acc.is_default" size="small" type="info">默认</el-tag>
          <el-icon v-if="selectedAccountId === acc.id" :size="16" class="account-check"><Select /></el-icon>
        </div>
      </div>
    </section>

    <!-- Job Status -->
    <section v-if="publishJob" class="publish-section">
      <div class="section-header">
        <el-icon :size="14" class="section-icon"><DataLine /></el-icon>
        <h4 class="section-label">发布状态</h4>
      </div>
      <PublishStatusCard
        :job="publishJob"
        @update="onJobUpdate"
      />
    </section>

    <!-- Actions -->
    <div class="action-section">
      <el-button
        type="primary"
        :loading="publishing"
        :disabled="!canPublish"
        class="publish-btn"
        @click="handlePublish"
      >
        <el-icon v-if="!publishing" class="publish-btn-icon"><component :is="publishButtonIcon" /></el-icon>
        {{ publishButtonLabel }}
      </el-button>
      <el-button v-if="publishJob && canCancel" plain @click="handleCancel">取消发布</el-button>
    </div>

    <!-- Warnings -->
    <div v-if="publishWarnings.length > 0" class="publish-warnings">
      <div v-for="(warn, idx) in publishWarnings" :key="idx" class="warning-item">
        <el-icon :size="12" color="#eab308"><WarningFilled /></el-icon>
        <span>{{ warn }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  CircleCheck, Promotion, User, Select, DataLine, WarningFilled,
} from '@element-plus/icons-vue'
import PublishStatusCard from '@/components/task/PublishStatusCard.vue'
import { usePublish, PUBLISH_MODES } from '@/composables/usePublish'
import type { DraftVO } from '@/types/draft'
import type { PublishJobVO } from '@/api/publish'

interface Props {
  draft: DraftVO | null
}

const props = defineProps<Props>()

const {
  publishMode, scheduleTime, selectedAccountId, accounts, accountsLoading,
  publishing, publishJob,
  reviewTagType, reviewLabel, qualityClass, riskAlertTitle, riskAlertDesc,
  canPublish, canCancel, publishButtonLabel, publishButtonIcon,
  publishWarnings,
  isPastDate, maskAppId, formatScheduleTime, handlePublish, handleCancel,
} = usePublish(() => props.draft)

function onJobUpdate(updated: PublishJobVO): void {
  publishJob.value = updated
}
</script>

<style lang="scss" scoped>
.publish-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.publish-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 6px;
}

.section-icon {
  color: #d4d4d4;
}

.section-label {
  font-size: 13px;
  font-weight: 600;
  color: #0a0a0a;
  line-height: 1;
}

.section-empty {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: #f5f5f5;
  border-radius: 8px;
}

.section-empty-text {
  font-size: 12px;
  color: #d4d4d4;
}

.section-loading {
  padding: 4px 0;
}

.review-row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.quality-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 10px;
  background: #f5f5f5;
  border-radius: 4px;
  font-size: 12px;
}

.quality-label {
  color: #d4d4d4;
}

.quality-value {
  font-weight: 600;
}

.quality--high { color: #22c55e; }
.quality--medium { color: #eab308; }
.quality--low { color: #ef4444; }

// Alerts
:deep(.el-alert--error) {
  background: #fef2f2 !important;
  border: 1px solid #fecaca !important;
  border-radius: 8px;
  .el-alert__title { color: #dc2626 !important; }
  .el-alert__description { color: #ef4444 !important; }
}

:deep(.el-alert--warning) {
  background: #fefce8 !important;
  border: 1px solid #fde68a !important;
  border-radius: 8px;
  .el-alert__title { color: #ca8a04 !important; }
  .el-alert__description { color: #eab308 !important; }
}

// Mode cards
.mode-cards {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.mode-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  background: var(--surface-card);

  &:hover:not(.mode-card--disabled) {
    border-color: #0a0a0a;
  }

  &--active {
    border-color: #0a0a0a !important;
    background: #fafafa !important;

    .mode-card-icon { color: #0a0a0a; }
    .mode-card-title { color: #0a0a0a; font-weight: 600; }
  }

  &--disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.mode-card-icon {
  flex-shrink: 0;
  color: #d4d4d4;
  transition: color 0.15s ease;
}

.mode-card-content {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.mode-card-title {
  font-size: 13px;
  font-weight: 500;
  color: #1a1a1a;
  transition: color 0.15s ease;
}

.mode-card-desc {
  font-size: 12px;
  color: #d4d4d4;
  line-height: 1.3;
}

.schedule-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 8px;
}

.field-label {
  font-size: 12px;
  color: #525252;
  font-weight: 500;
}

.schedule-picker { width: 100%; }

.schedule-hint {
  font-size: 12px;
  color: #525252;
  padding-left: 2px;
}

// Account list
.account-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.account-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  background: var(--surface-card);

  &:hover { border-color: #0a0a0a; }

  &--selected {
    border-color: #0a0a0a;
    background: #fafafa;
  }
}

.account-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: #0a0a0a;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.account-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.account-name {
  font-size: 13px;
  font-weight: 500;
  color: #0a0a0a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-appid {
  font-size: 11px;
  color: #d4d4d4;
  font-family: 'SF Mono', 'Menlo', monospace;
}

.account-check {
  flex-shrink: 0;
  color: #0a0a0a;
}

// Actions
.action-section {
  display: flex;
  gap: 10px;
  padding-top: 14px;
  border-top: 1px solid #e8e8e8;
}

.publish-btn { flex: 1; height: 36px; }
.publish-btn-icon { margin-right: 6px; }

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
  &:active { transform: scale(0.98); }
}

:deep(.el-button.is-plain) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  color: #0a0a0a !important;
  border-radius: 8px !important;
  &:hover { border-color: #0a0a0a !important; }
}

// Warnings
.publish-warnings {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.warning-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #525252;
  line-height: 1.4;
}

// Tag overrides
:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: #f5f5f5 !important; color: #525252 !important; }
:deep(.el-tag--success) { background: #f0fdf4 !important; color: #16a34a !important; }
:deep(.el-tag--danger) { background: #fef2f2 !important; color: #dc2626 !important; }
:deep(.el-tag--warning) { background: #fefce8 !important; color: #ca8a04 !important; }

:deep(.el-skeleton) {
  --el-skeleton-color: #f5f5f5;
  --el-skeleton-to-color: #e8e8e8;
}

:deep(.el-input__wrapper) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  box-shadow: none !important;
  border-radius: 8px !important;
  &:hover { border-color: #d4d4d4 !important; }
  &.is-focus { border-color: #0a0a0a !important; box-shadow: 0 0 0 2px rgba(10, 10, 10, 0.1) !important; }
}

:deep(.el-input__inner) {
  color: #0a0a0a !important;
  &::placeholder { color: #c4c4c4 !important; }
}

// Transitions
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.2s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@media (max-width: 1024px) {
  .mode-card { padding: 8px 10px; }
  .mode-card-desc { display: none; }
}

@media (max-width: 768px) {
  .publish-panel { gap: 12px; }

  .mode-cards { flex-direction: row; gap: 6px; }

  .mode-card {
    flex: 1;
    flex-direction: column;
    text-align: center;
    padding: 10px 6px;
    gap: 6px;
  }

  .mode-card-desc { display: none; }

  .account-list { flex-direction: row; overflow-x: auto; gap: 8px; }
  .account-item { min-width: 160px; flex-shrink: 0; }
}
</style>
