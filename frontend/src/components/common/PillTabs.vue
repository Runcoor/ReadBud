<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="pill-tabs" :class="{ 'pill-tabs--compact': compact }">
    <button
      v-for="opt in options"
      :key="String(opt.value)"
      type="button"
      class="pill-tabs__item"
      :class="{ 'is-active': opt.value === modelValue }"
      @click="$emit('update:modelValue', opt.value)"
    >
      {{ opt.label }}
    </button>
  </div>
</template>

<script setup lang="ts" generic="T extends string | number">
defineProps<{
  modelValue: T
  options: { label: string; value: T }[]
  compact?: boolean
}>()

defineEmits<{
  (e: 'update:modelValue', value: T): void
}>()
</script>

<style scoped lang="scss">
.pill-tabs {
  display: inline-flex;
  gap: 6px;

  &--compact {
    gap: 4px;
  }

  &__item {
    height: 32px;
    padding: 0 14px;
    border-radius: 4px;
    font-size: 12px;
    font-family: var(--font-sans);
    background: transparent;
    color: var(--text-body);
    border: 1px solid var(--border-hair);
    cursor: pointer;
    transition: all 120ms ease;
    line-height: 1;
    flex: 0 0 auto;
    white-space: nowrap;

    &:hover {
      border-color: var(--border-medium);
      color: var(--text-primary);
    }

    &.is-active {
      background: var(--brand-ink);
      color: var(--text-inverse);
      border-color: var(--brand-ink);
    }
  }

  &--compact &__item {
    height: 26px;
    padding: 0 12px;
    font-size: 11px;
    font-family: var(--font-mono);
    flex: 0 0 auto;
  }
}
</style>
