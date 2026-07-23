// Inventory API module — inventory transaction history.
// Calls GET /api/v1/inventory-transactions for global transaction queries.

import client from './client'
import type { InventoryTransaction, ListResponse } from './types'

export interface TransactionQueryParams {
  sku_id?: string
  warehouse_id?: string
  type?: string
  date_from?: string
  date_to?: string
  page?: number
  page_size?: number
}

export async function getTransactions(
  params: TransactionQueryParams = {},
): Promise<ListResponse<InventoryTransaction>> {
  const { data } = await client.get<ListResponse<InventoryTransaction>>(
    '/inventory-transactions',
    { params },
  )
  return data
}
