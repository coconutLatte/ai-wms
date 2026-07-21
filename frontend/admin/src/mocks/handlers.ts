// MSW handlers — mock all backend API endpoints for the GitHub Pages demo.
// Data is procedurally generated to reflect the current API structure.
// When API types change, these handlers should be updated to match.

import { http, HttpResponse, delay } from 'msw'

// ── Helpers ───────────────────────────────────────────────────

const now = new Date().toISOString()
const id = (n: number) => `00000000-0000-0000-0000-${String(n).padStart(12, '0')}`

function paginate<T>(items: T[], page: number, pageSize: number) {
  const total = items.length
  const totalPages = Math.ceil(total / pageSize)
  const start = (page - 1) * pageSize
  return {
    data: items.slice(start, start + pageSize),
    pagination: { total, page, page_size: pageSize, total_pages: totalPages },
  }
}

// ── Seed data ──────────────────────────────────────────────────

const warehouses = [
  { id: id(1), code: 'WH-SH-01', name: 'Shanghai Main Warehouse', address: 'No. 100, Zhangjiang Rd, Pudong', status: 'active', created_at: now, updated_at: now },
  { id: id(2), code: 'WH-SZ-01', name: 'Shenzhen Distribution Center', address: 'No. 200, Nanshan Ave, Shenzhen', status: 'active', created_at: now, updated_at: now },
]
const zones = [
  { id: id(10), warehouse_id: id(1), code: 'ZONE-RCV-01', name: 'Receiving Zone A', zone_type: 'receiving', status: 'active', created_at: now, updated_at: now },
  { id: id(11), warehouse_id: id(1), code: 'ZONE-STO-01', name: 'Storage Zone A', zone_type: 'storage', status: 'active', created_at: now, updated_at: now },
  { id: id(12), warehouse_id: id(1), code: 'ZONE-PCK-01', name: 'Picking Zone A', zone_type: 'picking', status: 'active', created_at: now, updated_at: now },
]
const skus = [
  { id: id(100), code: 'SKU-WIDGET-001', name: 'Widget Model A', barcode: '6901234567890', base_unit: 'EA', category: 'Widgets', status: 'active', created_at: now, updated_at: now },
  { id: id(101), code: 'SKU-BOLT-M8', name: 'M8 Hex Bolt', barcode: '6901234567891', base_unit: 'EA', category: 'Fasteners', status: 'active', created_at: now, updated_at: now },
  { id: id(102), code: 'SKU-GASKET-3MM', name: '3mm Rubber Gasket', barcode: '6901234567892', base_unit: 'EA', category: 'Seals', status: 'active', created_at: now, updated_at: now },
  { id: id(103), code: 'SKU-LABEL-TAG', name: 'RFID Label Tag', barcode: '6901234567893', base_unit: 'ROLL', category: 'Labels', status: 'active', created_at: now, updated_at: now },
]
const inventory = [
  { id: id(200), sku_id: id(100), location_id: id(1), warehouse_id: id(1), batch_no: 'B2026-001', qty: 1520, reserved_qty: 200, status: 'available', received_at: now, updated_at: now },
  { id: id(201), sku_id: id(101), location_id: id(2), warehouse_id: id(1), batch_no: 'B2026-002', qty: 8500, reserved_qty: 0, status: 'available', received_at: now, updated_at: now },
  { id: id(202), sku_id: id(102), location_id: id(3), warehouse_id: id(1), batch_no: 'B2026-003', qty: 320, reserved_qty: 50, status: 'available', received_at: now, updated_at: now },
  { id: id(203), sku_id: id(103), location_id: id(4), warehouse_id: id(1), batch_no: 'B2026-001', qty: 45, reserved_qty: 0, status: 'quarantine', received_at: now, updated_at: now },
]
const orders = [
  { id: id(300), order_no: 'IN-20260721-001', order_type: 'inbound', warehouse_id: id(1), status: 'processing', priority: 'high', external_ref: 'PO-2026-0042', created_at: now, updated_at: now },
  { id: id(301), order_no: 'OUT-20260721-001', order_type: 'outbound', warehouse_id: id(1), status: 'confirmed', priority: 'normal', external_ref: 'SO-2026-0157', created_at: now, updated_at: now },
  { id: id(302), order_no: 'OUT-20260721-002', order_type: 'outbound', warehouse_id: id(2), status: 'draft', priority: 'urgent', external_ref: 'SO-2026-0158', created_at: now, updated_at: now },
]
const tasks = [
  { id: id(400), task_no: 'TASK-20260721-000001', task_type: 'putaway', warehouse_id: id(1), sku_id: id(100), status: 'in_progress', priority: 'high', assigned_to: 'operator-zhang', expected_qty: 200, actual_qty: 0, created_at: now },
  { id: id(401), task_no: 'TASK-20260721-000002', task_type: 'pick', warehouse_id: id(1), sku_id: id(102), status: 'pending', priority: 'normal', assigned_to: '', expected_qty: 50, actual_qty: 0, created_at: now },
  { id: id(402), task_no: 'TASK-20260721-000003', task_type: 'cycle_count', warehouse_id: id(2), sku_id: id(101), status: 'assigned', priority: 'low', assigned_to: 'operator-li', expected_qty: 500, actual_qty: 0, created_at: now },
]

