<template>
  <div class="article-history-page">
    <header class="mono-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">文章历史</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Workbench' })">返回工作台</el-button>
      </div>
    </header>

    <main class="article-history-main">
      <div class="page-header">
        <h2 class="page-title">文章历史</h2>
        <p class="page-subtitle">管理所有已生成的文章</p>
      </div>

      <!-- Toolbar: search, filter, batch actions -->
      <div class="toolbar">
        <el-input v-model="searchKeyword" placeholder="搜索关键词..." clearable prefix-icon="Search" class="search-input" @input="handleSearch" />
        <el-select v-model="statusFilter" placeholder="状态" clearable class="status-filter" @change="loadTasks">
          <el-option label="全部" value="" />
          <el-option label="已完成" value="done" />
          <el-option label="失败" value="failed" />
          <el-option label="进行中" value="running" />
          <el-option label="排队中" value="pending" />
        </el-select>
        <el-button v-if="selectedIds.length > 0" type="danger" plain @click="handleBatchDelete">
          删除选中 ({{ selectedIds.length }})
        </el-button>
      </div>

      <!-- Table -->
      <div class="mono-card">
        <el-table
          :data="filteredTasks"
          v-loading="loading"
          @selection-change="handleSelectionChange"
          class="history-table"
          row-key="id"
        >
          <el-table-column type="selection" width="40" />
          <el-table-column prop="keyword" label="关键词" min-width="200">
            <template #default="{ row }">
              <router-link :to="`/task/${row.id}`" class="keyword-link">{{ row.keyword }}</router-link>
            </template>
          </el-table-column>
          <el-table-column prop="article_style" label="风格" width="100">
            <template #default="{ row }">
              <span v-if="row.article_style" class="style-tag">{{ ARTICLE_STYLE_LABELS[row.article_style as ArticleStyle] || row.article_style }}</span>
              <span v-else class="style-tag muted">自动</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="STATUS_TAG_TYPES[row.status as TaskStatus]" size="small">
                {{ STATUS_LABELS[row.status as TaskStatus] }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="target_words" label="字数" width="80" />
          <el-table-column prop="created_at" label="创建时间" width="170">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" text size="small" @click.stop="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- Pagination -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadTasks"
          @size-change="loadTasks"
        />
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

const router = useRouter()

const loading = ref(false)
const tasks = ref<TaskVO[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const searchKeyword = ref('')
const statusFilter = ref('')
const selectedIds = ref<string[]>([])

const filteredTasks = computed(() => {
  if (!searchKeyword.value) return tasks.value
  const kw = searchKeyword.value.toLowerCase()
  return tasks.value.filter(t => t.keyword.toLowerCase().includes(kw))
})

function formatTime(iso: string): string {
  return new Date(iso).toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

function handleSelectionChange(rows: TaskVO[]) {
  selectedIds.value = rows.map(r => r.id)
}

function handleSearch() {
  // Client-side filter via computed, no API call needed
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
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
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
      { confirmButtonText: '全部删除', cancelButtonText: '取消', type: 'warning' }
    )
    const res = await batchDeleteTasks(selectedIds.value)
    ElMessage.success(`已删除 ${res.data?.deleted || 0} 篇`)
    selectedIds.value = []
    await loadTasks()
  } catch {
    // cancelled or error
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<style lang="scss" scoped>
.article-history-page {
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
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-title {
  font-size: 20px;
  font-weight: 700;
  color: #0a0a0a;
}

.header-divider { color: #d4d4d4; }

.header-desc {
  font-size: 14px;
  color: #525252;
}

.header-actions {
  :deep(.el-button) {
    color: #525252 !important;
    &:hover { color: #0a0a0a !important; background: #f5f5f5 !important; }
  }
}

.article-history-main {
  max-width: 1000px;
  width: 100%;
  margin: 0 auto;
  padding: 32px;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: #0a0a0a;
  margin: 0 0 4px;
}

.page-subtitle {
  font-size: 13px;
  color: #a3a3a3;
  margin: 0;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.search-input {
  width: 260px;
}

.status-filter {
  width: 120px;
}

.mono-card {
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 12px;
  padding: 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.keyword-link {
  color: #0a0a0a;
  text-decoration: none;
  font-weight: 500;

  &:hover {
    color: #2563eb;
  }
}

.style-tag {
  font-size: 12px;
  color: #525252;

  &.muted {
    color: #a3a3a3;
  }
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

// Override Element Plus table styles to match app theme
:deep(.el-table) {
  --el-table-border-color: #f0f0f0;
  --el-table-header-bg-color: #fafafa;
  --el-table-row-hover-bg-color: #f9f9f9;
  font-size: 13px;
}

:deep(.el-table th.el-table__cell) {
  font-weight: 600;
  color: #737373;
  font-size: 12px;
}

:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: #f5f5f5 !important; color: #525252 !important; }
:deep(.el-tag--danger) { background: #fef2f2 !important; color: #dc2626 !important; }
:deep(.el-tag--success) { background: #f0fdf4 !important; color: #16a34a !important; }

:deep(.el-input__wrapper) {
  background: #fff !important;
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

:deep(.el-select__wrapper) {
  border-radius: 8px !important;
}

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
}

:deep(.el-button--danger) { color: #ef4444 !important; }

:deep(.el-button--default) {
  background: #fff !important;
  border: 1px solid #e8e8e8 !important;
  color: #0a0a0a !important;
  border-radius: 8px !important;
  &:hover { border-color: #0a0a0a !important; }
}

@media (max-width: 1024px) {
  .article-history-main { padding: 16px; max-width: 100%; }
}

@media (max-width: 768px) {
  .mono-header { height: 52px; padding: 0 12px; }
  .header-desc, .header-divider { display: none; }
  .article-history-main { padding: 12px; }

  .toolbar {
    flex-wrap: wrap;
  }

  .search-input {
    width: 100%;
  }
}
</style>
