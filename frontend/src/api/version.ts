// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Draft version and citation API client

import { get, post } from './request'
import type { ApiResponse } from '@/types/api'
import type {
  DraftVersionVO,
  DraftVersionDetailVO,
  CitationVO,
  CreateSnapshotRequest,
  AddCitationRequest,
} from '@/types/version'

export function listDraftVersions(draftId: string): Promise<ApiResponse<DraftVersionVO[]>> {
  return get(`/drafts/${draftId}/versions`)
}

export function getDraftVersion(draftId: string, versionId: string): Promise<ApiResponse<DraftVersionDetailVO>> {
  return get(`/drafts/${draftId}/versions/${versionId}`)
}

export function createSnapshot(draftId: string, data: CreateSnapshotRequest): Promise<ApiResponse<null>> {
  return post(`/drafts/${draftId}/versions/snapshot`, data)
}

export function rollbackVersion(draftId: string, versionId: string): Promise<ApiResponse<null>> {
  return post(`/drafts/${draftId}/versions/${versionId}/rollback`)
}

export function getDraftCitations(draftId: string): Promise<ApiResponse<CitationVO[]>> {
  return get(`/drafts/${draftId}/citations`)
}

export function getBlockCitations(draftId: string, blockId: string): Promise<ApiResponse<CitationVO[]>> {
  return get(`/drafts/${draftId}/blocks/${blockId}/citations`)
}

export function addCitation(draftId: string, data: AddCitationRequest): Promise<ApiResponse<null>> {
  return post(`/drafts/${draftId}/citations`, data)
}
