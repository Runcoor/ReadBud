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
  background-color: $color-bg;
  padding: $spacing-xl;
}

.not-found-card {
  text-align: center;
  max-width: 480px;
  padding: $spacing-4xl $spacing-3xl;
  background: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-xl;
  box-shadow: $shadow-card;
}

.not-found-code {
  font-size: 72px;
  font-weight: $font-weight-bold;
  color: $color-border;
  line-height: 1;
  margin-bottom: $spacing-lg;
  letter-spacing: -2px;
}

.not-found-title {
  font-size: $font-size-2xl;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0 0 $spacing-md;
}

.not-found-desc {
  font-size: $font-size-base;
  color: $color-text-secondary;
  line-height: $line-height-relaxed;
  margin: 0 0 $spacing-2xl;
}

.not-found-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-md;
}

@media (max-width: $breakpoint-sm) {
  .not-found-card {
    padding: $spacing-2xl $spacing-xl;
  }

  .not-found-code {
    font-size: 56px;
  }

  .not-found-actions {
    flex-direction: column;
  }
}
</style>
