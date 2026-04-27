<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="draft-preview" :class="`draft-preview--${mode}`">
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

    <!-- Markdown source mode -->
    <pre v-else-if="draft && mode === 'markdown'" class="md-source"><code>{{ markdownSource }}</code></pre>

    <!-- Mobile / Desktop rendered mode -->
    <article v-else-if="draft" class="rendered" :class="`rendered--${mode}`">
      <div class="cover-image" :class="{ 'cover-image--empty': !draft.cover_url }">
        <img v-if="draft.cover_url" :src="draft.cover_url" alt="cover" />
        <div v-else class="cover-image__placeholder">
          <el-icon :size="40" class="cover-image__placeholder-icon"><Picture /></el-icon>
          <p class="cover-image__placeholder-text">尚未生成封面</p>
          <p class="cover-image__placeholder-hint">公众号发布需要封面图</p>
        </div>
        <button
          v-if="!coverGenerating"
          type="button"
          class="cover-image__action"
          :class="{ 'cover-image__action--primary': !draft.cover_url }"
          @click="regenerateCover"
        >
          <el-icon><Refresh /></el-icon>
          <span>{{ draft.cover_url ? '重新生成' : '一键生成封面' }}</span>
        </button>
        <div v-if="coverGenerating" class="cover-image__overlay">
          <el-icon class="is-loading" :size="24"><Refresh /></el-icon>
          <span class="cover-image__overlay-text">AI 正在为这篇文章绘制封面…</span>
          <span class="cover-image__overlay-hint">通常需要 10–30 秒</span>
        </div>
      </div>
      <h1 class="draft-title">{{ draft.title }}</h1>
      <p v-if="draft.subtitle" class="draft-subtitle">{{ draft.subtitle }}</p>
      <div class="draft-meta">
        <span class="author">{{ draft.author_name }}</span>
        <span class="date">{{ formatDate(draft.created_at) }}</span>
      </div>
      <div v-if="draft.digest" class="draft-digest">{{ draft.digest }}</div>
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
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { DataLine, Refresh, Picture } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDraft, regenerateDraftCover } from '@/api/draft'
import type { DraftVO, BlockVO } from '@/types/draft'

interface Props {
  draftId: string | null
  mode?: 'mobile' | 'desktop' | 'markdown'
}

const props = withDefaults(defineProps<Props>(), { mode: 'mobile' })

const draft = ref<DraftVO | null>(null)
const loading = ref(false)
const fetchError = ref<string | null>(null)
const coverGenerating = ref(false)

async function regenerateCover(): Promise<void> {
  if (!draft.value || coverGenerating.value) return
  if (draft.value.cover_url) {
    try {
      await ElMessageBox.confirm(
        '将根据文章标题与风格预设重新生成一张 AI 封面，原封面会被替换。',
        '重新生成封面',
        { confirmButtonText: '继续', cancelButtonText: '取消', type: 'info' },
      )
    } catch {
      return
    }
  }
  coverGenerating.value = true
  try {
    const res = await regenerateDraftCover(draft.value.id)
    if (res.data?.url && draft.value) {
      draft.value.cover_url = res.data.url
      ElMessage.success('封面已生成')
    } else {
      ElMessage.warning('未拿到封面 URL，请重试')
    }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '封面生成失败'
    ElMessage.error(msg + '，可点击按钮重试')
  } finally {
    coverGenerating.value = false
  }
}

