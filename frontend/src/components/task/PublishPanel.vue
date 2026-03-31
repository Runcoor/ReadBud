<template>
  <div class="publish-panel">
    <!-- Review Status -->
    <section class="publish-section">
      <div class="section-header">
        <el-icon :size="14" class="section-icon"><CircleCheck /></el-icon>
        <h4 class="section-label">审核状态</h4>
      </div>
      <div v-if="draft" class="review-row">
        <el-tag :type="reviewTagType" size="small" effect="plain">{{ reviewLabel }}</el-tag>
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
          <el-tag v-if="acc.is_default" size="small" type="info" effect="plain">默认</el-tag>
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
        <el-icon :size="12" color="#FAAD14"><WarningFilled /></el-icon>
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
  gap: $spacing-base;
}

.publish-section {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.section-header {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
}

.section-icon { color: $color-metal; }

.section-label {
  font-size: $font-size-sm;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  line-height: 1;
}

.section-empty {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-md;
  background-color: $color-bg;
  border-radius: $radius-base;
}

.section-empty-text { font-size: $font-size-xs; color: $color-text-muted; }
.section-loading { padding: $spacing-xs 0; }

.review-row { display: flex; align-items: center; gap: $spacing-md; }

.quality-badge {
  display: inline-flex;
  align-items: center;
  gap: $spacing-xs;
  padding: 2px $spacing-sm;
  background-color: $color-bg;
  border-radius: $radius-sm;
  font-size: $font-size-xs;
}

.quality-label { color: $color-text-muted; }
.quality-value { font-weight: $font-weight-semibold; }
.quality--high { color: $color-success; }
.quality--medium { color: $color-warning; }
.quality--low { color: $color-error; }

.risk-alert {
  :deep(.el-alert__title) { font-size: $font-size-xs; font-weight: $font-weight-semibold; }
  :deep(.el-alert__description) { font-size: $font-size-xs; margin-top: 2px; }
}

.mode-cards { display: flex; flex-direction: column; gap: $spacing-xs; }

.mode-card {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-sm $spacing-md;
  border: 1px solid $color-border;
  border-radius: $radius-base;
  cursor: pointer;
  transition: all $transition-fast;
  background-color: $color-card-bg;

  &:hover:not(.mode-card--disabled) {
    border-color: $color-accent;
    background-color: rgba($color-accent, 0.02);
  }

  &--active {
    border-color: $color-accent;
    background-color: rgba($color-accent, 0.04);
    .mode-card-icon { color: $color-accent; }
    .mode-card-title { color: $color-primary; }
  }

  &--disabled { opacity: 0.5; cursor: not-allowed; }
}

.mode-card-icon { flex-shrink: 0; color: $color-metal; transition: color $transition-fast; }

.mode-card-content { display: flex; flex-direction: column; min-width: 0; }

.mode-card-title {
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  transition: color $transition-fast;
}

.mode-card-desc { font-size: $font-size-xs; color: $color-text-muted; line-height: $line-height-tight; }

.schedule-section { display: flex; flex-direction: column; gap: $spacing-xs; padding-top: $spacing-sm; }
.field-label { font-size: $font-size-xs; color: $color-text-secondary; font-weight: $font-weight-medium; }
.schedule-picker { width: 100%; }
.schedule-hint { font-size: $font-size-xs; color: $color-accent; padding-left: 2px; }

.account-list { display: flex; flex-direction: column; gap: $spacing-xs; }

.account-item {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-md;
  border: 1px solid $color-border;
  border-radius: $radius-base;
  cursor: pointer;
  transition: all $transition-fast;
  &:hover { border-color: $color-accent; }
  &--selected { border-color: $color-accent; background-color: rgba($color-accent, 0.04); }
}

.account-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: $radius-sm;
  background-color: $color-primary;
  color: #fff;
  font-size: $font-size-xs;
  font-weight: $font-weight-semibold;
  flex-shrink: 0;
}

.account-info { display: flex; flex-direction: column; min-width: 0; flex: 1; }

.account-name {
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-appid { font-size: $font-size-xs; color: $color-text-muted; font-family: 'SF Mono', 'Menlo', monospace; }
.account-check { flex-shrink: 0; color: $color-accent; }

.action-section {
  display: flex;
  gap: $spacing-sm;
  padding-top: $spacing-md;
  border-top: 1px solid $color-divider;
}

.publish-btn { flex: 1; height: 36px; }
.publish-btn-icon { margin-right: $spacing-xs; }

.publish-warnings { display: flex; flex-direction: column; gap: $spacing-xs; }

.warning-item {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  font-size: $font-size-xs;
  color: $color-text-secondary;
  line-height: $line-height-normal;
}

.slide-fade-enter-active,
.slide-fade-leave-active { transition: all $transition-base; }
.slide-fade-enter-from,
.slide-fade-leave-to { opacity: 0; transform: translateY(-8px); }

@media (max-width: $breakpoint-md) {
  .mode-card { padding: $spacing-xs $spacing-sm; }
  .mode-card-desc { display: none; }
}

@media (max-width: $breakpoint-sm) {
  .publish-panel { gap: $spacing-md; }
  .mode-cards { flex-direction: row; gap: $spacing-xs; }
  .mode-card { flex: 1; flex-direction: column; text-align: center; padding: $spacing-sm $spacing-xs; gap: $spacing-xs; }
  .mode-card-desc { display: none; }
  .account-list { flex-direction: row; overflow-x: auto; gap: $spacing-sm; }
  .account-item { min-width: 160px; flex-shrink: 0; }
}
</style>
