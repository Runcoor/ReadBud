<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="article-history-page">
    <AppTopBar crumb="文章历史">
      <template #right>
        <a class="article-history-page__back mono" @click="router.push({ name: 'Workbench' })">
          ← 返回工作台
        </a>
        <span class="topbar-sep" />
        <button class="btn-primary" @click="router.push({ name: 'Workbench' })">+ 新任务</button>
      </template>
    </AppTopBar>

    <main class="article-history-main">
      <!-- Hero row -->
      <div class="hero">
        <div class="hero__left">
          <h1 class="hero__title">文章历史</h1>
          <MonoChip>ARCHIVE · {{ total }} ITEMS</MonoChip>
          <p class="hero__sub">管理所有已生成的文章，支持二次编辑与重新发布。</p>
        </div>
        <div class="hero__stats">
          <div class="stat">
            <div class="stat__value" :style="{ color: 'var(--brand-sprout)' }">{{ statDone }}</div>
            <div class="stat__label">已完成</div>
          </div>
          <div class="stat">
            <div class="stat__value" :style="{ color: 'var(--brand-danger)' }">{{ statFailed }}</div>
            <div class="stat__label">失败</div>
          </div>
          <div class="stat">
            <div class="stat__value" :style="{ color: 'var(--brand-warn)' }">{{ statRunning }}</div>
            <div class="stat__label">执行中</div>
          </div>
          <div class="stat">
            <div class="stat__value" :style="{ color: 'var(--brand-ink)' }">{{ statWords }}</div>
            <div class="stat__label">总字数</div>
          </div>
        </div>
      </div>

      <!-- Batch action bar -->
      <div v-if="selectedIds.length > 0" class="batch-bar">
        <span class="batch-bar__text">已选 {{ selectedIds.length }} 项</span>
        <span class="batch-bar__sep">·</span>
        <button class="batch-bar__action" @click="handleBatchDelete">批量删除</button>
      </div>

      <!-- Filter row -->
      <div class="filters">
        <div class="search-box">
          <span class="search-box__icon">⌕</span>
          <input
            v-model="searchKeyword"
            class="search-box__input"
            placeholder="搜索关键词…"
            @input="handleSearch"
          />
        </div>
        <PillTabs v-model="statusFilter" :options="statusOptions" />
        <div class="filters__spacer" />
        <el-select v-model="sortOrder" class="sort-select" size="small">
          <el-option label="按时间倒序" value="desc" />
          <el-option label="按时间升序" value="asc" />
        </el-select>
      </div>

      <!-- Table card -->
      <div class="table-card" v-loading="loading">
        <div class="table-head">
          <div class="cell-cb">
            <span class="cb" :class="{ 'is-checked': allChecked, 'is-indeterminate': someChecked }" @click="toggleAll" />
          </div>
          <div class="cell">关键词</div>
          <div class="cell">风格</div>
          <div class="cell">状态</div>
          <div class="cell cell--right">字数</div>
          <div class="cell">创建时间</div>
          <div class="cell cell--right">操作</div>
        </div>

        <div v-if="filteredTasks.length === 0 && !loading" class="empty">
          <span class="empty__mono">EMPTY · 还没有文章 ·</span>
          <button class="empty__cta" @click="router.push({ name: 'Workbench' })">立即创建 →</button>
        </div>

        <div
          v-for="task in filteredTasks"
          :key="task.id"
          class="table-row"
        >
          <div class="cell-cb" @click.stop="toggleOne(task.id)">
            <span class="cb" :class="{ 'is-checked': selectedIds.includes(task.id) }" />
          </div>
          <div class="cell cell--keyword" @click="goToTask(task.id)">
            {{ task.keyword }}
          </div>
          <div class="cell cell--style">
            {{ styleLabel(task) }}
          </div>
          <div class="cell cell--status">
            <StatusDot :kind="statusKind(task.status)" />
            <span>{{ STATUS_LABELS[task.status] }}</span>
          </div>
          <div class="cell cell--right cell--mono">
            {{ task.target_words.toLocaleString() }}
          </div>
          <div class="cell cell--mono cell--mute">
            {{ formatTime(task.created_at) }}
          </div>
          <div class="cell cell--right cell--actions">
            <button class="row-action" @click.stop="goToTask(task.id)">查看</button>
            <button class="row-action row-action--danger" @click.stop="handleDelete(task)">删除</button>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div class="pagination">
        <span class="pagination__total">共 {{ total }} 条</span>
        <div class="pagination__right">
          <el-select v-model="pageSize" class="page-size-select" size="small" @change="onPageSizeChange">
            <el-option v-for="n in [20, 50, 100]" :key="n" :label="`${n} 条 / 页`" :value="n" />
          </el-select>
          <button
            class="page-btn"
            :disabled="currentPage <= 1"
            @click="goPage(currentPage - 1)"
          >‹</button>
          <button class="page-btn page-btn--active">{{ currentPage }}</button>
          <button
            class="page-btn"
            :disabled="currentPage >= totalPages"
            @click="goPage(currentPage + 1)"
          >›</button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { listTasks, deleteTask, batchDeleteTasks } from '@/api/task'
