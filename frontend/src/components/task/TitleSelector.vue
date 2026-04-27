<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="title-selector">
    <div
      v-for="(candidate, idx) in candidates"
      :key="idx"
      class="mono-title-card"
      :class="{ 'is-selected': idx === selectedIndex }"
      @click="handleSelect(idx)"
    >
      <div class="title-text">{{ candidate.title }}</div>
      <el-tag size="small" :type="typeTagColor(candidate.type)" disable-transitions>
        {{ candidate.type }}
      </el-tag>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TitleCandidate } from '@/types/draft'

interface Props {
  candidates: TitleCandidate[]
  selectedIndex: number
}

defineProps<Props>()

const emit = defineEmits<{
  select: [index: number]
}>()

const TYPE_COLORS: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
  seo: '',
  creative: 'success',
  question: 'warning',
  listicle: 'danger',
  how_to: 'info',
}

function typeTagColor(type: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  return TYPE_COLORS[type] || 'info'
}

function handleSelect(index: number): void {
  emit('select', index)
}
</script>

<style lang="scss" scoped>
.title-selector {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.mono-title-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 18px;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  background: var(--surface-card);
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    border-color: #525252;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }

  &.is-selected {
    border-color: #0a0a0a;
    background: #fafafa;
  }

  &:active {
    transform: scale(0.98);
  }
}

.title-text {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  line-height: 1.4;
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

@media (max-width: 768px) {
  .mono-title-card {
    padding: 10px 14px;
    gap: 10px;
  }

  .title-text {
    font-size: 13px;
  }
}
</style>
