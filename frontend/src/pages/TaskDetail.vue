<template>
  <div class="task-detail">
    <el-page-header @back="router.back()" title="返回工作台">
      <template #content>
        <span class="task-detail-title">任务详情</span>
      </template>
    </el-page-header>

    <div class="task-detail-body">
      <el-empty description="任务详情页待实现" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style lang="scss" scoped>
.task-detail {
  padding: $spacing-xl;
  min-height: 100vh;
  background-color: $color-bg;
}

.task-detail-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.task-detail-body {
  margin-top: $spacing-xl;
  background-color: $color-card-bg;
  border-radius: $radius-lg;
  padding: $spacing-2xl;
  border: 1px solid $color-border;
}
</style>
