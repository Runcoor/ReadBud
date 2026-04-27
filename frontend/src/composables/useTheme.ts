// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { ref, watchEffect, type Ref } from 'vue'

type Theme = 'light' | 'dark'

const STORAGE_KEY = 'readbud_theme'

const theme: Ref<Theme> = ref(getInitialTheme())

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'light'

  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') return stored

  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'

  return 'light'
}

watchEffect(() => {
  if (typeof document === 'undefined') return
  document.documentElement.setAttribute('data-theme', theme.value)
  localStorage.setItem(STORAGE_KEY, theme.value)
})

function toggle(): void {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
}

function isDark(): boolean {
  return theme.value === 'dark'
}

export function useTheme() {
  return { theme, toggle, isDark }
}
