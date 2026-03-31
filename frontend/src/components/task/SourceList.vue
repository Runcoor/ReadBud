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
                color="#5B8DEF"
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
  if (score >= 80) return '#FF4D4F'
  if (score >= 50) return '#FAAD14'
  return '#52C41A'
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
  gap: $spacing-md;
}

.source-collapse {
  border: none;

  :deep(.el-collapse-item__header) {
    height: auto;
    line-height: normal;
    background: transparent;
    border: none;
    padding: $spacing-xs 0;
    font-size: $font-size-xs;
    color: $color-accent;
  }

  :deep(.el-collapse-item__wrap) {
    border: none;
    background: transparent;
  }

  :deep(.el-collapse-item__content) {
    padding-bottom: 0;
  }
}

.source-card {
  padding: $spacing-md;
  border: 1px solid $color-border;
  border-radius: $radius-base;
  background: $color-card-bg;
  margin-bottom: $spacing-sm;
  transition: box-shadow $transition-fast;

  &:hover {
    box-shadow: $shadow-card-hover;
  }
}

.source-title-row {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  margin-bottom: $spacing-sm;
}

.source-title {
  flex: 1;
  font-size: $font-size-base;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.external-link {
  display: flex;
  align-items: center;
  color: $color-text-muted;
  transition: color $transition-fast;

  &:hover {
    color: $color-accent;
  }
}

.source-meta {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin-bottom: $spacing-md;
}

.score-bars {
  display: flex;
  gap: $spacing-lg;
}

.score-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  .score-label {
    font-size: $font-size-xs;
    color: $color-text-secondary;
    white-space: nowrap;
  }

  :deep(.el-progress) {
    flex: 1;
  }
}

.source-summary {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  line-height: $line-height-relaxed;
  margin: 0;
  padding: $spacing-sm 0;
}
</style>
