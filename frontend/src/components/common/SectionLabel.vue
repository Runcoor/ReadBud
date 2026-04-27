<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="section-label">
    <div class="section-label__head">
      <span class="section-label__title">{{ title }}</span>
      <span v-if="code" class="section-label__code">{{ code }}</span>
    </div>
    <p v-if="hint" class="section-label__hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  code?: string
  hint?: string
}>()
</script>

<style scoped lang="scss">
.section-label {
  margin-bottom: 18px;

  &__head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
  }

  &__title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  &__code {
    font-size: 10px;
    color: var(--text-tertiary);
    font-family: var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  &__hint {
    font-size: 11px;
    color: var(--text-tertiary);
    margin: 4px 0 0;
    line-height: 1.5;
  }
}
</style>
