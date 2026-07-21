// Task detail page — view and act on a single task.
// Shows task properties and provides action buttons for status transitions.
// Fetches task data from GET /api/v1/tasks/{id}.

import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import client from '@/api/client'
import type { Task, TaskStatus } from '@/api/types'

// ── Display helpers ───────────────────────────────────────────────────

const STATUS_LABELS: Record<TaskStatus, string> = {
  pending: 'Pending',
  assigned: 'Assigned',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
  paused: 'Paused',
  exception: 'Exception',
}

const TYPE_LABELS: Record<string, string> = {
  receiving: 'Receiving',
  putaway: 'Putaway',
  picking: 'Picking',
  cycle_count: 'Cycle Count',
  replenishment: 'Replenishment',
  pick: 'Picking',
  replenish: 'Replenish',
  transfer: 'Transfer',
  load: 'Load',
  unload: 'Unload',
}

const TYPE_ICONS: Record<string, string> = {
  receiving: '\u{1F4E5}',
  putaway: '\u{1F4E6}',
  picking: '\u{1F6D2}',
  cycle_count: '\u{1F522}',
  replenishment: '\u{1F504}',
  pick: '\u{1F6D2}',
  replenish: '\u{1F504}',
  transfer: '\u{2194}\u{FE0F}',
  load: '\u{2B06}\u{FE0F}',
  unload: '\u{2B07}\u{FE0F}',
}

// ── Available actions per current status ──────────────────────────────

function getAvailableActions(status: TaskStatus): { label: string; nextStatus: TaskStatus }[] {
  switch (status) {
    case 'assigned':
      return [{ label: 'Start Task', nextStatus: 'in_progress' }]
    case 'in_progress':
      return [{ label: 'Complete Task', nextStatus: 'completed' }]
    case 'pending':
      return [{ label: 'Accept & Start', nextStatus: 'in_progress' }]
    case 'paused':
      return [{ label: 'Resume', nextStatus: 'in_progress' }]
    default:
      return []
  }
}

