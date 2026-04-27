<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="brand-screen">
    <AppTopBar crumb="品牌风格管理">
      <template #right>
        <a class="brand-screen__back mono" @click="router.push({ name: 'Workbench' })">
          ← 返回工作台
        </a>
      </template>
    </AppTopBar>

    <main class="brand-screen__main">
      <div class="brand-hero">
        <div class="brand-hero__left">
          <div class="brand-hero__title-row">
            <h1 class="brand-hero__title">品牌风格管理</h1>
            <MonoChip kind="default">BRANDS · {{ profiles.length }} PROFILES</MonoChip>
          </div>
          <p class="brand-hero__sub">
            配置品牌语气、禁用词、偏好词与 CTA 规则，让每篇文章都像品牌自己在说话。
          </p>
        </div>
        <button class="brand-hero__cta" @click="startCreate">+ 添加品牌配置</button>
      </div>

      <div v-if="loading" class="brand-state">
        <el-skeleton :rows="3" animated />
      </div>

      <div v-else-if="error" class="brand-state brand-state--error">
        <el-alert type="error" :title="error" :closable="false" show-icon />
        <button class="link-btn" @click="loadProfiles">重试</button>
      </div>

      <div v-else-if="profiles.length === 0" class="brand-empty">
        <span class="brand-empty__label mono">EMPTY · 还没有品牌配置</span>
        <button class="brand-empty__cta mono" @click="startCreate">[ 创建第一个 → ]</button>
      </div>

      <div v-else class="brand-list">
        <div
          v-for="profile in profiles"
          :key="profile.id"
          class="brand-card"
          :class="{ 'brand-card--active': isDrawerOpen && editingId === profile.public_id }"
        >
          <div class="brand-card__col-name">
            <div class="brand-card__name-row">
              <span class="brand-card__name">{{ profile.name }}</span>
              <MonoChip v-if="isDefault(profile)" kind="default">DEFAULT</MonoChip>
            </div>
            <span class="brand-card__stat mono">
              {{ (profile.forbidden_words?.length || 0) }} 禁用词 ·
              {{ (profile.preferred_words?.length || 0) }} 偏好词
            </span>
          </div>

          <div class="brand-card__col-tone">
            <span class="brand-card__tone-label">品牌语气</span>
            <span class="brand-card__tone-text">
              {{ profile.brand_tone || '尚未填写' }}
            </span>
          </div>

          <div class="brand-card__col-actions">
            <span
              v-if="isDrawerOpen && editingId === profile.public_id"
              class="brand-card__active-tag"
            >
              <StatusDot kind="sprout" :size="6" />
              <span>使用中</span>
            </span>
            <a class="brand-card__link" @click="setAsDefault(profile)">设为默认</a>
            <a class="brand-card__link" @click="startEdit(profile)">编辑</a>
            <a class="brand-card__link brand-card__link--danger" @click="handleDelete(profile)">
              删除
            </a>
          </div>
        </div>
      </div>

      <p v-if="profiles.length > 0" class="brand-footer-note mono">
        品牌配置会作用于关键词扩展 / 文案撰写 / 自检校对 三个阶段
      </p>
    </main>

    <!-- Right-side drawer -->
    <teleport to="body">
      <transition name="brand-fade">
        <div
          v-if="isDrawerOpen"
          class="brand-drawer__backdrop"
          @click.self="cancelEdit"
        />
      </transition>
      <transition name="brand-slide">
        <aside v-if="isDrawerOpen" class="brand-drawer" role="dialog" aria-modal="true">
          <header class="brand-drawer__header">
            <div class="brand-drawer__title-block">
              <span class="brand-drawer__eyebrow mono">EDIT BRAND</span>
              <h2 class="brand-drawer__title">
                {{ editingId ? '编辑品牌配置' : '新增品牌配置' }}
              </h2>
            </div>
            <div class="brand-drawer__head-actions">
              <span v-if="isDirty" class="brand-drawer__dirty mono">未保存</span>
              <button class="brand-drawer__close" @click="cancelEdit" aria-label="关闭">×</button>
            </div>
          </header>

          <div class="brand-drawer__body">
            <!-- Name -->
            <div class="field">
              <label class="field__label">
                名称 <span class="field__required">*</span>
              </label>
              <input
                v-model="form.name"
                class="field__input"
                placeholder="品牌配置名称，例如：默认品牌"
              />
            </div>

            <!-- Tone -->
            <div class="field">
              <label class="field__label">品牌语气</label>
              <div class="textarea-wrap">
                <textarea
                  v-model="form.brand_tone"
                  class="field__textarea"
                  placeholder="描述品牌的沟通风格，例如：专业严谨、温和亲切、数据驱动…"
                  maxlength="200"
                />
                <span class="textarea-counter mono">
                  {{ (form.brand_tone || '').length }} / 200
                </span>
              </div>
            </div>

            <!-- Forbidden -->
            <div class="field">
              <div class="field__head">
                <div class="field__head-left">
                  <label class="field__label">禁用词</label>
                  <MonoChip kind="danger">{{ forbiddenWords.length }}</MonoChip>
                </div>
                <span class="field__hint">生成时如出现，将自动改写</span>
              </div>
              <div class="tag-wall">
                <span
                  v-for="(word, idx) in forbiddenWords"
                  :key="`f-${idx}-${word}`"
                  class="tag-pill tag-pill--danger"
                >
                  {{ word }}
                  <span class="tag-pill__close" @click="removeForbidden(idx)">×</span>
                </span>
                <span v-if="forbiddenWords.length === 0" class="tag-wall__empty mono">
                  还没有禁用词
                </span>
              </div>
              <div class="tag-input-row">
                <input
                  v-model="newForbiddenWord"
                  class="field__input field__input--inline"
                  placeholder="输入禁用词后回车"
                  @keyup.enter="addForbidden"
                />
                <button class="hairline-btn" @click="addForbidden">添加</button>
              </div>
            </div>

            <!-- Preferred -->
            <div class="field">
              <div class="field__head">
                <div class="field__head-left">
                  <label class="field__label">偏好词</label>
                  <MonoChip kind="sprout">{{ preferredWords.length }}</MonoChip>
                </div>
                <span class="field__hint">生成时优先采用</span>
              </div>
              <div class="tag-wall">
                <span
                  v-for="(word, idx) in preferredWords"
                  :key="`p-${idx}-${word}`"
                  class="tag-pill tag-pill--sprout"
                >
                  {{ word }}
                  <span class="tag-pill__close" @click="removePreferred(idx)">×</span>
                </span>
                <span v-if="preferredWords.length === 0" class="tag-wall__empty mono">
                  还没有偏好词
                </span>
              </div>
              <div class="tag-input-row">
                <input
                  v-model="newPreferredWord"
                  class="field__input field__input--inline"
                  placeholder="输入偏好词后回车"
                  @keyup.enter="addPreferred"
                />
                <button class="hairline-btn" @click="addPreferred">添加</button>
              </div>
            </div>

            <!-- CTA JSON -->
            <div class="field">
              <div class="field__head">
                <div class="field__head-left">
                  <label class="field__label">CTA 规则</label>
                  <span class="field__sub mono">(JSON)</span>
                </div>
                <button class="field__hint-btn mono" @click="fillCtaPreset">
                  ↻ 用预设填充
                </button>
              </div>
              <div v-if="ctaParseError" class="cta-warn mono">JSON 格式错误</div>
              <pre
                class="cta-code"
                contenteditable="true"
                spellcheck="false"
                @input="onCtaInput"
                @blur="onCtaBlur"
                v-text="ctaDisplay"
              />
            </div>
          </div>

          <footer class="brand-drawer__footer">
            <span class="brand-drawer__hint mono">⌘ + S 保存 · Esc 关闭</span>
            <div class="brand-drawer__footer-actions">
              <button class="hairline-btn" @click="cancelEdit">取消</button>
              <button
                class="ink-btn"
                :disabled="saving"
                @click="handleSave"
              >
                {{ saving ? '保存中…' : '保存' }}
              </button>
            </div>
            <span class="brand-drawer__note mono">
              封面样式 / 配图样式可在 API 中配置
            </span>
          </footer>
        </aside>
      </transition>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listBrandProfiles,
  createBrandProfile,
  updateBrandProfile,
  deleteBrandProfile,
} from '@/api/brand'
import type { BrandProfileVO } from '@/types/brand'
import AppTopBar from '@/components/common/AppTopBar.vue'
import MonoChip from '@/components/common/MonoChip.vue'
import StatusDot from '@/components/common/StatusDot.vue'

