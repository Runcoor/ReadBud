<template>
  <div class="workbench">
    <header class="workbench-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">内容工作台</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Settings' })">设置</el-button>
        <span v-if="authStore.user" class="header-user">{{ authStore.user.nickname }}</span>
        <el-button text @click="handleLogout">退出</el-button>
      </div>
    </header>

    <main class="workbench-main">
      <aside class="panel panel-left">
        <div class="panel-header">
          <h2 class="panel-title">任务配置</h2>
        </div>
        <div class="panel-body">
          <!-- TaskForm component will go here (HY-295) -->
          <el-empty description="任务表单待实现" :image-size="80" />
        </div>
      </aside>

      <section class="panel panel-center">
        <div class="panel-header">
          <h2 class="panel-title">执行流程</h2>
        </div>
        <div class="panel-body">
          <!-- TaskProgress component will go here (HY-296) -->
          <el-empty description="执行流程待实现" :image-size="80" />
        </div>
      </section>

      <aside class="panel panel-right">
        <div class="panel-header">
          <h2 class="panel-title">文章预览</h2>
        </div>
        <div class="panel-body">
          <!-- DraftPreview + PublishPanel will go here (HY-315, HY-324) -->
          <el-empty description="文章预览待实现" :image-size="80" />
        </div>
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

async function handleLogout() {
  try {
    await ElMessageBox.confirm('确定退出登录？', '退出', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info',
    })
    authStore.logout()
    router.push({ name: 'Login' })
  } catch {
    // User cancelled
  }
}
</script>

<style lang="scss" scoped>
.workbench {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $color-bg;
}

.workbench-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 $spacing-xl;
  background-color: $color-card-bg;
  border-bottom: 1px solid $color-border;
  box-shadow: $shadow-card;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.header-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-bold;
  color: $color-primary;
}

.header-divider {
  color: $color-border;
}

.header-desc {
  font-size: $font-size-base;
  color: $color-text-secondary;
}

.header-user {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  margin-right: $spacing-sm;
}

.workbench-main {
  display: grid;
  grid-template-columns: 320px 1fr 420px;
  gap: $spacing-base;
  flex: 1;
  padding: $spacing-base;
  overflow: hidden;
}

.panel {
  display: flex;
  flex-direction: column;
  background-color: $color-card-bg;
  border-radius: $radius-lg;
  border: 1px solid $color-border;
  overflow: hidden;
}

.panel-header {
  padding: $spacing-base $spacing-lg;
  border-bottom: 1px solid $color-divider;
}

.panel-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.panel-body {
  flex: 1;
  padding: $spacing-lg;
  overflow-y: auto;
}

@media (max-width: $breakpoint-md) {
  .workbench-main {
    grid-template-columns: 1fr;
    overflow-y: auto;
  }
}
</style>
