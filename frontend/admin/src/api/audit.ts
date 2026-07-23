// Audit log API module — system audit log viewer.
// Calls GET /api/v1/audit-logs for paginated audit log queries.

import client from './client'
import type { AuditLog, ListResponse } from './types'

export interface AuditLogQueryParams {
  user_id?: string
  action?: string
  resource?: string
  date_from?: string
  date_to?: string
  page?: number
  page_size?: number
}

export async function getAuditLogs(
  params: AuditLogQueryParams = {},
): Promise<ListResponse<AuditLog>> {
  const { data } = await client.get<ListResponse<AuditLog>>('/audit-logs', { params })
  return data
}
