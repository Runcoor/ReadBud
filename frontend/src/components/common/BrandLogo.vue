<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="brand-logo" :style="{ '--logo-size': `${size}px` }">
    <div class="brand-logo__mark" :class="{ 'brand-logo__mark--dark': dark }">
      <div class="brand-logo__dot" />
    </div>
    <span v-if="showLabel" class="brand-logo__text">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    size?: number
    showLabel?: boolean
    label?: string
    dark?: boolean
  }>(),
  {
    size: 18,
    showLabel: true,
    label: '阅芽',
    dark: false,
  },
)
</script>

<style scoped lang="scss">
.brand-logo {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);

  &__mark {
    width: var(--logo-size);
    height: var(--logo-size);
    border: 1px solid currentColor;
    display: grid;
    place-items: center;
    flex-shrink: 0;

    &--dark {
      color: var(--text-inverse);
    }
  }

  &__dot {
    width: calc(var(--logo-size) / 6);
    height: calc(var(--logo-size) / 6);
    background: currentColor;
  }

  &__text {
    font-size: 14px;
    font-weight: 600;
    letter-spacing: 0.02em;
    color: inherit;
  }
}
</style>
