<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    :disabled="disabled"
    label-position="top"
    class="task-form"
  >
    <div class="form-section">
      <h3 class="form-section-title">基本信息</h3>
      <el-form-item label="关键词" prop="keyword">
        <el-input
          v-model="form.keyword"
          placeholder="输入内容关键词，如：新能源汽车"
          maxlength="255"
          show-word-limit
          clearable
        />
      </el-form-item>

      <el-form-item label="目标受众" prop="audience">
        <el-input
          v-model="form.audience"
          placeholder="如：科技爱好者、30-45岁职场人群"
          maxlength="255"
          clearable
        />
      </el-form-item>

      <el-form-item label="写作风格" prop="tone">
        <el-select v-model="form.tone" placeholder="选择文章风格" class="w-full">
          <el-option
            v-for="opt in toneOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="排版预设">
        <div class="style-grid">
          <button
            type="button"
            class="style-card"
            :class="{ active: form.article_style === '' }"
            @click="form.article_style = ''"
          >
            <span class="style-name">自动推荐</span>
            <span class="style-detail">交给 AI 根据主题与受众自动挑一种。</span>
          </button>
          <button
            v-for="(label, key) in ARTICLE_STYLE_LABELS"
            :key="key"
            type="button"
            class="style-card"
            :class="[`style-card--${key}`, { active: form.article_style === key }]"
            @click="form.article_style = key as ArticleStyle"
          >
            <span class="style-swatch"><span /><span /><span /></span>
            <span class="style-name">{{ label }}</span>
            <span class="style-detail">{{ ARTICLE_STYLE_DETAILS[key as ArticleStyle] }}</span>
          </button>
        </div>
      </el-form-item>

      <el-form-item label="目标字数" prop="target_words">
        <el-slider
          v-model="form.target_words"
          :min="500"
          :max="10000"
          :step="500"
          :marks="wordMarks"
          show-input
          :show-input-controls="false"
          input-size="small"
        />
      </el-form-item>
    </div>

    <el-divider />

    <div class="form-section">
      <h3 class="form-section-title">内容选项</h3>
      <el-form-item label="图片模式" prop="image_mode">
        <el-radio-group v-model="form.image_mode">
          <el-radio-button
            v-for="(label, key) in IMAGE_MODE_LABELS"
            :key="key"
            :value="key"
          >
            {{ label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="数据图表">
        <el-switch
          v-model="chartEnabled"
          active-text="自动生成"
          inactive-text="不需要"
        />
      </el-form-item>

      <el-form-item label="视觉增强">
        <el-switch
          v-model="form.visual_enhance"
          active-text="AI排版美化"
          inactive-text="预设模板"
        />
      </el-form-item>

      <el-form-item v-if="brandProfiles.length > 0" label="品牌档案">
        <el-select v-model="form.brand_profile_id" placeholder="默认品牌" class="w-full" clearable>
          <el-option
            v-for="bp in brandProfiles"
            :key="bp.public_id || bp.id"
            :label="bp.name"
            :value="bp.public_id || String(bp.id)"
          />
        </el-select>
      </el-form-item>
    </div>

    <el-divider />

    <div class="form-section">
      <h3 class="form-section-title">发布设置</h3>
      <el-form-item label="发布方式" prop="publish_mode">
        <el-radio-group v-model="form.publish_mode">
          <el-radio-button
            v-for="(label, key) in PUBLISH_MODE_LABELS"
            :key="key"
            :value="key"
          >
            {{ label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <el-form-item
        v-if="form.publish_mode === 'schedule'"
        label="发布时间"
        prop="publish_at"
      >
        <el-date-picker
          v-model="form.publish_at"
          type="datetime"
          placeholder="选择发布时间"
          :disabled-date="disabledDate"
          class="w-full"
        />
      </el-form-item>
    </div>

    <div class="form-actions">
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="disabled"
        size="large"
        class="submit-btn"
        @click="handleSubmit"
      >
        开始生成
      </el-button>
    </div>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { CreateTaskRequest, ImageMode, PublishMode, ArticleStyle } from '@/types/task'
import { IMAGE_MODE_LABELS, PUBLISH_MODE_LABELS, ARTICLE_STYLE_LABELS, ARTICLE_STYLE_DETAILS } from '@/types/task'
import { listBrandProfiles } from '@/api/brand'
import type { BrandProfileVO } from '@/types/brand'

interface Props {
  disabled?: boolean
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  submitting: false,
})

const emit = defineEmits<{
  submit: [payload: CreateTaskRequest]
}>()

const formRef = ref<FormInstance>()

const form = reactive({
  keyword: '',
  audience: '',
  tone: 'professional',
  target_words: 2000,
  image_mode: 'auto' as ImageMode,
  publish_mode: 'manual' as PublishMode,
  publish_at: undefined as string | undefined,
  article_style: '' as ArticleStyle | '',
  visual_enhance: false,
  brand_profile_id: '' as string,
})

const chartEnabled = ref(true)

const brandProfiles = ref<BrandProfileVO[]>([])

onMounted(async () => {
  try {
    const res = await listBrandProfiles()
    brandProfiles.value = res.data || []
  } catch {
    // ignore - brand profiles are optional
  }
})

const toneOptions = [
  { value: 'professional', label: '专业严谨' },
  { value: 'casual', label: '轻松日常' },
  { value: 'storytelling', label: '故事叙事' },
  { value: 'analytical', label: '数据分析' },
  { value: 'educational', label: '科普教育' },
  { value: 'news', label: '新闻资讯' },
]

const wordMarks = {
  500: '500',
  2000: '2000',
  5000: '5000',
  10000: '10000',
}

const rules: FormRules = {
  keyword: [
    { required: true, message: '请输入关键词', trigger: 'blur' },
    { min: 1, max: 255, message: '关键词长度 1-255 个字符', trigger: 'blur' },
  ],
  image_mode: [
    { required: true, message: '请选择图片模式', trigger: 'change' },
  ],
  publish_mode: [
    { required: true, message: '请选择发布方式', trigger: 'change' },
  ],
  publish_at: [
    {
      required: true,
      message: '请选择发布时间',
      trigger: 'change',
      validator: (_rule: unknown, value: string | undefined, callback: (error?: Error) => void) => {
        if (form.publish_mode === 'schedule' && !value) {
          callback(new Error('请选择发布时间'))
        } else {
          callback()
        }
      },
    },
  ],
}

const _disabled = computed(() => props.disabled)

function disabledDate(date: Date): boolean {
  return date.getTime() < Date.now() - 60 * 1000
}

async function handleSubmit(): Promise<void> {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  const payload: CreateTaskRequest = {
    keyword: form.keyword.trim(),
    audience: form.audience.trim() || undefined,
    tone: form.tone || undefined,
    target_words: form.target_words,
    image_mode: form.image_mode,
    chart_mode: chartEnabled.value ? 1 : 0,
    publish_mode: form.publish_mode,
    publish_at: form.publish_mode === 'schedule' ? form.publish_at : undefined,
    article_style: form.article_style || undefined,
    visual_enhance: form.visual_enhance,
    brand_profile_id: form.brand_profile_id || undefined,
  }

  emit('submit', payload)
}
</script>

<style lang="scss" scoped>
.task-form {
  :deep(.el-form-item__label) {
    font-weight: 500;
    color: var(--text-secondary) !important;
    padding-bottom: 4px;
  }
}

.form-section {
  margin-bottom: 8px;
}

.form-section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--border-medium);
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 14px;
}

.w-full {
  width: 100%;
}

.form-actions {
  padding-top: 14px;
}

.submit-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  font-weight: 600;
  background: var(--text-primary) !important;
  border-color: var(--text-primary) !important;
  color: var(--text-inverse) !important;
  border-radius: 8px !important;
  transition: all 0.15s ease;

  &:hover:not(:disabled) {
    background: var(--surface-inverse) !important;
    border-color: var(--surface-inverse) !important;
  }

  &:active:not(:disabled) {
    transform: scale(0.98);
  }
}

