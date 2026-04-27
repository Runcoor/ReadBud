// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Draft version type definitions

import type { BlockVO } from './draft'

export interface DraftVersionVO {
  id: string
  version_no: number
  title: string
  digest: string
  operator_id?: number
  change_reason: string
  created_at: string
}

export interface DraftVersionDetailVO extends DraftVersionVO {
  blocks: BlockVO[]
}

export interface CitationVO {
  id: string
  block_id: string
  source_document_id: string
  citation_type: string
  citation_text: string
  source_link: string
  source_note: string
}

export interface CreateSnapshotRequest {
  change_reason: string
}

export interface AddCitationRequest {
  block_id: string
  source_document_id: string
  citation_type: string
  citation_text: string
  source_link?: string
  source_note?: string
}
