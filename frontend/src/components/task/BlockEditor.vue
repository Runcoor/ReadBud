<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="block-editor">
    <div
      v-for="block in blocks"
      :key="block.id"
      class="block-item"
      :class="{ 'is-editable': editable }"
    >
      <div class="block-header">
        <el-tag size="small" :type="blockTagType(block.block_type)" disable-transitions>
          <el-icon class="block-type-icon"><component :is="blockIcon(block.block_type)" /></el-icon>
          {{ blockLabel(block.block_type) }}
        </el-tag>
        <span v-if="block.heading" class="block-heading">{{ block.heading }}</span>
        <div v-if="editable" class="block-toolbar">
          <el-tooltip content="编辑" placement="top">
            <el-button :icon="EditPen" text size="small" @click="openEdit(block)" />
          </el-tooltip>
          <el-tooltip content="重新生成" placement="top">
            <el-button :icon="RefreshRight" text size="small" @click="handleRegenerate(block)" />
          </el-tooltip>
          <el-tooltip content="删除" placement="top">
            <el-button :icon="Delete" text size="small" type="danger" @click="handleDelete(block)" />
          </el-tooltip>
        </div>
      </div>
      <div class="block-content">
        <div
          v-if="block.html_fragment"
          v-html="block.html_fragment"
        />
        <p v-else-if="block.text_md">{{ block.text_md }}</p>
        <div v-else-if="block.asset_url" class="block-asset">
          <img :src="block.asset_url" :alt="block.attribution_text || ''" />
        </div>
        <span v-else class="text-muted">暂无内容</span>
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      title="编辑区块内容"
      width="600px"
      class="block-edit-dialog"
      destroy-on-close
    >
      <el-input
        v-model="editText"
        type="textarea"
        :rows="10"
        placeholder="输入 Markdown 内容"
      />
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, type Component as VueComponent } from 'vue'
import {
  EditPen,
  RefreshRight,
  Delete,
  Document,
  Promotion,
  Picture,
  DataLine,
  ChatDotSquare,
  Flag,
  Tickets,
} from '@element-plus/icons-vue'
import type { BlockVO } from '@/types/draft'

interface Props {
  blocks: BlockVO[]
  editable: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:block': [block: BlockVO]
  'regenerate:block': [block: BlockVO]
  'delete:block': [block: BlockVO]
}>()

const dialogVisible = ref(false)
const editText = ref('')
const editingBlock = ref<BlockVO | null>(null)

const BLOCK_ICONS: Record<string, VueComponent> = {
  title: Flag,
  lead: Promotion,
  section: Document,
  image: Picture,
  chart: DataLine,
  quote: ChatDotSquare,
  cta: Tickets,
}

const BLOCK_LABELS: Record<string, string> = {
  title: '标题',
  lead: '导语',
  section: '段落',
  image: '图片',
  chart: '图表',
  quote: '引用',
  cta: '行动号召',
}

const BLOCK_TAG_TYPES: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
  title: '',
  lead: 'warning',
  section: 'info',
  image: 'success',
  chart: 'success',
  quote: 'info',
  cta: 'danger',
}

function blockIcon(blockType: string): VueComponent {
  return BLOCK_ICONS[blockType] || Document
}

function blockLabel(blockType: string): string {
  return BLOCK_LABELS[blockType] || blockType
}

function blockTagType(blockType: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  return BLOCK_TAG_TYPES[blockType] || 'info'
}

function openEdit(block: BlockVO): void {
  editingBlock.value = block
  editText.value = block.text_md || block.html_fragment || ''
  dialogVisible.value = true
}

function saveEdit(): void {
  if (editingBlock.value) {
    emit('update:block', {
      ...editingBlock.value,
      text_md: editText.value,
    })
  }
  dialogVisible.value = false
}

function handleRegenerate(block: BlockVO): void {
  emit('regenerate:block', block)
}

function handleDelete(block: BlockVO): void {
  emit('delete:block', block)
}
</script>

<style lang="scss" scoped>
.block-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.block-item {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  background: var(--surface-card);
  transition: all 0.15s ease;

  &.is-editable:hover {
    border-color: #0a0a0a;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .block-toolbar {
      opacity: 1;
    }
  }
}

.block-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-bottom: 1px solid #f5f5f5;
  background: #fafafa;
  border-radius: 8px 8px 0 0;
}

.block-type-icon {
  margin-right: 2px;
  vertical-align: middle;
}

.block-heading {
  font-size: 13px;
  font-weight: 500;
  color: #1a1a1a;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.block-toolbar {
  display: flex;
  gap: 4px;
  margin-left: auto;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.block-content {
  padding: 14px;
  font-size: 14px;
  color: #1a1a1a;
  line-height: 1.6;

  p {
    margin: 0;
  }

  .block-asset img {
    max-width: 100%;
    border-radius: 8px;
  }

  .text-muted {
    color: #d4d4d4;
    font-style: italic;
  }
}

// Tag overrides
:deep(.el-tag) {
  border-radius: 4px !important;
  border: none !important;
}
:deep(.el-tag--info) { background: #f5f5f5 !important; color: #525252 !important; }
:deep(.el-tag--success) { background: #f0fdf4 !important; color: #16a34a !important; }
:deep(.el-tag--warning) { background: #fefce8 !important; color: #ca8a04 !important; }
:deep(.el-tag--danger) { background: #fef2f2 !important; color: #dc2626 !important; }

// Button overrides
:deep(.el-button) {
  color: #525252 !important;
  &:hover {
    color: #0a0a0a !important;
    background: #f5f5f5 !important;
  }
}

:deep(.el-button--danger) {
  color: #ef4444 !important;
}

// Dialog
:deep(.el-dialog) {
  border-radius: 12px !important;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.12) !important;
  .el-dialog__header { border-bottom: 1px solid #e8e8e8; }
  .el-dialog__title { color: #0a0a0a !important; font-weight: 600; }
  .el-dialog__footer { border-top: 1px solid #e8e8e8; }
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

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
}

:deep(.el-button--default) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  color: #0a0a0a !important;
  border-radius: 8px !important;
  &:hover { border-color: #0a0a0a !important; }
}

@media (max-width: 1024px) {
  .block-toolbar {
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .block-header {
    flex-wrap: wrap;
    gap: 4px;
    padding: 8px 10px;
  }

  .block-heading {
    flex-basis: 100%;
    order: 3;
    white-space: normal;
    font-size: 12px;
  }

  .block-toolbar {
    opacity: 1;
    margin-left: auto;
  }

  .block-content {
    padding: 10px;
    font-size: 13px;
  }
}
</style>
