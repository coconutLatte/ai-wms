// Order API module — fetch orders for the PDA order lookup page.
// Uses the shared axios client with JWT auth interceptors.

import client from './client'
import type { Order, OrderSummary, ListResponse } from './types'

export async function fetchOrders(params?: {
  order_no?: string
  warehouse_id?: string
  order_type?: string
  status?: string
  page?: number
  page_size?: number
}): Promise<ListResponse<OrderSummary>> {
  const { data } = await client.get<ListResponse<OrderSummary>>('/orders', { params })
  return data
}

export async function fetchOrder(id: string): Promise<Order> {
  const { data } = await client.get<Order>(`/orders/${id}`)
  return data
}
