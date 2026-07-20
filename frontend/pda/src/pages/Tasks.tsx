// Tasks list page — primary PDA screen for warehouse operators.
// Shows assigned tasks with status, type, and priority indicators.
// P3-09 will implement full API integration with pull-to-refresh.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Task, TaskStatus, TaskType } from '@/api/types'

// ── Mock data for scaffold ────────────────────────────────────────────
// P3-09 will replace with real API calls (GET /api/v1/tasks).

const MOCK_TASKS: Task[] = [
  {
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
  {
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
  {
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
  {
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
  {
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
]

// ── Status label map ──────────────────────────────────────────────────

const STATUS_LABELS: Record<TaskStatus, string> = {
  pending: 'Pending',
  assigned: 'Assigned',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
}

// ── Type label map ────────────────────────────────────────────────────

const TYPE_LABELS: Record<TaskType, string> = {
  receiving: 'Receiving',
  putaway: 'Putaway',
  picking: 'Picking',
  cycle_count: 'Cycle Count',
  replenishment: 'Replenishment',
}

// ── Type icons ────────────────────────────────────────────────────────

const TYPE_ICONS: Record<TaskType, string> = {
  receiving: '📥',
  putaway: '📦',
  picking: '🛒',
  cycle_count: '🔢',
  replenishment: '🔄',
}

// ── Filter tabs ───────────────────────────────────────────────────────

type FilterKey = 'all' | 'pending' | 'in_progress' | 'completed'

const FILTERS: { key: FilterKey; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'pending', label: 'Pending' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'completed', label: 'Done' },
]

export default function TasksPage() {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<FilterKey>('all')
  const [refreshing, setRefreshing] = useState(false)

  // Filter tasks
  const tasks = MOCK_TASKS.filter((task) => {
    if (filter === 'all') return true
    if (filter === 'in_progress') return task.status === 'in_progress' || task.status === 'assigned'
    return task.status === filter
  })

  // Pull-to-refresh simulation
  const handleRefresh = useCallback(async () => {
    setRefreshing(true)
    // P3-09: replace with actual API re-fetch
    await new Promise((resolve) => setTimeout(resolve, 1000))
    setRefreshing(false)
  }, [])

  const handleTaskClick = useCallback(
    (task: Task) => {
      navigate(`/tasks/${task.id}`)
    },
    [navigate],
  )

  return (
    <div>
      {/* Filter tabs */}
      <div
        style={{
          display: 'flex',
          gap: 8,
          marginBottom: 16,
          overflowX: 'auto',
          paddingBottom: 4,
        }}
      >
        {FILTERS.map((f) => (
          <button
            key={f.key}
            onClick={() => setFilter(f.key)}
            style={{
              flexShrink: 0,
              padding: '8px 16px',
              fontSize: 13,
              fontWeight: filter === f.key ? 600 : 400,
              color: filter === f.key ? '#fff' : '#595959',
              background: filter === f.key ? '#1677ff' : '#f0f0f0',
              border: 'none',
              borderRadius: 20,
              cursor: 'pointer',
              transition: 'all 0.2s',
            }}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Pull indicator */}
      {refreshing && (
        <div className="pda-pull-indicator">Refreshing...</div>
      )}

      {/* Refresh button */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          style={{
            padding: '6px 14px',
            fontSize: 13,
            color: '#1677ff',
            background: 'transparent',
            border: '1px solid #1677ff',
            borderRadius: 6,
            cursor: refreshing ? 'not-allowed' : 'pointer',
          }}
        >
          {refreshing ? '...' : '↻ Refresh'}
        </button>
      </div>

      {/* Task cards */}
      {tasks.length === 0 ? (
        <div className="pda-empty">
          <span className="empty-icon">📋</span>
          <p className="empty-text">No tasks found</p>
          <button
            onClick={handleRefresh}
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
            Refresh
          </button>
        </div>
      ) : (
        tasks.map((task) => (
          <div
            key={task.id}
            className="pda-task-card"
            onClick={() => handleTaskClick(task)}
          >
            <div className="task-header">
              <div>
                <div className="task-no">{task.task_no}</div>
                <div style={{ marginTop: 4 }}>
                  <span className="pda-type-badge">
                    {TYPE_ICONS[task.type]} {TYPE_LABELS[task.type]}
                  </span>
                </div>
              </div>
              <span className={`pda-status-badge ${task.status.replace(/_/g, '-')}`}>
                {STATUS_LABELS[task.status]}
              </span>
            </div>
            <div className="task-meta">
              <span>Ref: {task.reference_id}</span>
              <span>P{task.priority}</span>
            </div>
          </div>
        ))
      )}
    </div>
  )
}
