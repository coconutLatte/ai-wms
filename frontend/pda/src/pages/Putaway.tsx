// Putaway page — putaway workflow for warehouse operators.
// Operator scans/enters a location barcode to identify the target putaway
// location, then selects a pending putaway task, enters the actual quantity
// put away, and confirms completion. All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import client from '@/api/client'
import BarcodeScanner from '@/components/BarcodeScanner'
import type { Task, TaskStatus, TaskPriority, ListResponse } from '@/api/types'

// ── Task status helpers ────────────────────────────────────────────────

function getTaskStatusLabelKey(status: TaskStatus): string {
  const map: Record<TaskStatus, string> = {
    pending: 'task.statusPending',
    assigned: 'task.statusAssigned',
    in_progress: 'task.statusInProgress',
    completed: 'task.statusCompleted',
    cancelled: 'task.statusCancelled',
    paused: 'task.statusPaused',
    exception: 'task.statusException',
  }
  return map[status] || 'task.statusPending'
}

function getTaskStatusColor(status: TaskStatus): string {
  const map: Record<TaskStatus, string> = {
    pending: '#d48806',
    assigned: '#1677ff',
    in_progress: '#389e0d',
    completed: '#595959',
    cancelled: '#cf1322',
    paused: '#fa8c16',
    exception: '#cf1322',
  }
  return map[status] || '#595959'
}

function getPriorityLabelKey(priority: TaskPriority): string {
  const map: Record<TaskPriority, string> = {
    low: 'putaway.priorityLow',
    normal: 'putaway.priorityNormal',
    high: 'putaway.priorityHigh',
    urgent: 'putaway.priorityUrgent',
  }
  return map[priority] || 'putaway.priorityNormal'
}

function getPriorityColor(priority: TaskPriority): string {
  const map: Record<TaskPriority, string> = {
    low: '#8c8c8c',
    normal: '#1677ff',
    high: '#fa8c16',
    urgent: '#cf1322',
  }
  return map[priority] || '#8c8c8c'
}

