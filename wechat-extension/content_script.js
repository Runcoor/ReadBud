// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// content_script.js — runs on https://mp.weixin.qq.com/cgi-bin/appmsg*
//
// Workflow:
//   1. Detect ?readbud_draft=<id>&readbud_job=<id> on the URL.
//   2. Inject a floating "📰 从 ReadBud 导入" button so the user can also
//      trigger fill manually (the URL params trigger an auto-attempt once).
//   3. On trigger: ask background to fetch the wechat-package, then populate
//      the editor's title / author / digest / source-url / body / cover.
//
// WeChat's editor DOM is undocumented and changes from time to time. The
// selectors below were observed on the 2025-Q1 layout. If WeChat ships a
// breaking change, the failure surfaces as a toast pointing at the missing
// selector — that's the signal to update FILLERS.
//
// The body is dropped into the rich-text iframe via a synthetic `paste` event
// carrying text/html — that's the path WeChat treats as "imported HTML" and
// preserves inline styling. Setting innerHTML directly tends to lose styles.

const RB_LOG_PREFIX = '[ReadBud]'

function rbLog(...args) {
  console.log(RB_LOG_PREFIX, ...args)
}

function rbWarn(...args) {
  console.warn(RB_LOG_PREFIX, ...args)
}

// ----- Toast UI (lightweight, no framework) -----
let toastTimer = null
function toast(title, message, kind = 'ok', durationMs = 5000) {
  const existing = document.querySelector('.readbud-toast')
  if (existing) existing.remove()
  const el = document.createElement('div')
  el.className = 'readbud-toast'
  el.dataset.kind = kind
  el.innerHTML = `<div class="readbud-toast__title"></div><div class="readbud-toast__body"></div>`
  el.querySelector('.readbud-toast__title').textContent = title
  el.querySelector('.readbud-toast__body').textContent = message
  document.body.appendChild(el)
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => el.remove(), durationMs)
}

// ----- Background message helper -----
function ask(message) {
  return new Promise((resolve, reject) => {
    chrome.runtime.sendMessage(message, (resp) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message))
        return
      }
      if (!resp?.ok) {
        reject(new Error(resp?.error || 'unknown error'))
        return
      }
      resolve(resp)
    })
  })
}

// ----- Field fillers -----
//
// Each filler accepts the package payload and returns true on success, false
// if the target element wasn't found. We try multiple selectors in order to
// stay tolerant of small DOM shifts.

function setNativeInputValue(el, value) {
  // Setting `el.value = ...` on a React/Vue-managed input doesn't trigger the
  // framework's change handlers. Calling the native setter then dispatching
  // an input event reliably notifies any framework that's listening.
  const proto = el.tagName === 'TEXTAREA'
    ? HTMLTextAreaElement.prototype
    : HTMLInputElement.prototype
  const desc = Object.getOwnPropertyDescriptor(proto, 'value')
  desc?.set?.call(el, value)
  el.dispatchEvent(new Event('input', { bubbles: true }))
  el.dispatchEvent(new Event('change', { bubbles: true }))
}

function findFirst(selectors) {
  for (const sel of selectors) {
    const el = document.querySelector(sel)
    if (el) return el
  }
  return null
}

function fillTitle(pkg) {
  const el = findFirst([
    '#title',
    'input#title',
    '[name="title"]',
    'input.js_title',
  ])
  if (!el) return false
  setNativeInputValue(el, pkg.title || '')
  return true
}

function fillAuthor(pkg) {
  const el = findFirst([
    '#author',
    'input#author',
    '[name="author"]',
  ])
  if (!el || !pkg.author) return false
  setNativeInputValue(el, pkg.author)
  return true
}

function fillDigest(pkg) {
  const el = findFirst([
    '#js_description',
    'textarea#js_description',
    '[name="digest"]',
    'textarea.js_desc',
  ])
  if (!el || !pkg.digest) return false
  setNativeInputValue(el, pkg.digest)
  return true
}

function fillSourceURL(pkg) {
  if (!pkg.source_url) return false
  const el = findFirst([
    '#js_url',
    'input#js_url',
    '[name="content_source_url"]',
    'input.js_link',
  ])
  if (!el) return false
  setNativeInputValue(el, pkg.source_url)
  return true
}

// Locate the rich-text editor iframe. WeChat uses an iframe for the body to
// isolate styles. We attempt to access its document — same-origin with
// mp.weixin.qq.com so this should work.
function findEditorIframe() {
  const iframes = document.querySelectorAll('iframe')
  for (const f of iframes) {
    try {
      const doc = f.contentDocument
      if (!doc) continue
      const body = doc.body
      // Heuristic: the editor body has contenteditable="true" or class .editor
      if (body && (body.isContentEditable || body.querySelector('[contenteditable="true"]'))) {
        return f
      }
    } catch {
      // cross-origin frame; skip
    }
  }
  return null
}

function fillBodyHTML(pkg) {
  if (!pkg.content_html) return false
  const iframe = findEditorIframe()
  if (!iframe) {
    rbWarn('editor iframe not found')
    return false
  }
  const doc = iframe.contentDocument
  const editable = doc.body.isContentEditable
    ? doc.body
    : doc.querySelector('[contenteditable="true"]')
  if (!editable) {
    rbWarn('editable region not found')
    return false
  }
  // Focus + clear + paste-event with text/html. Going through ClipboardEvent
  // makes WeChat process the import the same way as if the user pasted into
  // the editor — that path preserves inline styles best.
  editable.focus()
  // Replace existing content. innerHTML is safe here because we control the
  // payload (rendered by our own backend) and it's a fresh editor anyway.
  editable.innerHTML = ''

  const dt = new DataTransfer()
  dt.setData('text/html', pkg.content_html)
  dt.setData('text/plain', pkg.content_html.replace(/<[^>]+>/g, ''))
  const evt = new ClipboardEvent('paste', {
    bubbles: true,
    cancelable: true,
    clipboardData: dt,
  })
  // Fallback if the paste event doesn't get processed (some Chromium builds
  // strip clipboardData from synthesized events): fall back to innerHTML.
  const handled = editable.dispatchEvent(evt)
  if (!handled || !editable.innerHTML) {
    editable.innerHTML = pkg.content_html
  }
  // Notify any change listeners.
  editable.dispatchEvent(new Event('input', { bubbles: true }))
  return true
}

