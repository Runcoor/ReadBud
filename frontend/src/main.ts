// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import './styles/main.scss'

// Stale-chunk recovery: after a deploy the old index.html in the user's tab
// references chunk hashes that no longer exist. Vite emits `vite:preloadError`
// and vue-router surfaces "Failed to fetch dynamically imported module" — in
// both cases reload once to pick up the fresh index.html. Sentinel guards
// against infinite reload if the network really is broken.
const RELOAD_FLAG = 'readbud_chunk_reload'
function recoverFromStaleChunk(reason: unknown): void {
  if (sessionStorage.getItem(RELOAD_FLAG)) return
  sessionStorage.setItem(RELOAD_FLAG, '1')
  console.warn('[stale-chunk] reloading to recover', reason)
  window.location.reload()
}
window.addEventListener('vite:preloadError', (e) => recoverFromStaleChunk(e))
window.addEventListener('load', () => sessionStorage.removeItem(RELOAD_FLAG))
router.onError((err) => {
  const msg = err instanceof Error ? err.message : String(err)
  if (/Failed to fetch dynamically imported module|Importing a module script failed/i.test(msg)) {
    recoverFromStaleChunk(err)
  }
})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')
