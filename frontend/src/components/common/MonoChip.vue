<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <span class="mono-chip" :class="kindClass">
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    kind?: 'default' | 'sprout' | 'danger' | 'warn'
  }>(),
  { kind: 'default' },
)

const kindClass = computed(() => `mono-chip--${props.kind}`)
</script>

<style scoped lang="scss">
.mono-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  padding: 2px 6px;
  border-radius: 3px;
  border: 1px solid var(--border-hair);
  color: var(--text-tertiary);
  background: transparent;
  text-transform: uppercase;
  line-height: 1.5;

  &--sprout {
    color: var(--brand-sprout);
    background: var(--brand-sprout-soft);
    border-color: transparent;
  }

  &--danger {
    color: var(--brand-danger);
    background: var(--brand-danger-soft);
    border-color: transparent;
  }

  &--warn {
    color: var(--brand-warn);
    background: var(--brand-warn-soft);
    border-color: transparent;
  }
}
</style>
