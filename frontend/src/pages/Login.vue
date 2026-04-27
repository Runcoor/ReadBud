<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="login-page">
    <!-- Left brand panel -->
    <aside class="brand-panel">
      <svg
        class="brand-motif"
        width="100%"
        height="100%"
        viewBox="0 0 540 800"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        <path
          d="M 270 800 Q 270 500 270 360 Q 270 300 320 280 Q 360 265 380 240"
          fill="none"
          stroke="#F5F3ED"
          stroke-width="1"
        />
        <path
          d="M 270 460 Q 220 440 195 400 Q 178 370 175 340"
          fill="none"
          stroke="#F5F3ED"
          stroke-width="1"
        />
        <circle cx="380" cy="240" r="3" fill="#F5F3ED" />
        <circle cx="175" cy="340" r="3" fill="#F5F3ED" />
      </svg>

      <div class="brand-mark">
        <div class="brand-mark__box">
          <div class="brand-mark__dot"></div>
        </div>
        <span class="brand-mark__name">YUEYA / 阅芽</span>
      </div>

      <div class="brand-hero">
        <h1 class="brand-hero__title">
          让写作<br />从一个词<br />开始生长
        </h1>
        <p class="brand-hero__sub">
          输入一个关键词，阅芽会替你完成扩展、采集、撰写、配图、发布的整条链路。
        </p>
      </div>

      <div class="brand-footer">
        <span>v 2.4.0</span>
        <span>© 2026 YUEYA</span>
      </div>
    </aside>

    <!-- Right form panel -->
    <main class="form-panel">
      <form @submit.prevent="handleLogin" class="login-form" novalidate>
        <header class="form-heading">
          <h2 class="form-heading__title">欢迎回来</h2>
          <p class="form-heading__sub">登录后继续你的写作任务</p>
        </header>

        <div class="form-fields">
          <div class="field">
            <label class="field__label" for="login-username">用户名</label>
            <div class="field__control">
              <input
                id="login-username"
                v-model="form.username"
                type="text"
                placeholder="请输入用户名"
                autocomplete="username"
                class="field__input"
              />
            </div>
          </div>

          <div class="field">
            <label class="field__label" for="login-password">密码</label>
            <div class="field__control field__control--with-action">
              <input
                id="login-password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="请输入密码"
                autocomplete="current-password"
                class="field__input"
              />
              <button
                type="button"
                class="field__toggle"
                aria-label="切换密码可见"
                tabindex="-1"
                @click="showPassword = !showPassword"
              >
                <svg
                  v-if="!showPassword"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
                <svg
                  v-else
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
                  />
                  <line x1="1" y1="1" x2="23" y2="23" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="form-row">
          <label class="remember">
            <input v-model="remember" type="checkbox" />
            <span>记住我</span>
          </label>
          <span class="form-row__forgot">忘记密码？</span>
        </div>

        <p v-if="errorMsg" class="form-error">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
          <span>{{ errorMsg }}</span>
        </p>

        <button
          type="submit"
          class="submit-btn"
          :class="{ 'is-loading': loading }"
          :disabled="loading"
        >
          <span v-if="loading" class="submit-btn__spinner" aria-hidden="true"></span>
          <span class="submit-btn__label">{{ loading ? '登录中…' : '登 录' }}</span>
        </button>

        <p class="form-hint">NEED ACCESS?&nbsp;&nbsp;联系管理员</p>
      </form>
    </main>
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
const remember = ref(false)

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
    // Extract backend message from AxiosError if present, fall back to generic.
    const ax = err as { response?: { data?: { message?: string } }; message?: string }
    errorMsg.value =
      ax?.response?.data?.message ||
      (err instanceof Error ? err.message : '') ||
      '登录失败，请检查用户名或密码'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
@use '@/styles/tokens' as *;

// Local design constants from the design canvas
$cream: #F5F3ED;
$cream-dim: #A8A49A;
$cream-faint: #6F6B62;

.login-page {
  display: flex;
  width: 100vw;
  height: 100vh;
  min-height: 100vh;
  background: var(--brand-paper);
  color: var(--text-primary);
  font-family: var(--font-sans);
  overflow: hidden;
}

// ---------------------------------------------------------------
// Left brand panel
// ---------------------------------------------------------------
.brand-panel {
  position: relative;
  flex: 0 0 540px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 56px 56px 48px;
  background: #0A0A0A;
  color: $cream;
  overflow: hidden;
}

.brand-motif {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0.08;
  pointer-events: none;
}

