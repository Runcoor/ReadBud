<template>
  <div class="draft-preview">
    <div v-if="!draftId" class="empty-state">
      <el-empty description="任务完成后将在此显示文章预览" />
    </div>
    <div v-else-if="loading" class="loading-state">
      <el-skeleton :rows="8" animated />
    </div>
    <div v-else-if="fetchError" class="error-state">
      <el-result icon="warning" title="加载失败" :sub-title="fetchError">
        <template #extra>
          <el-button size="small" type="primary" @click="retryFetch">重试</el-button>
        </template>
      </el-result>
    </div>
    <div v-else-if="draft" class="phone-frame">
      <div class="phone-notch" />
      <div class="phone-screen">
        <div v-if="draft.cover_url" class="cover-image">
          <img :src="draft.cover_url" alt="cover" />
        </div>
        <h1 class="draft-title">{{ draft.title }}</h1>
        <p v-if="draft.subtitle" class="draft-subtitle">{{ draft.subtitle }}</p>
        <div class="draft-meta">
          <span class="author">{{ draft.author_name }}</span>
          <span class="date">{{ formatDate(draft.created_at) }}</span>
        </div>
        <div class="draft-digest">{{ draft.digest }}</div>
        <div class="draft-blocks">
          <template v-for="block in sortedBlocks" :key="block.id">
            <div v-if="block.block_type === 'title'" class="block block-title">
              <h1>{{ block.heading || block.text_md }}</h1>
            </div>
            <div v-else-if="block.block_type === 'lead'" class="block block-lead">
              <div class="lead-box">{{ block.text_md }}</div>
            </div>
            <div v-else-if="block.block_type === 'section'" class="block block-section">
              <h2 v-if="block.heading">{{ block.heading }}</h2>
              <div
                v-if="block.html_fragment"
                class="section-body"
                v-html="block.html_fragment"
              />
              <p v-else-if="block.text_md" class="section-body">{{ block.text_md }}</p>
            </div>
            <div v-else-if="block.block_type === 'image'" class="block block-image">
              <img
                v-if="block.asset_url"
                :src="block.asset_url"
                :alt="block.attribution_text || ''"
              />
              <span v-if="block.attribution_text" class="image-caption">
                {{ block.attribution_text }}
              </span>
            </div>
            <div v-else-if="block.block_type === 'chart'" class="block block-chart">
              <div class="chart-placeholder">
                <el-icon :size="32"><DataLine /></el-icon>
                <span>{{ block.heading || '数据图表' }}</span>
              </div>
            </div>
            <div v-else-if="block.block_type === 'quote'" class="block block-quote">
              <blockquote>
                <p>{{ block.text_md }}</p>
                <footer v-if="block.attribution_text">
                  &mdash; {{ block.attribution_text }}
                </footer>
              </blockquote>
            </div>
            <div v-else-if="block.block_type === 'cta'" class="block block-cta">
              <div class="cta-card">
                <h3 v-if="block.heading">{{ block.heading }}</h3>
                <p v-if="block.text_md">{{ block.text_md }}</p>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { DataLine } from '@element-plus/icons-vue'
import { getDraft } from '@/api/draft'
import type { DraftVO, BlockVO } from '@/types/draft'

interface Props {
  draftId: string | null
}

const props = defineProps<Props>()

const draft = ref<DraftVO | null>(null)
const loading = ref(false)
const fetchError = ref<string | null>(null)

const sortedBlocks = computed<BlockVO[]>(() => {
  if (!draft.value) return []
  return [...draft.value.blocks].sort((a, b) => a.sort_no - b.sort_no)
})

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function fetchDraft(id: string): Promise<void> {
  loading.value = true
  fetchError.value = null
  try {
    const res = await getDraft(id)
    draft.value = res.data
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载文章预览失败'
    fetchError.value = msg
    draft.value = null
  } finally {
    loading.value = false
  }
}

function retryFetch(): void {
  if (props.draftId) {
    fetchDraft(props.draftId)
  }
}

