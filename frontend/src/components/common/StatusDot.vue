<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <span class="status-dot" :style="dotStyle" />
</template>

<script setup lang="ts">
import { computed } from 'vue'

type StatusKind = 'sprout' | 'danger' | 'warn' | 'ink' | 'mute'

const props = withDefaults(
  defineProps<{
    kind?: StatusKind
    color?: string
    size?: number
  }>(),
  {
    kind: 'sprout',
    size: 6,
  },
)

const colorMap: Record<StatusKind, string> = {
  sprout: 'var(--brand-sprout)',
  danger: 'var(--brand-danger)',
  warn: 'var(--brand-warn)',
  ink: 'var(--brand-ink)',
  mute: 'var(--text-tertiary)',
}

const dotStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  background: props.color ?? colorMap[props.kind],
}))
</script>

<style scoped lang="scss">
.status-dot {
  display: inline-block;
  border-radius: 50%;
  flex-shrink: 0;
}
</style>
