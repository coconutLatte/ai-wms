// Cycle Count API — PDA operator workflow for physical inventory counting.
// Operators start counts, scan location barcodes, enter counted quantities per SKU/batch,
// and submit the variance report for review.

import client from './client'

// ── Response Types ────────────────────────────────────────────────────────

export interface CycleCountLine {
  id: string
  sku_id: string
  location_id: string
  batch_no: string
  system_qty: number
  counted_qty?: number
  variance?: number
  status: CycleCountLineStatus
  counted_at?: string
  created_at: string
}

export type CycleCountLineStatus = 'pending' | 'counted' | 'reviewed'

export type CycleCountStatus = 'draft' | 'in_progress' | 'pending_review' | 'approved' | 'adjusted' | 'cancelled'

export interface CycleCount {
  id: string
  count_no: string
  warehouse_id: string
  location_id?: string
  zone_id?: string
  status: CycleCountStatus
  counted_by: string
  notes: string
  total_lines: number
  matched_lines: number
  created_at: string
  started_at?: string
  completed_at?: string
  approved_at?: string
  approved_by?: string
  lines?: CycleCountLine[]
}

export interface StartCycleCountRequest {
  warehouse_id: string
  location_id?: string
  zone_id?: string
  counted_by?: string
  notes?: string
}

export interface SubmitLineRequest {
  line_id: string
  counted_qty: number
}

export interface FinalizeCountRequest {
  notes?: string
}

export interface ApproveCountRequest {
  approved_by?: string
  action: 'approve' | 'adjust'
}

// ── API Functions ─────────────────────────────────────────────────────────

export async function startCycleCount(input: StartCycleCountRequest): Promise<CycleCount> {
  const { data } = await client.post<CycleCount>('/cycle-counts', input)
  return data
}

export async function getCycleCount(id: string): Promise<CycleCount> {
  const { data } = await client.get<CycleCount>(`/cycle-counts/${id}`)
  return data
}

export async function listCycleCounts(params?: {
  warehouse_id?: string
  status?: string
}): Promise<CycleCount[]> {
  const { data } = await client.get<{ data: CycleCount[] }>('/cycle-counts', { params })
  return data.data
}

export async function submitLine(cycleCountId: string, input: SubmitLineRequest): Promise<CycleCountLine> {
  const { data } = await client.post<CycleCountLine>(`/cycle-counts/${cycleCountId}/lines`, input)
  return data
}

export async function finalizeCount(id: string, input?: FinalizeCountRequest): Promise<CycleCount> {
  const { data } = await client.post<CycleCount>(`/cycle-counts/${id}/finalize`, input || {})
  return data
}

export async function approveCount(id: string, input: ApproveCountRequest): Promise<CycleCount> {
  const { data } = await client.put<CycleCount>(`/cycle-counts/${id}/approve`, input)
  return data
}

export async function cancelCycleCount(id: string): Promise<CycleCount> {
  const { data } = await client.put<CycleCount>(`/cycle-counts/${id}/cancel`)
  return data
}
