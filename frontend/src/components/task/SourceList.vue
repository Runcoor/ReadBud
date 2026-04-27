<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="source-list">
    <el-collapse v-model="expandedIds" class="source-collapse">
      <div v-for="source in sortedSources" :key="source.id" class="source-card">
        <div class="source-header">
          <div class="source-title-row">
            <el-tag
              size="small"
              :type="sourceTypeBadge(source.source_type)"
              disable-transitions
            >
              {{ sourceTypeLabel(source.source_type) }}
            </el-tag>
            <span class="source-title">{{ source.title }}</span>
            <a
              :href="source.source_url"
              target="_blank"
              rel="noopener noreferrer"
              class="external-link"
              title="打开原文"
            >
              <el-icon><Link /></el-icon>
            </a>
          </div>
          <div class="source-meta">
            <span class="site-name">{{ source.site_name }}</span>
            <span v-if="source.author" class="author">{{ source.author }}</span>
            <span v-if="source.published_at" class="date">
              {{ formatDate(source.published_at) }}
            </span>
          </div>
          <div class="score-bars">
            <div class="score-item">
              <span class="score-label">热度</span>
              <el-progress
                :percentage="source.hot_score"
                :stroke-width="8"
                :color="hotScoreColor(source.hot_score)"
                :show-text="true"
                :format="(val: number) => `${val}`"
              />
            </div>
            <div class="score-item">
              <span class="score-label">相关性</span>
              <el-progress
                :percentage="source.relevance_score"
                :stroke-width="8"
                color="#0a0a0a"
                :show-text="true"
                :format="(val: number) => `${val}`"
              />
            </div>
          </div>
        </div>
        <el-collapse-item
          v-if="source.summary"
          :name="source.id"
          title="查看摘要"
          class="summary-collapse"
        >
          <p class="source-summary">{{ source.summary }}</p>
        </el-collapse-item>
      </div>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Link } from '@element-plus/icons-vue'
import type { SourceVO } from '@/types/draft'

interface Props {
  sources: SourceVO[]
}

const props = defineProps<Props>()

const expandedIds = ref<string[]>([])

const sortedSources = computed<SourceVO[]>(() => {
  return [...props.sources].sort((a, b) => b.hot_score - a.hot_score)
})

const SOURCE_TYPE_LABELS: Record<string, string> = {
  web: '网页',
  news: '新闻',
  wechat: '微信',
  blog: '博客',
}

const SOURCE_TYPE_BADGES: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
  web: 'info',
  news: '',
  wechat: 'success',
  blog: 'warning',
}

function sourceTypeLabel(type: string): string {
  return SOURCE_TYPE_LABELS[type] || type
}

function sourceTypeBadge(type: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  return SOURCE_TYPE_BADGES[type] || 'info'
}

function hotScoreColor(score: number): string {
  if (score >= 80) return '#ef4444'
  if (score >= 50) return '#eab308'
  return '#22c55e'
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}
</script>

<style lang="scss" scoped>
.source-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

// Collapse overrides
.source-collapse {
  border: none;

  :deep(.el-collapse-item__header) {
    height: auto;
    line-height: normal;
    background: transparent;
    border: none;
    padding: 6px 0;
    font-size: 12px;
    color: #525252;
  }

  :deep(.el-collapse-item__wrap) {
    border: none;
    background: transparent;
  }

  :deep(.el-collapse-item__content) {
    padding-bottom: 0;
    color: #525252;
  }

  :deep(.el-collapse-item__arrow) {
    color: #d4d4d4;
  }
}

// Source card
.source-card {
  padding: 14px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  background: var(--surface-card);
  margin-bottom: 8px;
  transition: all 0.15s ease;

  &:hover {
    border-color: #0a0a0a;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }
}

.source-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.source-title {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.external-link {
  display: flex;
  align-items: center;
  color: #d4d4d4;
  transition: color 0.15s ease;

  &:hover {
    color: #0a0a0a;
  }
}

.source-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #d4d4d4;
  margin-bottom: 12px;
}

.score-bars {
  display: flex;
  gap: 20px;
}

.score-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;

  .score-label {
    font-size: 12px;
    color: #525252;
    white-space: nowrap;
  }

  :deep(.el-progress) {
    flex: 1;
  }

  :deep(.el-progress-bar__outer) {
    background-color: #e8e8e8 !important;
  }

  :deep(.el-progress__text) {
    color: #525252 !important;
    font-size: 12px !important;
  }
}

.source-summary {
  font-size: 13px;
  color: #525252;
  line-height: 1.6;
  margin: 0;
  padding: 8px 0;
}

// Tag overrides
:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: #f5f5f5 !important; color: #525252 !important; }
:deep(.el-tag--success) { background: #f0fdf4 !important; color: #16a34a !important; }
:deep(.el-tag--warning) { background: #fefce8 !important; color: #ca8a04 !important; }

@media (max-width: 768px) {
  .source-card {
    padding: 10px;
  }

  .source-title-row {
    gap: 6px;
  }

  .source-title {
    font-size: 13px;
  }

  .source-meta {
    gap: 8px;
    flex-wrap: wrap;
  }

  .score-bars {
    flex-direction: column;
    gap: 8px;
  }
}
</style>
