<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="not-found">
    <div class="not-found-card">
      <div class="not-found-code">404</div>
      <h1 class="not-found-title">页面未找到</h1>
      <p class="not-found-desc">
        抱歉，您访问的页面不存在或已被移动。
      </p>
      <div class="not-found-actions">
        <el-button type="primary" @click="router.push({ name: 'Workbench' })">
          返回工作台
        </el-button>
        <el-button @click="router.back()">
          返回上一页
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style lang="scss" scoped>
.not-found {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 32px;
  background: #fafafa;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', sans-serif;
}

.not-found-card {
  text-align: center;
  max-width: 480px;
  padding: 64px 48px;
}

.not-found-code {
  font-size: 96px;
  font-weight: 800;
  line-height: 1;
  margin-bottom: 20px;
  letter-spacing: -4px;
  color: #0a0a0a;
}

.not-found-title {
  font-size: 24px;
  font-weight: 600;
  color: #0a0a0a;
  margin: 0 0 12px;
}

.not-found-desc {
  font-size: 15px;
  color: #525252;
  line-height: 1.6;
  margin: 0 0 36px;
}

.not-found-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

:deep(.el-button--primary) {
  background: #0a0a0a !important;
  border-color: #0a0a0a !important;
  color: #fff !important;
  border-radius: 8px !important;
  padding: 10px 24px !important;
  &:hover { background: #333 !important; border-color: #333 !important; }
  &:active { transform: scale(0.98); }
}

:deep(.el-button:not(.el-button--primary)) {
  background: var(--surface-card) !important;
  border: 1px solid #e8e8e8 !important;
  color: #0a0a0a !important;
  border-radius: 8px !important;
  padding: 10px 24px !important;
  &:hover {
    border-color: #0a0a0a !important;
  }
}

@media (max-width: 768px) {
  .not-found-card {
    padding: 40px 28px;
  }

  .not-found-code {
    font-size: 72px;
  }

  .not-found-actions {
    flex-direction: column;
  }
}
</style>
