<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <header class="app-topbar">
    <BrandLogo @click="goHome" style="cursor: pointer" />
    <span class="app-topbar__sep" />
    <span class="app-topbar__crumb">{{ crumb }}</span>
    <div class="app-topbar__spacer" />
    <slot name="right" />
    <slot name="user">
      <div v-if="userInitial" class="app-topbar__avatar">{{ userInitial }}</div>
    </slot>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import BrandLogo from './BrandLogo.vue'

defineProps<{
  crumb?: string
  userInitial?: string
}>()

const router = useRouter()
const goHome = () => router.push('/')
</script>

<style scoped lang="scss">
.app-topbar {
  height: 48px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--border-hair);
  background: var(--surface-card);
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 16px;
  font-family: var(--font-sans);
  color: var(--text-primary);
  position: sticky;
  top: 0;
  z-index: 10;

  &__sep {
    width: 1px;
    height: 16px;
    background: var(--border-hair);
  }

  &__crumb {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  &__spacer {
    flex: 1;
  }

  &__avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--brand-ink);
    color: var(--text-inverse);
    display: grid;
    place-items: center;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
  }
}
</style>
