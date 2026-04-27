<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="review-page">
    <header class="mono-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">审核规则</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Workbench' })">返回工作台</el-button>
      </div>
    </header>

    <main class="review-main">
      <div class="mono-card section-card">
        <div class="section-toolbar">
          <div>
            <h3 class="section-title">内容审核规则</h3>
            <p class="section-desc">配置关键词黑名单、正则匹配、内容策略规则，用于文章质量检查</p>
          </div>
          <el-button type="primary" @click="openCreateDialog">添加规则</el-button>
        </div>

        <div v-if="loading" class="section-loading">
          <el-skeleton :rows="4" animated />
        </div>

        <div v-else-if="error" class="section-error">
          <el-alert type="error" :title="error" :closable="false" show-icon />
          <el-button size="small" type="primary" plain @click="loadRules">重试</el-button>
        </div>

        <el-empty
          v-else-if="rules.length === 0"
          description="暂无审核规则，点击上方按钮添加"
        />

        <el-table
          v-else
          :data="rules"
          stripe
          class="mono-table"
        >
          <el-table-column label="规则类型" width="140">
            <template #default="{ row }">
              <el-tag size="small" :type="getRuleTypeTag(row.rule_type)">
                {{ RULE_TYPE_LABELS[row.rule_type as RuleType] || row.rule_type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="rule_content" label="规则内容" min-width="280">
            <template #default="{ row }">
              <span class="rule-content-text">{{ truncate(row.rule_content, 80) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="风险等级" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="getRiskTag(row.risk_level)">
                {{ RISK_LEVEL_LABELS[row.risk_level as RiskLevel] || row.risk_level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-switch
                :model-value="row.is_enabled === 1"
                :loading="togglingId === row.id"
                @change="(val: boolean) => handleToggle(row, val)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text @click="openEditDialog(row)">编辑</el-button>
              <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Content Evaluation -->
      <div class="mono-card evaluate-card">
        <h3 class="section-title">内容检测</h3>
        <p class="section-desc">输入内容文本，检测是否触发审核规则</p>
        <el-input
          v-model="evalContent"
          type="textarea"
          :rows="4"
          placeholder="粘贴或输入需要检测的文章内容..."
          class="eval-input"
        />
        <el-button
          type="primary"
          plain
          :loading="evaluating"
          :disabled="!evalContent.trim()"
          @click="handleEvaluate"
        >
          检测内容
        </el-button>

        <div v-if="evalResults !== null" class="eval-results">
          <el-alert
            v-if="evalResults.length === 0"
            type="success"
            title="内容检测通过，未触发任何审核规则"
            :closable="false"
            show-icon
          />
          <template v-else>
            <el-alert
              type="warning"
              :title="`检测到 ${evalResults.length} 条规则触发`"
              :closable="false"
              show-icon
              class="eval-alert"
            />
            <div
              v-for="(v, idx) in evalResults"
              :key="idx"
              class="violation-item"
            >
              <el-tag size="small" :type="getRiskTag(v.risk_level)">
                {{ RISK_LEVEL_LABELS[v.risk_level as RiskLevel] || v.risk_level }}
              </el-tag>
              <el-tag size="small" type="info">
                {{ RULE_TYPE_LABELS[v.rule_type as RuleType] || v.rule_type }}
              </el-tag>
              <span class="violation-detail">{{ v.detail }}</span>
            </div>
          </template>
        </div>
      </div>
    </main>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showDialog"
      :title="editingId ? '编辑审核规则' : '添加审核规则'"
      width="540"
      class="rule-dialog"
      destroy-on-close
    >
      <el-form :model="form" label-position="top">
        <el-form-item label="规则类型" required>
          <el-select v-model="form.rule_type" style="width: 100%">
            <el-option
              v-for="(label, key) in RULE_TYPE_LABELS"
              :key="key"
              :label="label"
              :value="key"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="规则内容" required>
          <el-input
            v-model="form.rule_content"
            type="textarea"
            :rows="4"
            :placeholder="getContentPlaceholder(form.rule_type)"
          />
        </el-form-item>

        <el-form-item label="风险等级" required>
          <el-radio-group v-model="form.risk_level">
            <el-radio value="low">低风险</el-radio>
            <el-radio value="medium">中风险</el-radio>
            <el-radio value="high">高风险</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listReviewRules,
  createReviewRule,
  updateReviewRule,
  deleteReviewRule,
  toggleReviewRule,
  evaluateContent,
} from '@/api/reviewRule'
import type { ReviewRuleVO, RuleViolation } from '@/types/review'
import type { RuleType, RiskLevel } from '@/types/review'
import { RULE_TYPE_LABELS, RISK_LEVEL_LABELS } from '@/types/review'

const router = useRouter()

const loading = ref(false)
const error = ref<string | null>(null)
const rules = ref<ReviewRuleVO[]>([])

const showDialog = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const togglingId = ref<string | null>(null)

const form = reactive({
  rule_type: 'keyword_blacklist' as RuleType,
  rule_content: '',
  risk_level: 'medium' as RiskLevel,
})

const evalContent = ref('')
const evaluating = ref(false)
const evalResults = ref<RuleViolation[] | null>(null)

function truncate(str: string, max: number): string {
  return str.length > max ? str.slice(0, max) + '...' : str
}

function getRuleTypeTag(type: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const map: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    keyword_blacklist: 'danger',
    pattern_match: 'warning',
    content_policy: 'info',
  }
  return map[type] || 'info'
}

function getRiskTag(level: string): '' | 'success' | 'warning' | 'danger' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger'> = {
    low: 'success',
    medium: 'warning',
    high: 'danger',
  }
  return map[level] || ''
}

function getContentPlaceholder(type: string): string {
  switch (type) {
    case 'keyword_blacklist':
      return '输入禁用关键词，用逗号分隔，例如：广告,推广,联系方式'
    case 'pattern_match':
      return '输入正则表达式，例如：\\d{11}（匹配手机号）'
    case 'content_policy':
      return '描述内容策略规则，例如：文章不得包含未经验证的统计数据'
    default:
      return '输入规则内容'
  }
}

async function loadRules() {
  loading.value = true
  error.value = null
  try {
    const resp = await listReviewRules()
    if (resp.code === 0) {
      rules.value = resp.data || []
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载审核规则失败'
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingId.value = null
  form.rule_type = 'keyword_blacklist'
  form.rule_content = ''
  form.risk_level = 'medium'
  showDialog.value = true
}

function openEditDialog(rule: ReviewRuleVO) {
  editingId.value = rule.id
  form.rule_type = rule.rule_type
  form.rule_content = rule.rule_content
  form.risk_level = rule.risk_level
  showDialog.value = true
}

async function handleSave() {
  if (!form.rule_content.trim()) {
    ElMessage.warning('请输入规则内容')
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await updateReviewRule(editingId.value, {
        rule_type: form.rule_type,
        rule_content: form.rule_content,
        risk_level: form.risk_level,
      })
      ElMessage.success('规则已更新')
    } else {
      await createReviewRule({
        rule_type: form.rule_type,
        rule_content: form.rule_content,
        risk_level: form.risk_level,
        is_enabled: 1,
      })
      ElMessage.success('规则已创建')
    }
    showDialog.value = false
    await loadRules()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleToggle(rule: ReviewRuleVO, enabled: boolean) {
  togglingId.value = rule.id
  try {
    const resp = await toggleReviewRule(rule.id, enabled ? 1 : 0)
    if (resp.code === 0) {
      rule.is_enabled = enabled ? 1 : 0
      ElMessage.success(enabled ? '规则已启用' : '规则已停用')
    }
  } catch {
    ElMessage.error('操作失败')
  } finally {
    togglingId.value = null
  }
}

async function handleDelete(rule: ReviewRuleVO) {
  try {
    await ElMessageBox.confirm(
      '确定要删除此审核规则吗？',
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' },
    )
    await deleteReviewRule(rule.id)
    ElMessage.success('已删除')
    await loadRules()
  } catch (e: unknown) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error('删除失败')
    }
  }
}

async function handleEvaluate() {
  evaluating.value = true
  evalResults.value = null
  try {
    const resp = await evaluateContent(evalContent.value)
    if (resp.code === 0) {
      evalResults.value = resp.data || []
    }
  } catch {
    ElMessage.error('检测失败')
  } finally {
    evaluating.value = false
  }
}

onMounted(() => {
  loadRules()
})
</script>

<style lang="scss" scoped>
.review-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #fafafa;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', sans-serif;
}

.mono-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 32px;
  background: var(--surface-card);
  border-bottom: 1px solid #e8e8e8;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-brand { display: flex; align-items: center; gap: 12px; }
.header-title { font-size: 20px; font-weight: 700; color: #0a0a0a; }
.header-divider { color: #d4d4d4; }
.header-desc { font-size: 14px; color: #525252; }

.header-actions {
  :deep(.el-button) {
    color: #525252 !important;
    &:hover { color: #0a0a0a !important; background: #f5f5f5 !important; }
  }
}

.review-main {
  max-width: 1100px;
  width: 100%;
  margin: 0 auto;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.mono-card {
  background: var(--surface-card);
  border: 1px solid #e8e8e8;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.section-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.section-title { font-size: 18px; font-weight: 600; color: #0a0a0a; margin: 0 0 4px 0; }
.section-desc { font-size: 13px; color: #d4d4d4; margin: 0; }

.section-loading, .section-error { padding: 24px 0; }
.section-error { display: flex; flex-direction: column; gap: 10px; align-items: flex-start; }

// Table
:deep(.el-table) {
  --el-table-border-color: #e8e8e8;
  --el-table-header-bg-color: #fafafa;
  --el-table-row-hover-bg-color: #f5f5f5;
  th { font-weight: 600 !important; color: #525252 !important; }
}

.rule-content-text {
  font-size: 13px;
  color: #525252;
  word-break: break-all;
}

// Switch
:deep(.el-switch) {
  --el-switch-off-color: #d4d4d4;
  --el-switch-on-color: #0a0a0a;
}

// Evaluate card
.evaluate-card {
  .eval-input { margin: 14px 0; }
}

.eval-results { margin-top: 20px; }
.eval-alert { margin-bottom: 14px; }

.violation-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  margin-bottom: 8px;
  background: #fafafa;
}

.violation-detail {
  font-size: 13px;
  color: #1a1a1a;
  flex: 1;
}

// Tags
:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: #f5f5f5 !important; color: #525252 !important; }
:deep(.el-tag--danger) { background: #fef2f2 !important; color: #dc2626 !important; }
:deep(.el-tag--success) { background: #f0fdf4 !important; color: #16a34a !important; }
:deep(.el-tag--warning) { background: #fefce8 !important; color: #ca8a04 !important; }

// Buttons
:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
  &:active { transform: scale(0.98); }
}

:deep(.el-button--danger) { color: #ef4444 !important; }

:deep(.el-skeleton) { --el-skeleton-color: #f5f5f5; --el-skeleton-to-color: #e8e8e8; }
:deep(.el-empty__description p) { color: #525252 !important; }

// Alerts
:deep(.el-alert) {
  border-radius: 8px;
  &.el-alert--error {
    background: #fef2f2 !important;
    border: 1px solid #fecaca !important;
    .el-alert__title { color: #dc2626 !important; }
  }
  &.el-alert--success {
    background: #f0fdf4 !important;
    border: 1px solid #bbf7d0 !important;
    .el-alert__title { color: #16a34a !important; }
  }
  &.el-alert--warning {
    background: #fefce8 !important;
    border: 1px solid #fde68a !important;
    .el-alert__title { color: #ca8a04 !important; }
  }
}

// Dialog
:deep(.el-dialog) {
  border-radius: 12px !important;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12) !important;
  .el-dialog__header { border-bottom: 1px solid #e8e8e8; padding: 20px 24px; }
  .el-dialog__body { padding: 24px; }
  .el-dialog__title { font-weight: 600; color: #0a0a0a; }
  .el-dialog__footer { border-top: 1px solid #e8e8e8; padding: 16px 24px; }
}

// Form
:deep(.el-form-item__label) { color: #525252 !important; }

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

:deep(.el-textarea__inner) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  box-shadow: none !important;
  border-radius: 8px !important;
  color: #0a0a0a !important;
  &::placeholder { color: #c4c4c4 !important; }
  &:focus { border-color: #0a0a0a !important; box-shadow: 0 0 0 2px rgba(10, 10, 10, 0.1) !important; }
}

:deep(.el-select__wrapper) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  box-shadow: none !important;
  border-radius: 8px !important;
}

:deep(.el-radio) {
  --el-radio-text-color: #1a1a1a;
  --el-radio-input-border-color: #d4d4d4;
}

:deep(.el-button--default) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  color: #0a0a0a !important;
  border-radius: 8px !important;
  &:hover { border-color: #0a0a0a !important; }
}

@media (max-width: 1024px) {
  .review-main { padding: 16px; max-width: 100%; }
  .mono-card { padding: 16px; }
  .section-toolbar { flex-direction: column; gap: 10px; }
}

@media (max-width: 768px) {
  .mono-header { height: 52px; padding: 0 12px; }
  .header-desc, .header-divider { display: none; }
  .review-main { padding: 12px; }
  .mono-card { padding: 12px; }
  .violation-item { flex-wrap: wrap; }
}
</style>
