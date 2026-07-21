// Tasks list page — primary PDA screen for warehouse operators.
// Shows assigned tasks with status, type, and priority indicators.
// Fetches tasks from GET /api/v1/tasks with swipable filter tabs.
// All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import client from '@/api/client'
import type { Task, TaskStatus, ListResponse } from '@/api/types'

// ── Filter tabs ────────────────────────────────────────────────────────

type FilterKey = 'all' | 'pending' | 'in_progress' | 'completed'

export default function TasksPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [filter, setFilter] = useState<FilterKey>('all')

  // ── Localized maps ───────────────────────────────────────────────────

  const STATUS_LABELS: Record<TaskStatus, string> = {
    pending: t('task.statusPending'),
    assigned: t('task.statusAssigned'),
    in_progress: t('task.statusInProgress'),
    completed: t('task.statusCompleted'),
    cancelled: t('task.statusCancelled'),
    paused: t('task.statusPaused'),
    exception: t('task.statusException'),
  }

  const TYPE_LABELS: Record<string, string> = {
    receiving: t('task.typeReceiving'),
    putaway: t('task.typePutaway'),
    picking: t('task.typePicking'),
    cycle_count: t('task.typeCycleCount'),
    replenishment: t('task.typeReplenishment'),
    pick: t('task.typePicking'),
    replenish: t('task.typeReplenishment'),
    transfer: t('task.typeTransfer'),
    load: t('task.typeLoad'),
    unload: t('task.typeUnload'),
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

  const FILTERS: { key: FilterKey; label: string }[] = [
    { key: 'all', label: t('task.all') },
    { key: 'pending', label: t('task.pending') },
    { key: 'in_progress', label: t('task.inProgress') },
    { key: 'completed', label: t('task.done') },
  ]

  // ── Fetch tasks from API ─────────────────────────────────────────────

  const { data, isLoading, isError, error, refetch, isFetching } = useQuery<ListResponse<Task>>({
    queryKey: ['tasks'],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<Task>>('/tasks')
      return data
    },
    staleTime: 10_000,
  })

  const tasks = (data?.data ?? []).filter((task) => {
    if (filter === 'all') return true
    if (filter === 'in_progress') return task.status === 'in_progress' || task.status === 'assigned'
    return task.status === filter
  })

  // ── Callbacks ────────────────────────────────────────────────────────

  const handleRefresh = useCallback(() => {
    refetch()
  }, [refetch])

  const handleTaskClick = useCallback(
    (task: Task) => {
      navigate(`/tasks/${task.id}`)
    },
    [navigate],
  )

  // ── Render ───────────────────────────────────────────────────────────

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

      {/* Refresh indicator */}
      {isFetching && (
        <div className="pda-pull-indicator">{t('task.refreshing')}</div>
      )}

      {/* Refresh button */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
        <button
          onClick={handleRefresh}
          disabled={isFetching}
          style={{
            padding: '6px 14px',
            fontSize: 13,
            color: '#1677ff',
            background: 'transparent',
            border: '1px solid #1677ff',
            borderRadius: 6,
            cursor: isFetching ? 'not-allowed' : 'pointer',
          }}
        >
          {isFetching ? '...' : t('task.refresh')}
        </button>
      </div>

      {/* Loading state */}
      {isLoading && (
        <div className="pda-empty">
          <span className="empty-icon">{'⏳'}</span>
          <p className="empty-text">{t('task.loadingTasks')}</p>
        </div>
      )}

      {/* Error state */}
      {isError && (
        <div className="pda-empty">
          <span className="empty-icon">{'⚠️'}</span>
          <p className="empty-text">{t('task.failedToLoad')}</p>
          <p style={{ fontSize: 12, color: '#bfbfbf', marginTop: 8 }}>
            {(error as Error)?.message || t('task.unknownError')}
          </p>
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
            {t('task.retry')}
          </button>
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !isError && tasks.length === 0 && (
        <div className="pda-empty">
          <span className="empty-icon">{'\u{1F4CB}'}</span>
          <p className="empty-text">{t('task.noTasks')}</p>
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
            {t('task.refreshTasks')}
          </button>
        </div>
      )}

      {/* Task cards */}
      {!isLoading && !isError && tasks.map((task) => (
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
                  {TYPE_ICONS[task.task_type] || '\u{1F4CB}'} {TYPE_LABELS[task.task_type] || task.task_type}
                </span>
              </div>
            </div>
            <span className={`pda-status-badge ${task.status.replace(/_/g, '-')}`}>
              {STATUS_LABELS[task.status]}
            </span>
          </div>
          <div className="task-meta">
            <span>{t('task.qty')}: {task.expected_qty} {task.uom}</span>
            <span>{t('task.prio')}: {task.priority}</span>
          </div>
        </div>
      ))}
    </div>
  )
}