export default function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // ── Fetch task ──────────────────────────────────────────────────────

  const { data: task, isLoading, isError } = useQuery<Task>({
    queryKey: ['task', taskId],
    queryFn: async () => {
      const { data } = await client.get<Task>(`/tasks/${taskId}`)
      return data
    },
    enabled: !!taskId,
  })

  // ── Update status mutation ──────────────────────────────────────────

  const statusMutation = useMutation({
    mutationFn: async (status: TaskStatus) => {
      await client.put(`/tasks/${taskId}/status`, { status })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
  })

  const handleAction = async (nextStatus: TaskStatus) => {
    statusMutation.mutate(nextStatus)
  }

  // ── Loading state ───────────────────────────────────────────────────

  if (isLoading) {
    return (
      <div className="pda-empty">
        <span className="empty-icon">{'⏳'}</span>
        <p className="empty-text">Loading task...</p>
      </div>
    )
  }

  // ── Not found / error state ─────────────────────────────────────────

  if (isError || !task) {
    return (
      <div className="pda-empty">
        <span className="empty-icon">{'🔍'}</span>
        <p className="empty-text">Task not found</p>
        <button
          onClick={() => navigate('/tasks')}
          style={{
            marginTop: 16,
            padding: '10px 24px',
            fontSize: 14,
            color: '#1677ff',
            background: '#fff',
            border: '1px solid #1677ff',
            borderRadius: 8,
            cursor: 'pointer',
          }}
        >
          Back to Tasks
        </button>
      </div>
    )
  }

  const currentStatus: TaskStatus = task.status
  const actions = getAvailableActions(currentStatus)
  const isSubmitting = statusMutation.isPending

  return (
    <div>
      {/* Back button */}
      <button
        onClick={() => navigate('/tasks')}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 4,
          padding: 0,
          marginBottom: 16,
          fontSize: 14,
          color: '#1677ff',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
        }}
      >
        ← Back to Tasks
      </button>

      {/* Task card */}
      <div
        style={{
          background: '#fff',
          borderRadius: 12,
          padding: 20,
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
          marginBottom: 16,
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
          <div>
            <h2 style={{ fontSize: 20, fontWeight: 700, color: '#262626', marginBottom: 4 }}>
              {task.task_no}
            </h2>
            <span className="pda-type-badge">
              {TYPE_ICONS[task.task_type] || '\u{1F4CB}'} {TYPE_LABELS[task.task_type] || task.task_type}
            </span>
          </div>
          <span className={`pda-status-badge ${currentStatus.replace(/_/g, '-')}`}>
            {STATUS_LABELS[currentStatus]}
          </span>
        </div>

        {/* Detail rows */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <DetailRow label="Type" value={TYPE_LABELS[task.task_type] || task.task_type} />
          <DetailRow label="Priority" value={task.priority} />
          <DetailRow label="Warehouse" value={task.warehouse_id} />
          <DetailRow label="Assigned To" value={task.assigned_to || '—'} />
          <DetailRow label="SKU" value={task.sku_id} />
          <DetailRow label="Expected Qty" value={`${task.expected_qty} ${task.uom}`} />
          <DetailRow label="Actual Qty" value={`${task.actual_qty} ${task.uom}`} />
          {task.batch_no && <DetailRow label="Batch" value={task.batch_no} />}
          {task.instructions && <DetailRow label="Instructions" value={task.instructions} />}
          <DetailRow label="Created" value={new Date(task.created_at).toLocaleString()} />
          {task.started_at && <DetailRow label="Started" value={new Date(task.started_at).toLocaleString()} />}
          {task.completed_at && <DetailRow label="Completed" value={new Date(task.completed_at).toLocaleString()} />}
        </div>
      </div>

      {/* Error message from mutation */}
      {statusMutation.isError && (
        <div
          style={{
            color: '#cf1322',
            fontSize: 13,
            marginBottom: 12,
            textAlign: 'center',
            padding: '8px 12px',
            background: '#fff1f0',
            borderRadius: 6,
          }}
        >
          {(statusMutation.error as Error)?.message || 'Failed to update task status'}
        </div>
      )}

      {/* Action buttons */}
      {actions.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {actions.map((action) => (
            <button
              key={action.nextStatus}
              onClick={() => handleAction(action.nextStatus)}
              disabled={isSubmitting}
              style={{
                width: '100%',
                padding: '16px 0',
                fontSize: 16,
                fontWeight: 600,
                color: '#fff',
                background: isSubmitting ? '#91caff' : '#1677ff',
                border: 'none',
                borderRadius: 10,
                cursor: isSubmitting ? 'not-allowed' : 'pointer',
                transition: 'background 0.2s',
              }}
            >
              {isSubmitting ? 'Processing...' : action.label}
            </button>
          ))}
        </div>
      )}

      {actions.length === 0 && (currentStatus === 'completed' || currentStatus === 'cancelled') && (
        <div
          style={{
            padding: '16px',
            background: currentStatus === 'completed' ? '#f6ffed' : '#fff1f0',
            border: currentStatus === 'completed' ? '1px solid #b7eb8f' : '1px solid #ffa39e',
            borderRadius: 8,
            textAlign: 'center',
            color: currentStatus === 'completed' ? '#389e0d' : '#cf1322',
            fontSize: 14,
          }}
        >
          {currentStatus === 'completed' ? '✓ This task has been completed' : '✗ This task has been cancelled'}
        </div>
      )}
    </div>
  )
}

// ── Detail row helper ─────────────────────────────────────────────────

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <span style={{ fontSize: 13, color: '#8c8c8c' }}>{label}</span>
      <span style={{ fontSize: 14, fontWeight: 500, color: '#262626', textAlign: 'right', maxWidth: '60%', overflow: 'hidden', textOverflow: 'ellipsis' }}>{value}</span>
    </div>
  )
}
