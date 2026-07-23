// System configuration API client.
// Handles fetching and updating application-level settings (site name,
// default warehouse, inventory thresholds, pagination defaults, JWT TTL).

import client from './client'
import type { AppConfig, UpdateAppConfigRequest } from './types'

export function getSettings(): Promise<AppConfig> {
  return client.get('/settings').then((res) => res.data)
}

export function updateSettings(
  input: UpdateAppConfigRequest,
): Promise<AppConfig> {
  return client.put('/settings', input).then((res) => res.data)
}
