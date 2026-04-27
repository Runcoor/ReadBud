<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="timeline">
    <div
      v-for="(stage, idx) in TASK_STAGES"
      :key="stage.key"
      class="timeline-row"
    >
      <div class="timeline-rail">
        <span :class="['timeline-dot', `timeline-dot--${stageState(idx)}`]">
          <span v-if="stageState(idx) === 'done'" class="timeline-dot__inner" />
          <span v-else-if="stageState(idx) === 'active'" class="timeline-dot__pulse" />
        </span>
        <span v-if="idx < TASK_STAGES.length - 1" class="timeline-line" />
      </div>
      <div class="timeline-content">
        <span :class="['timeline-name', { 'is-mute': stageState(idx) === 'pending' }]">{{ stage.label }}</span>
        <span class="timeline-meta mono">{{ stageMeta(idx) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { TaskVO } from '@/types/task'

interface Props {
  task: TaskVO | null
}

const props = defineProps<Props>()

const TASK_STAGES = [
  { key: 'keyword_expand', label: '关键词扩展' },
  { key: 'source_search',  label: '素材搜集' },
  { key: 'content_crawl',  label: '内容采集' },
  { key: 'hot_score',      label: '热度评分' },
  { key: 'article_write',  label: '文案撰写' },
  { key: 'image_match',    label: '图片匹配' },
  { key: 'chart_gen',      label: '图表生成' },
  { key: 'html_compile',   label: 'HTML 编译' },
  { key: 'self_check',     label: '自检校对' },
  { key: 'publish',        label: '发布' },
]

// Some backend stages have legacy names; map them to the canonical key.
const STAGE_ALIAS: Record<string, string> = {
  review: 'self_check',
}

function normalize(stage?: string | null): string | null {
  if (!stage) return null
  return STAGE_ALIAS[stage] ?? stage
}

const currentStageIndex = computed(() => {
  const s = normalize(props.task?.current_stage)
  return s ? TASK_STAGES.findIndex(t => t.key === s) : -1
})

type StageState = 'done' | 'active' | 'pending'

function stageState(idx: number): StageState {
  const t = props.task
  if (!t) return 'pending'
  if (t.status === 'done') return 'done'
  const cur = currentStageIndex.value
  if (cur < 0) return 'pending'
  if (idx < cur) return 'done'
  if (idx === cur && t.status === 'running') return 'active'
  return 'pending'
}

function stageMeta(idx: number): string {
  const s = stageState(idx)
  if (s === 'done') return '完成'
  if (s === 'active') return '执行中'
  return '—'
}
</script>

<style lang="scss" scoped>
.timeline {
  position: relative;
}

.mono {
  font-family: var(--font-mono);
  letter-spacing: 0.04em;
}

.timeline-row {
  display: grid;
  grid-template-columns: 18px 1fr;
  align-items: center;
  gap: 10px;
  min-height: 28px;
}

.timeline-rail {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
}

.timeline-dot {
  width: 11px;
  height: 11px;
  flex-shrink: 0;
  border: 1px solid var(--border-medium);
  border-radius: 50%;
  background: var(--surface-card);
  display: grid;
  place-items: center;
  z-index: 1;

  &--done   { border-color: var(--brand-sprout); }
  &--active { border-color: var(--brand-warn); }

  &__inner {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--brand-sprout);
  }

  &__pulse {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--brand-warn);
    animation: timeline-pulse 1.4s ease-in-out infinite;
  }
}

@keyframes timeline-pulse {
  0%, 100% { opacity: 1; }
  50%      { opacity: 0.35; }
}

.timeline-line {
  flex: 1;
  width: 1px;
  min-height: 16px;
  margin-top: -1px;
  background: var(--border-hair);
}

.timeline-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  color: var(--text-body);
  font-size: 12px;
}

.timeline-name.is-mute {
  color: var(--text-tertiary);
}

.timeline-meta {
  font-size: 10px;
  color: var(--text-tertiary);
  text-align: right;
  flex-shrink: 0;
}
</style>