const router = useRouter()

// === Preserved state ===
const loading = ref(false)
const error = ref<string | null>(null)
const profiles = ref<BrandProfileVO[]>([])

const isDrawerOpen = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const form = reactive({
  name: '',
  brand_tone: '',
})
const forbiddenWords = ref<string[]>([])
const preferredWords = ref<string[]>([])
const newForbiddenWord = ref('')
const newPreferredWord = ref('')
const ctaRulesStr = ref('')
const coverStyleStr = ref('')
const imageStyleStr = ref('')

// Snapshot for dirty detection
interface FormSnapshot {
  name: string
  brand_tone: string
  forbiddenWords: string[]
  preferredWords: string[]
  ctaRulesStr: string
  coverStyleStr: string
  imageStyleStr: string
}
const snapshot = ref<FormSnapshot | null>(null)

const isDirty = computed(() => {
  if (!snapshot.value) return false
  const s = snapshot.value
  if (form.name !== s.name) return true
  if (form.brand_tone !== s.brand_tone) return true
  if (forbiddenWords.value.join('') !== s.forbiddenWords.join('')) return true
  if (preferredWords.value.join('') !== s.preferredWords.join('')) return true
  if (ctaRulesStr.value !== s.ctaRulesStr) return true
  if (coverStyleStr.value !== s.coverStyleStr) return true
  if (imageStyleStr.value !== s.imageStyleStr) return true
  return false
})

