import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface UserInfo {
  id: string
  username: string
  nickname: string
  role: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('readbud_token'))
  const user = ref<UserInfo | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('readbud_token', newToken)
  }

  function clearAuth() {
    token.value = null
    user.value = null
    localStorage.removeItem('readbud_token')
  }

  function setUser(userInfo: UserInfo) {
    user.value = userInfo
  }

  return {
    token,
    user,
    isAuthenticated,
    setToken,
    clearAuth,
    setUser,
  }
})
