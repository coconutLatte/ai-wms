// Dashboard API module — admin dashboard aggregated data.
// Calls GET /api/v1/dashboard for the comprehensive admin landing page.

import client from './client'
import type { AdminDashboardResponse } from './types'

export async function getDashboard(): Promise<AdminDashboardResponse> {
  const { data } = await client.get<AdminDashboardResponse>('/dashboard')
  return data
}
