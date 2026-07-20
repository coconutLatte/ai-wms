// ── API Response Envelope ──────────────────────────────────────────────────
// Matches the Go backend's ListResponse[T] generic envelope.

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

// ── Warehouse ──────────────────────────────────────────────────────────────

export interface Warehouse {
  id: string
  code: string
  name: string
  address: string
  status: string
  created_at: string
  updated_at: string
}

export interface CreateWarehouseRequest {
  code: string
  name: string
  address: string
}

export interface UpdateWarehouseRequest {
  name?: string
  address?: string
  status?: string
}

export interface Zone {
  id: string
  warehouse_id: string
  code: string
  name: string
  zone_type: string
  status: string
  created_at: string
  updated_at: string
}

export interface CreateZoneRequest {
  code: string
  name: string
  zone_type: string
}

export interface Location {
  id: string
  zone_id: string
  warehouse_id: string
  code: string
  barcode: string
  location_type: string
  capacity?: {
    max_weight: number
    max_volume: number
    max_qty: number
  }
  status: string
  created_at: string
  updated_at: string
}

export interface CreateLocationRequest {
  code: string
  barcode?: string
  location_type: string
  capacity?: {
    max_weight: number
    max_volume: number
    max_qty: number
  }
}

// ── SKU ────────────────────────────────────────────────────────────────────

export interface SKU {
  id: string
  code: string
  name: string
  category: string
  unit: string
  barcode: string
  attributes: Record<string, unknown>
  created_at: string
  updated_at: string
}

// ── Inventory ──────────────────────────────────────────────────────────────

export interface Inventory {
  id: string
  sku_id: string
  location_id: string
  batch_no: string
  quantity: number
  reserved_quantity: number
  created_at: string
  updated_at: string
}

export interface InventoryTransaction {
  id: string
  inventory_id: string
  type: string
  quantity: number
  reference_type: string
  reference_id: string
  created_at: string
}

// ── Orders ─────────────────────────────────────────────────────────────────

export interface Order {
  id: string
  order_no: string
  type: string
  status: string
  warehouse_id: string
  priority: number
  lines?: OrderLine[]
  created_at: string
  updated_at: string
}

export interface OrderLine {
  id: string
  order_id: string
  sku_id: string
  quantity: number
  fulfilled_quantity: number
  status: string
}

// ── Tasks ──────────────────────────────────────────────────────────────────

export interface Task {
  id: string
  task_no: string
  type: string
  status: string
  warehouse_id: string
  assignee_id: string
  reference_type: string
  reference_id: string
  priority: number
  created_at: string
  updated_at: string
}
