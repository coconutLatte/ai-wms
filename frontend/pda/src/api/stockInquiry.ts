// Stock Inquiry API — barcode-based inventory lookup for PDA operators.
// Resolves a scanned barcode to a location or SKU, then returns current
// inventory levels (qty, available, reserved, batch_no, status).

import client from './client'

// ── Response Types ────────────────────────────────────────────────────────

export interface StockInquiryLocation {
  id: string
  code: string
  barcode: string
  location_type: string
  status: string
}

export interface StockInquirySKU {
  id: string
  code: string
  name: string
  barcode: string
  status: string
}

export interface StockInquiryInventory {
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

export interface StockInquiryResponse {
  barcode: string
  entity_type: string // "location" | "sku" | ""
  location?: StockInquiryLocation
  sku?: StockInquirySKU
  inventory: StockInquiryInventory[]
  total_qty: number
  total_reserved: number
  total_available: number
}

// ── API Functions ─────────────────────────────────────────────────────────

export async function inquireStock(barcode: string): Promise<StockInquiryResponse> {
  const { data } = await client.get<StockInquiryResponse>('/stock-inquiry', {
    params: { barcode },
  })
  return data
}