import type { TaskVO, TaskStatus, ArticleStyle } from '@/types/task'
import { STATUS_LABELS, STATUS_TAG_TYPES, ARTICLE_STYLE_LABELS } from '@/types/task'
import AppTopBar from '@/components/common/AppTopBar.vue'
import MonoChip from '@/components/common/MonoChip.vue'
import PillTabs from '@/components/common/PillTabs.vue'
import StatusDot from '@/components/common/StatusDot.vue'

// Suppress unused import warning — kept per instructions
void STATUS_TAG_TYPES

const router = useRouter()

const loading = ref(false)
const tasks = ref<TaskVO[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const searchKeyword = ref('')
const statusFilter = ref<'' | 'done' | 'failed' | 'running'>('')
const selectedIds = ref<string[]>([])
const sortOrder = ref<'desc' | 'asc'>('desc')

const statusOptions = [
  { label: '全部', value: '' as const },
  { label: '已完成', value: 'done' as const },
  { label: '失败', value: 'failed' as const },
  { label: '执行中', value: 'running' as const },
]

const filteredTasks = computed(() => {
  if (!searchKeyword.value) return tasks.value
  const kw = searchKeyword.value.toLowerCase()
  return tasks.value.filter(t => t.keyword.toLowerCase().includes(kw))
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const statDone = computed(() => tasks.value.filter(t => t.status === 'done').length)
const statFailed = computed(() => tasks.value.filter(t => t.status === 'failed').length)
const statRunning = computed(() => tasks.value.filter(t => t.status === 'running' || t.status === 'pending').length)
const statWords = computed(() => {
  const sum = tasks.value
    .filter(t => t.status === 'done')
    .reduce((acc, t) => acc + (t.target_words || 0), 0)
  if (sum >= 1000) return `${(sum / 1000).toFixed(1)}k`
  return String(sum)
})

const allChecked = computed(() =>
  filteredTasks.value.length > 0 && filteredTasks.value.every(t => selectedIds.value.includes(t.id)),
)
const someChecked = computed(() =>
  !allChecked.value && filteredTasks.value.some(t => selectedIds.value.includes(t.id)),
)

function statusKind(status: TaskStatus): 'sprout' | 'danger' | 'warn' | 'mute' {
  if (status === 'done') return 'sprout'
  if (status === 'failed') return 'danger'
  if (status === 'running' || status === 'pending') return 'warn'
  return 'mute'
}

function styleLabel(task: TaskVO): string {
  if (task.article_style && ARTICLE_STYLE_LABELS[task.article_style as ArticleStyle]) {
    return ARTICLE_STYLE_LABELS[task.article_style as ArticleStyle]
  }
  return task.tone || '自动'
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

function handleSearch() {
  // Client-side filter via computed, no API call needed
}

function toggleOne(id: string) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function toggleAll() {
  if (allChecked.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = filteredTasks.value.map(t => t.id)
  }
}

function goToTask(id: string) {
  router.push({ name: 'TaskDetail', params: { id } })
}

function goPage(p: number) {
  if (p < 1 || p > totalPages.value) return
  currentPage.value = p
  loadTasks()
}

function onPageSizeChange() {
  currentPage.value = 1
  loadTasks()
}

async function loadTasks() {
  loading.value = true
  try {
    const res = await listTasks(currentPage.value, pageSize.value, statusFilter.value || undefined)
    tasks.value = res.data?.items || []
    total.value = res.data?.total || 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

async function handleDelete(row: TaskVO) {
  try {
    await ElMessageBox.confirm(
      `确定要删除「${row.keyword}」吗？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
    )
    await deleteTask(row.id)
    ElMessage.success('已删除')
    await loadTasks()
  } catch {
    // cancelled or error
  }
}

async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 篇文章吗？`,
      '批量删除',
      { confirmButtonText: '全部删除', cancelButtonText: '取消', type: 'warning' },
    )
    const res = await batchDeleteTasks(selectedIds.value)
    ElMessage.success(`已删除 ${res.data?.deleted || 0} 篇`)
    selectedIds.value = []
    await loadTasks()
  } catch {
    // cancelled or error
  }
}

// React to filter changes
import { watch } from 'vue'
watch(statusFilter, () => {
  currentPage.value = 1
  loadTasks()
})

onMounted(() => {
  loadTasks()
})
</script>

<style lang="scss" scoped>
.article-history-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--brand-paper);
  font-family: var(--font-sans);
  color: var(--text-primary);
}

.btn-primary {
  height: 30px;
  padding: 0 14px;
  font-size: 12px;
  letter-spacing: 0.08em;
  background: var(--brand-ink);
  color: var(--text-inverse);
  border: 1px solid var(--brand-ink);
  border-radius: 4px;
  cursor: pointer;
  font-family: var(--font-sans);
  transition: background 120ms ease;

  &:hover { background: var(--brand-ink-soft); }
}

.article-history-page__back {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-tertiary);
  cursor: pointer;
  letter-spacing: 0.04em;
  transition: color 120ms ease;

  &:hover { color: var(--text-primary); }
}

.topbar-sep {
  width: 1px;
  height: 16px;
  background: var(--border-hair);
  margin: 0 4px;
}

.article-history-main {
  flex: 1;
  padding: 32px 48px;
  overflow: auto;
}

// ===== Hero =====
.hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  margin-bottom: 24px;

  &__left {
    min-width: 0;
  }

  &__title {
    font-family: var(--font-serif);
    font-size: 28px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 8px;
    line-height: 1.2;
  }

  &__sub {
    font-size: 13px;
    color: var(--text-tertiary);
    margin: 6px 0 0;
    line-height: 1.5;
  }

  &__stats {
    display: flex;
    gap: 32px;
    flex-shrink: 0;
  }
}

.stat {
  text-align: right;

  &__value {
    font-family: var(--font-mono);
    font-size: 22px;
    font-weight: 600;
    line-height: 1.1;
  }

  &__label {
    font-size: 11px;
    color: var(--text-tertiary);
    margin-top: 4px;
    letter-spacing: 0.04em;
  }
}

// ===== Batch action bar =====
.batch-bar {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  margin-bottom: 12px;
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-body);

  &__sep {
    color: var(--text-tertiary);
  }

  &__action {
    background: transparent;
    border: none;
    color: var(--brand-danger);
    cursor: pointer;
    font-size: 12px;
    padding: 0;
    font-family: inherit;

    &:hover { text-decoration: underline; }
  }
}

// ===== Filter row =====
.filters {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.search-box {
  position: relative;
  width: 320px;
  height: 34px;

  &__icon {
    position: absolute;
    left: 10px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 12px;
    color: var(--text-tertiary);
    pointer-events: none;
  }

  &__input {
    width: 100%;
    height: 100%;
    padding: 0 12px 0 28px;
    background: var(--surface-card);
    border: 1px solid var(--border-hair);
    border-radius: 5px;
    font-size: 13px;
    color: var(--text-primary);
    font-family: inherit;
    outline: none;
    transition: border-color 120ms ease;

    &::placeholder { color: var(--text-tertiary); }
    &:focus { border-color: var(--brand-ink); }
  }
}

.filters__spacer {
  flex: 1;
}

.sort-select {
  width: 130px;
}

// ===== Table =====
.table-card {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  overflow: hidden;
}

.table-head,
.table-row {
  display: grid;
  grid-template-columns: 32px 1fr 120px 120px 80px 140px 100px;
  align-items: center;
  gap: 0;
}

.table-head {
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-hair);

  .cell {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.15em;
    color: var(--text-tertiary);
  }
}

.table-row {
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-hair-soft);
  font-size: 13px;
  color: var(--text-body);
  transition: background 120ms ease;

  &:last-child { border-bottom: none; }
  &:hover { background: var(--brand-paper); }
}

