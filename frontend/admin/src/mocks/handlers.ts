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

// ── Seed data (Chinese by default, matches zh-CN i18n) ──────────

const warehouses = [
  { id: id(1), code: 'WH-SH-01', name: '上海主仓', address: '上海市浦东新区张江路100号', status: 'active', created_at: now, updated_at: now },
  { id: id(2), code: 'WH-SZ-01', name: '深圳配送中心', address: '深圳市南山区南山大道200号', status: 'active', created_at: now, updated_at: now },
]
const zones = [
  { id: id(10), warehouse_id: id(1), code: 'ZONE-RCV-01', name: '收货区A', zone_type: 'receiving', status: 'active', created_at: now, updated_at: now },
  { id: id(11), warehouse_id: id(1), code: 'ZONE-STO-01', name: '存储区A', zone_type: 'storage', status: 'active', created_at: now, updated_at: now },
  { id: id(12), warehouse_id: id(1), code: 'ZONE-PCK-01', name: '拣货区A', zone_type: 'picking', status: 'active', created_at: now, updated_at: now },
]
const skus = [
  { id: id(100), code: 'SKU-WIDGET-001', name: '标准型零件A', barcode: '6901234567890', base_unit: 'EA', category: '零件', status: 'active', created_at: now, updated_at: now },
  { id: id(101), code: 'SKU-BOLT-M8', name: 'M8六角螺栓', barcode: '6901234567891', base_unit: 'EA', category: '紧固件', status: 'active', created_at: now, updated_at: now },
  { id: id(102), code: 'SKU-GASKET-3MM', name: '3mm橡胶垫片', barcode: '6901234567892', base_unit: 'EA', category: '密封件', status: 'active', created_at: now, updated_at: now },
  { id: id(103), code: 'SKU-LABEL-TAG', name: 'RFID电子标签', barcode: '6901234567893', base_unit: 'ROLL', category: '标签', status: 'active', created_at: now, updated_at: now },
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
  { id: id(400), task_no: 'TASK-20260721-000001', task_type: 'putaway', warehouse_id: id(1), sku_id: id(100), status: 'in_progress', priority: 'high', assigned_to: '操作员-张', expected_qty: 200, actual_qty: 0, created_at: now },
  { id: id(401), task_no: 'TASK-20260721-000002', task_type: 'pick', warehouse_id: id(1), sku_id: id(102), status: 'pending', priority: 'normal', assigned_to: '', expected_qty: 50, actual_qty: 0, created_at: now },
  { id: id(402), task_no: 'TASK-20260721-000003', task_type: 'cycle_count', warehouse_id: id(2), sku_id: id(101), status: 'assigned', priority: 'low', assigned_to: '操作员-李', expected_qty: 500, actual_qty: 0, created_at: now },
]

const asns = [
  { id: id(500), asn_no: 'ASN-20260721-000001', warehouse_id: id(1), carrier: '顺丰速运', tracking_no: 'SF1234567890', expected_at: '2026-07-25T10:00:00Z', status: 'pending', created_at: now },
  { id: id(501), asn_no: 'ASN-20260721-000002', warehouse_id: id(1), carrier: '德邦物流', tracking_no: 'DP9876543210', expected_at: '2026-07-22T14:00:00Z', status: 'arrived', arrived_at: '2026-07-22T13:45:00Z', created_at: now },
  { id: id(502), asn_no: 'ASN-20260720-000001', warehouse_id: id(2), carrier: '中通快递', tracking_no: 'ZT5555666677', expected_at: '2026-07-21T08:00:00Z', status: 'receiving', arrived_at: '2026-07-21T07:30:00Z', created_at: now },
]
const asnLines: Record<string, { id: string; asn_id: string; sku_id: string; expected_qty: number; received_qty: number; batch_no: string; status: string }[]> = {
  [id(500)]: [
    { id: id(510), asn_id: id(500), sku_id: id(100), expected_qty: 200, received_qty: 0, batch_no: 'B2026-004', status: 'pending' },
    { id: id(511), asn_id: id(500), sku_id: id(101), expected_qty: 1000, received_qty: 0, batch_no: 'B2026-005', status: 'pending' },
  ],
  [id(501)]: [
    { id: id(520), asn_id: id(501), sku_id: id(102), expected_qty: 500, received_qty: 0, batch_no: 'B2026-006', status: 'pending' },
  ],
  [id(502)]: [
    { id: id(530), asn_id: id(502), sku_id: id(103), expected_qty: 300, received_qty: 150, batch_no: 'B2026-007', status: 'partial' },
    { id: id(531), asn_id: id(502), sku_id: id(100), expected_qty: 100, received_qty: 100, batch_no: 'B2026-008', status: 'received' },
  ],
}

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
      display_name: body.username || '演示管理员',
      role_names: ['admin'],
    })
  }),
  http.post('/api/v1/auth/refresh', async () => {
    await delay(100)
    return HttpResponse.json({ access_token: 'demo-access-token', refresh_token: 'demo-refresh-token', token_type: 'Bearer', expires_in: 900 })
  }),
  http.get('/api/v1/auth/me', async () => {
    return HttpResponse.json({ id: id(999), username: 'admin', display_name: '演示管理员', role_ids: [id(998)], role_names: ['admin'], created_at: now, updated_at: now })
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

  // ASNs
  http.get('/api/v1/asns', async ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const status = url.searchParams.get('status')
    const filtered = status ? asns.filter(a => a.status === status) : asns
    return HttpResponse.json(paginate(filtered, page, 10))
  }),
  http.get('/api/v1/asns/:id', async ({ params }) => {
    const a = asns.find(a => a.id === params.id)
    if (!a) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json({ ...a, lines: asnLines[a.id] ?? [] })
  }),
  http.post('/api/v1/asns', async ({ request }) => {
    await delay(200)
    const body = await request.json() as { warehouse_id: string; carrier?: string }
    const newAsn = {
      id: id(Date.now() % 100000),
      asn_no: `ASN-${new Date().toISOString().slice(0, 10).replace(/-/g, '')}-${String(Math.floor(Math.random() * 1000000)).padStart(6, '0')}`,
      warehouse_id: body.warehouse_id,
      carrier: body.carrier ?? '',
      tracking_no: '',
      expected_at: new Date().toISOString(),
      status: 'pending',
      created_at: new Date().toISOString(),
    }
    asns.unshift(newAsn)
    return HttpResponse.json(newAsn, { status: 201 })
  }),
  http.put('/api/v1/asns/:id/status', async ({ params, request }) => {
    const body = await request.json() as { status: string }
    const a = asns.find(a => a.id === params.id)
    if (!a) return new HttpResponse(null, { status: 404 })
    a.status = body.status
    if (body.status === 'arrived') (a as Record<string, unknown>).arrived_at = new Date().toISOString()
    return HttpResponse.json(a)
  }),
]