// ── Handlers ───────────────────────────────────────────────────

export const handlers = [
  // Auth
  http.post('/api/v1/auth/login', async ({ request }) => {
    await delay(300)
    const body = await request.json() as { username: string }
    return HttpResponse.json({
      access_token: 'demo-access-token',
      refresh_token: 'demo-refresh-token',
      token_type: 'Bearer',
      expires_in: 900,
      display_name: body.username || 'Demo Admin',
      role_names: ['admin'],
    })
  }),
  http.post('/api/v1/auth/refresh', async () => {
    await delay(100)
    return HttpResponse.json({ access_token: 'demo-access-token', refresh_token: 'demo-refresh-token', token_type: 'Bearer', expires_in: 900 })
  }),
  http.get('/api/v1/auth/me', async () => {
    return HttpResponse.json({ id: id(999), username: 'admin', display_name: 'Demo Admin', role_ids: [id(998)], role_names: ['admin'], created_at: now, updated_at: now })
  }),

  // Warehouses
  http.get('/api/v1/warehouses', async ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    return HttpResponse.json(paginate(warehouses, page, 10))
  }),
  http.get('/api/v1/warehouses/:id', async ({ params }) => {
    const w = warehouses.find(w => w.id === params.id)
    return w ? HttpResponse.json(w) : new HttpResponse(null, { status: 404 })
  }),
  http.get('/api/v1/warehouses/:id/zones', async () => {
    return HttpResponse.json(paginate(zones, 1, 10))
  }),

  // Inventory
  http.get('/api/v1/inventory', async ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const status = url.searchParams.get('status')
    const filtered = status ? inventory.filter(i => i.status === status) : inventory
    return HttpResponse.json(paginate(filtered, page, 10))
  }),
  http.get('/api/v1/inventory/dashboard', async () => {
    return HttpResponse.json({
      total_skus: skus.length,
      total_qty: inventory.reduce((s, i) => s + i.qty, 0),
      low_stock_count: inventory.filter(i => i.qty < 100).length,
      warehouse_breakdown: warehouses.map(w => ({
        warehouse_id: w.id,
        warehouse_name: w.name,
        sku_count: Math.floor(Math.random() * 100) + 10,
        total_qty: Math.floor(Math.random() * 50000) + 1000,
      })),
    })
  }),

  // SKUs
  http.get('/api/v1/skus', async ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    return HttpResponse.json(paginate(skus, page, 10))
  }),
  http.get('/api/v1/skus/:id', async ({ params }) => {
    const s = skus.find(s => s.id === params.id)
    return s ? HttpResponse.json(s) : new HttpResponse(null, { status: 404 })
  }),

  // Orders
  http.get('/api/v1/orders', async ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const type = url.searchParams.get('order_type')
    const filtered = type ? orders.filter(o => o.order_type === type) : orders
    return HttpResponse.json(paginate(filtered, page, 10))
  }),
  http.get('/api/v1/orders/:id', async ({ params }) => {
    const o = orders.find(o => o.id === params.id)
    return o ? HttpResponse.json({ ...o, lines: [] }) : new HttpResponse(null, { status: 404 })
  }),

  // Tasks
  http.get('/api/v1/tasks', async ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const status = url.searchParams.get('status')
    const filtered = status ? tasks.filter(t => t.status === status) : tasks
    return HttpResponse.json(paginate(filtered, page, 10))
  }),
  http.get('/api/v1/tasks/:id', async ({ params }) => {
    const t = tasks.find(t => t.id === params.id)
    return t ? HttpResponse.json(t) : new HttpResponse(null, { status: 404 })
  }),
]