watch(
  () => props.draftId,
  (newId) => {
    if (newId) {
      fetchDraft(newId)
    } else {
      draft.value = null
    }
  },
  { immediate: true },
)
</script>

<style lang="scss" scoped>
.draft-preview {
  display: flex;
  justify-content: center;
  padding: $spacing-lg;
}

.empty-state,
.loading-state,
.error-state {
  width: 100%;
  max-width: 375px;
  margin: 0 auto;
  padding: $spacing-3xl $spacing-base;
}

.phone-frame {
  position: relative;
  width: 375px;
  min-height: 600px;
  background: $color-card-bg;
  border: 2px solid $color-border;
  border-radius: $radius-xl * 3;
  box-shadow: $shadow-card-hover;
  overflow: hidden;
}

.phone-notch {
  width: 120px;
  height: 24px;
  background: $color-border;
  border-radius: 0 0 $radius-lg $radius-lg;
  margin: 0 auto;
}

.phone-screen {
  padding: $spacing-base;
  max-height: 80vh;
  overflow-y: auto;
}

.cover-image {
  margin: 0 (-$spacing-base) $spacing-base;

  img {
    width: 100%;
    display: block;
  }
}

.draft-title {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
  line-height: $line-height-tight;
  margin-bottom: $spacing-sm;
}

.draft-subtitle {
  font-size: $font-size-base;
  color: $color-text-secondary;
  margin-bottom: $spacing-md;
}

.draft-meta {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin-bottom: $spacing-base;
  padding-bottom: $spacing-sm;
  border-bottom: 1px solid $color-divider;
}

.draft-digest {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  line-height: $line-height-relaxed;
  padding: $spacing-md;
  background: rgba($color-accent, 0.06);
  border-left: 3px solid $color-accent;
  border-radius: 0 $radius-sm $radius-sm 0;
  margin-bottom: $spacing-lg;
}

.block {
  margin-bottom: $spacing-base;
}

.block-title h1 {
  font-size: $font-size-lg;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.block-lead .lead-box {
  padding: $spacing-md;
  background: rgba($color-accent, 0.08);
  border-radius: $radius-base;
  font-size: $font-size-base;
  color: $color-primary;
  line-height: $line-height-relaxed;
}

.block-section {
  h2 {
    font-size: $font-size-md;
    font-weight: $font-weight-semibold;
    color: $color-text-primary;
    margin-bottom: $spacing-sm;
  }

  .section-body {
    font-size: $font-size-base;
    color: $color-text-primary;
    line-height: $line-height-relaxed;
  }
}

.block-image {
  text-align: center;

  img {
    max-width: 100%;
    border-radius: $radius-base;
  }

  .image-caption {
    display: block;
    font-size: $font-size-xs;
    color: $color-text-muted;
    margin-top: $spacing-xs;
  }
}

.block-chart .chart-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-2xl;
  background: $color-bg;
  border: 1px dashed $color-border;
  border-radius: $radius-base;
  color: $color-text-muted;
  font-size: $font-size-sm;
}

.block-quote blockquote {
  margin: 0;
  padding: $spacing-md $spacing-base;
  border-left: 3px solid $color-metal;
  background: $color-bg;
  border-radius: 0 $radius-sm $radius-sm 0;

  p {
    font-size: $font-size-base;
    color: $color-text-secondary;
    line-height: $line-height-relaxed;
    font-style: italic;
    margin: 0;
  }

  footer {
    font-size: $font-size-xs;
    color: $color-text-muted;
    margin-top: $spacing-sm;
  }
}

.block-cta .cta-card {
  padding: $spacing-base;
  background: linear-gradient(135deg, $color-primary, $color-accent);
  border-radius: $radius-lg;
  color: #fff;

  h3 {
    font-size: $font-size-md;
    font-weight: $font-weight-semibold;
    margin-bottom: $spacing-xs;
  }

  p {
    font-size: $font-size-sm;
    opacity: 0.9;
    margin: 0;
  }
}
</style>
