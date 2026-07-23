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

// ── Shipments ─────────────────────────────────────────────────────────────────

export type ShipmentStatus = 'pending' | 'in_transit' | 'delivered' | 'cancelled'

export interface Shipment {
  id: string
  shipment_no: string
  order_id: string
  warehouse_id: string
  status: ShipmentStatus
  carrier: string
  tracking_no?: string
  carrier_service?: string
  estimated_delivery?: string
  actual_delivery?: string
  notes?: string
  created_at: string
  updated_at: string
  shipped_at?: string
  delivered_at?: string
}

export interface CreateShipmentRequest {
  order_id: string
  warehouse_id: string
  carrier: string
  tracking_no?: string
  carrier_service?: string
  estimated_delivery?: string
  notes?: string
}

// ── ASN (Advanced Shipping Notice) ──────────────────────────────────────

export type ASNStatus = 'pending' | 'arrived' | 'receiving' | 'partial' | 'received'
export type ASNLineStatus = 'pending' | 'partial' | 'received'

export interface ASNLine {
  id: string
  asn_id: string
  sku_id: string
  expected_qty: number
  received_qty: number
  batch_no?: string
  status: ASNLineStatus
}

export interface ASN {
  id: string
  asn_no: string
  warehouse_id: string
  order_id?: string
  carrier?: string
  tracking_no?: string
  expected_at: string
  arrived_at?: string
  lines: ASNLine[]
  status: ASNStatus
  created_at: string
}

export interface ASNSummary {
  id: string
  asn_no: string
  warehouse_id: string
  order_id?: string
  carrier?: string
  tracking_no?: string
  expected_at: string
  arrived_at?: string
  status: ASNStatus
  created_at: string
}

export interface ReceiveLineRequest {
  received_qty: number
}

// ── Orders ──────────────────────────────────────────────────────────────────

export type OrderType = 'inbound' | 'outbound' | 'transfer' | 'return'
export type OrderStatus = 'draft' | 'confirmed' | 'processing' | 'partial' | 'completed' | 'cancelled'
export type OrderPriority = 'low' | 'normal' | 'high' | 'urgent'
export type OrderLineStatus = 'pending' | 'allocated' | 'partial' | 'fulfilled' | 'cancelled'

export interface OrderLine {
  id: string
  order_id: string
  line_no: number
  sku_id: string
  ordered_qty: number
  fulfilled_qty: number
  uom: string
  batch_no?: string
  status: OrderLineStatus
  notes?: string
}

export interface Order {
  id: string
  order_no: string
  order_type: OrderType
  warehouse_id: string
  status: OrderStatus
  priority: OrderPriority
  external_ref?: string
  external_type?: string
  lines: OrderLine[]
  notes?: string
  created_at: string
  updated_at: string
  completed_at?: string
  created_by: string
}

export interface OrderSummary {
  id: string
  order_no: string
  order_type: OrderType
  warehouse_id: string
  status: OrderStatus
  priority: OrderPriority
  external_ref?: string
  external_type?: string
  notes?: string
  created_at: string
  updated_at: string
  completed_at?: string
  created_by: string
}