.cell {
  min-width: 0;
  padding: 0 8px;

  &--right {
    text-align: right;
    justify-self: end;
  }

  &--mono {
    font-family: var(--font-mono);
    font-size: 12px;
  }

  &--mute {
    color: var(--text-tertiary);
  }

  &--keyword {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    cursor: pointer;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &:hover { color: var(--brand-ink-soft); text-decoration: underline; }
  }

  &--style {
    font-size: 13px;
    color: var(--text-body);
  }

  &--status {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
  }

  &--actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }
}

.cell-cb {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.cb {
  width: 12px;
  height: 12px;
  border: 1px solid var(--border-hair);
  border-radius: 2px;
  background: var(--surface-card);
  display: inline-block;
  position: relative;
  transition: all 120ms ease;

  &:hover { border-color: var(--brand-ink); }

  &.is-checked {
    background: var(--brand-ink);
    border-color: var(--brand-ink);

    &::after {
      content: '';
      position: absolute;
      top: 1px;
      left: 3px;
      width: 3px;
      height: 6px;
      border: solid var(--text-inverse);
      border-width: 0 1.5px 1.5px 0;
      transform: rotate(45deg);
    }
  }

  &.is-indeterminate {
    background: var(--brand-ink);
    border-color: var(--brand-ink);

    &::after {
      content: '';
      position: absolute;
      top: 5px;
      left: 2px;
      width: 6px;
      height: 1.5px;
      background: var(--text-inverse);
    }
  }
}

.row-action {
  background: transparent;
  border: none;
  font-size: 12px;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 0;
  font-family: inherit;
  transition: color 120ms ease;

  &:hover { color: var(--text-primary); }

  &--danger {
    color: var(--brand-danger);
    opacity: 0.7;

    &:hover { opacity: 1; color: var(--brand-danger); }
  }
}

.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 60px 20px;
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.1em;
  color: var(--text-tertiary);

  &__cta {
    background: transparent;
    border: none;
    color: var(--brand-ink);
    font-family: inherit;
    font-size: 12px;
    letter-spacing: 0.1em;
    cursor: pointer;
    padding: 0;

    &:hover { text-decoration: underline; }
  }
}

