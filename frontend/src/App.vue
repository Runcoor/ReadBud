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
