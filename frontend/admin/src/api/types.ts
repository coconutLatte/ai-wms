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
  pagination: PaginationMeta
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

export interface UOM {
  base_unit: string
  pack_unit: string
  pack_qty: number
  weight: number
  volume: number
  length: number
  width: number
  height: number
}

export interface SKU {
  id: string
  code: string
  name: string
  description: string
  barcode: string
  uom: UOM
  attributes: Record<string, string>
  category: string
  status: string
  created_at: string
  updated_at: string
}

export interface CreateSKURequest {
  code: string
  name: string
  description?: string
  barcode?: string
  uom: UOM
  attributes?: Record<string, string>
  category?: string
}

export interface UpdateSKURequest {
  name?: string
  description?: string
  barcode?: string
  uom?: UOM
  attributes?: Record<string, string>
  category?: string
  status?: string
}

// ── Inventory ──────────────────────────────────────────────────────────────

export interface Inventory {
  id: string
  sku_id: string
  location_id: string
  warehouse_id: string
  batch_no: string
  qty: number
  reserved_qty: number
  available_qty: number
  status: string
  production_date?: string
  expiry_date?: string
  received_at: string
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

// ── Inventory Dashboard ────────────────────────────────────────────────────

export interface DashboardStats {
  total_records: number
  total_qty: number
  total_reserved_qty: number
  total_available_qty: number
  available_count: number
  quarantine_count: number
  damaged_count: number
  expired_count: number
  low_stock_count: number
}

export interface WarehouseBreakdown {
  warehouse_id: string
  warehouse_name: string
  warehouse_code: string
  total_qty: number
  reserved_qty: number
  available_qty: number
  record_count: number
}

export interface DashboardResponse {
  stats: DashboardStats
  low_stock: Inventory[]
  by_warehouse: WarehouseBreakdown[]
}

// ── Admin Dashboard (comprehensive) ─────────────────────────────────────────

export interface AdminDashboardResponse {
  warehouse_count: number
  sku_count: number
  inventory_stats: DashboardStats | null
  order_summary: Record<string, number> | null
  task_summary: Record<string, number> | null
}

// ── Orders ─────────────────────────────────────────────────────────────────

export interface OrderSummary {
  id: string
  order_no: string
  order_type: string
  warehouse_id: string
  status: string
  priority: string
  external_ref: string
  external_type: string
  notes: string
  created_at: string
  updated_at: string
  completed_at?: string
  created_by: string
}

export interface OrderLine {
  id: string
  order_id: string
  line_no: number
  sku_id: string
  ordered_qty: number
  fulfilled_qty: number
  uom: string
  batch_no: string
  status: string
  notes: string
}

export interface Order extends OrderSummary {
  lines: OrderLine[]
}

export interface UpdateOrderStatusRequest {
  status: string
  notes?: string
}

export interface CreateOrderLineRequest {
  sku_id: string
  ordered_qty: number
  uom?: string
  batch_no?: string
  notes?: string
}

export interface CreateOrderRequest {
  order_no?: string
  order_type: string
  warehouse_id: string
  priority?: string
  external_ref?: string
  external_type?: string
  lines: CreateOrderLineRequest[]
  notes?: string
  created_by: string
}

// ── Tasks ──────────────────────────────────────────────────────────────────

export interface Task {
  id: string
  task_no: string
  task_type: string
  warehouse_id: string
  order_id?: string
  order_line_id?: string
  priority: string
  status: string
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

export interface AssignTaskRequest {
  assigned_to: string
}

export interface UpdateTaskStatusRequest {
  status: string
}

export interface CompleteTaskRequest {
  actual_qty: number
  batch_no?: string
  from_location_id?: string
  to_location_id?: string
}

// ── Users ────────────────────────────────────────────────────────────────────

export interface User {
  id: string
  username: string
  email: string
  display_name: string
  role_ids: string[]
  status: string
  last_login?: string
  created_at: string
  updated_at: string
}

export interface CreateUserRequest {
  username: string
  email: string
  password: string
  display_name?: string
  role_ids?: string[]
}

export interface UpdateUserRequest {
  email?: string
  display_name?: string
  role_ids?: string[]
}

export interface UpdateUserStatusRequest {
  status: string
}

// ── Roles ────────────────────────────────────────────────────────────────────

export interface Permission {
  resource: string
  actions: string[]
}

export interface Role {
  id: string
  name: string
  description: string
  permissions: Permission[]
  created_at: string
}

export interface CreateRoleRequest {
  name: string
  description?: string
  permissions?: Permission[]
}

export interface UpdateRoleRequest {
  name?: string
  description?: string
  permissions?: Permission[]
}
