// Task detail page — view and act on a single task.
// Shows task properties and provides action buttons for status transitions.
// P3-09+ will implement full task execution flows.

import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import type { Task, TaskType } from '@/api/types'

// ── Mock data ─────────────────────────────────────────────────────────

const MOCK_TASKS: Record<string, Task> = {
  '1': {
    id: '1',
    task_no: 'TASK-20260720-001',
    type: 'picking',
    status: 'assigned',
    warehouse_id: 'wh-001',
    assignee_id: 'op-001',
    reference_type: 'order',
    reference_id: 'ORD-001',
    priority: 1,
    created_at: '2026-07-20T08:00:00Z',
    updated_at: '2026-07-20T08:30:00Z',
  },
  '2': {
    id: '2',
    task_no: 'TASK-20260720-002',
    type: 'receiving',
    status: 'in_progress',
    warehouse_id: 'wh-001',
    assignee_id: 'op-001',
    reference_type: 'asn',
    reference_id: 'ASN-20260720-001',
    priority: 2,
    created_at: '2026-07-20T07:00:00Z',
    updated_at: '2026-07-20T07:45:00Z',
  },
  '3': {
    id: '3',
    task_no: 'TASK-20260720-003',
    type: 'putaway',
    status: 'pending',
    warehouse_id: 'wh-001',
    assignee_id: 'op-001',
    reference_type: 'asn',
    reference_id: 'ASN-20260719-003',
    priority: 1,
    created_at: '2026-07-20T06:30:00Z',
    updated_at: '2026-07-20T06:30:00Z',
  },
  '4': {
    id: '4',
    task_no: 'TASK-20260720-004',
    type: 'cycle_count',
    status: 'assigned',
    warehouse_id: 'wh-001',
    assignee_id: 'op-001',
    reference_type: 'location',
    reference_id: 'LOC-Z1-A01',
    priority: 3,
    created_at: '2026-07-20T05:00:00Z',
    updated_at: '2026-07-20T05:00:00Z',
  },
  '5': {
    id: '5',
    task_no: 'TASK-20260719-005',
    type: 'picking',
    status: 'completed',
    warehouse_id: 'wh-001',
    assignee_id: 'op-001',
    reference_type: 'order',
    reference_id: 'ORD-002',
    priority: 1,
    created_at: '2026-07-19T15:00:00Z',
    updated_at: '2026-07-19T15:45:00Z',
  },
}

// ── Display helpers ───────────────────────────────────────────────────

const STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  assigned: 'Assigned',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
}

const TYPE_LABELS: Record<string, string> = {
  receiving: 'Receiving',
  putaway: 'Putaway',
  picking: 'Picking',
  cycle_count: 'Cycle Count',
  replenishment: 'Replenishment',
}

const TYPE_ICONS: Record<string, string> = {
  receiving: '📥',
  putaway: '📦',
  picking: '🛒',
  cycle_count: '🔢',
  replenishment: '🔄',
}

// ── Available actions per current status ──────────────────────────────

function getAvailableActions(task: Task): { label: string; nextStatus: string }[] {
  switch (task.status) {
    case 'assigned':
      return [{ label: 'Start Task', nextStatus: 'in_progress' }]
    case 'in_progress':
      return [{ label: 'Complete Task', nextStatus: 'completed' }]
    case 'pending':
      return [{ label: 'Accept & Start', nextStatus: 'in_progress' }]
    default:
      return []
  }
}

export default function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const [submitting, setSubmitting] = useState(false)
  const [localStatus, setLocalStatus] = useState<string | null>(null)

  const task = taskId ? MOCK_TASKS[taskId] : null

  if (!task) {
    return (
      <div className="pda-empty">
        <span className="empty-icon">🔍</span>
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

  const currentStatus = localStatus || task.status
  const actions = getAvailableActions({ ...task, status: currentStatus as Task['status'] })

  const handleAction = async (nextStatus: string) => {
    setSubmitting(true)
    // P3-09+: replace with actual API call (PUT /api/v1/tasks/:id/status)
    await new Promise((resolve) => setTimeout(resolve, 500))
    setLocalStatus(nextStatus)
    setSubmitting(false)
  }

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
              {TYPE_ICONS[task.type]} {TYPE_LABELS[task.type]}
            </span>
          </div>
          <span className={`pda-status-badge ${currentStatus.replace(/_/g, '-')}`}>
            {STATUS_LABELS[currentStatus]}
          </span>
        </div>

        {/* Detail rows */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <DetailRow label="Type" value={TYPE_LABELS[task.type]} />
          <DetailRow label="Reference" value={`${task.reference_type}: ${task.reference_id}`} />
          <DetailRow label="Priority" value={`P${task.priority}`} />
          <DetailRow label="Warehouse" value={task.warehouse_id} />
          <DetailRow label="Assigned To" value={task.assignee_id} />
          <DetailRow label="Created" value={new Date(task.created_at).toLocaleString()} />
          <DetailRow label="Updated" value={new Date(task.updated_at).toLocaleString()} />
        </div>
      </div>

      {/* Action buttons */}
      {actions.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {actions.map((action) => (
            <button
              key={action.nextStatus}
              onClick={() => handleAction(action.nextStatus)}
              disabled={submitting}
              style={{
                width: '100%',
                padding: '16px 0',
                fontSize: 16,
                fontWeight: 600,
                color: '#fff',
                background: submitting ? '#91caff' : '#1677ff',
                border: 'none',
                borderRadius: 10,
                cursor: submitting ? 'not-allowed' : 'pointer',
                transition: 'background 0.2s',
              }}
            >
              {submitting ? 'Processing...' : action.label}
            </button>
          ))}
        </div>
      )}

      {actions.length === 0 && currentStatus === 'completed' && (
        <div
          style={{
            padding: '16px',
            background: '#f6ffed',
            border: '1px solid #b7eb8f',
            borderRadius: 8,
            textAlign: 'center',
            color: '#389e0d',
            fontSize: 14,
          }}
        >
          ✓ This task has been completed
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
      <span style={{ fontSize: 14, fontWeight: 500, color: '#262626', textAlign: 'right' }}>{value}</span>
    </div>
  )
}
