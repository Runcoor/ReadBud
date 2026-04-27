// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Draft-related type definitions

export interface BlockVO {
  id: string
  sort_no: number
  block_type: string
  heading?: string
  text_md?: string
  html_fragment?: string
  asset_url?: string
  attribution_text?: string
  prompt_text?: string
  status: string
}

export interface DraftVO {
  id: string
  task_id: string
  title: string
  subtitle?: string
  digest: string
  author_name: string
  cover_url?: string
  outline_json?: unknown
  quality_score: number
  similarity_score: number
  risk_level: string
  review_status: string
  version: number
  blocks: BlockVO[]
  created_at: string
  updated_at: string
}

export interface TitleCandidate {
  title: string
  type: string
}

export interface CoverVO {
  asset_id?: string
  url: string
  width?: number
  height?: number
  is_ai_generated: boolean
  prompt?: string
}

export interface SourceVO {
  id: string
  title: string
  source_type: string
  site_name: string
  source_url: string
  author?: string
  published_at?: string
  hot_score: number
  relevance_score: number
  summary?: string
}
