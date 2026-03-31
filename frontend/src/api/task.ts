import { get, post } from './request'
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

/** List recent tasks with pagination */
export function listTasks(page = 1, pageSize = 20): Promise<ApiResponse<TaskListResponse>> {
  return get<ApiResponse<TaskListResponse>>('/tasks', { params: { page, page_size: pageSize } })
}

/** Retry a failed task */
export function retryTask(id: string): Promise<ApiResponse<TaskVO>> {
  return post<ApiResponse<TaskVO>>(`/tasks/${id}/retry`)
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
