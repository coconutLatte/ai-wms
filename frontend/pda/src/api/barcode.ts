// Barcode lookup API — resolves scanned barcodes to location, SKU, or task entities.
// Used by the PDA scanner page to provide result navigation and info display.

import client from './client'

// ── Response Types ────────────────────────────────────────────────────────

export interface BarcodeLocation {
  type: 'location'
  id: string
  code: string
  barcode: string
  zone_id: string
  warehouse_id: string
  location_type: string
  status: string
}

export interface BarcodeSKU {
  type: 'sku'
  id: string
  code: string
  name: string
  barcode: string
  category: string
  status: string
  uom: {
    base_unit: string
    pack_unit: string
    pack_qty: number
  }
}

export type BarcodeResult = BarcodeLocation | BarcodeSKU

// ── Lookup Functions ──────────────────────────────────────────────────────

// lookupBarcode resolves a scanned barcode to a location, SKU, or returns null.
// It tries location-by-barcode first, then SKU-by-code.
export async function lookupBarcode(barcode: string): Promise<BarcodeResult | null> {
  // Try location barcode lookup first (GET /api/v1/locations?barcode=X)
  try {
    const { data: loc } = await client.get<BarcodeLocation>('/locations', {
      params: { barcode },
    })
    if (loc) {
      return { ...loc, type: 'location' }
    }
  } catch {
    // Not a location barcode — try SKU next
  }

  // Try SKU code lookup (GET /api/v1/skus?code=X)
  try {
    const { data: skuResp } = await client.get<{
      data?: BarcodeSKU[]
    } & BarcodeSKU>('/skus', {
      params: { code: barcode },
    })
    // SKU lookup by code returns either a single SKU or a list wrapper
    const sku = 'data' in skuResp && Array.isArray(skuResp.data) && skuResp.data.length === 1
      ? skuResp.data[0]
      : skuResp as unknown as BarcodeSKU
    if (sku && sku.id) {
      return { ...sku, type: 'sku' }
    }
  } catch {
    // Not recognized
  }

  return null
}
