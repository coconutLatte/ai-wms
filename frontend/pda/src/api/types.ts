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
  | 'pick' | 'replenish' | 'transfer' | 'load' | 'unload'
export type TaskStatus = 'pending' | 'assigned' | 'in_progress' | 'completed' | 'cancelled' | 'paused' | 'exception'
export type TaskPriority = 'low' | 'normal' | 'high' | 'urgent'

export interface Task {
  id: string
  task_no: string
  task_type: TaskType
  warehouse_id: string
  order_id?: string
  order_line_id?: string
  priority: TaskPriority
  status: TaskStatus
  assigned_to?: string
  from_location_id?: string
  to_location_id?: string
  sku_id: string
  expected_qty: number
  actual_qty: number
  uom: string
  batch_no?: string
  instructions?: string
  created_at: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
}

export interface CreateTaskRequest {
  task_type: TaskType
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
