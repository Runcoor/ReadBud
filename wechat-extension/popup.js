// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// popup.js — handles the extension popup form.
//
// Reads/writes the API base + plugin token via the background service worker
// (which owns chrome.storage). Keeping that responsibility in one place means
// the content script can also pick up config changes without coordination.

const $apiBase = document.getElementById('apiBase')
const $token   = document.getElementById('token')
const $save    = document.getElementById('save')
const $test    = document.getElementById('test')
const $status  = document.getElementById('status')

function showStatus(message, kind) {
  $status.hidden = false
  $status.dataset.kind = kind
  $status.textContent = message
}

function clearStatus() {
  $status.hidden = true
  $status.textContent = ''
}

function send(message) {
  return new Promise((resolve, reject) => {
    chrome.runtime.sendMessage(message, (resp) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message))
        return
      }
      if (!resp?.ok) {
        reject(new Error(resp?.error || '操作失败'))
        return
      }
      resolve(resp)
    })
  })
}

async function loadConfig() {
  try {
    const { config } = await send({ type: 'get-config' })
    $apiBase.value = config?.apiBase || ''
    $token.value   = config?.token || ''
  } catch (e) {
    showStatus(e.message, 'err')
  }
}

$save.addEventListener('click', async () => {
  clearStatus()
  try {
    await send({
      type: 'set-config',
      config: { apiBase: $apiBase.value, token: $token.value },
    })
    showStatus('已保存。打开 mp.weixin.qq.com 编辑器即可使用。', 'ok')
  } catch (e) {
    showStatus(e.message, 'err')
  }
})

$test.addEventListener('click', async () => {
  clearStatus()
  // Save first so the test uses the latest input.
  try {
    await send({
      type: 'set-config',
      config: { apiBase: $apiBase.value, token: $token.value },
    })
  } catch (e) {
    showStatus(`保存失败: ${e.message}`, 'err')
    return
  }
  // We don't have a dedicated /ping endpoint yet — fetch a known-bad draft id
  // and surface the resulting error. A 401/403 means the URL/token are wrong;
  // a 404 means we got through auth, which is the "good" outcome here.
  try {
    await send({ type: 'fetch-package', draftId: '__plugin_self_test__' })
    showStatus('意外：自检 draft id 居然存在 ?', 'err')
  } catch (e) {
    const msg = e.message || ''
    if (/401|403|令牌|token/i.test(msg)) {
      showStatus(`认证失败: ${msg}`, 'err')
    } else if (/404|不存在/i.test(msg)) {
      showStatus('连接成功 ✓ (令牌有效)', 'ok')
    } else {
      showStatus(msg, 'err')
    }
  }
})

loadConfig()
