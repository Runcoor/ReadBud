// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { ref } from 'vue'
import type { DistributionVO } from '@/types/distribution'
import {
  generateDistribution,
  getDistributionByDraft,
  deleteDistribution,
} from '@/api/distribution'

export function useDistribution() {
  const distribution = ref<DistributionVO | null>(null)
  const loading = ref(false)
  const generating = ref(false)
  const error = ref<string | null>(null)

  async function loadByDraft(draftPublicId: string) {
    loading.value = true
    error.value = null
    try {
      const res = await getDistributionByDraft(draftPublicId)
      if (res.code === 0 && res.data) {
        distribution.value = res.data
      } else {
        distribution.value = null
      }
    } catch {
      // 404 is expected when no package exists yet
      distribution.value = null
    } finally {
      loading.value = false
    }
  }

  async function generate(draftPublicId: string) {
    generating.value = true
    error.value = null
    try {
      const res = await generateDistribution({ draft_public_id: draftPublicId })
      if (res.code === 0 && res.data) {
        distribution.value = res.data
      } else {
        error.value = res.message || '生成分发素材包失败'
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '生成分发素材包失败'
      error.value = msg
    } finally {
      generating.value = false
    }
  }

  async function remove(publicId: string) {
    loading.value = true
    error.value = null
    try {
      const res = await deleteDistribution(publicId)
      if (res.code === 0) {
        distribution.value = null
      } else {
        error.value = res.message || '删除分发素材包失败'
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '删除分发素材包失败'
      error.value = msg
    } finally {
      loading.value = false
    }
  }

  return {
    distribution,
    loading,
    generating,
    error,
    loadByDraft,
    generate,
    remove,
  }
}
