<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-brand">
        <h1 class="brand-name">阅芽</h1>
        <p class="brand-tagline">让写作从一个词开始生长</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label class="form-label">用户名</label>
          <div class="input-wrapper" :class="{ 'is-focus': usernameFocused }">
            <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
            <input
              v-model="form.username"
              type="text"
              placeholder="请输入用户名"
              class="mono-input"
              @focus="usernameFocused = true"
              @blur="usernameFocused = false"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">密码</label>
          <div class="input-wrapper" :class="{ 'is-focus': passwordFocused }">
            <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入密码"
              class="mono-input"
              @focus="passwordFocused = true"
              @blur="passwordFocused = false"
            />
            <button
              type="button"
              class="password-toggle"
              @click="showPassword = !showPassword"
              tabindex="-1"
            >
              <svg v-if="!showPassword" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
              <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                <line x1="1" y1="1" x2="23" y2="23" />
              </svg>
            </button>
          </div>
        </div>

        <div v-if="errorMsg" class="error-message">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
          {{ errorMsg }}
        </div>

        <button
          type="submit"
          class="login-btn"
          :disabled="loading"
        >
          <span v-if="loading" class="btn-spinner"></span>
          <span v-else>登录</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const errorMsg = ref('')
const showPassword = ref(false)
const usernameFocused = ref(false)
const passwordFocused = ref(false)

const form = reactive({
  username: '',
  password: '',
})

async function handleLogin() {
  errorMsg.value = ''

  if (!form.username || form.username.length < 2) {
    errorMsg.value = '请输入用户名（至少 2 个字符）'
    return
  }
  if (!form.password || form.password.length < 6) {
    errorMsg.value = '请输入密码（至少 6 个字符）'
    return
  }

  loading.value = true
  try {
    await authStore.login({
      username: form.username,
      password: form.password,
    })
    ElMessage.success('登录成功')
    router.push({ name: 'Workbench' })
  } catch (err: unknown) {
    errorMsg.value = err instanceof Error ? err.message : '登录失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
@use '@/styles/tokens' as *;

.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: $spacing-xl;
  background: var(--surface-bg);
}

.login-container {
  width: 360px;
  max-width: 100%;
}

// --- Brand ---
.login-brand {
  text-align: center;
  margin-bottom: 48px;
}

.brand-name {
  font-size: $font-size-3xl;
  font-weight: $font-weight-bold;
  color: var(--text-primary);
  letter-spacing: 0.04em;
  margin-bottom: 8px;
}

.brand-tagline {
  font-size: $font-size-base;
  color: $text-tertiary;
  letter-spacing: 0.02em;
}

// --- Form ---
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: $text-secondary;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--border-light);
  border-radius: $radius-md;
  padding: 0 12px;
  transition: all $transition-base;
  background: var(--surface-bg);

  &:hover {
    border-color: var(--border-medium);
  }

  &.is-focus {
    border-color: var(--text-primary);
    box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.06);
  }
}

.input-icon {
  flex-shrink: 0;
  color: $text-tertiary;
  transition: color $transition-base;

  .is-focus & {
    color: var(--text-primary);
  }
}

.mono-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  padding: 12px 0;
  color: var(--text-primary);
  font-size: $font-size-base;
  line-height: 1.5;

  &::placeholder {
    color: $text-placeholder;
  }
}

.password-toggle {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  padding: 4px;
  color: $text-tertiary;
  cursor: pointer;
  transition: color $transition-fast;
  border-radius: $radius-sm;

  &:hover {
    color: $text-secondary;
  }
}

// --- Error ---
.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: $radius-md;
  color: $status-danger;
  font-size: $font-size-sm;
}

// --- Button ---
.login-btn {
  width: 100%;
  padding: 12px 24px;
  font-size: $font-size-md;
  font-weight: $font-weight-medium;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 4px;
  background: var(--text-primary);
  color: var(--text-inverse);
  border: 1px solid var(--text-primary);
  border-radius: $radius-md;
  transition: all $transition-base;

  &:hover {
    background: var(--surface-inverse);
    border-color: var(--surface-inverse);
  }

  &:active {
    transform: scale(0.98);
  }

  &:focus-visible {
    outline: 2px solid #000;
    outline-offset: 2px;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
}

.btn-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// --- Responsive ---
@media (max-width: $breakpoint-sm) {
  .login-page {
    align-items: flex-start;
    padding-top: 12vh;
    padding-left: $spacing-base;
    padding-right: $spacing-base;
  }

  .login-container {
    width: 100%;
  }

  .login-brand {
    margin-bottom: 32px;
  }

  .brand-name {
    font-size: $font-size-2xl;
  }

  .brand-tagline {
    font-size: $font-size-sm;
  }
}
</style>
