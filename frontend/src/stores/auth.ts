// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi } from '@/api/auth'
import type { LoginRequest } from '@/types/api'

export interface UserInfo {
  id: string
  username: string
  nickname: string
  role: string
}

const TOKEN_KEY = 'readbud_token'
const USER_KEY = 'readbud_user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<UserInfo | null>(restoreUser())

  const isAuthenticated = computed(() => !!token.value)

  function restoreUser(): UserInfo | null {
    try {
      const raw = localStorage.getItem(USER_KEY)
      return raw ? (JSON.parse(raw) as UserInfo) : null
    } catch {
      return null
    }
  }

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
  }

  function setUser(userInfo: UserInfo) {
    user.value = userInfo
    localStorage.setItem(USER_KEY, JSON.stringify(userInfo))
  }

  function clearAuth() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  async function login(req: LoginRequest): Promise<void> {
    const resp = await loginApi(req)
    if (resp.code === 0 && resp.data) {
      setToken(resp.data.token)
      setUser(resp.data.user)
    } else {
      throw new Error(resp.message || '登录失败')
    }
  }

  function logout() {
    clearAuth()
  }

  return {
    token,
    user,
    isAuthenticated,
    setToken,
    setUser,
    clearAuth,
    login,
    logout,
  }
})