// ===== Pagination =====
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 0 4px;

  &__total {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  &__right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.page-size-select {
  width: 110px;
}

.page-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-hair);
  background: var(--surface-card);
  color: var(--text-body);
  border-radius: 4px;
  font-family: var(--font-mono);
  font-size: 12px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 120ms ease;

  &:hover:not(:disabled) {
    border-color: var(--brand-ink);
    color: var(--text-primary);
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  &--active {
    background: var(--brand-ink);
    color: var(--text-inverse);
    border-color: var(--brand-ink);

    &:hover { color: var(--text-inverse); }
  }
}

// ===== Element Plus overrides for embedded selects =====
:deep(.sort-select),
:deep(.page-size-select) {
  .el-select__wrapper {
    background: var(--surface-card) !important;
    border: 1px solid var(--border-hair) !important;
    box-shadow: none !important;
    border-radius: 4px !important;
    min-height: 28px !important;
    padding: 0 8px !important;
    font-size: 12px !important;

    &:hover { border-color: var(--brand-ink) !important; }
    &.is-focused { border-color: var(--brand-ink) !important; box-shadow: none !important; }
  }
}

// ===== Responsive =====
@media (max-width: 1024px) {
  .article-history-main { padding: 24px; }
  .hero {
    flex-direction: column;
  }
  .hero__stats { gap: 24px; }
}

@media (max-width: 768px) {
  .article-history-main { padding: 16px; }
  .filters { flex-wrap: wrap; }
  .search-box { width: 100%; }
  .table-head,
  .table-row {
    grid-template-columns: 32px 1fr 100px 100px;
    .cell:nth-child(5),
    .cell:nth-child(6),
    .cell:nth-child(7) { display: none; }
  }
}
</style>
