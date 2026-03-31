<template>
  <div class="brand-page">
    <header class="brand-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">品牌配置</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Workbench' })">返回工作台</el-button>
      </div>
    </header>

    <main class="brand-main">
      <div class="section-card">
        <div class="section-toolbar">
          <h3 class="section-title">品牌风格管理</h3>
          <p class="section-desc">配置品牌语气、禁用词、偏好表达，让生成内容更像品牌自己写的</p>
        </div>
        <div class="section-actions">
          <el-button type="primary" @click="openCreateDialog">添加品牌配置</el-button>
        </div>

        <div v-if="loading" class="section-loading">
          <el-skeleton :rows="4" animated />
        </div>

        <div v-else-if="error" class="section-error">
          <el-alert type="error" :title="error" :closable="false" show-icon />
          <el-button size="small" type="primary" plain @click="loadProfiles">重试</el-button>
        </div>

        <el-empty
          v-else-if="profiles.length === 0"
          description="暂无品牌配置，点击上方按钮添加"
        />

        <div v-else class="profile-grid">
          <div
            v-for="profile in profiles"
            :key="profile.id"
            class="profile-card"
          >
            <div class="profile-card-header">
              <h4 class="profile-name">{{ profile.name }}</h4>
              <div class="profile-card-actions">
                <el-button size="small" text type="primary" @click="openEditDialog(profile)">
                  编辑
                </el-button>
                <el-button size="small" text type="danger" @click="handleDelete(profile)">
                  删除
                </el-button>
              </div>
            </div>

            <div v-if="profile.brand_tone" class="profile-field">
              <span class="field-label">品牌语气</span>
              <span class="field-value">{{ profile.brand_tone }}</span>
            </div>

            <div class="profile-tags-row">
              <div v-if="profile.forbidden_words?.length" class="profile-field">
                <span class="field-label">禁用词</span>
                <div class="tag-group">
                  <el-tag
                    v-for="word in profile.forbidden_words.slice(0, 5)"
                    :key="word"
                    size="small"
                    type="danger"
                    class="word-tag"
                  >
                    {{ word }}
                  </el-tag>
                  <el-tag
                    v-if="profile.forbidden_words.length > 5"
                    size="small"
                    type="info"
                  >
                    +{{ profile.forbidden_words.length - 5 }}
                  </el-tag>
                </div>
              </div>

              <div v-if="profile.preferred_words?.length" class="profile-field">
                <span class="field-label">偏好词</span>
                <div class="tag-group">
                  <el-tag
                    v-for="word in profile.preferred_words.slice(0, 5)"
                    :key="word"
                    size="small"
                    type="success"
                    class="word-tag"
                  >
                    {{ word }}
                  </el-tag>
                  <el-tag
                    v-if="profile.preferred_words.length > 5"
                    size="small"
                    type="info"
                  >
                    +{{ profile.preferred_words.length - 5 }}
                  </el-tag>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showDialog"
      :title="editingId ? '编辑品牌配置' : '添加品牌配置'"
      width="640"
      class="brand-dialog"
      destroy-on-close
    >
      <el-form :model="form" label-position="top">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="品牌配置名称，例如：默认品牌" />
        </el-form-item>

        <el-form-item label="品牌语气">
          <el-input
            v-model="form.brand_tone"
            type="textarea"
            :rows="3"
            placeholder="描述品牌的沟通风格，例如：专业严谨、温和亲切、数据驱动..."
          />
        </el-form-item>

        <el-form-item label="禁用词">
          <div class="tag-input-area">
            <div class="tag-group">
              <el-tag
                v-for="(word, idx) in forbiddenWords"
                :key="idx"
                closable
                type="danger"
                class="word-tag"
                @close="forbiddenWords.splice(idx, 1)"
              >
                {{ word }}
              </el-tag>
            </div>
            <div class="tag-input-row">
              <el-input
                v-model="newForbiddenWord"
                size="small"
                placeholder="输入禁用词后回车"
                @keyup.enter="addForbiddenWord"
              />
              <el-button size="small" @click="addForbiddenWord">添加</el-button>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="偏好词">
          <div class="tag-input-area">
            <div class="tag-group">
              <el-tag
                v-for="(word, idx) in preferredWords"
                :key="idx"
                closable
                type="success"
                class="word-tag"
                @close="preferredWords.splice(idx, 1)"
              >
                {{ word }}
              </el-tag>
            </div>
            <div class="tag-input-row">
              <el-input
                v-model="newPreferredWord"
                size="small"
                placeholder="输入偏好词后回车"
                @keyup.enter="addPreferredWord"
              />
              <el-button size="small" @click="addPreferredWord">添加</el-button>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="CTA 规则 (JSON)">
          <el-input
            v-model="ctaRulesStr"
            type="textarea"
            :rows="3"
            placeholder='{"default_cta": "关注获取更多内容", "style": "soft"}'
          />
        </el-form-item>

        <el-form-item label="封面图风格规则 (JSON)">
          <el-input
            v-model="coverStyleStr"
            type="textarea"
            :rows="2"
            placeholder='{"prefer_style": "minimal", "color_scheme": "brand"}'
          />
        </el-form-item>

        <el-form-item label="配图风格规则 (JSON)">
          <el-input
            v-model="imageStyleStr"
            type="textarea"
            :rows="2"
            placeholder='{"prefer_type": "photography", "avoid": "cartoon"}'
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listBrandProfiles,
  createBrandProfile,
  updateBrandProfile,
  deleteBrandProfile,
} from '@/api/brand'
import type { BrandProfileVO } from '@/types/brand'