function makeSnapshot(): FormSnapshot {
  return {
    name: form.name,
    brand_tone: form.brand_tone,
    forbiddenWords: [...forbiddenWords.value],
    preferredWords: [...preferredWords.value],
    ctaRulesStr: ctaRulesStr.value,
    coverStyleStr: coverStyleStr.value,
    imageStyleStr: imageStyleStr.value,
  }
}

// === CTA code-block display ===
const ctaParseError = ref(false)

const ctaDisplay = computed(() => {
  const raw = ctaRulesStr.value.trim()
  if (!raw) return '{}'
  try {
    const obj = JSON.parse(raw)
    ctaParseError.value = false
    return JSON.stringify(obj, null, 2)
  } catch {
    ctaParseError.value = true
    return raw
  }
})

function onCtaInput(e: Event) {
  const el = e.target as HTMLElement
  ctaRulesStr.value = el.innerText
  // re-validate
  try {
    if (ctaRulesStr.value.trim()) JSON.parse(ctaRulesStr.value)
    ctaParseError.value = false
  } catch {
    ctaParseError.value = true
  }
}

function onCtaBlur(e: Event) {
  const el = e.target as HTMLElement
  const raw = el.innerText.trim()
  ctaRulesStr.value = raw
  if (!raw) {
    ctaParseError.value = false
    return
  }
  try {
    const obj = JSON.parse(raw)
    ctaRulesStr.value = JSON.stringify(obj, null, 2)
    ctaParseError.value = false
  } catch {
    ctaParseError.value = true
  }
}

const CTA_PRESET = `{
  "style": "具体可执行的行动建议或引发思考的好问题，不要空洞的关注我们",
  "avoid": ["点赞 + 在看", "转发支持一下"]
}`