// Cover image — WeChat exposes a hidden <input type="file"> behind the
// "封面图片" button. We construct a File from the base64 payload, drop it onto
// the input, and dispatch change. This mirrors what happens when the user
// picks a file in the native file dialog.
function findCoverFileInput() {
  // Multiple candidates: WeChat has both cover and inline-image upload inputs;
  // we want the one inside the cover-area component. The cover input usually
  // sits in #js_cover_area or has an aria/title hinting "封面".
  const inputs = document.querySelectorAll('input[type="file"][accept*="image"]')
  for (const el of inputs) {
    if (!el.offsetParent && getComputedStyle(el).display === 'none') {
      // some are hidden — that's expected for cover input
    }
    const ctx = el.closest('[id*="cover"], [class*="cover"], [class*="thumb"]')
    if (ctx) return el
  }
  // Fallback: first hidden image file input on the page.
  return inputs[0] || null
}

function base64ToFile(base64, mime, filename) {
  const binary = atob(base64)
  const len = binary.length
  const bytes = new Uint8Array(len)
  for (let i = 0; i < len; i++) bytes[i] = binary.charCodeAt(i)
  const blob = new Blob([bytes], { type: mime || 'image/png' })
  return new File([blob], filename || 'cover.png', { type: mime || 'image/png' })
}

async function fillCover(pkg) {
  if (!pkg.cover_base64 || !pkg.cover_mime_type) return false
  const input = findCoverFileInput()
  if (!input) return false
  const file = base64ToFile(pkg.cover_base64, pkg.cover_mime_type, pkg.cover_filename)
  const dt = new DataTransfer()
  dt.items.add(file)
  Object.defineProperty(input, 'files', {
    value: dt.files,
    writable: false,
    configurable: true,
  })
  input.dispatchEvent(new Event('input', { bubbles: true }))
  input.dispatchEvent(new Event('change', { bubbles: true }))
  return true
}

// ----- Orchestration -----
async function runFill({ draftId, jobId }) {
  if (!draftId) {
    toast('未指定草稿', '请通过 ReadBud 的「通过插件发布」按钮跳转到此页面。', 'err')
    return
  }

  const fab = document.querySelector('.readbud-fab')
  if (fab) {
    fab.disabled = true
    fab.textContent = '加载中…'
  }

  toast('正在拉取数据', '从 ReadBud 获取标题、正文与封面…', 'ok', 3000)

  let pkg
  try {
    const resp = await ask({ type: 'fetch-package', draftId })
    pkg = resp.data
  } catch (e) {
    toast('拉取失败', e.message, 'err', 8000)
    if (fab) { fab.disabled = false; fab.textContent = '📰 从 ReadBud 导入' }
    return
  }

  const results = {
    title:     fillTitle(pkg),
    author:    fillAuthor(pkg),
    digest:    fillDigest(pkg),
    sourceURL: fillSourceURL(pkg),
    body:      fillBodyHTML(pkg),
    cover:     await fillCover(pkg),
  }
  rbLog('fill results', results)

  const failed = Object.entries(results)
    .filter(([k, ok]) => !ok && (k !== 'cover' || pkg.cover_base64))
    .map(([k]) => k)

  if (failed.length === 0) {
    toast('填充完成 ✓', '请在 WeChat 编辑器中检查后点「群发」。完成后回 ReadBud 标记已发布。', 'ok', 8000)
  } else {
    toast(
      '部分填充未成功',
      `失败字段: ${failed.join(', ')}。可能 WeChat DOM 已变更, 请手动复制对应内容。`,
      'err',
      10000,
    )
  }

  if (fab) { fab.disabled = false; fab.textContent = '📰 重新导入' }

  // Stash jobId so the user can later mark it fulfilled from this tab too,
  // if we add a "已完成发布" button. For V1 they go back to ReadBud.
  if (jobId) {
    sessionStorage.setItem('readbud:job', jobId)
  }
}

// ----- Init -----
function injectFloatingButton(initialParams) {
  if (document.querySelector('.readbud-fab')) return
  const btn = document.createElement('button')
  btn.className = 'readbud-fab'
  btn.innerHTML = `<span class="readbud-fab__icon">📰</span> 从 ReadBud 导入`
  btn.addEventListener('click', () => {
    const params = readParamsFromURL() || initialParams
    runFill(params || {})
  })
  document.body.appendChild(btn)
}

function readParamsFromURL() {
  try {
    const u = new URL(location.href)
    const draftId = u.searchParams.get('readbud_draft')
    const jobId   = u.searchParams.get('readbud_job')
    if (!draftId) return null
    return { draftId, jobId }
  } catch {
    return null
  }
}

function init() {
  rbLog('content script ready')
  const params = readParamsFromURL()
  injectFloatingButton(params)
  if (params?.draftId) {
    // Auto-run once on first load. Wait a beat for WeChat's React app to
    // mount its inputs (otherwise our setNativeInputValue lands on nothing).
    setTimeout(() => runFill(params), 1500)
  }
}

if (document.readyState === 'complete' || document.readyState === 'interactive') {
  init()
} else {
  document.addEventListener('DOMContentLoaded', init)
}
