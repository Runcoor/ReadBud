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
  gap: $spacing-md;
}

.block-item {
  border: 1px solid $color-border;
  border-radius: $radius-base;
  background: $color-card-bg;
  transition: box-shadow $transition-fast;

  &.is-editable:hover {
    box-shadow: $shadow-card-hover;

    .block-toolbar {
      opacity: 1;
    }
  }
}

.block-header {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-md;
  border-bottom: 1px solid $color-divider;
  background: $color-bg;
  border-radius: $radius-base $radius-base 0 0;
}

.block-type-icon {
  margin-right: 2px;
  vertical-align: middle;
}

.block-heading {
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.block-toolbar {
  display: flex;
  gap: $spacing-xs;
  margin-left: auto;
  opacity: 0;
  transition: opacity $transition-fast;
}

.block-content {
  padding: $spacing-md;
  font-size: $font-size-base;
  color: $color-text-primary;
  line-height: $line-height-relaxed;

  p {
    margin: 0;
  }

  .block-asset img {
    max-width: 100%;
    border-radius: $radius-sm;
  }

  .text-muted {
    color: $color-text-muted;
    font-style: italic;
  }
}
</style>