function fillCtaPreset() {
  ctaRulesStr.value = CTA_PRESET
  ctaParseError.value = false
}

// === Tag actions ===
function addForbidden() {
  const word = newForbiddenWord.value.trim()
  if (word && !forbiddenWords.value.includes(word)) {
    forbiddenWords.value.push(word)
  }
  newForbiddenWord.value = ''
}

function removeForbidden(idx: number) {
  forbiddenWords.value.splice(idx, 1)
}

function addPreferred() {
  const word = newPreferredWord.value.trim()
  if (word && !preferredWords.value.includes(word)) {
    preferredWords.value.push(word)
  }
  newPreferredWord.value = ''
}

function removePreferred(idx: number) {
  preferredWords.value.splice(idx, 1)
}

// === Default detection (best-effort: look at unknown shape on the VO) ===
function isDefault(p: BrandProfileVO): boolean {
  const anyP = p as unknown as Record<string, unknown>
  return Boolean(anyP.is_default || anyP.isDefault)
}

function setAsDefault(_profile: BrandProfileVO) {
  // No dedicated API yet — placeholder
  ElMessage.info('即将上线')
}

// === Loaders ===
function safeParseJSON(str: string): Record<string, unknown> | undefined {
  if (!str.trim()) return undefined
  try {
    return JSON.parse(str) as Record<string, unknown>
  } catch {
    ElMessage.warning('JSON 格式不正确，请检查')
    return undefined
  }
}