export default function PutawayPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ────────────────────────────────────────────────────────────

  const [locationBarcode, setLocationBarcode] = useState('')
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)
  const [actualQty, setActualQty] = useState('')
  const [isCompleting, setIsCompleting] = useState(false)

  // ── Location lookup by barcode ───────────────────────────────────────

  const {
    data: locationData,
    isLoading: isLookingUpLocation,
    isError: isLocationError,
    error: locationError,
  } = useQuery<{ data?: { id: string; code: string; barcode: string; status: string; location_type: string; zone_id: string; warehouse_id: string }[] } & { id: string; code: string; barcode: string; status: string; location_type: string; zone_id: string; warehouse_id: string }>({
    queryKey: ['locationBarcode', locationBarcode],
    queryFn: async () => {
      const { data } = await client.get('/locations', {
        params: { barcode: locationBarcode },
      })
      return data
    },
    enabled: locationBarcode.length > 0,
    retry: false,
  })

  // Extract single location from response
  const location = (() => {
    if (!locationData) return null
    if ('data' in locationData && Array.isArray(locationData.data) && locationData.data.length === 1) {
      return locationData.data[0]
    }
    if (locationData.id && locationData.barcode) return locationData
    return null
  })()

  // ── Fetch pending putaway tasks ──────────────────────────────────────

  const {
    data: tasksList,
    isLoading: isLoadingTasks,
    isError: isTasksError,
    error: tasksError,
    refetch: refetchTasks,
  } = useQuery<ListResponse<Task>>({
    queryKey: ['putawayTasks'],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<Task>>('/tasks', {
        params: {
          task_type: 'putaway',
          status: 'pending',
          page_size: 50,
        },
      })
      return data
    },
    enabled: !!location,
    retry: false,
  })

  const pendingTasks = tasksList?.data ?? []

  // ── Fetch selected task detail ───────────────────────────────────────

  const selectedTask = pendingTasks.find((t) => t.id === selectedTaskId) ?? null

  // ── Handle barcode scan / manual input ───────────────────────────────

  const handleScan = useCallback((barcode: string) => {
    setLocationBarcode(barcode)
    setSelectedTaskId(null)
    setActualQty('')
    setIsCompleting(false)
  }, [])

  // ── Handle clear / rescan ────────────────────────────────────────────

  const handleClear = useCallback(() => {
    setLocationBarcode('')
    setSelectedTaskId(null)
    setActualQty('')
    setIsCompleting(false)
    queryClient.removeQueries({ queryKey: ['locationBarcode', locationBarcode] })
    queryClient.removeQueries({ queryKey: ['putawayTasks'] })
  }, [locationBarcode, queryClient])

  // ── Start task mutation (pending → in_progress) ──────────────────────

  const startTaskMutation = useMutation({
    mutationFn: async (taskId: string) => {
      await client.put(`/tasks/${taskId}/status`, { status: 'in_progress' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['putawayTasks'] })
    },
  })

  const handleSelectTask = async (task: Task) => {
    setSelectedTaskId(task.id)
    setActualQty(task.expected_qty.toString())
    try {
      await startTaskMutation.mutateAsync(task.id)
    } catch {
      // Error handled by mutation state
    }
  }

  // ── Complete task mutation ───────────────────────────────────────────

  const completeTaskMutation = useMutation({
    mutationFn: async ({ taskId, qty, toLocationId }: { taskId: string; qty: number; toLocationId: string }) => {
      await client.post(`/tasks/${taskId}/complete`, {
        actual_qty: qty,
        to_location_id: toLocationId,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['putawayTasks'] })
    },
  })

  const handleComplete = async () => {
    if (!selectedTask || !location) return
    const qty = parseFloat(actualQty)
    if (isNaN(qty) || qty <= 0) return

    setIsCompleting(true)
    try {
      await completeTaskMutation.mutateAsync({
        taskId: selectedTask.id,
        qty,
        toLocationId: location.id,
      })
      setSelectedTaskId(null)
      setActualQty('')
    } catch {
      // Error handled by mutation state
    } finally {
      setIsCompleting(false)
    }
  }

  // ── Handle back from task detail to task list ────────────────────────

  const handleBackToTaskList = useCallback(() => {
    setSelectedTaskId(null)
    setActualQty('')
  }, [])

  // ── Handle scan another location after completion ────────────────────

  const handleScanAnother = useCallback(() => {
    queryClient.removeQueries({ queryKey: ['locationBarcode', locationBarcode] })
    queryClient.removeQueries({ queryKey: ['putawayTasks'] })
    setLocationBarcode('')
    setSelectedTaskId(null)
    setActualQty('')
    setIsCompleting(false)
  }, [locationBarcode, queryClient])

  // ── Qty validation ───────────────────────────────────────────────────

  let qtyError: string | null = null
  if (selectedTask && actualQty) {
    const parsed = parseFloat(actualQty)
    if (!isNaN(parsed)) {
      if (parsed > selectedTask.expected_qty) {
        qtyError = t('putaway.qtyExceedsExpected', {
          total: actualQty,
          expected: selectedTask.expected_qty.toString(),
        })
      } else if (parsed <= 0) {
        qtyError = t('putaway.qtyMustBePositive')
      }
    }
  }

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      {/* Location lookup section — shown when no location scanned yet */}
      {!location && (
        <>
          <div style={{ marginBottom: 16 }}>
            <h3
              style={{
                fontSize: 15,
                fontWeight: 600,
                color: '#595959',
                marginBottom: 10,
              }}
            >
              {t('putaway.title')}
            </h3>
            <BarcodeScanner
              onScan={handleScan}
              placeholder={t('putaway.scanLocationPlaceholder')}
            />
          </div>

          {/* Lookup in progress */}
          {isLookingUpLocation && (
            <div
              style={{
                padding: '16px',
                textAlign: 'center',
                color: '#1677ff',
                fontSize: 14,
              }}
            >
              {'\u{1F50D}'} {t('putaway.lookingUpLocation')}
            </div>
          )}

          {/* Location not found or error */}
          {(isLocationError ||
            (!isLookingUpLocation && locationBarcode && !location)) && (
            <div
              style={{
                marginBottom: 16,
                padding: '14px 18px',
                background: '#fff1f0',
                border: '1px solid #ffa39e',
                borderRadius: 10,
                fontSize: 14,
                color: '#cf1322',
              }}
            >
              {locationError
                ? t('putaway.lookupFailed', { error: (locationError as Error)?.message })
                : t('putaway.locationNotFound', { barcode: locationBarcode })}
            </div>
          )}
        </>
      )}

      {/* Location confirmed + task selection / execution */}
      {location && (
        <>
          {/* Back / clear button */}
          <button
            onClick={handleClear}
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
            {t('putaway.backToScan')}
          </button>

          {/* Location info card */}
          {!selectedTask && (
            <div
              style={{
                background: '#fff',
                borderRadius: 12,
                padding: 20,
                boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                marginBottom: 16,
              }}
            >
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'flex-start',
                  marginBottom: 16,
                }}
              >
                <div>
                  <h2
                    style={{
                      fontSize: 18,
                      fontWeight: 700,
                      color: '#262626',
                      marginBottom: 4,
                    }}
                  >
                    {t('putaway.targetLocation')}
                  </h2>
                </div>
                <span
                  style={{
                    display: 'inline-block',
                    padding: '4px 10px',
                    fontSize: 12,
                    fontWeight: 600,
                    color: '#fff',
                    background: location.status === 'active' ? '#389e0d' : '#fa8c16',
                    borderRadius: 4,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {location.status}
                </span>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <DetailRow label={t('putaway.locationCode')} value={location.code || '—'} />
                <DetailRow
                  label={t('putaway.locationBarcode')}
                  value={location.barcode || '—'}
                />
                <DetailRow label={t('putaway.locationType')} value={location.location_type || '—'} />
              </div>
            </div>
          )}

          {/* Error states */}
          {isLoadingTasks && (
            <div className="pda-empty">
              <span className="empty-icon">{'⏳'}</span>
              <p className="empty-text">{t('putaway.loadingTasks')}</p>
            </div>
          )}

          {isTasksError && (
            <div className="pda-empty">
              <span className="empty-icon">{'⚠️'}</span>
              <p className="empty-text">
                {t('putaway.tasksLoadFailed', {
                  error: (tasksError as Error)?.message,
                })}
              </p>
              <button
                onClick={() => refetchTasks()}
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

          {/* Error from mutations */}
          {startTaskMutation.isError && (
            <div
              style={{
                marginBottom: 12,
                padding: '10px 14px',
                background: '#fff1f0',
                border: '1px solid #ffa39e',
                borderRadius: 8,
                fontSize: 13,
                color: '#cf1322',
                textAlign: 'center',
              }}
            >
              {t('putaway.startTaskError', {
                error: (startTaskMutation.error as Error)?.message,
              })}
            </div>
          )}

          {completeTaskMutation.isError && (
            <div
              style={{
                marginBottom: 12,
                padding: '10px 14px',
                background: '#fff1f0',
                border: '1px solid #ffa39e',
                borderRadius: 8,
                fontSize: 13,
                color: '#cf1322',
                textAlign: 'center',
              }}
            >
              {t('putaway.completeTaskError', {
                error: (completeTaskMutation.error as Error)?.message,
              })}
            </div>
          )}

          {/* Task list — shown when location is found but no task selected yet */}
          {!selectedTask && !isLoadingTasks && !isTasksError && (
            <>
              {pendingTasks.length === 0 ? (
                <div
                  style={{
                    padding: '32px 16px',
                    textAlign: 'center',
                    background: '#fff',
                    borderRadius: 12,
                    boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                  }}
                >
                  <div style={{ fontSize: 40, marginBottom: 12 }}>{'✅'}</div>
                  <div
                    style={{
                      fontSize: 15,
                      fontWeight: 600,
                      color: '#389e0d',
                      marginBottom: 8,
                    }}
                  >
                    {t('putaway.noPendingTasks')}
                  </div>
                  <div style={{ fontSize: 13, color: '#8c8c8c', marginBottom: 20 }}>
                    {t('putaway.noPendingTasksDesc')}
                  </div>
                  <button
                    onClick={handleScanAnother}
                    style={{
                      padding: '12px 24px',
                      fontSize: 14,
                      fontWeight: 600,
                      color: '#fff',
                      background: '#1677ff',
                      border: 'none',
                      borderRadius: 8,
                      cursor: 'pointer',
                    }}
                  >
                    {t('putaway.scanAnother')}
                  </button>
                </div>
              ) : (
                <>
                  <div
                    style={{
                      fontSize: 14,
                      fontWeight: 600,
                      color: '#595959',
                      marginBottom: 10,
                    }}
                  >
                    {t('putaway.selectTask', { count: pendingTasks.length })}
                  </div>

                  {/* Task cards */}
                  {pendingTasks.map((task) => (
                    <div
                      key={task.id}
                      onClick={() => handleSelectTask(task)}
                      style={{
                        background: '#fff',
                        borderRadius: 12,
                        padding: 16,
                        boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                        marginBottom: 12,
                        cursor: 'pointer',
                        transition: 'box-shadow 0.2s',
                        border: '2px solid transparent',
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.12)'
                        e.currentTarget.style.borderColor = '#1677ff'
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.08)'
                        e.currentTarget.style.borderColor = 'transparent'
                      }}
                    >
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'flex-start',
                          marginBottom: 10,
                        }}
                      >
                        <div>
                          <div
                            style={{
                              fontSize: 15,
                              fontWeight: 600,
                              color: '#262626',
                              marginBottom: 2,
                            }}
                          >
                            {task.task_no}
                          </div>
                          <div style={{ fontSize: 12, color: '#8c8c8c', fontFamily: 'monospace' }}>
                            {t('task.sku')}: {task.sku_id.substring(0, 8)}...
                          </div>
                        </div>
                        <span
                          style={{
                            display: 'inline-block',
                            padding: '3px 8px',
                            fontSize: 11,
                            fontWeight: 600,
                            color: getPriorityColor(task.priority as TaskPriority),
                            background: '#fff',
                            border: `1px solid ${getPriorityColor(task.priority as TaskPriority)}`,
                            borderRadius: 4,
                            whiteSpace: 'nowrap',
                          }}
                        >
                          {t(getPriorityLabelKey(task.priority as TaskPriority))}
                        </span>
                      </div>

                      <div
                        style={{
                          display: 'grid',
                          gridTemplateColumns: '1fr 1fr',
                          gap: 6,
                          fontSize: 13,
                          color: '#595959',
                        }}
                      >
                        <div>
                          <span style={{ color: '#8c8c8c' }}>{t('task.expectedQty')}: </span>
                          <span style={{ fontWeight: 600, color: '#262626' }}>
                            {task.expected_qty} {task.uom}
                          </span>
                        </div>
                        {task.batch_no && (
                          <div>
                            <span style={{ color: '#8c8c8c' }}>{t('task.batch')}: </span>
                            <span style={{ fontWeight: 500, color: '#262626' }}>{task.batch_no}</span>
                          </div>
                        )}
                      </div>

                      {task.instructions && (
                        <div
                          style={{
                            marginTop: 8,
                            padding: '8px 12px',
                            background: '#fafafa',
                            borderRadius: 6,
                            fontSize: 12,
                            color: '#8c8c8c',
                          }}
                        >
                          {task.instructions}
                        </div>
                      )}
                    </div>
                  ))}
                </>
              )}

              {/* Scan another button */}
              {pendingTasks.length > 0 && (
                <div style={{ marginTop: 16 }}>
                  <button
                    onClick={handleScanAnother}
                    style={{
                      width: '100%',
                      padding: '14px 0',
                      fontSize: 15,
                      fontWeight: 500,
                      color: '#595959',
                      background: '#f0f0f0',
                      border: 'none',
                      borderRadius: 10,
                      cursor: 'pointer',
                    }}
                  >
                    {t('putaway.scanAnother')}
                  </button>
                </div>
              )}
            </>
          )}

          {/* Task detail + putaway confirmation — shown when a task is selected */}
          {selectedTask && (
            <>
              {/* Back to task list */}
              <button
                onClick={handleBackToTaskList}
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
                {t('putaway.backToTaskList')}
              </button>

              {/* Task detail card */}
              <div
                style={{
                  background: '#fff',
                  borderRadius: 12,
                  padding: 20,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                  marginBottom: 16,
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    marginBottom: 16,
                  }}
                >
                  <div>
                    <h2
                      style={{
                        fontSize: 18,
                        fontWeight: 700,
                        color: '#262626',
                        marginBottom: 4,
                      }}
                    >
                      {selectedTask.task_no}
                    </h2>
                    <div style={{ fontSize: 13, color: '#8c8c8c' }}>
                      {t(getTaskStatusLabelKey(selectedTask.status as TaskStatus))}
                    </div>
                  </div>
                  <span
                    style={{
                      display: 'inline-block',
                      padding: '4px 10px',
                      fontSize: 12,
                      fontWeight: 600,
                      color: '#fff',
                      background: getTaskStatusColor(selectedTask.status as TaskStatus),
                      borderRadius: 4,
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {t(getTaskStatusLabelKey(selectedTask.status as TaskStatus))}
                  </span>
                </div>

                {/* Detail rows */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                  <DetailRow label={t('task.sku')} value={selectedTask.sku_id.substring(0, 8) + '...'} />
                  <DetailRow
                    label={t('task.expectedQty')}
                    value={`${selectedTask.expected_qty} ${selectedTask.uom}`}
                  />
                  {selectedTask.batch_no && (
                    <DetailRow label={t('task.batch')} value={selectedTask.batch_no} />
                  )}
                  {selectedTask.from_location_id && (
                    <DetailRow
                      label={t('putaway.fromLocation')}
                      value={selectedTask.from_location_id.substring(0, 8) + '...'}
                    />
                  )}
                  <DetailRow
                    label={t('putaway.toLocation')}
                    value={`${location.code} (${location.barcode})`}
                  />
                </div>
              </div>

              {/* Putaway confirmation card */}
              <div
                style={{
                  background: '#fff',
                  borderRadius: 12,
                  padding: 20,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                  marginBottom: 16,
                }}
              >
                <h3
                  style={{
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#262626',
                    marginBottom: 12,
                  }}
                >
                  {t('putaway.confirmPutaway')}
                </h3>

                {/* Location confirmation */}
                <div
                  style={{
                    padding: '12px 14px',
                    background: '#f6ffed',
                    border: '1px solid #b7eb8f',
                    borderRadius: 8,
                    marginBottom: 12,
                    fontSize: 13,
                    color: '#389e0d',
                  }}
                >
                  <div style={{ marginBottom: 4 }}>
                    <strong>{t('putaway.puttingAwayTo')}:</strong>
                  </div>
                  <div style={{ display: 'flex', gap: 16, fontSize: 13 }}>
                    <span>
                      {t('putaway.locationCode')}: <strong>{location.code}</strong>
                    </span>
                    <span>
                      {t('putaway.locationBarcode')}: <strong>{location.barcode}</strong>
                    </span>
                  </div>
                </div>

                {/* Qty input */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  <div style={{ display: 'flex', gap: 8, width: '100%' }}>
                    <input
                      type="number"
                      inputMode="decimal"
                      min="0"
                      max={selectedTask.expected_qty}
                      step="any"
                      value={actualQty}
                      onChange={(e) => setActualQty(e.target.value)}
                      placeholder={`${t('putaway.actualQty')} (max ${selectedTask.expected_qty})`}
                      disabled={isCompleting}
                      style={{
                        flex: 1,
                        padding: '10px 14px',
                        fontSize: 15,
                        border: qtyError
                          ? '2px solid #ff4d4f'
                          : '2px solid #d9d9d9',
                        borderRadius: 8,
                        outline: 'none',
                        background: isCompleting ? '#f5f5f5' : '#fff',
                      }}
                      onFocus={(e) => {
                        if (!qtyError) e.currentTarget.style.borderColor = '#1677ff'
                      }}
                      onBlur={(e) => {
                        if (!qtyError) e.currentTarget.style.borderColor = '#d9d9d9'
                      }}
                    />
                    <button
                      onClick={handleComplete}
                      disabled={
                        isCompleting ||
                        !actualQty ||
                        isNaN(parseFloat(actualQty)) ||
                        parseFloat(actualQty) <= 0 ||
                        !!qtyError
                      }
                      style={{
                        padding: '10px 18px',
                        fontSize: 14,
                        fontWeight: 600,
                        color: '#fff',
                        background:
                          isCompleting || !actualQty || isNaN(parseFloat(actualQty)) ||
                          parseFloat(actualQty) <= 0 || !!qtyError
                            ? '#d9d9d9'
                            : '#389e0d',
                        border: 'none',
                        borderRadius: 8,
                        cursor:
                          isCompleting || !actualQty || isNaN(parseFloat(actualQty)) ||
                          parseFloat(actualQty) <= 0 || !!qtyError
                            ? 'not-allowed'
                            : 'pointer',
                        whiteSpace: 'nowrap',
                        transition: 'background 0.2s',
                      }}
                    >
                      {isCompleting ? t('putaway.completing') : t('putaway.confirm')}
                    </button>
                  </div>

                  {/* Qty error */}
                  {qtyError && (
                    <div
                      style={{
                        width: '100%',
                        fontSize: 12,
                        color: '#ff4d4f',
                        paddingLeft: 2,
                      }}
                    >
                      {qtyError}
                    </div>
                  )}
                </div>
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}

// ── Detail row helper ─────────────────────────────────────────────────

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}
    >
      <span style={{ fontSize: 13, color: '#8c8c8c' }}>{label}</span>
      <span
        style={{
          fontSize: 14,
          fontWeight: 500,
          color: '#262626',
          textAlign: 'right',
          maxWidth: '60%',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {value}
      </span>
    </div>
  )
}