// Input overrides
:deep(.el-input__wrapper) {
  background: var(--surface-bg) !important;
  border: 1px solid var(--border-light) !important;
  box-shadow: none !important;
  border-radius: 8px !important;
  transition: all 0.15s ease;

  &:hover {
    border-color: var(--border-medium) !important;
  }

  &.is-focus {
    border-color: var(--text-primary) !important;
    box-shadow: 0 0 0 2px rgba(10, 10, 10, 0.1) !important;
  }
}

:deep(.el-input__inner) {
  color: var(--text-primary) !important;
  &::placeholder {
    color: var(--text-placeholder) !important;
  }
}

:deep(.el-input__count-inner) {
  background: transparent !important;
  color: var(--border-medium) !important;
}

// Select
:deep(.el-select__wrapper) {
  background: var(--surface-bg) !important;
  border: 1px solid var(--border-light) !important;
  box-shadow: none !important;
  border-radius: 8px !important;
}

// Divider
:deep(.el-divider) {
  margin: 18px 0;
  border-color: var(--border-light);
}

// Radio button group
:deep(.el-radio-group) {
  flex-wrap: wrap;
}

:deep(.el-radio-button__inner) {
  font-size: 13px;
  background: var(--surface-bg) !important;
  border-color: var(--border-light) !important;
  color: var(--text-secondary) !important;
  transition: all 0.15s ease;

  &:hover {
    color: var(--text-primary) !important;
  }
}

:deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: var(--text-primary) !important;
  border-color: var(--text-primary) !important;
  color: var(--text-inverse) !important;
  box-shadow: -1px 0 0 0 var(--text-primary) !important;
}

// Slider
:deep(.el-slider) {
  padding: 0 8px;

  .el-slider__runway {
    background-color: var(--border-light);
  }

  .el-slider__bar {
    background: var(--text-primary);
  }

  .el-slider__button {
    border-color: var(--text-primary);
    background: var(--text-primary);
  }

  .el-slider__marks-text {
    color: var(--border-medium);
    font-size: 11px;
  }
}

:deep(.el-slider__input) {
  .el-input__wrapper {
    background: var(--surface-bg) !important;
    border: 1px solid var(--border-light) !important;
    box-shadow: none !important;
  }
}

// Switch
:deep(.el-switch) {
  --el-switch-off-color: var(--border-medium);
  --el-switch-on-color: var(--text-primary);

  .el-switch__label {
    color: var(--text-secondary) !important;

    &.is-active {
      color: var(--text-primary) !important;
    }
  }
}

.style-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  width: 100%;
}

.style-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
  padding: 12px 14px;
  border: 1px solid var(--border-light);
  background: var(--surface-bg);
  border-radius: 10px;
  cursor: pointer;
  transition: border-color 0.15s ease, transform 0.1s ease, box-shadow 0.15s ease;
  font-family: inherit;
  color: var(--text-primary);

  &:hover {
    border-color: var(--text-primary);
  }

  &.active {
    border-color: var(--text-primary);
    box-shadow: 0 0 0 2px rgba(10, 10, 10, 0.06);
  }

  .style-swatch {
    display: flex;
    gap: 4px;
    margin-bottom: 2px;

    span {
      width: 14px;
      height: 14px;
      border-radius: 4px;
      border: 1px solid rgba(0, 0, 0, 0.06);
    }
  }

  &--minimal .style-swatch span:nth-child(1) { background: #ffffff; }
  &--minimal .style-swatch span:nth-child(2) { background: #111111; }
  &--minimal .style-swatch span:nth-child(3) { background: #ffe94d; }

  &--magazine .style-swatch span:nth-child(1) { background: #f2efe8; }
  &--magazine .style-swatch span:nth-child(2) { background: #0a0a0a; }
  &--magazine .style-swatch span:nth-child(3) { background: #e63946; }

  &--stitch .style-swatch span:nth-child(1) { background: #fcfaf5; }
  &--stitch .style-swatch span:nth-child(2) { background: #fdf2e5; }
  &--stitch .style-swatch span:nth-child(3) { background: #d2691e; }

  .style-name {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
    line-height: 1.3;
  }

  .style-detail {
    font-size: 11px;
    color: var(--text-secondary);
    line-height: 1.5;
  }
}

@media (max-width: 768px) {
  .style-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .form-section-title {
    font-size: 10px;
    margin-bottom: 10px;
  }

  :deep(.el-radio-group) {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  :deep(.el-radio-button) {
    flex: 0 0 auto;
  }

  :deep(.el-radio-button__inner) {
    padding: 4px 10px;
    font-size: 12px;
  }

  :deep(.el-slider) {
    padding: 0;
  }

  :deep(.el-slider__marks-text) {
    font-size: 10px;
  }

  .submit-btn {
    height: 40px;
    font-size: 14px;
  }

  :deep(.el-divider) {
    margin: 12px 0;
  }
}
</style>