const sortedBlocks = computed<BlockVO[]>(() => {
  if (!draft.value) return []
  return [...draft.value.blocks].sort((a, b) => a.sort_no - b.sort_no)
})

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function stripHtml(html: string): string {
  return html
    .replace(/<\/(p|div|h[1-6]|li|br)\s*\/?>/gi, '\n')
    .replace(/<br\s*\/?>/gi, '\n')
    .replace(/<[^>]+>/g, '')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

function blockToMarkdown(b: BlockVO): string {
  switch (b.block_type) {
    case 'title':
      return `# ${b.heading || b.text_md || ''}`
    case 'lead': {
      const body = b.text_md || (b.html_fragment ? stripHtml(b.html_fragment) : '')
      return body
        .split('\n')
        .map((l) => `> ${l}`)
        .join('\n')
    }
    case 'section': {
      const head = b.heading ? `## ${b.heading}\n\n` : ''
      const body = b.text_md || (b.html_fragment ? stripHtml(b.html_fragment) : '')
      return head + body
    }
    case 'image': {
      const alt = b.attribution_text || ''
      const url = b.asset_url || ''
      return url ? `![${alt}](${url})` : ''
    }
    case 'chart':
      return `> [图表] ${b.heading || '数据图表'}`
    case 'quote': {
      const lines = (b.text_md || '')
        .split('\n')
        .map((l) => `> ${l}`)
        .join('\n')
      return b.attribution_text ? `${lines}\n> — ${b.attribution_text}` : lines
    }
    case 'cta': {
      const head = b.heading ? `**${b.heading}**\n\n` : ''
      const body = b.text_md
        ? b.text_md.includes('<')
          ? stripHtml(b.text_md)
          : b.text_md
        : b.html_fragment
          ? stripHtml(b.html_fragment)
          : ''
      return head + body
    }
    default:
      return ''
  }
}

const markdownSource = computed<string>(() => {
  if (!draft.value) return ''
  const d = draft.value
  const parts: string[] = []
  parts.push(`# ${d.title}`)
  if (d.subtitle) parts.push(`*${d.subtitle}*`)
  parts.push(`> ${d.author_name} · ${formatDate(d.created_at)}`)
  if (d.digest) parts.push(`> ${d.digest}`)
  if (d.cover_url) parts.push(`![cover](${d.cover_url})`)
  for (const b of sortedBlocks.value) {
    const md = blockToMarkdown(b)
    if (md) parts.push(md)
  }
  return parts.join('\n\n') + '\n'
})

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

defineExpose({
  getMarkdown: (): string => markdownSource.value,
  getTitle: (): string => draft.value?.title || '',
})
</script>

<style lang="scss" scoped>
.draft-preview {
  width: 100%;
}

.empty-state,
.loading-state,
.error-state {
  width: 100%;
  padding: 32px 16px;
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

// === Markdown source mode ===
.md-source {
  margin: 0;
  padding: 0;
  font-family: var(--font-mono);
  font-size: 12.5px;
  line-height: 1.7;
  color: var(--text-primary);
  background: transparent;
  white-space: pre-wrap;
  word-break: break-word;
  user-select: text;

  code {
    font-family: inherit;
    background: transparent;
    padding: 0;
  }
}

// === Rendered article (mobile + desktop) ===
.rendered {
  width: 100%;
  color: var(--text-primary);
}

.cover-image {
  position: relative;
  margin: 0 0 16px;
  border-radius: 8px;
  overflow: hidden;

  img {
    width: 100%;
    display: block;
    aspect-ratio: 16 / 9;
    object-fit: cover;
  }

  &__placeholder {
    aspect-ratio: 16 / 9;
    width: 100%;
    background: linear-gradient(135deg, #f7f7f8 0%, #ececef 100%);
    border: 1px dashed var(--border-color, #d8d8de);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    color: var(--text-secondary, #707078);

    &-icon {
      color: var(--text-tertiary, #a8a8b0);
    }

    &-text {
      margin: 0;
      font-size: 14px;
      font-weight: 500;
      color: var(--text-primary, #1f1f23);
    }

    &-hint {
      margin: 0;
      font-size: 12px;
      color: var(--text-tertiary, #909098);
    }
  }

  &__action {
    position: absolute;
    bottom: 12px;
    right: 12px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 14px;
    border-radius: 999px;
    border: 1px solid rgba(255, 255, 255, 0.6);
    background: rgba(0, 0, 0, 0.62);
    backdrop-filter: blur(8px);
    color: #fff;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.18s ease;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18);

    &:hover {
      background: rgba(0, 0, 0, 0.78);
      transform: translateY(-1px);
    }

    &--primary {
      position: static;
      margin-top: -36px;
      transform: none;
      background: var(--color-primary, #2c2c34);
      border-color: transparent;
      padding: 9px 18px;
      font-size: 13px;

      &:hover {
        background: var(--color-primary-hover, #18181d);
        transform: translateY(-1px);
      }
    }
  }

  &--empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    border: none;
  }

  &__overlay {
    position: absolute;
    inset: 0;
    background: rgba(255, 255, 255, 0.82);
    backdrop-filter: blur(6px);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: var(--text-primary, #1f1f23);
    font-size: 13px;

    &-text {
      font-weight: 500;
    }

    &-hint {
      font-size: 11px;
      color: var(--text-tertiary, #909098);
    }
  }
}

.draft-title {
  font-family: var(--font-serif);
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
  margin: 0 0 8px;
}

.draft-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0 0 12px;
}

.draft-meta {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-hair);
}

.draft-digest {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  padding: 12px;
  background: var(--surface-secondary);
  border-left: 3px solid var(--text-primary);
  border-radius: 0 6px 6px 0;
  margin-bottom: 20px;
}

.block {
  margin-bottom: 14px;
}

.block-title h1 {
  font-family: var(--font-serif);
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.block-lead .lead-box {
  padding: 12px;
  background: var(--surface-secondary);
  border-radius: 6px;
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.block-section {
  h2 {
    font-family: var(--font-serif);
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 8px;
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
    border-radius: 6px;
  }

  .image-caption {
    display: block;
    font-size: 12px;
    color: var(--text-tertiary);
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
  border: 1px dashed var(--border-hair);
  border-radius: 6px;
  color: var(--text-tertiary);
  font-size: 13px;
}

.block-quote blockquote {
  margin: 0;
  padding: 12px 16px;
  border-left: 3px solid var(--border-medium);
  background: var(--surface-secondary);
  border-radius: 0 6px 6px 0;

  p {
    font-size: 14px;
    color: var(--text-secondary);
    line-height: 1.6;
    font-style: italic;
    margin: 0;
  }

  footer {
    font-size: 12px;
    color: var(--text-tertiary);
    margin-top: 8px;
  }
}

.block-cta .cta-card {
  padding: 16px;
  background: var(--brand-ink);
  border-radius: 6px;
  color: var(--text-inverse);

  h3 {
    font-size: 15px;
    font-weight: 600;
    margin: 0 0 4px;
  }

  p {
    font-size: 13px;
    opacity: 0.85;
    margin: 0;
  }
}

// === Desktop typographic upgrade ===
.rendered--desktop {
  max-width: 720px;
  margin: 0 auto;

  .draft-title { font-size: 32px; line-height: 1.25; margin-bottom: 12px; }
  .draft-subtitle { font-size: 16px; margin-bottom: 16px; }
  .draft-meta { font-size: 13px; margin-bottom: 20px; padding-bottom: 14px; }
  .draft-digest { font-size: 14px; padding: 14px 16px; margin-bottom: 28px; }
  .block { margin-bottom: 22px; }
  .block-title h1 { font-size: 24px; }
  .block-section h2 { font-size: 20px; margin-bottom: 10px; }
  .block-section .section-body { font-size: 15px; line-height: 1.85; }
  .block-lead .lead-box { font-size: 15px; padding: 14px 16px; }
  .block-quote blockquote p { font-size: 15px; }
}
</style>