const router = useRouter()

const loading = ref(false)
const error = ref<string | null>(null)
const profiles = ref<BrandProfileVO[]>([])

const showDialog = ref(false)
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

function addForbiddenWord() {
  const word = newForbiddenWord.value.trim()
  if (word && !forbiddenWords.value.includes(word)) {
    forbiddenWords.value.push(word)
  }
  newForbiddenWord.value = ''
}

function addPreferredWord() {
  const word = newPreferredWord.value.trim()
  if (word && !preferredWords.value.includes(word)) {
    preferredWords.value.push(word)
  }
  newPreferredWord.value = ''
}

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

function openCreateDialog() {
  editingId.value = null
  form.name = ''
  form.brand_tone = ''
  forbiddenWords.value = []
  preferredWords.value = []
  ctaRulesStr.value = ''
  coverStyleStr.value = ''
  imageStyleStr.value = ''
  showDialog.value = true
}

function openEditDialog(profile: BrandProfileVO) {
  editingId.value = profile.id
  form.name = profile.name
  form.brand_tone = profile.brand_tone || ''
  forbiddenWords.value = [...(profile.forbidden_words || [])]
  preferredWords.value = [...(profile.preferred_words || [])]
  ctaRulesStr.value = profile.cta_rules ? JSON.stringify(profile.cta_rules, null, 2) : ''
  coverStyleStr.value = profile.cover_style_rules ? JSON.stringify(profile.cover_style_rules, null, 2) : ''
  imageStyleStr.value = profile.image_style_rules ? JSON.stringify(profile.image_style_rules, null, 2) : ''
  showDialog.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入名称')
    return
  }

  const ctaRules = ctaRulesStr.value ? safeParseJSON(ctaRulesStr.value) : undefined
  const coverRules = coverStyleStr.value ? safeParseJSON(coverStyleStr.value) : undefined
  const imageRules = imageStyleStr.value ? safeParseJSON(imageStyleStr.value) : undefined

  // If any JSON parse returned undefined (but had content), skip save
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
      await updateBrandProfile(editingId.value, payload)
      ElMessage.success('品牌配置已更新')
    } else {
      await createBrandProfile(payload)
      ElMessage.success('品牌配置已创建')
    }
    showDialog.value = false
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
    await deleteBrandProfile(profile.id)
    ElMessage.success('已删除')
    await loadProfiles()
  } catch (e: unknown) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  loadProfiles()
})
</script>

<style lang="scss" scoped>
.brand-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $color-bg;
}

.brand-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 $spacing-xl;
  background-color: $color-card-bg;
  border-bottom: 1px solid $color-border;
  box-shadow: $shadow-card;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.header-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-bold;
  color: $color-primary;
}

.header-divider {
  color: $color-border;
}

.header-desc {
  font-size: $font-size-base;
  color: $color-text-secondary;
}

.brand-main {
  max-width: 960px;
  width: 100%;
  margin: 0 auto;
  padding: $spacing-xl;
}

.section-card {
  background: $color-card-bg;
  border: 1px solid $color-border;
  border-radius: $radius-lg;
  padding: $spacing-xl;
}

.section-toolbar {
  margin-bottom: $spacing-sm;
}

.section-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0 0 $spacing-xs 0;
}

.section-desc {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin: 0;
}

.section-actions {
  margin-bottom: $spacing-lg;
}

.section-loading,
.section-error {
  padding: $spacing-xl 0;
}

.section-error {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  align-items: flex-start;
}

.profile-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: $spacing-base;
}

.profile-card {
  border: 1px solid $color-border;
  border-radius: $radius-base;
  padding: $spacing-lg;
  transition: border-color $transition-fast, box-shadow $transition-fast;

  &:hover {
    border-color: $color-accent;
    box-shadow: $shadow-card-hover;
  }
}

.profile-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-md;
}

.profile-name {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-primary;
  margin: 0;
}

.profile-card-actions {
  display: flex;
  gap: $spacing-xs;
}

.profile-field {
  margin-bottom: $spacing-sm;
}

.field-label {
  display: block;
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin-bottom: $spacing-xs;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.field-value {
  font-size: $font-size-base;
  color: $color-text-secondary;
  line-height: $line-height-relaxed;
}

.profile-tags-row {
  display: flex;
  gap: $spacing-xl;
  flex-wrap: wrap;
}

.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-xs;
}

.word-tag {
  margin: 0;
}

.tag-input-area {
  width: 100%;
}

.tag-input-row {
  display: flex;
  gap: $spacing-sm;
  margin-top: $spacing-sm;
}

// Responsive
@media (max-width: $breakpoint-md) {
  .brand-main {
    padding: $spacing-base;
    max-width: 100%;
  }

  .section-card {
    padding: $spacing-base;
  }
}

@media (max-width: $breakpoint-sm) {
  .brand-header {
    height: 48px;
    padding: 0 $spacing-sm;
  }

  .header-desc,
  .header-divider {
    display: none;
  }

  .brand-main {
    padding: $spacing-sm;
  }

  .section-card {
    padding: $spacing-sm;
  }

  .profile-tags-row {
    flex-direction: column;
    gap: $spacing-sm;
  }
}
</style>
