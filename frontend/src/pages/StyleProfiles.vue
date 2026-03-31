<template>
  <div class="style-page">
    <header class="style-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">写作模板</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Workbench' })">返回工作台</el-button>
      </div>
    </header>

    <main class="style-main">
      <div class="section-card">
        <div class="section-toolbar">
          <div>
            <h3 class="section-title">写作模板管理</h3>
            <p class="section-desc">配置文章结构模板、开头模板、结尾模板，统一写作风格</p>
          </div>
          <el-button type="primary" @click="openCreateDialog">添加模板</el-button>
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
          description="暂无写作模板，点击上方按钮添加"
        />

        <el-table
          v-else
          :data="profiles"
          stripe
          class="style-table"
        >
          <el-table-column prop="name" label="模板名称" min-width="140">
            <template #default="{ row }">
              <span class="profile-link" @click="openEditDialog(row)">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="applicable_scene" label="适用场景" min-width="180">
            <template #default="{ row }">
              <span class="scene-text">{{ row.applicable_scene || '通用' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="开头模板" min-width="200">
            <template #default="{ row }">
              <span class="template-preview">{{ truncate(row.opening_template, 60) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="结尾模板" min-width="200">
            <template #default="{ row }">
              <span class="template-preview">{{ truncate(row.closing_template, 60) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="openEditDialog(row)">编辑</el-button>
              <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </main>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showDialog"
      :title="editingId ? '编辑写作模板' : '添加写作模板'"
      width="720"
      class="style-dialog"
      destroy-on-close
    >
      <el-form :model="form" label-position="top">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="模板名称" required>
              <el-input v-model="form.name" placeholder="例如：深度分析模板" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="适用场景">
              <el-input v-model="form.applicable_scene" placeholder="例如：行业分析、产品评测..." />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="开头模板">
          <el-input
            v-model="form.opening_template"
            type="textarea"
            :rows="4"
            placeholder="文章开头的写作模板，支持变量如 {keyword}、{audience}..."
          />
        </el-form-item>

        <el-form-item label="结构模板 (JSON)">
          <el-input
            v-model="structureTemplateStr"
            type="textarea"
            :rows="6"
            placeholder='{"sections": [{"type": "intro", "length": "short"}, {"type": "analysis", "length": "long"}, {"type": "conclusion"}]}'
          />
        </el-form-item>

        <el-form-item label="结尾模板">
          <el-input
            v-model="form.closing_template"
            type="textarea"
            :rows="4"
            placeholder="文章结尾的写作模板，包含 CTA 设计..."
          />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="句式偏好 (JSON)">
              <el-input
                v-model="sentencePrefStr"
                type="textarea"
                :rows="3"
                placeholder='{"avg_length": "medium", "style": "formal"}'
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="标题偏好 (JSON)">
              <el-input
                v-model="titlePrefStr"
                type="textarea"
                :rows="3"
                placeholder='{"types": ["数字型", "问题型"], "max_length": 30}'
              />
            </el-form-item>
          </el-col>
        </el-row>
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
  listStyleProfiles,
  createStyleProfile,
  updateStyleProfile,
  deleteStyleProfile,
} from '@/api/brand'
import type { StyleProfileVO } from '@/types/brand'

const router = useRouter()

const loading = ref(false)
const error = ref<string | null>(null)
const profiles = ref<StyleProfileVO[]>([])

const showDialog = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const form = reactive({
  name: '',
  applicable_scene: '',
  opening_template: '',
  closing_template: '',
})
const structureTemplateStr = ref('')
const sentencePrefStr = ref('')
const titlePrefStr = ref('')

function truncate(str: string | undefined, max: number): string {
  if (!str) return '—'
  return str.length > max ? str.slice(0, max) + '...' : str
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
    const resp = await listStyleProfiles()
    if (resp.code === 0) {
      profiles.value = resp.data || []
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载写作模板失败'
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingId.value = null
  form.name = ''
  form.applicable_scene = ''
  form.opening_template = ''
  form.closing_template = ''
  structureTemplateStr.value = ''
  sentencePrefStr.value = ''
  titlePrefStr.value = ''
  showDialog.value = true
}

function openEditDialog(profile: StyleProfileVO) {
  editingId.value = profile.id
  form.name = profile.name
  form.applicable_scene = profile.applicable_scene || ''
  form.opening_template = profile.opening_template || ''
  form.closing_template = profile.closing_template || ''
  structureTemplateStr.value = profile.structure_template
    ? JSON.stringify(profile.structure_template, null, 2)
    : ''
  sentencePrefStr.value = profile.sentence_preference
    ? JSON.stringify(profile.sentence_preference, null, 2)
    : ''
  titlePrefStr.value = profile.title_preference
    ? JSON.stringify(profile.title_preference, null, 2)
    : ''
  showDialog.value = true
}

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入模板名称')
    return
  }

  const structureTemplate = structureTemplateStr.value ? safeParseJSON(structureTemplateStr.value) : undefined
  const sentencePref = sentencePrefStr.value ? safeParseJSON(sentencePrefStr.value) : undefined
  const titlePref = titlePrefStr.value ? safeParseJSON(titlePrefStr.value) : undefined

  if (
    (structureTemplateStr.value && structureTemplate === undefined) ||
    (sentencePrefStr.value && sentencePref === undefined) ||
    (titlePrefStr.value && titlePref === undefined)
  ) {
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.name,
      applicable_scene: form.applicable_scene,
      opening_template: form.opening_template,
      closing_template: form.closing_template,
      structure_template: structureTemplate,
      sentence_preference: sentencePref,
      title_preference: titlePref,
    }

    if (editingId.value) {
      await updateStyleProfile(editingId.value, payload)
      ElMessage.success('模板已更新')
    } else {
      await createStyleProfile(payload)
      ElMessage.success('模板已创建')
    }
    showDialog.value = false
    await loadProfiles()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(profile: StyleProfileVO) {
  try {
    await ElMessageBox.confirm(
      `确定要删除写作模板「${profile.name}」吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' },
    )
    await deleteStyleProfile(profile.id)
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
.style-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $color-bg;
}

.style-header {
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

.style-main {
  max-width: 1100px;
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
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: $spacing-lg;
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

.style-table {
  width: 100%;
}

.profile-link {
  color: $color-accent;
  font-weight: $font-weight-medium;
  cursor: pointer;

  &:hover {
    text-decoration: underline;
  }
}

.scene-text {
  color: $color-text-secondary;
  font-size: $font-size-sm;
}

.template-preview {
  font-size: $font-size-sm;
  color: $color-text-muted;
  line-height: $line-height-normal;
}

// Responsive
@media (max-width: $breakpoint-md) {
  .style-main {
    padding: $spacing-base;
    max-width: 100%;
  }

  .section-card {
    padding: $spacing-base;
  }

  .section-toolbar {
    flex-direction: column;
    gap: $spacing-sm;
  }
}

@media (max-width: $breakpoint-sm) {
  .style-header {
    height: 48px;
    padding: 0 $spacing-sm;
  }

  .header-desc,
  .header-divider {
    display: none;
  }

  .style-main {
    padding: $spacing-sm;
  }

  .section-card {
    padding: $spacing-sm;
  }
}
</style>
