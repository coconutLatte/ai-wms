// ── API Response Envelope ──────────────────────────────────────────────────
// Shared types matching the Go backend's response structures.
// Reuses the same envelope as the Admin frontend.

export interface ApiError {
  type: string
  title: string
  status: number
  detail: string
  instance: string
}

export interface PaginationMeta {
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ListResponse<T> {
  data: T[]
  meta: PaginationMeta
}

// ── Auth ───────────────────────────────────────────────────────────────────

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface RefreshRequest {
  refresh_token: string
}

export interface UserProfile {
  id: string
  username: string
  display_name: string
  role_ids: string[]
  created_at: string
  updated_at: string
}

// ── Tasks ──────────────────────────────────────────────────────────────────

export type TaskType = 'receiving' | 'putaway' | 'picking' | 'cycle_count' | 'replenishment'
export type TaskStatus = 'pending' | 'assigned' | 'in_progress' | 'completed' | 'cancelled'

export interface Task {
  id: string
  task_no: string
  type: TaskType
  status: TaskStatus
  warehouse_id: string
  assignee_id: string
  reference_type: string
  reference_id: string
  priority: number
  created_at: string
  updated_at: string
}

export interface CreateTaskRequest {
  type: TaskType
  warehouse_id: string
  reference_type: string
  reference_id: string
  priority?: number
}

export interface AssignTaskRequest {
  assignee_id: string
}

export interface UpdateTaskStatusRequest {
  status: TaskStatus
}

export interface CompleteTaskRequest {
  notes?: string
}