.brand-mark {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-mark__box {
  width: 24px;
  height: 24px;
  border: 1px solid $cream;
  display: grid;
  place-items: center;
}

.brand-mark__dot {
  width: 4px;
  height: 4px;
  background: $cream;
}

.brand-mark__name {
  font-family: var(--font-mono);
  font-size: 13px;
  letter-spacing: 1.5px;
  color: $cream;
}

.brand-hero {
  position: relative;
}

.brand-hero__title {
  margin: 0;
  font-family: var(--font-serif);
  font-size: 64px;
  line-height: 1.1;
  font-weight: 500;
  letter-spacing: 2px;
  color: $cream;
}

.brand-hero__sub {
  margin: 28px 0 0;
  max-width: 340px;
  font-size: 13px;
  line-height: 1.7;
  color: $cream-dim;
}

.brand-footer {
  position: relative;
  display: flex;
  justify-content: space-between;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 1px;
  color: $cream-faint;
}

// ---------------------------------------------------------------
// Right form panel
// ---------------------------------------------------------------
.form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
  background: var(--brand-paper);
}

.login-form {
  width: 360px;
  max-width: 100%;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.form-heading__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: var(--text-primary);
}

.form-heading__sub {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

.form-fields {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-body);
  letter-spacing: 0.3px;
}

.field__control {
  display: flex;
  align-items: center;
  height: 36px;
  padding: 0 12px;
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 6px;
  transition: border-color $transition-base, box-shadow $transition-base;

  &:hover {
    border-color: var(--border-medium);
  }

  &:focus-within {
    border-color: var(--brand-ink);
    box-shadow: 0 0 0 2px rgba(10, 10, 10, 0.06);
  }
}

.field__control--with-action {
  padding-right: 4px;
}

.field__input {
  flex: 1;
  height: 100%;
  background: transparent;
  border: none;
  outline: none;
  padding: 0;
  font-size: 13px;
  color: var(--text-primary);
  font-family: inherit;

  &::placeholder {
    color: var(--text-placeholder);
  }
}

.field__toggle {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  border-radius: 4px;
  transition: color $transition-fast, background-color $transition-fast;

  &:hover {
    color: var(--text-primary);
    background: var(--surface-secondary);
  }
}

.form-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: var(--text-tertiary);
}

.remember {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  user-select: none;

  input {
    width: 13px;
    height: 13px;
    accent-color: var(--brand-ink);
    cursor: pointer;
  }
}

.form-row__forgot {
  cursor: pointer;
  transition: color $transition-fast;

  &:hover {
    color: var(--text-primary);
  }
}

.form-error {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0;
  padding: 8px 10px;
  font-size: 12px;
  color: var(--brand-danger);
  background: var(--brand-danger-soft);
  border: 1px solid var(--brand-danger-soft);
  border-radius: 6px;

  svg {
    flex-shrink: 0;
  }
}

.submit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 42px;
  padding: 0 16px;
  background: var(--brand-ink);
  color: $cream;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  letter-spacing: 4px;
  cursor: pointer;
  transition: opacity $transition-base, transform $transition-base;

  &:hover:not(:disabled) {
    opacity: 0.88;
  }

  &:active:not(:disabled) {
    transform: scale(0.99);
  }

  &:focus-visible {
    outline: 2px solid var(--brand-ink);
    outline-offset: 2px;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  &.is-loading .submit-btn__label {
    letter-spacing: 1px;
  }
}

.submit-btn__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(245, 243, 237, 0.3);
  border-top-color: $cream;
  border-radius: 50%;
  animation: login-spin 0.6s linear infinite;
}

@keyframes login-spin {
  to {
    transform: rotate(360deg);
  }
}

.form-hint {
  margin: 0;
  text-align: center;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 1px;
  color: var(--text-tertiary);
}

// ---------------------------------------------------------------
// Responsive — stack vertically below md breakpoint
// ---------------------------------------------------------------
@media (max-width: $breakpoint-md) {
  .login-page {
    flex-direction: column;
    height: auto;
    min-height: 100vh;
  }

  .brand-panel {
    flex: 0 0 auto;
    padding: 40px 32px 32px;
    gap: 32px;
  }

  .brand-hero__title {
    font-size: 44px;
  }

  .brand-hero__sub {
    margin-top: 20px;
  }

  .form-panel {
    flex: 1;
    padding: 40px 24px;
  }
}

@media (max-width: $breakpoint-sm) {
  .brand-panel {
    padding: 32px 24px 24px;
  }

  .brand-hero__title {
    font-size: 36px;
    letter-spacing: 1px;
  }

  .login-form {
    width: 100%;
  }
}
</style>
