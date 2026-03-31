import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { createTask, getTask, listTasks, retryTask, cancelTask, createTaskSSE } from '@/api/task'
import type { TaskVO, CreateTaskRequest, TaskStatus } from '@/types/task'

export const useTaskStore = defineStore('task', () => {
  // --- State ---
  const currentTask = ref<TaskVO | null>(null)
  const tasks = ref<TaskVO[]>([])
  const total = ref(0)
  const loading = ref(false)
  const creating = ref(false)
  const sseConnection = ref<EventSource | null>(null)

  // --- Getters ---
  const isRunning = computed(() =>
    currentTask.value?.status === 'running' || currentTask.value?.status === 'pending',
  )

  const isDone = computed(() => currentTask.value?.status === 'done')
  const isFailed = computed(() => currentTask.value?.status === 'failed')

  // --- Actions ---

  /** Create a new task and start listening for progress */
  async function create(req: CreateTaskRequest): Promise<TaskVO> {
    creating.value = true
    try {
      const res = await createTask(req)
      currentTask.value = res.data
      connectSSE(res.data.id)
      return res.data
    } finally {
      creating.value = false
    }
  }

  /** Fetch a task by ID */
  async function fetchTask(id: string): Promise<void> {
    loading.value = true
    try {
      const res = await getTask(id)
      currentTask.value = res.data
    } finally {
      loading.value = false
    }
  }

  /** Fetch the task list */
  async function fetchList(page = 1, pageSize = 20): Promise<void> {
    loading.value = true
    try {
      const res = await listTasks(page, pageSize)
      tasks.value = res.data.items
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  /** Retry a failed task */
  async function retry(id: string): Promise<void> {
    const res = await retryTask(id)
    currentTask.value = res.data
    connectSSE(res.data.id)
  }

  /** Cancel a pending or running task */
  async function cancel(id: string): Promise<void> {
    await cancelTask(id)
    disconnectSSE()
    currentTask.value = null
  }

  /** Auto-recovery: find an active or recent task on page load */
  async function init(): Promise<void> {
    try {
      // 1. Check for a running task
      const runningRes = await listTasks(1, 1, 'running')
      if (runningRes.data.items.length > 0) {
        currentTask.value = runningRes.data.items[0]
        connectSSE(runningRes.data.items[0].id)
        return
      }

      // 2. Check for a pending task
      const pendingRes = await listTasks(1, 1, 'pending')
      if (pendingRes.data.items.length > 0) {
        currentTask.value = pendingRes.data.items[0]
        connectSSE(pendingRes.data.items[0].id)
        return
      }

      // 3. Show most recent task (no SSE)
      const recentRes = await listTasks(1, 1)
      if (recentRes.data.items.length > 0) {
        currentTask.value = recentRes.data.items[0]
      }
    } catch {
      // Silently catch errors during auto-recovery
    }
  }

  /** Connect to SSE for real-time progress updates */
  function connectSSE(taskId: string): void {
    disconnectSSE()

    const es = createTaskSSE(taskId)

    es.addEventListener('progress', (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data) as {
          data: { status: TaskStatus; stage: string; progress: number }
        }
        if (currentTask.value) {
          currentTask.value.status = payload.data.status
          currentTask.value.current_stage = payload.data.stage
          currentTask.value.progress = payload.data.progress
        }
      } catch {
        // Ignore malformed events
      }
    })

    es.addEventListener('failed', (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data) as {
          data: { status: TaskStatus; message: string }
        }
        if (currentTask.value) {
          currentTask.value.status = payload.data.status
          currentTask.value.error_message = payload.data.message
        }
      } catch {
        // Ignore malformed events
      }
      disconnectSSE()
    })

    es.addEventListener('done', () => {
      if (currentTask.value) {
        currentTask.value.status = 'done'
        currentTask.value.progress = 100
      }
      disconnectSSE()
    })

    es.onerror = () => {
      // EventSource will auto-reconnect, but we close on terminal states
      if (
        currentTask.value?.status === 'done' ||
        currentTask.value?.status === 'failed' ||
        currentTask.value?.status === 'cancelled'
      ) {
        disconnectSSE()
      }
    }

    sseConnection.value = es
  }

  /** Disconnect the SSE stream */
  function disconnectSSE(): void {
    if (sseConnection.value) {
      sseConnection.value.close()
      sseConnection.value = null
    }
  }

  /** Reset the current task state */
  function resetCurrent(): void {
    disconnectSSE()
    currentTask.value = null
  }

  return {
    // State
    currentTask,
    tasks,
    total,
    loading,
    creating,
    // Getters
    isRunning,
    isDone,
    isFailed,
    // Actions
    create,
    fetchTask,
    fetchList,
    retry,
    cancel,
    init,
    connectSSE,
    disconnectSSE,
    resetCurrent,
  }
})
