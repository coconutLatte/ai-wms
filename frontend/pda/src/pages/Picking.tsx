// Picking page — picking workflow for warehouse operators.
// Operator views pending pick tasks, selects one, scans the source location
// barcode, scans the SKU barcode for verification, enters picked quantity,
// and confirms completion. All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import client from '@/api/client'
import BarcodeScanner from '@/components/BarcodeScanner'
import type { Task, TaskStatus, TaskPriority, ListResponse } from '@/api/types'

// ── Step enum for workflow state ───────────────────────────────────────

type Step = 'task_list' | 'scan_location' | 'scan_sku' | 'confirm'

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

export default function PickingPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ────────────────────────────────────────────────────────────

  const [step, setStep] = useState<Step>('task_list')
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)
  const [locationBarcode, setLocationBarcode] = useState('')
  const [skuBarcode, setSkuBarcode] = useState('')
  const [actualQty, setActualQty] = useState('')
  const [isCompleting, setIsCompleting] = useState(false)

  // ── Fetch pending picking tasks ──────────────────────────────────────

  const {
    data: tasksList,
    isLoading: isLoadingTasks,
    isError: isTasksError,
    error: tasksError,
    refetch: refetchTasks,
  } = useQuery<ListResponse<Task>>({
    queryKey: ['pickingTasks'],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<Task>>('/tasks', {
        params: {
          task_type: 'picking',
          status: 'pending',
          page_size: 50,
        },
      })
      return data
    },
    retry: false,
  })

  const pendingTasks = tasksList?.data ?? []

  // ── Selected task ────────────────────────────────────────────────────

  const selectedTask = pendingTasks.find((t) => t.id === selectedTaskId) ?? null

  // ── Location lookup by barcode ───────────────────────────────────────

  const {
    data: locationData,
    isLoading: isLookingUpLocation,
    isError: isLocationError,
    error: locationError,
  } = useQuery<
    | { data?: { id: string; code: string; barcode: string; status: string; location_type: string }[] }
    | { id: string; code: string; barcode: string; status: string; location_type: string }
  >({
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
    if ('id' in locationData && locationData.barcode) return locationData
    return null
  })()

  // ── SKU lookup by barcode ────────────────────────────────────────────

  const {
    data: skuData,
    isLoading: isLookingUpSku,
    isError: isSkuError,
    error: skuError,
  } = useQuery<
    | { data?: { id: string; code: string; name: string; barcode: string; status: string }[] }
    | { id: string; code: string; name: string; barcode: string; status: string }
  >({
    queryKey: ['skuBarcode', skuBarcode],
    queryFn: async () => {
      const { data } = await client.get('/skus', {
        params: { code: skuBarcode },
      })
      return data
    },
    enabled: skuBarcode.length > 0,
    retry: false,
  })

  // Extract single SKU from response
  const sku = (() => {
    if (!skuData) return null
    if ('data' in skuData && Array.isArray(skuData.data) && skuData.data.length === 1) {
      return skuData.data[0]
    }
    if ('id' in skuData && skuData.barcode) return skuData
    return null
  })()

  // ── Start task mutation (pending → in_progress) ──────────────────────

  const startTaskMutation = useMutation({
    mutationFn: async (taskId: string) => {
      await client.put(`/tasks/${taskId}/status`, { status: 'in_progress' })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pickingTasks'] })
    },
  })

  // ── Complete task mutation ───────────────────────────────────────────

  const completeTaskMutation = useMutation({
    mutationFn: async ({
      taskId,
      qty,
      fromLocationId,
    }: {
      taskId: string
      qty: number
      fromLocationId: string
    }) => {
      await client.post(`/tasks/${taskId}/complete`, {
        actual_qty: qty,
        from_location_id: fromLocationId,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pickingTasks'] })
    },
  })

  // ── Handle select task → start it ────────────────────────────────────

  const handleSelectTask = async (task: Task) => {
    setSelectedTaskId(task.id)
    setStep('scan_location')
    setLocationBarcode('')
    setSkuBarcode('')
    setActualQty('')
    try {
      await startTaskMutation.mutateAsync(task.id)
    } catch {
      // Error handled by mutation state
    }
  }

  // ── Handle location barcode scan ─────────────────────────────────────

  const handleLocationScan = useCallback((barcode: string) => {
    setLocationBarcode(barcode)
  }, [])

  // Location verified — advance to next step
  const handleLocationConfirmed = useCallback(() => {
    if (location) {
      setStep('scan_sku')
      setSkuBarcode('')
    }
  }, [location])

  // ── Handle SKU barcode scan ──────────────────────────────────────────

  const handleSkuScan = useCallback((barcode: string) => {
    setSkuBarcode(barcode)
  }, [])

  // SKU verified — advance to confirm step
  const handleSkuConfirmed = useCallback(() => {
    if (sku) {
      setStep('confirm')
      if (selectedTask) {
        setActualQty(selectedTask.expected_qty.toString())
      }
    }
  }, [sku, selectedTask])

  // ── Complete the pick ────────────────────────────────────────────────

  const handleComplete = async () => {
    if (!selectedTask || !location) return
    const qty = parseFloat(actualQty)
    if (isNaN(qty) || qty <= 0) return

    setIsCompleting(true)
    try {
      await completeTaskMutation.mutateAsync({
        taskId: selectedTask.id,
        qty,
        fromLocationId: location.id,
      })
      // All done — reset everything
      setStep('task_list')
      setSelectedTaskId(null)
      setLocationBarcode('')
      setSkuBarcode('')
      setActualQty('')
    } catch {
      // Error handled by mutation state
    } finally {
      setIsCompleting(false)
    }
  }

  // ── Reset entirely ───────────────────────────────────────────────────

  const handleResetAll = useCallback(() => {
    setStep('task_list')
    setSelectedTaskId(null)
    setLocationBarcode('')
    setSkuBarcode('')
    setActualQty('')
    setIsCompleting(false)
    queryClient.removeQueries({ queryKey: ['locationBarcode', locationBarcode] })
    queryClient.removeQueries({ queryKey: ['skuBarcode', skuBarcode] })
  }, [locationBarcode, skuBarcode, queryClient])

  // ── Back one step ────────────────────────────────────────────────────

  const handleBack = useCallback(() => {
    if (step === 'scan_location') {
      setStep('task_list')
      setSelectedTaskId(null)
      setLocationBarcode('')
    } else if (step === 'scan_sku') {
      setStep('scan_location')
      setLocationBarcode('')
    } else if (step === 'confirm') {
      setStep('scan_sku')
      setSkuBarcode('')
      setActualQty('')
    }
  }, [step])

  // ── Qty validation ───────────────────────────────────────────────────

  let qtyError: string | null = null
  if (step === 'confirm' && selectedTask && actualQty) {
    const parsed = parseFloat(actualQty)
    if (!isNaN(parsed)) {
      if (parsed > selectedTask.expected_qty) {
        qtyError = t('picking.qtyExceedsExpected', {
          total: actualQty,
          expected: selectedTask.expected_qty.toString(),
        })
      } else if (parsed <= 0) {
        qtyError = t('picking.qtyMustBePositive')
      }
    }
  }

  // ── Verify location matches task's from_location ─────────────────────

  let locationWarning: string | null = null
  if (location && selectedTask?.from_location_id) {
    if (location.id !== selectedTask.from_location_id) {
      locationWarning = t('picking.locationMismatch', {
        scanned: location.code || location.barcode,
        expected: selectedTask.from_location_id.substring(0, 8) + '...',
      })
    }
  }

  // ── Verify SKU matches task's SKU ────────────────────────────────────

  let skuWarning: string | null = null
  if (sku && selectedTask) {
    if (sku.id !== selectedTask.sku_id) {
      skuWarning = t('picking.skuMismatch', {
        scanned: sku.code || sku.barcode,
        expected: selectedTask.sku_id.substring(0, 8) + '...',
      })
    }
  }

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      {/* ── Step indicator ────────────────────────────────────────────── */}
      {step !== 'task_list' && selectedTask && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            marginBottom: 16,
            padding: '10px 14px',
            background: '#e6f7ff',
            borderRadius: 10,
            fontSize: 13,
          }}
        >
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 22,
              height: 22,
              borderRadius: '50%',
              background: step === 'scan_location' ? '#1677ff' : '#b7eb8f',
              color: '#fff',
              fontSize: 12,
              fontWeight: 700,
            }}
          >
            {step === 'scan_location' ? '1' : '✓'}
          </span>
          <span style={{ color: step === 'scan_location' ? '#1677ff' : '#389e0d' }}>
            {t('picking.locationVerification')}
          </span>
          <span style={{ color: '#d9d9d9' }}>→</span>
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 22,
              height: 22,
              borderRadius: '50%',
              background: step === 'scan_sku' ? '#1677ff' : step === 'confirm' ? '#b7eb8f' : '#f0f0f0',
              color: step === 'scan_sku' ? '#fff' : step === 'confirm' ? '#fff' : '#8c8c8c',
              fontSize: 12,
              fontWeight: 700,
            }}
          >
            {step === 'scan_sku' ? '2' : step === 'confirm' ? '✓' : '2'}
          </span>
          <span
            style={{
              color:
                step === 'scan_sku' ? '#1677ff' : step === 'confirm' ? '#389e0d' : '#8c8c8c',
            }}
          >
            {t('picking.skuVerification')}
          </span>
          <span style={{ color: '#d9d9d9' }}>→</span>
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 22,
              height: 22,
              borderRadius: '50%',
              background: step === 'confirm' ? '#1677ff' : '#f0f0f0',
              color: step === 'confirm' ? '#fff' : '#8c8c8c',
              fontSize: 12,
              fontWeight: 700,
            }}
          >
            3
          </span>
          <span style={{ color: step === 'confirm' ? '#1677ff' : '#8c8c8c' }}>
            {t('picking.confirmPick')}
          </span>
        </div>
      )}

      {/* ── Back button (when not on task_list) ───────────────────────── */}
      {step !== 'task_list' && (
        <button
          onClick={handleBack}
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
          {t('picking.back')}
        </button>
      )}

      {/* ── Mutation errors ───────────────────────────────────────────── */}
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
          {t('picking.startTaskError', {
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
          {t('picking.completeTaskError', {
            error: (completeTaskMutation.error as Error)?.message,
          })}
        </div>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 0: Task list */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'task_list' && (
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
              {t('picking.title')}
            </h3>
          </div>

          {/* Loading */}
          {isLoadingTasks && (
            <div className="pda-empty">
              <span className="empty-icon">{'⏳'}</span>
              <p className="empty-text">{t('picking.loadingTasks')}</p>
            </div>
          )}

          {/* Error */}
          {isTasksError && (
            <div className="pda-empty">
              <span className="empty-icon">{'⚠️'}</span>
              <p className="empty-text">
                {t('picking.tasksLoadFailed', {
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

          {/* Empty state */}
          {!isLoadingTasks && !isTasksError && pendingTasks.length === 0 && (
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
                {t('picking.noPendingTasks')}
              </div>
              <div style={{ fontSize: 13, color: '#8c8c8c', marginBottom: 20 }}>
                {t('picking.noPendingTasksDesc')}
              </div>
            </div>
          )}

          {/* Task list */}
          {!isLoadingTasks && !isTasksError && pendingTasks.length > 0 && (
            <>
              <div
                style={{
                  fontSize: 14,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 10,
                }}
              >
                {t('picking.selectTask', { count: pendingTasks.length })}
              </div>

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
                    {task.from_location_id && (
                      <div>
                        <span style={{ color: '#8c8c8c' }}>{t('picking.pickFrom')}: </span>
                        <span style={{ fontWeight: 500, color: '#262626', fontFamily: 'monospace' }}>
                          {task.from_location_id.substring(0, 8)}...
                        </span>
                      </div>
                    )}
                    {task.batch_no && (
                      <div>
                        <span style={{ color: '#8c8c8c' }}>{t('task.batch')}: </span>
                        <span style={{ fontWeight: 500, color: '#262626' }}>{task.batch_no}</span>
                      </div>
                    )}
                    {task.order_id && (
                      <div>
                        <span style={{ color: '#8c8c8c' }}>{t('picking.order')}: </span>
                        <span style={{ fontWeight: 500, color: '#262626', fontFamily: 'monospace' }}>
                          {task.order_id.substring(0, 8)}...
                        </span>
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
        </>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 1: Scan location barcode (verify pick-from location) */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'scan_location' && selectedTask && (
        <>
          {/* Task summary card */}
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
                marginBottom: 12,
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

            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <DetailRow
                label={t('task.expectedQty')}
                value={`${selectedTask.expected_qty} ${selectedTask.uom}`}
              />
              {selectedTask.batch_no && (
                <DetailRow label={t('task.batch')} value={selectedTask.batch_no} />
              )}
              {selectedTask.from_location_id && (
                <DetailRow
                  label={t('picking.pickFrom')}
                  value={selectedTask.from_location_id.substring(0, 8) + '...'}
                />
              )}
              {selectedTask.instructions && (
                <div
                  style={{
                    marginTop: 4,
                    padding: '8px 12px',
                    background: '#fafafa',
                    borderRadius: 6,
                    fontSize: 12,
                    color: '#8c8c8c',
                  }}
                >
                  {selectedTask.instructions}
                </div>
              )}
            </div>
          </div>

          {/* Location scan section */}
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
              {t('picking.scanLocationTitle')}
            </h3>
            <p
              style={{
                fontSize: 13,
                color: '#8c8c8c',
                marginBottom: 12,
                lineHeight: 1.5,
              }}
            >
              {t('picking.scanLocationHint')}
            </p>

            <BarcodeScanner
              onScan={handleLocationScan}
              placeholder={t('picking.scanLocationPlaceholder')}
            />

            {/* Lookup in progress */}
            {isLookingUpLocation && (
              <div
                style={{
                  marginTop: 12,
                  padding: '12px',
                  textAlign: 'center',
                  color: '#1677ff',
                  fontSize: 14,
                }}
              >
                {'\u{1F50D}'} {t('picking.lookingUpLocation')}
              </div>
            )}

            {/* Location not found or error */}
            {(isLocationError || (!isLookingUpLocation && locationBarcode && !location)) && (
              <div
                style={{
                  marginTop: 12,
                  padding: '14px 18px',
                  background: '#fff1f0',
                  border: '1px solid #ffa39e',
                  borderRadius: 10,
                  fontSize: 14,
                  color: '#cf1322',
                }}
              >
                {locationError
                  ? t('picking.lookupFailed', { error: (locationError as Error)?.message })
                  : t('picking.locationNotFound', { barcode: locationBarcode })}
              </div>
            )}

            {/* Location found card */}
            {location && (
              <div style={{ marginTop: 12 }}>
                <div
                  style={{
                    padding: '14px 18px',
                    background: locationWarning ? '#fffbe6' : '#f6ffed',
                    border: locationWarning ? '1px solid #ffe58f' : '1px solid #b7eb8f',
                    borderRadius: 10,
                  }}
                >
                  <div
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: locationWarning ? '#d48806' : '#389e0d',
                      marginBottom: 8,
                    }}
                  >
                    {locationWarning
                      ? t('picking.warning')
                      : t('picking.locationVerified')}
                  </div>
                  {locationWarning && (
                    <div
                      style={{
                        fontSize: 13,
                        color: '#d48806',
                        marginBottom: 8,
                      }}
                    >
                      {locationWarning}
                    </div>
                  )}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ color: '#8c8c8c' }}>{t('picking.locationCode')}</span>
                      <span style={{ fontWeight: 500, color: '#262626' }}>{location.code}</span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ color: '#8c8c8c' }}>{t('picking.locationBarcode')}</span>
                      <span
                        style={{
                          fontWeight: 500,
                          color: '#262626',
                          fontFamily: 'monospace',
                          fontSize: 12,
                        }}
                      >
                        {location.barcode}
                      </span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ color: '#8c8c8c' }}>{t('picking.locationType')}</span>
                      <span style={{ fontWeight: 500, color: '#262626' }}>{location.location_type}</span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={handleLocationConfirmed}
                  style={{
                    width: '100%',
                    marginTop: 12,
                    padding: '14px 0',
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#fff',
                    background: '#1677ff',
                    border: 'none',
                    borderRadius: 8,
                    cursor: 'pointer',
                  }}
                >
                  {t('picking.confirmLocation')}
                </button>
              </div>
            )}
          </div>
        </>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 2: Scan SKU barcode (verify correct SKU) */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'scan_sku' && selectedTask && (
        <>
          {/* Location confirmation banner */}
          {location && (
            <div
              style={{
                padding: '10px 14px',
                background: '#f6ffed',
                border: '1px solid #b7eb8f',
                borderRadius: 8,
                marginBottom: 16,
                fontSize: 13,
                color: '#389e0d',
              }}
            >
              {t('picking.locationConfirmedBanner', {
                location: location.code || location.barcode,
              })}
            </div>
          )}

          {/* SKU scan card */}
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
              {t('picking.scanSkuTitle')}
            </h3>
            <p
              style={{
                fontSize: 13,
                color: '#8c8c8c',
                marginBottom: 12,
                lineHeight: 1.5,
              }}
            >
              {t('picking.scanSkuHint')}
            </p>

            <BarcodeScanner
              onScan={handleSkuScan}
              placeholder={t('picking.scanSkuPlaceholder')}
            />

            {/* SKU lookup in progress */}
            {isLookingUpSku && (
              <div
                style={{
                  marginTop: 12,
                  padding: '12px',
                  textAlign: 'center',
                  color: '#1677ff',
                  fontSize: 14,
                }}
              >
                {'\u{1F50D}'} {t('picking.lookingUpSku')}
              </div>
            )}

            {/* SKU not found or error */}
            {(isSkuError || (!isLookingUpSku && skuBarcode && !sku)) && (
              <div
                style={{
                  marginTop: 12,
                  padding: '14px 18px',
                  background: '#fff1f0',
                  border: '1px solid #ffa39e',
                  borderRadius: 10,
                  fontSize: 14,
                  color: '#cf1322',
                }}
              >
                {skuError
                  ? t('picking.skuLookupFailed', { error: (skuError as Error)?.message })
                  : t('picking.skuNotFound', { barcode: skuBarcode })}
              </div>
            )}

            {/* SKU found card */}
            {sku && (
              <div style={{ marginTop: 12 }}>
                <div
                  style={{
                    padding: '14px 18px',
                    background: skuWarning ? '#fffbe6' : '#f6ffed',
                    border: skuWarning ? '1px solid #ffe58f' : '1px solid #b7eb8f',
                    borderRadius: 10,
                  }}
                >
                  <div
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: skuWarning ? '#d48806' : '#389e0d',
                      marginBottom: 8,
                    }}
                  >
                    {skuWarning
                      ? t('picking.warning')
                      : t('picking.skuVerified')}
                  </div>
                  {skuWarning && (
                    <div
                      style={{
                        fontSize: 13,
                        color: '#d48806',
                        marginBottom: 8,
                      }}
                    >
                      {skuWarning}
                    </div>
                  )}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ color: '#8c8c8c' }}>{t('picking.skuCode')}</span>
                      <span style={{ fontWeight: 500, color: '#262626', fontFamily: 'monospace' }}>
                        {sku.code}
                      </span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ color: '#8c8c8c' }}>{t('picking.skuName')}</span>
                      <span style={{ fontWeight: 500, color: '#262626' }}>{sku.name || '—'}</span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span style={{ color: '#8c8c8c' }}>{t('scan.barcode')}</span>
                      <span
                        style={{
                          fontWeight: 500,
                          color: '#262626',
                          fontFamily: 'monospace',
                          fontSize: 12,
                        }}
                      >
                        {sku.barcode}
                      </span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={handleSkuConfirmed}
                  style={{
                    width: '100%',
                    marginTop: 12,
                    padding: '14px 0',
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#fff',
                    background: '#1677ff',
                    border: 'none',
                    borderRadius: 8,
                    cursor: 'pointer',
                  }}
                >
                  {t('picking.confirmSku')}
                </button>
              </div>
            )}
          </div>
        </>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 3: Confirm — enter qty and complete */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'confirm' && selectedTask && (
        <>
          {/* Verification summary */}
          <div
            style={{
              padding: '14px 18px',
              background: '#f6ffed',
              border: '1px solid #b7eb8f',
              borderRadius: 10,
              marginBottom: 16,
              fontSize: 13,
            }}
          >
            <div style={{ color: '#389e0d', fontWeight: 600, marginBottom: 8 }}>
              {t('picking.verificationPassed')}
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 4, color: '#595959' }}>
              {location && (
                <div>
                  {t('picking.pickFrom')}: <strong>{location.code}</strong> ({location.barcode})
                </div>
              )}
              {sku && (
                <div>
                  {t('task.sku')}: <strong>{sku.code}</strong>
                </div>
              )}
            </div>
          </div>

          {/* Task summary card */}
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
                marginBottom: 12,
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

            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <DetailRow
                label={t('task.expectedQty')}
                value={`${selectedTask.expected_qty} ${selectedTask.uom}`}
              />
              {selectedTask.batch_no && (
                <DetailRow label={t('task.batch')} value={selectedTask.batch_no} />
              )}
            </div>
          </div>

          {/* Confirmation card */}
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
              {t('picking.confirmPick')}
            </h3>

            {/* Qty input row */}
            <div style={{ display: 'flex', gap: 8, width: '100%' }}>
              <input
                type="number"
                inputMode="decimal"
                min="0"
                max={selectedTask.expected_qty}
                step="any"
                value={actualQty}
                onChange={(e) => setActualQty(e.target.value)}
                placeholder={`${t('picking.actualQty')} (max ${selectedTask.expected_qty})`}
                disabled={isCompleting}
                style={{
                  flex: 1,
                  padding: '10px 14px',
                  fontSize: 15,
                  border: qtyError ? '2px solid #ff4d4f' : '2px solid #d9d9d9',
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
                    isCompleting ||
                    !actualQty ||
                    isNaN(parseFloat(actualQty)) ||
                    parseFloat(actualQty) <= 0 ||
                    !!qtyError
                      ? '#d9d9d9'
                      : '#389e0d',
                  border: 'none',
                  borderRadius: 8,
                  cursor:
                    isCompleting ||
                    !actualQty ||
                    isNaN(parseFloat(actualQty)) ||
                    parseFloat(actualQty) <= 0 ||
                    !!qtyError
                      ? 'not-allowed'
                      : 'pointer',
                  whiteSpace: 'nowrap',
                  transition: 'background 0.2s',
                }}
              >
                {isCompleting ? t('picking.completing') : t('picking.confirm')}
              </button>
            </div>

            {/* Qty error */}
            {qtyError && (
              <div
                style={{
                  width: '100%',
                  marginTop: 8,
                  fontSize: 12,
                  color: '#ff4d4f',
                  paddingLeft: 2,
                }}
              >
                {qtyError}
              </div>
            )}
          </div>

          {/* Cancel / start over button */}
          <div style={{ marginTop: 16 }}>
            <button
              onClick={handleResetAll}
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
              {t('picking.cancelAndReturn')}
            </button>
          </div>
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
