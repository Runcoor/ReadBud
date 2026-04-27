<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <router-view />
</template>

<script setup lang="ts">
import { useTheme } from '@/composables/useTheme'

// Initialize theme on mount (triggers watchEffect to sync data-theme attribute)
useTheme()
</script>

<style lang="scss">
#app {
  min-height: 100vh;
  background: var(--surface-bg);
}
</style>
