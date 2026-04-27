// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

import { del, get, post } from './request'
import type { ApiResponse } from '@/types/api'
import type { CreateTaskRequest, TaskVO, TaskListResponse } from '@/types/task'

/** Create a new content task */
export function createTask(data: CreateTaskRequest): Promise<ApiResponse<TaskVO>> {
  return post<ApiResponse<TaskVO>>('/tasks', data)
}

/** Get a task by public ID */
export function getTask(id: string): Promise<ApiResponse<TaskVO>> {
  return get<ApiResponse<TaskVO>>(`/tasks/${id}`)
}

/** List recent tasks with pagination and optional status filter */
export function listTasks(page = 1, pageSize = 20, status?: string): Promise<ApiResponse<TaskListResponse>> {
  const params: Record<string, unknown> = { page, page_size: pageSize }
  if (status) params.status = status
  return get<ApiResponse<TaskListResponse>>('/tasks', { params })
}

/** Retry a failed task */
export function retryTask(id: string): Promise<ApiResponse<TaskVO>> {
  return post<ApiResponse<TaskVO>>(`/tasks/${id}/retry`)
}

/** Delete a task */
export function deleteTask(id: string): Promise<ApiResponse<{ message: string }>> {
  return del<ApiResponse<{ message: string }>>(`/tasks/${id}`)
}

/** Batch delete tasks */
export function batchDeleteTasks(ids: string[]): Promise<ApiResponse<{ deleted: number }>> {
  return post<ApiResponse<{ deleted: number }>>('/tasks/batch-delete', { ids })
}

/** Cancel a pending or running task */
export function cancelTask(id: string): Promise<ApiResponse<null>> {
  return post<ApiResponse<null>>(`/tasks/${id}/cancel`)
}

/**
 * Create an SSE connection for real-time task progress.
 * Returns an EventSource instance — caller must close it when done.
 */
export function createTaskSSE(taskId: string): EventSource {
  const token = localStorage.getItem('readbud_token')
  const url = `/api/v1/tasks/${taskId}/events`
  // SSE doesn't support Authorization header natively;
  // we pass token as query param and handle it server-side
  return new EventSource(`${url}${token ? `?token=${token}` : ''}`)
}
