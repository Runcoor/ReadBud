<template>
  <div class="title-selector">
    <div
      v-for="(candidate, idx) in candidates"
      :key="idx"
      class="title-card"
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
  gap: $spacing-sm;
}

.title-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $spacing-md;
  padding: $spacing-md $spacing-base;
  border: 2px solid $color-border;
  border-radius: $radius-lg;
  background: $color-card-bg;
  cursor: pointer;
  transition: all $transition-fast;

  &:hover {
    border-color: rgba($color-accent, 0.4);
    box-shadow: $shadow-card-hover;
  }

  &.is-selected {
    border-color: $color-accent;
    background: rgba($color-accent, 0.04);
  }
}

.title-text {
  flex: 1;
  font-size: $font-size-base;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  line-height: $line-height-normal;
}

// Responsive: Mobile — compact title cards
@media (max-width: $breakpoint-sm) {
  .title-card {
    padding: $spacing-sm $spacing-md;
    gap: $spacing-sm;
  }

  .title-text {
    font-size: $font-size-sm;
  }
}
</style>
