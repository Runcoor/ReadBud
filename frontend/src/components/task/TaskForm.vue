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
import { ref, reactive, computed } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { CreateTaskRequest, ImageMode, PublishMode } from '@/types/task'
import { IMAGE_MODE_LABELS, PUBLISH_MODE_LABELS } from '@/types/task'

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
})

const chartEnabled = ref(true)

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
  }

  emit('submit', payload)
}
</script>

<style lang="scss" scoped>
.task-form {
  :deep(.el-form-item__label) {
    font-weight: $font-weight-medium;
    color: $color-text-primary;
    padding-bottom: $spacing-xs;
  }
}

.form-section {
  margin-bottom: $spacing-sm;
}

.form-section-title {
  font-size: $font-size-sm;
  font-weight: $font-weight-semibold;
  color: $color-text-secondary;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: $spacing-base;
}

.w-full {
  width: 100%;
}

.form-actions {
  padding-top: $spacing-base;
}

.submit-btn {
  width: 100%;
  height: 44px;
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  background-color: $color-primary;
  border-color: $color-primary;

  &:hover:not(:disabled) {
    background-color: lighten($color-primary, 8%);
    border-color: lighten($color-primary, 8%);
  }
}

:deep(.el-divider) {
  margin: $spacing-lg 0;
  border-color: $color-divider;
}

:deep(.el-radio-button__inner) {
  font-size: $font-size-sm;
}

:deep(.el-slider) {
  padding: 0 $spacing-sm;
}

// Responsive: Mobile — wrap radio buttons, compact slider
@media (max-width: $breakpoint-sm) {
  .form-section-title {
    font-size: $font-size-xs;
    margin-bottom: $spacing-sm;
  }

  :deep(.el-radio-group) {
    display: flex;
    flex-wrap: wrap;
    gap: $spacing-xs;
  }

  :deep(.el-radio-button) {
    flex: 0 0 auto;
  }

  :deep(.el-radio-button__inner) {
    padding: $spacing-xs $spacing-sm;
    font-size: $font-size-xs;
  }

  :deep(.el-slider) {
    padding: 0;
  }

  :deep(.el-slider__marks-text) {
    font-size: 10px;
  }

  .submit-btn {
    height: 40px;
    font-size: $font-size-base;
  }

  :deep(.el-divider) {
    margin: $spacing-base 0;
  }
}
</style>
