// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Distribution package types

export interface DistributionVO {
  public_id: string
  draft_id: number
  community_copy: string
  moments_copy: string
  summary_card: string
  comment_guide: string
  next_topic_suggestion: string
  created_at: string
  updated_at: string
}

export interface GenerateDistributionRequest {
  draft_public_id: string
}
