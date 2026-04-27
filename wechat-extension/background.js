// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// background.js — ReadBud Extension Service Worker
//
// Responsibilities:
//   - Holds the user's API base URL + plugin token in chrome.storage.sync
//     (sync so the same config follows them across browsers signed into the
//     same Google account).
//   - Performs all cross-origin fetches against the ReadBud API on behalf of
//     content scripts. Doing the fetch here (rather than in the content script
//     running on mp.weixin.qq.com) avoids CORS friction — the manifest's
//     host_permissions whitelist gives the service worker direct access.
//   - Routes messages from popup + content script.
//
// Message protocol (JSON):
//   { type: "get-config" }                            → { ok, config }
//   { type: "set-config", config: { apiBase, token }} → { ok }
//   { type: "fetch-package", draftId }                → { ok, data }
//   { type: "mark-fulfilled", jobId, articleUrl }     → { ok, data }

const STORAGE_KEY = 'readbud:config'

const DEFAULT_CONFIG = {
  apiBase: 'http://localhost:19881/api/v1',
  token: '',
}

async function getConfig() {
  const result = await chrome.storage.sync.get(STORAGE_KEY)
  return { ...DEFAULT_CONFIG, ...(result[STORAGE_KEY] || {}) }
}

async function setConfig(config) {
  // Defensive: trim and reject obvious bad input.
  const clean = {
    apiBase: (config.apiBase || '').trim().replace(/\/+$/, ''),
    token: (config.token || '').trim(),
  }
  await chrome.storage.sync.set({ [STORAGE_KEY]: clean })
  return clean
}

async function apiFetch(path, init = {}) {
  const config = await getConfig()
  if (!config.apiBase) throw new Error('未配置 ReadBud API 地址')
  if (!config.token) throw new Error('未配置插件令牌')

  const url = `${config.apiBase}${path}`
  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${config.token}`,
    ...(init.headers || {}),
  }
  const resp = await fetch(url, { ...init, headers })
  const text = await resp.text()
  let json
  try { json = text ? JSON.parse(text) : {} } catch { json = { raw: text } }
  if (!resp.ok) {
    const msg = json?.message || `HTTP ${resp.status}`
    throw new Error(msg)
  }
  if (json.code !== undefined && json.code !== 0) {
    throw new Error(json.message || `API error code ${json.code}`)
  }
  return json.data ?? json
}

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  // Async handlers must return true synchronously for sendResponse to work.
  ;(async () => {
    try {
      switch (msg?.type) {
        case 'get-config': {
          const config = await getConfig()
          sendResponse({ ok: true, config })
          return
        }
        case 'set-config': {
          const config = await setConfig(msg.config || {})
          sendResponse({ ok: true, config })
          return
        }
        case 'fetch-package': {
          if (!msg.draftId) throw new Error('missing draftId')
          const data = await apiFetch(`/drafts/${encodeURIComponent(msg.draftId)}/wechat-package`)
          sendResponse({ ok: true, data })
          return
        }
        case 'mark-fulfilled': {
          if (!msg.jobId) throw new Error('missing jobId')
          const body = msg.articleUrl ? { article_url: msg.articleUrl } : {}
          const data = await apiFetch(`/publish/jobs/${encodeURIComponent(msg.jobId)}/fulfilled`, {
            method: 'POST',
            body: JSON.stringify(body),
          })
          sendResponse({ ok: true, data })
          return
        }
        default:
          sendResponse({ ok: false, error: `unknown message type ${msg?.type}` })
      }
    } catch (e) {
      sendResponse({ ok: false, error: String(e?.message || e) })
    }
  })()
  return true
})

// Mark "installed" by setting a sentinel cookie-like value the webapp can read
// via chrome.runtime.id (only available when the webapp injects a content
// script). For now keep this lightweight and rely on the popup for status.
chrome.runtime.onInstalled.addListener(() => {
  chrome.storage.sync.get(STORAGE_KEY).then((cur) => {
    if (!cur[STORAGE_KEY]) {
      chrome.storage.sync.set({ [STORAGE_KEY]: DEFAULT_CONFIG })
    }
  })
})
