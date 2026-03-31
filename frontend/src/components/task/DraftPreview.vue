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
              <div v-if="block.html_fragment" class="lead-box" v-html="block.html_fragment" />
              <div v-else class="lead-box" v-html="block.text_md" />
            </div>
            <div v-else-if="block.block_type === 'section'" class="block block-section">
              <h2 v-if="block.heading">{{ block.heading }}</h2>
              <div
                v-if="block.html_fragment"
                class="section-body"
                v-html="block.html_fragment"
              />
              <div v-else-if="block.text_md" class="section-body" v-html="block.text_md" />
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
              <div v-if="block.html_fragment" v-html="block.html_fragment" />
              <div v-else-if="block.text_md && block.text_md.includes('<')" v-html="block.text_md" />
              <div v-else class="cta-card">
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
  padding: 20px;
}

.empty-state,
.loading-state,
.error-state {
  width: 100%;
  max-width: 375px;
  margin: 0 auto;
  padding: 48px 16px;
}

:deep(.el-empty__description p) {
  color: var(--text-secondary) !important;
}

:deep(.el-skeleton) {
  --el-skeleton-color: var(--surface-tertiary);
  --el-skeleton-to-color: var(--border-light);
}

:deep(.el-result__title) {
  color: var(--text-primary) !important;
}

:deep(.el-result__subtitle) {
  color: var(--text-secondary) !important;
}

:deep(.el-button--primary) {
  background: var(--text-primary) !important;
  border-color: var(--text-primary) !important;
  color: var(--text-inverse) !important;
  border-radius: 8px !important;
}

// Phone frame
.phone-frame {
  position: relative;
  width: 375px;
  min-height: 600px;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: 40px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.phone-notch {
  width: 120px;
  height: 24px;
  background: var(--surface-tertiary);
  border-radius: 0 0 14px 14px;
  margin: 0 auto;
}

.phone-screen {
  padding: 16px;
  max-height: 80vh;
  overflow-y: auto;
}

.cover-image {
  margin: 0 (-16px) 16px;

  img {
    width: 100%;
    display: block;
  }
}

.draft-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
  margin-bottom: 8px;
}

.draft-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.draft-meta {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 12px;
  color: var(--border-medium);
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-light);
}

.draft-digest {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  padding: 12px;
  background: var(--surface-tertiary);
  border-left: 3px solid var(--text-primary);
  border-radius: 0 8px 8px 0;
  margin-bottom: 20px;
}

.block {
  margin-bottom: 14px;
}

.block-title h1 {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.block-lead .lead-box {
  padding: 12px;
  background: var(--surface-tertiary);
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.block-section {
  h2 {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 8px;
  }

  .section-body {
    font-size: 14px;
    color: var(--text-primary);
    line-height: 1.7;
  }
}

.block-image {
  text-align: center;

  img {
    max-width: 100%;
    border-radius: 8px;
  }

  .image-caption {
    display: block;
    font-size: 12px;
    color: var(--border-medium);
    margin-top: 6px;
  }
}

.block-chart .chart-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 32px;
  background: var(--surface-secondary);
  border: 1px dashed var(--border-light);
  border-radius: 8px;
  color: var(--border-medium);
  font-size: 13px;
}

.block-quote blockquote {
  margin: 0;
  padding: 12px 16px;
  border-left: 3px solid var(--border-light);
  background: var(--surface-secondary);
  border-radius: 0 8px 8px 0;

  p {
    font-size: 14px;
    color: var(--text-secondary);
    line-height: 1.6;
    font-style: italic;
    margin: 0;
  }

  footer {
    font-size: 12px;
    color: var(--border-medium);
    margin-top: 8px;
  }
}

.block-cta .cta-card {
  padding: 16px;
  background: var(--text-primary);
  border-radius: 8px;
  color: var(--text-inverse);

  h3 {
    font-size: 15px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  p {
    font-size: 13px;
    opacity: 0.8;
    margin: 0;
  }
}

@media (max-width: 1024px) {
  .draft-preview {
    padding: 14px;
  }

  .phone-frame {
    width: 100%;
    max-width: 375px;
    border-radius: 32px;
  }

  .empty-state,
  .loading-state,
  .error-state {
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .draft-preview {
    padding: 8px;
  }

  .phone-frame {
    border-radius: 20px;
    min-height: 400px;
    border-width: 1px;
  }

  .phone-notch {
    width: 80px;
    height: 16px;
  }

  .phone-screen {
    padding: 12px;
    max-height: 60vh;
  }

  .draft-title {
    font-size: 17px;
  }
}
</style>