async function loadProfiles() {
  loading.value = true
  error.value = null
  try {
    const resp = await listBrandProfiles()
    if (resp.code === 0) {
      profiles.value = resp.data || []
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载品牌配置失败'
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.name = ''
  form.brand_tone = ''
  forbiddenWords.value = []
  preferredWords.value = []
  newForbiddenWord.value = ''
  newPreferredWord.value = ''
  ctaRulesStr.value = ''
  coverStyleStr.value = ''
  imageStyleStr.value = ''
  ctaParseError.value = false
}

function startCreate() {
  editingId.value = null
  resetForm()
  snapshot.value = makeSnapshot()
  isDrawerOpen.value = true
}

function startEdit(profile: BrandProfileVO) {
  editingId.value = profile.public_id
  form.name = profile.name
  form.brand_tone = profile.brand_tone || ''
  forbiddenWords.value = [...(profile.forbidden_words || [])]
  preferredWords.value = [...(profile.preferred_words || [])]
  ctaRulesStr.value = profile.cta_rules
    ? JSON.stringify(profile.cta_rules, null, 2)
    : ''
  coverStyleStr.value = profile.cover_style_rules
    ? JSON.stringify(profile.cover_style_rules, null, 2)
    : ''
  imageStyleStr.value = profile.image_style_rules
    ? JSON.stringify(profile.image_style_rules, null, 2)
    : ''
  ctaParseError.value = false
  newForbiddenWord.value = ''
  newPreferredWord.value = ''
  snapshot.value = makeSnapshot()
  isDrawerOpen.value = true
}

function cancelEdit() {
  isDrawerOpen.value = false
  editingId.value = null
  snapshot.value = null
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入名称')
    return
  }

  const ctaRules = ctaRulesStr.value ? safeParseJSON(ctaRulesStr.value) : undefined
  const coverRules = coverStyleStr.value ? safeParseJSON(coverStyleStr.value) : undefined
  const imageRules = imageStyleStr.value ? safeParseJSON(imageStyleStr.value) : undefined

  if (
    (ctaRulesStr.value && ctaRules === undefined) ||
    (coverStyleStr.value && coverRules === undefined) ||
    (imageStyleStr.value && imageRules === undefined)
  ) {
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.name,
      brand_tone: form.brand_tone,
      forbidden_words: forbiddenWords.value,
      preferred_words: preferredWords.value,
      cta_rules: ctaRules,
      cover_style_rules: coverRules,
      image_style_rules: imageRules,
    }

    if (editingId.value) {
      // The API takes the public_id string
      await updateBrandProfile(editingId.value, payload)
      ElMessage.success('品牌配置已更新')
    } else {
      await createBrandProfile(payload)
      ElMessage.success('品牌配置已创建')
    }
    isDrawerOpen.value = false
    editingId.value = null
    snapshot.value = null
    await loadProfiles()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(profile: BrandProfileVO) {
  try {
    await ElMessageBox.confirm(
      `确定要删除品牌配置「${profile.name}」吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' },
    )
    await deleteBrandProfile(profile.public_id)
    ElMessage.success('已删除')
    if (editingId.value === profile.public_id) {
      cancelEdit()
    }
    await loadProfiles()
  } catch (e: unknown) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error('删除失败')
    }
  }
}

// === Keyboard shortcuts ===
function onKeydown(e: KeyboardEvent) {
  if (!isDrawerOpen.value) return
  if (e.key === 'Escape') {
    e.preventDefault()
    cancelEdit()
  } else if ((e.metaKey || e.ctrlKey) && (e.key === 's' || e.key === 'S')) {
    e.preventDefault()
    handleSave()
  }
}

onMounted(() => {
  loadProfiles()
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})

// Lock body scroll while drawer open
watch(isDrawerOpen, (open) => {
  if (typeof document === 'undefined') return
  document.body.style.overflow = open ? 'hidden' : ''
})
</script>

<style scoped lang="scss">
@use '@/styles/tokens' as *;

.brand-screen {
  min-height: 100vh;
  background: var(--brand-paper);
  display: flex;
  flex-direction: column;
  font-family: var(--font-sans);
  color: var(--text-primary);

  &__back {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-tertiary);
    cursor: pointer;
    letter-spacing: 0.04em;
    transition: color 120ms ease;

    &:hover { color: var(--text-primary); }
  }

  &__main {
    flex: 1;
    padding: 32px 48px 48px;
    max-width: 1280px;
    width: 100%;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
}

// === Hero ===
.brand-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 8px;

  &__left {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  &__title-row {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  &__title {
    font-family: var(--font-serif);
    font-size: 28px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    line-height: 1.2;
  }

  &__sub {
    font-size: 13px;
    color: var(--text-body);
    margin: 0;
    max-width: 640px;
    line-height: 1.55;
  }

  &__cta {
    height: 36px;
    padding: 0 18px;
    background: var(--brand-ink);
    color: var(--text-inverse);
    border: 1px solid var(--brand-ink);
    font-family: var(--font-sans);
    font-size: 13px;
    letter-spacing: 1px;
    cursor: pointer;
    border-radius: 0;
    transition: opacity 120ms ease;

    &:hover { opacity: 0.85; }
    &:active { transform: scale(0.98); }
  }
}

// === States ===
.brand-state {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  padding: 24px;

  &--error {
    display: flex;
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}

.link-btn {
  background: transparent;
  border: 1px solid var(--border-hair);
  font-family: var(--font-sans);
  font-size: 12px;
  padding: 6px 12px;
  cursor: pointer;
  color: var(--text-primary);
  border-radius: 4px;

  &:hover { border-color: var(--brand-ink); }
}

.brand-empty {
  background: var(--surface-card);
  border: 1px dashed var(--border-hair);
  border-radius: 8px;
  padding: 36px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;

  &__label {
    font-size: 11px;
    color: var(--text-tertiary);
    letter-spacing: 0.08em;
  }

  &__cta {
    background: transparent;
    border: none;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--brand-ink);
    cursor: pointer;
    letter-spacing: 0.04em;

    &:hover { text-decoration: underline; }
  }
}

// === Cards ===
.brand-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.brand-card {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  padding: 20px 24px;
  display: grid;
  grid-template-columns: 220px 1fr 200px;
  gap: 24px;
  align-items: center;
  transition: border-color 150ms ease;

  &--active {
    border-color: var(--brand-ink);
  }

  &__col-name {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }

  &__name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__name {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }

  &__stat {
    font-size: 11px;
    color: var(--text-tertiary);
    letter-spacing: 0.04em;
  }

  &__col-tone {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  &__tone-label {
    font-size: 11px;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  &__tone-text {
    font-size: 13px;
    color: var(--text-body);
    line-height: 1.55;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  &__col-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 14px;
    flex-wrap: wrap;
  }

  &__active-tag {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--brand-sprout);
  }

  &__link {
    font-size: 13px;
    color: var(--text-body);
    cursor: pointer;
    transition: color 120ms ease;

    &:hover { color: var(--brand-ink); }

    &--danger {
      color: var(--brand-danger);
      opacity: 0.7;

      &:hover { opacity: 1; }
    }
  }
}

.brand-footer-note {
  text-align: center;
  font-size: 11px;
  color: var(--text-tertiary);
  letter-spacing: 0.04em;
  margin: 16px 0 0;
}

// === Drawer ===
.brand-fade-enter-active,
.brand-fade-leave-active {
  transition: opacity 200ms ease;
}
.brand-fade-enter-from,
.brand-fade-leave-to { opacity: 0; }

.brand-slide-enter-active,
.brand-slide-leave-active {
  transition: transform 250ms ease;
}
.brand-slide-enter-from,
.brand-slide-leave-to { transform: translateX(100%); }

.brand-drawer__backdrop {
  position: fixed;
  inset: 0;
  background: rgba(10, 10, 10, 0.32);
  z-index: 600;
}

.brand-drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 560px;
  max-width: 100vw;
  background: var(--surface-card);
  z-index: 601;
  display: flex;
  flex-direction: column;
  box-shadow: -32px 0 64px -16px rgba(0, 0, 0, 0.18);

  &__header {
    padding: 22px 28px;
    border-bottom: 1px solid var(--border-hair);
    display: flex;
    align-items: center;
    gap: 16px;
  }

  &__title-block {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }

  &__eyebrow {
    font-size: 10px;
    color: var(--text-tertiary);
    letter-spacing: 0.12em;
  }

  &__title {
    font-family: var(--font-serif);
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }

  &__head-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__dirty {
    font-size: 10px;
    color: var(--text-tertiary);
    border: 1px solid var(--border-hair);
    border-radius: 999px;
    padding: 3px 10px;
    letter-spacing: 0.08em;
  }

  &__close {
    width: 28px;
    height: 28px;
    background: transparent;
    border: none;
    font-size: 22px;
    line-height: 1;
    color: var(--text-tertiary);
    cursor: pointer;
    border-radius: 4px;
    transition: background 120ms ease, color 120ms ease;

    &:hover {
      background: var(--surface-secondary);
      color: var(--text-primary);
    }
  }

  &__body {
    flex: 1;
    overflow-y: auto;
    padding: 24px 28px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  &__footer {
    padding: 16px 28px;
    border-top: 1px solid var(--border-hair);
    background: var(--brand-paper-warm);
    display: grid;
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    align-items: center;
    column-gap: 12px;
    row-gap: 6px;
  }

  &__hint {
    font-size: 11px;
    color: var(--text-tertiary);
    letter-spacing: 0.04em;
  }

  &__footer-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  &__note {
    grid-column: 1 / -1;
    font-size: 10px;
    color: var(--text-faint);
    text-align: left;
    letter-spacing: 0.04em;
  }
}

// === Fields ===
.field {
  display: flex;
  flex-direction: column;
  gap: 8px;

  &__label {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }

  &__required {
    color: var(--brand-danger);
    margin-left: 2px;
  }

  &__sub {
    font-size: 10px;
    color: var(--text-tertiary);
    letter-spacing: 0.06em;
  }

  &__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  &__head-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__hint {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  &__hint-btn {
    background: transparent;
    border: none;
    font-size: 11px;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 0;
    letter-spacing: 0.04em;

    &:hover { color: var(--brand-ink); }
  }

  &__input {
    width: 100%;
    height: 36px;
    padding: 0 12px;
    font-family: var(--font-sans);
    font-size: 13px;
    color: var(--text-primary);
    background: var(--surface-card);
    border: 1px solid var(--border-hair);
    border-radius: 4px;
    outline: none;
    transition: border-color 120ms ease;

    &::placeholder { color: var(--text-placeholder); }
    &:hover { border-color: var(--border-medium); }
    &:focus { border-color: var(--brand-ink); }

    &--inline { height: 34px; flex: 1; }
  }

  &__textarea {
    width: 100%;
    min-height: 88px;
    padding: 12px;
    font-family: var(--font-sans);
    font-size: 13px;
    color: var(--text-primary);
    background: var(--brand-paper-warm);
    border: 1px solid var(--border-hair);
    border-radius: 4px;
    outline: none;
    resize: vertical;
    transition: border-color 120ms ease;
    line-height: 1.55;

    &::placeholder { color: var(--text-placeholder); }
    &:hover { border-color: var(--border-medium); }
    &:focus { border-color: var(--brand-ink); }
  }
}

.textarea-wrap {
  position: relative;
}

.textarea-counter {
  position: absolute;
  top: 8px;
  right: 12px;
  font-size: 10px;
  color: var(--text-faint);
  letter-spacing: 0.04em;
  pointer-events: none;
}

// === Tag wall ===
.tag-wall {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 12px;
  border: 1px solid var(--border-hair-soft);
  border-radius: 4px;
  background: var(--brand-paper);
  min-height: 44px;
  align-items: center;

  &__empty {
    font-size: 11px;
    color: var(--text-faint);
    letter-spacing: 0.04em;
  }
}

.tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid transparent;
  line-height: 1.4;

  &--danger {
    background: var(--brand-danger-soft);
    color: var(--brand-danger);
    border-color: oklch(0.92 0.04 25);
  }

  &--sprout {
    background: var(--brand-sprout-soft);
    color: var(--brand-sprout);
    border-color: oklch(0.92 0.04 145);
  }

  &__close {
    cursor: pointer;
    opacity: 0.5;
    margin-left: 2px;
    font-size: 13px;
    line-height: 1;
    transition: opacity 120ms ease;

    &:hover { opacity: 1; }
  }
}

.tag-input-row {
  display: flex;
  gap: 8px;
}

.hairline-btn {
  height: 34px;
  padding: 0 14px;
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  font-family: var(--font-sans);
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
  border-radius: 4px;
  transition: border-color 120ms ease;

  &:hover { border-color: var(--brand-ink); }
}

.ink-btn {
  height: 34px;
  padding: 0 18px;
  background: var(--brand-ink);
  border: 1px solid var(--brand-ink);
  font-family: var(--font-sans);
  font-size: 12px;
  letter-spacing: 0.04em;
  color: var(--text-inverse);
  cursor: pointer;
  border-radius: 4px;
  transition: opacity 120ms ease;

  &:hover:not(:disabled) { opacity: 0.85; }
  &:active:not(:disabled) { transform: scale(0.98); }
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

// === CTA code-view ===
.cta-warn {
  font-size: 11px;
  color: var(--brand-danger);
  letter-spacing: 0.04em;
  margin-bottom: 4px;
}

.cta-code {
  background: #0F0F0E;
  color: #E8E5DC;
  font-family: var(--font-mono);
  font-size: 12.5px;
  line-height: 1.55;
  padding: 14px;
  border-radius: 6px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  outline: none;
  min-height: 90px;
  border: 1px solid #0F0F0E;
  transition: border-color 120ms ease;

  &:focus { border-color: var(--brand-sprout); }
}

// === Responsive ===
@media (max-width: 960px) {
  .brand-screen__main { padding: 24px; }
  .brand-card {
    grid-template-columns: 1fr;
    gap: 14px;
  }
  .brand-card__col-actions { justify-content: flex-start; }
}

@media (max-width: 640px) {
  .brand-drawer { width: 100vw; }
}
</style>
