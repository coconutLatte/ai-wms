// Transfer page — transfer workflow for warehouse operators.
// Operator scans a source location, then a destination location, then enters
// the SKU (by code/barcode), optional batch number, and quantity. The system
// creates, starts, and completes a transfer task to move inventory between
// locations within the same warehouse.
// All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import client from '@/api/client'
import BarcodeScanner from '@/components/BarcodeScanner'
import type { Task, ListResponse } from '@/api/types'

// ── Types ────────────────────────────────────────────────────────────────

interface LocationData {
  id: string
  code: string
  barcode: string
  status: string
  warehouse_id: string
  location_type: string
}

interface SkuData {
  id: string
  code: string
  name: string
  barcode: string
  status: string
}

interface InventoryRecord {
  id: string
  sku_id: string
  location_id: string
  warehouse_id: string
  qty: number
  reserved_qty: number
  available_qty?: number
  batch_no: string
  status: string
}

// ── Helpers ──────────────────────────────────────────────────────────────

function getInventoryStatusColor(status: string): string {
  const map: Record<string, string> = {
    available: '#389e0d',
    quarantine: '#fa8c16',
    damaged: '#cf1322',
    expired: '#8c8c8c',
  }
  return map[status] || '#595959'
}

// ── Steps ────────────────────────────────────────────────────────────────

type Step = 'source' | 'destination' | 'details' | 'confirm' | 'done'

// ── Main component ───────────────────────────────────────────────────────

export default function TransferPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ──────────────────────────────────────────────────────────────

  const [step, setStep] = useState<Step>('source')
  const [sourceBarcode, setSourceBarcode] = useState('')
  const [destBarcode, setDestBarcode] = useState('')
  const [skuInput, setSkuInput] = useState('')
  const [skuLookupBarcode, setSkuLookupBarcode] = useState('')
  const [batchNo, setBatchNo] = useState('')
  const [transferQty, setTransferQty] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [completedTask, setCompletedTask] = useState<Task | null>(null)

  // ── Step 1: Lookup source location by barcode ─────────────────────────

  const {
    data: sourceLocationData,
    isLoading: isLoadingSource,
    isError: isSourceError,
    error: sourceError,
  } = useQuery<{ data?: LocationData[] } & LocationData>({
    queryKey: ['transferSourceLocation', sourceBarcode],
    queryFn: async () => {
      const { data } = await client.get('/locations', {
        params: { barcode: sourceBarcode },
      })
      return data
    },
    enabled: sourceBarcode.length > 0,
    retry: false,
  })

  const sourceLocation = (() => {
    if (!sourceLocationData) return null
    if ('data' in sourceLocationData && Array.isArray(sourceLocationData.data) && sourceLocationData.data.length === 1) {
      return sourceLocationData.data[0]
    }
    if (sourceLocationData.id && sourceLocationData.barcode) return sourceLocationData as LocationData
    return null
  })()

  // ── Step 2: Lookup destination location by barcode ─────────────────────

  const {
    data: destLocationData,
    isLoading: isLoadingDest,
    isError: isDestError,
    error: destError,
  } = useQuery<{ data?: LocationData[] } & LocationData>({
    queryKey: ['transferDestLocation', destBarcode],
    queryFn: async () => {
      const { data } = await client.get('/locations', {
        params: { barcode: destBarcode },
      })
      return data
    },
    enabled: destBarcode.length > 0,
    retry: false,
  })

  const destLocation = (() => {
    if (!destLocationData) return null
    if ('data' in destLocationData && Array.isArray(destLocationData.data) && destLocationData.data.length === 1) {
      return destLocationData.data[0]
    }
    if (destLocationData.id && destLocationData.barcode) return destLocationData as LocationData
    return null
  })()

  // ── Step 3: SKU lookup by code or barcode ──────────────────────────────

  const skuQueryParam = skuLookupBarcode || skuInput
  const skuLookupMode = skuLookupBarcode ? 'barcode' : 'code'

  const {
    data: skuData,
    isLoading: isLoadingSku,
    isError: isSkuError,
    error: skuError,
  } = useQuery<{ data?: SkuData[] } & SkuData>({
    queryKey: ['transferSku', skuQueryParam, skuLookupMode],
    queryFn: async () => {
      const param = skuLookupBarcode ? { barcode: skuQueryParam } : { code: skuQueryParam }
      const { data } = await client.get('/skus', { params: param })
      return data
    },
    enabled: skuQueryParam.length > 0,
    retry: false,
  })

  const sku = (() => {
    if (!skuData) return null
    if ('data' in skuData && Array.isArray(skuData.data) && skuData.data.length === 1) {
      return skuData.data[0]
    }
    if (skuData.id && skuData.code) return skuData as SkuData
    return null
  })()

  // ── Source inventory (shown after source location confirmed) ───────────

  const {
    data: sourceInventoryData,
    isLoading: isLoadingSourceInv,
    isError: isSourceInvError,
    error: sourceInvError,
  } = useQuery<ListResponse<InventoryRecord>>({
    queryKey: ['transferSourceInventory', sourceLocation?.id],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<InventoryRecord>>('/inventory', {
        params: { location_id: sourceLocation!.id, status: 'available', page_size: 50 },
      })
      return data
    },
    enabled: !!sourceLocation,
    retry: false,
  })

  const sourceInventory = sourceInventoryData?.data ?? []

  // ── Handlers ───────────────────────────────────────────────────────────

  const handleScanSource = useCallback((barcode: string) => {
    setSourceBarcode(barcode)
    // Don't auto-advance; wait for user to confirm
  }, [])

  const handleScanDest = useCallback((barcode: string) => {
    setDestBarcode(barcode)
  }, [])

  const handleConfirmSource = useCallback(() => {
    if (sourceLocation) {
      setStep('destination')
    }
  }, [sourceLocation])

  const handleConfirmDest = useCallback(() => {
    if (destLocation) {
      // Validate: source and destination must be different
      if (sourceLocation && destLocation.id === sourceLocation.id) {
        return // will show error
      }
      setStep('details')
    }
  }, [destLocation, sourceLocation])

  const handleScanSku = useCallback((barcode: string) => {
    setSkuLookupBarcode(barcode)
    setSkuInput('')
  }, [])

  const handleReviewTransfer = useCallback(() => {
    if (sku && sourceLocation && destLocation && transferQty) {
      setStep('confirm')
    }
  }, [sku, sourceLocation, destLocation, transferQty])

  const handleClear = useCallback(() => {
    queryClient.removeQueries({ queryKey: ['transferSourceLocation'] })
    queryClient.removeQueries({ queryKey: ['transferDestLocation'] })
    queryClient.removeQueries({ queryKey: ['transferSku'] })
    queryClient.removeQueries({ queryKey: ['transferSourceInventory'] })
    setStep('source')
    setSourceBarcode('')
    setDestBarcode('')
    setSkuInput('')
    setSkuLookupBarcode('')
    setBatchNo('')
    setTransferQty('')
    setIsProcessing(false)
    setCompletedTask(null)
  }, [queryClient])

  // ── Mutations: create → start → complete transfer task ────────────────

  const createTaskMutation = useMutation({
    mutationFn: async (input: {
      fromLocationId: string
      toLocationId: string
      skuId: string
      qty: number
      warehouseId: string
      batchNo: string
    }) => {
      const { data } = await client.post('/tasks', {
        task_type: 'transfer',
        warehouse_id: input.warehouseId,
        sku_id: input.skuId,
        from_location_id: input.fromLocationId,
        to_location_id: input.toLocationId,
        expected_qty: input.qty,
        uom: 'EA',
        batch_no: input.batchNo || undefined,
        priority: 'normal',
        instructions: t('transfer.taskInstructions', {
          from: input.fromLocationId.substring(0, 8),
          to: input.toLocationId.substring(0, 8),
          qty: input.qty,
        }),
      })
      return data as Task
    },
  })

  const startTaskMutation = useMutation({
    mutationFn: async (taskId: string) => {
      await client.put(`/tasks/${taskId}/status`, { status: 'in_progress' })
    },
  })

  const completeTaskMutation = useMutation({
    mutationFn: async (input: { taskId: string; qty: number; toLocationId: string }) => {
      await client.post(`/tasks/${input.taskId}/complete`, {
        actual_qty: input.qty,
        to_location_id: input.toLocationId,
      })
    },
  })

  const handleConfirmTransfer = async () => {
    if (!sourceLocation || !destLocation || !sku) return
    const qty = parseFloat(transferQty)
    if (isNaN(qty) || qty <= 0) return

    setIsProcessing(true)
    try {
      // 1. Create transfer task
      const task = await createTaskMutation.mutateAsync({
        fromLocationId: sourceLocation.id,
        toLocationId: destLocation.id,
        skuId: sku.id,
        qty,
        warehouseId: sourceLocation.warehouse_id,
        batchNo,
      })

      // 2. Start task (pending → in_progress)
      await startTaskMutation.mutateAsync(task.id)

      // 3. Complete task with inventory transfer
      await completeTaskMutation.mutateAsync({
        taskId: task.id,
        qty,
        toLocationId: destLocation.id,
      })

      // Success
      setCompletedTask(task)
      setStep('done')
      queryClient.invalidateQueries({ queryKey: ['transferSourceInventory', sourceLocation.id] })
    } catch {
      // Error handled by mutation state
    } finally {
      setIsProcessing(false)
    }
  }

  // ── Validation ─────────────────────────────────────────────────────────

  let qtyError: string | null = null
  if (transferQty) {
    const parsed = parseFloat(transferQty)
    if (!isNaN(parsed) && parsed <= 0) {
      qtyError = t('transfer.qtyMustBePositive')
    }
  }

  const isSameLocation = sourceLocation && destLocation && sourceLocation.id === destLocation.id

  const anyMutationError =
    createTaskMutation.isError || startTaskMutation.isError || completeTaskMutation.isError
  const mutationErrorMessage =
    (createTaskMutation.error as Error)?.message ||
    (startTaskMutation.error as Error)?.message ||
    (completeTaskMutation.error as Error)?.message

  // ── Render ─────────────────────────────────────────────────────────────

  return (
    <div>
      {/* Header + back/clear */}
      {step !== 'source' && (
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
          {t('transfer.backToStart')}
        </button>
      )}

      {/* Mutation error banner */}
      {anyMutationError && (
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
          {t('transfer.processError', { error: mutationErrorMessage || '' })}
        </div>
      )}

      {/* ── Step 1: Scan source location ──────────────────────────────── */}
      {step === 'source' && (
        <>
          <h3 style={{ fontSize: 15, fontWeight: 600, color: '#595959', marginBottom: 10 }}>
            {t('transfer.title')}
          </h3>
          <p style={{ fontSize: 13, color: '#8c8c8c', marginBottom: 16, lineHeight: 1.5 }}>
            {t('transfer.description')}
          </p>

          {/* Step indicator */}
          <div style={{
            display: 'flex', gap: 6, marginBottom: 16, fontSize: 12, color: '#8c8c8c',
          }}>
            <StepBadge active label={t('transfer.stepSource')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge label={t('transfer.stepDest')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge label={t('transfer.stepDetails')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge label={t('transfer.stepConfirm')} />
          </div>

          <div style={{
            background: '#e6f7ff',
            borderRadius: 10,
            padding: 16,
            marginBottom: 16,
            border: '1px solid #91d5ff',
          }}>
            <div style={{ fontSize: 14, fontWeight: 600, color: '#1677ff', marginBottom: 8 }}>
              {t('transfer.scanSourceLocation')}
            </div>
            <BarcodeScanner
              onScan={handleScanSource}
              placeholder={t('transfer.scanSourcePlaceholder')}
            />
          </div>

          {/* Loading */}
          {isLoadingSource && (
            <div style={{ padding: '16px', textAlign: 'center', color: '#1677ff', fontSize: 14 }}>
              {'🔍'} {t('transfer.lookingUpLocation')}
            </div>
          )}

          {/* Error */}
          {(isSourceError || (!isLoadingSource && sourceBarcode && !sourceLocation)) && (
            <div style={{
              marginTop: 16, padding: '14px 18px', background: '#fff1f0',
              border: '1px solid #ffa39e', borderRadius: 10, fontSize: 14, color: '#cf1322',
            }}>
              {sourceError
                ? t('transfer.lookupFailed', { error: (sourceError as Error)?.message })
                : t('transfer.locationNotFound', { barcode: sourceBarcode })}
            </div>
          )}

          {/* Found — show source location card */}
          {sourceLocation && (
            <>
              <LocationCard
                location={sourceLocation}
                roleLabel={t('transfer.sourceLocation')}
                color="#1677ff"
                background="#e6f7ff"
                borderColor="#91d5ff"
                t={t}
              />

              <button
                onClick={handleConfirmSource}
                style={{
                  width: '100%', padding: '14px 0', fontSize: 15, fontWeight: 600,
                  color: '#fff', background: '#1677ff', border: 'none', borderRadius: 10,
                  cursor: 'pointer', marginTop: 4,
                }}
              >
                {t('transfer.confirmSource')}
              </button>
            </>
          )}
        </>
      )}

      {/* ── Step 2: Scan destination location ──────────────────────────── */}
      {step === 'destination' && sourceLocation && (
        <>
          <h3 style={{ fontSize: 15, fontWeight: 600, color: '#595959', marginBottom: 10 }}>
            {t('transfer.title')}
          </h3>

          {/* Step indicator */}
          <div style={{
            display: 'flex', gap: 6, marginBottom: 16, fontSize: 12, color: '#8c8c8c',
          }}>
            <StepBadge done label={t('transfer.stepSource')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge active label={t('transfer.stepDest')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge label={t('transfer.stepDetails')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge label={t('transfer.stepConfirm')} />
          </div>

          {/* Source location summary */}
          <LocationCard
            location={sourceLocation}
            roleLabel={t('transfer.sourceLocation')}
            color="#1677ff"
            background="#e6f7ff"
            borderColor="#91d5ff"
            compact
            t={t}
          />

          <div style={{
            background: '#fff7e6',
            borderRadius: 10,
            padding: 16,
            marginBottom: 16,
            marginTop: 16,
            border: '1px solid #ffd591',
          }}>
            <div style={{ fontSize: 14, fontWeight: 600, color: '#fa8c16', marginBottom: 8 }}>
              {t('transfer.scanDestLocation')}
            </div>
            <BarcodeScanner
              onScan={handleScanDest}
              placeholder={t('transfer.scanDestPlaceholder')}
            />
          </div>

          {/* Loading */}
          {isLoadingDest && (
            <div style={{ padding: '16px', textAlign: 'center', color: '#1677ff', fontSize: 14 }}>
              {'🔍'} {t('transfer.lookingUpLocation')}
            </div>
          )}

          {/* Error */}
          {(isDestError || (!isLoadingDest && destBarcode && !destLocation)) && (
            <div style={{
              marginTop: 16, padding: '14px 18px', background: '#fff1f0',
              border: '1px solid #ffa39e', borderRadius: 10, fontSize: 14, color: '#cf1322',
            }}>
              {destError
                ? t('transfer.lookupFailed', { error: (destError as Error)?.message })
                : t('transfer.locationNotFound', { barcode: destBarcode })}
            </div>
          )}

          {/* Same location warning */}
          {isSameLocation && (
            <div style={{
              marginTop: 16, padding: '14px 18px', background: '#fff7e6',
              border: '1px solid #ffd591', borderRadius: 10, fontSize: 14, color: '#ad6800',
            }}>
              {t('transfer.sameLocationError')}
            </div>
          )}

          {/* Found — show destination location card */}
          {destLocation && !isSameLocation && (
            <>
              <LocationCard
                location={destLocation}
                roleLabel={t('transfer.destLocation')}
                color="#389e0d"
                background="#f6ffed"
                borderColor="#b7eb8f"
                t={t}
              />

              <button
                onClick={handleConfirmDest}
                style={{
                  width: '100%', padding: '14px 0', fontSize: 15, fontWeight: 600,
                  color: '#fff', background: '#389e0d', border: 'none', borderRadius: 10,
                  cursor: 'pointer',
                }}
              >
                {t('transfer.confirmDest')}
              </button>
            </>
          )}
        </>
      )}

      {/* ── Step 3: Enter transfer details ─────────────────────────────── */}
      {step === 'details' && sourceLocation && destLocation && (
        <>
          <h3 style={{ fontSize: 15, fontWeight: 600, color: '#595959', marginBottom: 10 }}>
            {t('transfer.enterDetails')}
          </h3>

          {/* Step indicator */}
          <div style={{
            display: 'flex', gap: 6, marginBottom: 16, fontSize: 12, color: '#8c8c8c',
          }}>
            <StepBadge done label={t('transfer.stepSource')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge done label={t('transfer.stepDest')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge active label={t('transfer.stepDetails')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge label={t('transfer.stepConfirm')} />
          </div>

          {/* Transfer route summary */}
          <div style={{
            background: '#fff',
            borderRadius: 12,
            padding: 16,
            boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
            marginBottom: 16,
          }}>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 12,
              fontSize: 14,
            }}>
              <span style={{ fontWeight: 600, color: '#1677ff', fontFamily: 'monospace' }}>
                {sourceLocation.code || sourceLocation.barcode.substring(0, 8)}
              </span>
              <span style={{ fontSize: 18, color: '#8c8c8c' }}>→</span>
              <span style={{ fontWeight: 600, color: '#389e0d', fontFamily: 'monospace' }}>
                {destLocation.code || destLocation.barcode.substring(0, 8)}
              </span>
            </div>
          </div>

          {/* Source inventory summary */}
          {isLoadingSourceInv && (
            <div style={{ padding: '16px', textAlign: 'center', color: '#1677ff', fontSize: 14 }}>
              {'⏳'} {t('transfer.loadingSourceInventory')}
            </div>
          )}
          {isSourceInvError && (
            <div style={{
              marginBottom: 12, padding: '10px 14px', background: '#fff1f0',
              border: '1px solid #ffa39e', borderRadius: 8, fontSize: 13, color: '#cf1322', textAlign: 'center',
            }}>
              {t('transfer.sourceInventoryFailed', { error: (sourceInvError as Error)?.message })}
            </div>
          )}
          {!isLoadingSourceInv && !isSourceInvError && sourceInventory.length > 0 && (
            <div style={{
              background: '#fff',
              borderRadius: 12,
              padding: 16,
              boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              marginBottom: 16,
            }}>
              <div style={{ fontSize: 14, fontWeight: 600, color: '#595959', marginBottom: 12 }}>
                {t('transfer.sourceInventory', { count: sourceInventory.length })}
              </div>
              {sourceInventory.map((inv) => {
                const available = inv.available_qty ?? (inv.qty - (inv.reserved_qty || 0))
                return (
                  <div key={inv.id} style={{
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    padding: '10px 0', borderBottom: '1px solid #f0f0f0', fontSize: 13,
                  }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = '#fafafa' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                  >
                    <div>
                      <span style={{ fontFamily: 'monospace', fontWeight: 500, color: '#262626' }}>
                        {inv.sku_id.substring(0, 8)}...
                      </span>
                      {inv.batch_no && (
                        <span style={{ marginLeft: 8, color: '#8c8c8c', fontSize: 12 }}>
                          [{inv.batch_no}]
                        </span>
                      )}
                    </div>
                    <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                      <span style={{ color: '#8c8c8c' }}>
                        {t('transfer.available')}: <strong style={{ color: '#262626' }}>{available}</strong>
                      </span>
                      <span style={{
                        display: 'inline-block', padding: '2px 6px', fontSize: 11, fontWeight: 600,
                        color: getInventoryStatusColor(inv.status),
                        background: `${getInventoryStatusColor(inv.status)}15`,
                        borderRadius: 4,
                        border: `1px solid ${getInventoryStatusColor(inv.status)}40`,
                      }}>
                        {inv.status}
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          )}

          {/* SKU input section */}
          <div style={{
            background: '#fff',
            borderRadius: 12,
            padding: 20,
            boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
            marginBottom: 16,
          }}>
            <div style={{ fontSize: 14, fontWeight: 600, color: '#262626', marginBottom: 12 }}>
              {t('transfer.skuDetails')}
            </div>

            {/* SKU barcode scan */}
            <div style={{ marginBottom: 12 }}>
              <div style={{ fontSize: 12, color: '#8c8c8c', marginBottom: 6 }}>
                {t('transfer.scanSkuBarcode')}
              </div>
              <BarcodeScanner
                onScan={handleScanSku}
                placeholder={t('transfer.scanSkuPlaceholder')}
              />
            </div>

            {/* SKU code manual input */}
            <div style={{ marginBottom: 12 }}>
              <div style={{ fontSize: 12, color: '#8c8c8c', marginBottom: 6 }}>
                {t('transfer.orEnterSkuCode')}
              </div>
              <input
                type="text"
                value={skuInput}
                onChange={(e) => {
                  setSkuInput(e.target.value)
                  setSkuLookupBarcode('')
                }}
                placeholder={t('transfer.skuCodePlaceholder')}
                style={{
                  width: '100%', padding: '10px 14px', fontSize: 15,
                  border: '2px solid #d9d9d9', borderRadius: 8, outline: 'none',
                  boxSizing: 'border-box',
                }}
                onFocus={(e) => { e.currentTarget.style.borderColor = '#1677ff' }}
                onBlur={(e) => { e.currentTarget.style.borderColor = '#d9d9d9' }}
              />
            </div>

            {/* SKU loading / found / error */}
            {isLoadingSku && (
              <div style={{ padding: '8px', textAlign: 'center', color: '#1677ff', fontSize: 13 }}>
                {'🔍'} {t('transfer.lookingUpSku')}
              </div>
            )}
            {isSkuError && (
              <div style={{
                padding: '8px 12px', background: '#fff1f0',
                border: '1px solid #ffa39e', borderRadius: 6, fontSize: 12, color: '#cf1322',
              }}>
                {t('transfer.skuLookupFailed', { error: (skuError as Error)?.message })}
              </div>
            )}
            {(skuQueryParam && !isLoadingSku && !isSkuError && !sku) && (
              <div style={{
                padding: '8px 12px', background: '#fff7e6',
                border: '1px solid #ffd591', borderRadius: 6, fontSize: 12, color: '#ad6800',
              }}>
                {t('transfer.skuNotFound', { query: skuQueryParam })}
              </div>
            )}
            {sku && !isLoadingSku && (
              <div style={{
                padding: '12px 14px', background: '#f6ffed',
                border: '1px solid #b7eb8f', borderRadius: 8,
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 600, color: '#262626', fontFamily: 'monospace' }}>
                      {sku.code}
                    </div>
                    {sku.name && (
                      <div style={{ fontSize: 12, color: '#8c8c8c', marginTop: 2 }}>
                        {sku.name}
                      </div>
                    )}
                  </div>
                  <span style={{
                    display: 'inline-block', padding: '2px 8px', fontSize: 11, fontWeight: 600,
                    color: sku.status === 'active' ? '#389e0d' : '#8c8c8c',
                    background: sku.status === 'active' ? '#f6ffed' : '#fafafa',
                    borderRadius: 4,
                  }}>
                    {sku.status}
                  </span>
                </div>
              </div>
            )}

            {/* Batch number */}
            <div style={{ marginTop: 12 }}>
              <div style={{ fontSize: 12, color: '#8c8c8c', marginBottom: 6 }}>
                {t('transfer.batchNoOptional')}
              </div>
              <input
                type="text"
                value={batchNo}
                onChange={(e) => setBatchNo(e.target.value)}
                placeholder={t('transfer.batchNoPlaceholder')}
                style={{
                  width: '100%', padding: '10px 14px', fontSize: 15,
                  border: '2px solid #d9d9d9', borderRadius: 8, outline: 'none',
                  boxSizing: 'border-box',
                }}
                onFocus={(e) => { e.currentTarget.style.borderColor = '#1677ff' }}
                onBlur={(e) => { e.currentTarget.style.borderColor = '#d9d9d9' }}
              />
            </div>

            {/* Quantity */}
            <div style={{ marginTop: 12 }}>
              <div style={{ fontSize: 12, color: '#8c8c8c', marginBottom: 6 }}>
                {t('transfer.transferQty')}
              </div>
              <input
                type="number"
                inputMode="decimal"
                min="0"
                step="any"
                value={transferQty}
                onChange={(e) => setTransferQty(e.target.value)}
                placeholder={t('transfer.transferQtyPlaceholder')}
                style={{
                  width: '100%', padding: '10px 14px', fontSize: 15,
                  border: qtyError ? '2px solid #ff4d4f' : '2px solid #d9d9d9',
                  borderRadius: 8, outline: 'none', boxSizing: 'border-box',
                }}
                onFocus={(e) => { if (!qtyError) e.currentTarget.style.borderColor = '#1677ff' }}
                onBlur={(e) => { if (!qtyError) e.currentTarget.style.borderColor = '#d9d9d9' }}
              />
              {qtyError && (
                <div style={{ fontSize: 12, color: '#ff4d4f', marginTop: 4, paddingLeft: 2 }}>
                  {qtyError}
                </div>
              )}
            </div>
          </div>

          {/* Review button */}
          <button
            onClick={handleReviewTransfer}
            disabled={!sku || !transferQty || isNaN(parseFloat(transferQty)) || parseFloat(transferQty) <= 0}
            style={{
              width: '100%', padding: '14px 0', fontSize: 15, fontWeight: 600,
              color: '#fff',
              background: (!sku || !transferQty || isNaN(parseFloat(transferQty)) || parseFloat(transferQty) <= 0) ? '#d9d9d9' : '#1677ff',
              border: 'none', borderRadius: 10,
              cursor: (!sku || !transferQty || isNaN(parseFloat(transferQty)) || parseFloat(transferQty) <= 0) ? 'not-allowed' : 'pointer',
            }}
          >
            {t('transfer.review')}
          </button>
        </>
      )}

      {/* ── Step 4: Confirm & execute ──────────────────────────────────── */}
      {step === 'confirm' && sourceLocation && destLocation && sku && (
        <>
          <h3 style={{ fontSize: 15, fontWeight: 600, color: '#595959', marginBottom: 10 }}>
            {t('transfer.confirmTitle')}
          </h3>

          {/* Step indicator */}
          <div style={{
            display: 'flex', gap: 6, marginBottom: 16, fontSize: 12, color: '#8c8c8c',
          }}>
            <StepBadge done label={t('transfer.stepSource')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge done label={t('transfer.stepDest')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge done label={t('transfer.stepDetails')} />
            <span style={{ color: '#d9d9d9' }}>→</span>
            <StepBadge active label={t('transfer.stepConfirm')} />
          </div>

          {/* Transfer summary card */}
          <div style={{
            background: '#fff',
            borderRadius: 12,
            padding: 20,
            boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
            marginBottom: 16,
          }}>
            {/* Transfer route visual */}
            <div style={{
              padding: '14px 16px',
              background: '#f0f5ff',
              border: '1px solid #adc6ff',
              borderRadius: 8,
              marginBottom: 16,
              textAlign: 'center',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
                <div>
                  <div style={{ fontSize: 11, color: '#8c8c8c', marginBottom: 2 }}>{t('transfer.source')}</div>
                  <div style={{ fontSize: 15, fontWeight: 700, color: '#1677ff', fontFamily: 'monospace' }}>
                    {sourceLocation.code || sourceLocation.barcode.substring(0, 8)}
                  </div>
                </div>
                <span style={{ fontSize: 24, color: '#1677ff' }}>→</span>
                <div>
                  <div style={{ fontSize: 11, color: '#8c8c8c', marginBottom: 2 }}>{t('transfer.destination')}</div>
                  <div style={{ fontSize: 15, fontWeight: 700, color: '#389e0d', fontFamily: 'monospace' }}>
                    {destLocation.code || destLocation.barcode.substring(0, 8)}
                  </div>
                </div>
              </div>
            </div>

            {/* Details grid */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <DetailRow label={t('transfer.sourceLocation')} value={sourceLocation.code || sourceLocation.barcode} />
              <DetailRow label={t('transfer.destLocation')} value={destLocation.code || destLocation.barcode} />
              <DetailRow label={t('task.sku')} value={sku.code} />
              {sku.name && <DetailRow label={t('transfer.skuName')} value={sku.name} />}
              {batchNo && <DetailRow label={t('task.batch')} value={batchNo} />}
              <div style={{ borderTop: '1px dashed #d9d9d9', paddingTop: 10, marginTop: 4 }}>
                <DetailRow
                  label={t('transfer.transferQty')}
                  value={parseFloat(transferQty).toString()}
                  highlight
                />
              </div>
            </div>
          </div>

          {/* Confirm + back buttons */}
          <div style={{ display: 'flex', gap: 10 }}>
            <button
              onClick={() => setStep('details')}
              disabled={isProcessing}
              style={{
                flex: 1, padding: '14px 0', fontSize: 15, fontWeight: 500,
                color: '#595959', background: '#f0f0f0', border: 'none', borderRadius: 10,
                cursor: isProcessing ? 'not-allowed' : 'pointer',
              }}
            >
              {t('transfer.back')}
            </button>
            <button
              onClick={handleConfirmTransfer}
              disabled={isProcessing}
              style={{
                flex: 2, padding: '14px 0', fontSize: 15, fontWeight: 600,
                color: '#fff',
                background: isProcessing ? '#d9d9d9' : '#389e0d',
                border: 'none', borderRadius: 10,
                cursor: isProcessing ? 'not-allowed' : 'pointer',
                transition: 'background 0.2s',
              }}
            >
              {isProcessing ? t('transfer.processing') : t('transfer.confirmTransfer')}
            </button>
          </div>
        </>
      )}

      {/* ── Step 5: Completion ─────────────────────────────────────────── */}
      {step === 'done' && completedTask && (
        <div style={{
          padding: '32px 16px',
          textAlign: 'center',
          background: '#fff',
          borderRadius: 12,
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        }}>
          <div style={{ fontSize: 48, marginBottom: 16 }}>{'✅'}</div>
          <div style={{ fontSize: 16, fontWeight: 700, color: '#389e0d', marginBottom: 8 }}>
            {t('transfer.transferComplete')}
          </div>
          <div style={{ fontSize: 14, color: '#595959', marginBottom: 4 }}>
            {t('transfer.completedMessage', {
              qty: completedTask.expected_qty,
              from: sourceLocation?.code || sourceLocation?.barcode.substring(0, 8) || '',
              to: destLocation?.code || destLocation?.barcode.substring(0, 8) || '',
            })}
          </div>
          {completedTask.task_no && (
            <div style={{ fontSize: 12, color: '#8c8c8c', marginBottom: 24, fontFamily: 'monospace' }}>
              {t('transfer.taskNo')}: {completedTask.task_no}
            </div>
          )}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <button
              onClick={handleClear}
              style={{
                width: '100%', padding: '14px 0', fontSize: 15, fontWeight: 600,
                color: '#fff', background: '#1677ff', border: 'none', borderRadius: 10, cursor: 'pointer',
              }}
            >
              {t('transfer.newTransfer')}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

// ── Step badge helper ───────────────────────────────────────────────────

function StepBadge({ label, active, done }: { label: string; active?: boolean; done?: boolean }) {
  const bg = active ? '#1677ff' : done ? '#389e0d' : '#f0f0f0'
  const color = active || done ? '#fff' : '#8c8c8c'
  return (
    <span style={{
      display: 'inline-block', padding: '3px 8px', fontSize: 11, fontWeight: 600,
      color, background: bg, borderRadius: 4,
    }}>
      {label}
    </span>
  )
}

// ── Location card helper ────────────────────────────────────────────────

function LocationCard({
  location, roleLabel, color, background, borderColor, compact, t,
}: {
  location: LocationData
  roleLabel: string
  color: string
  background: string
  borderColor: string
  compact?: boolean
  t: (key: string) => string
}) {
  return (
    <div style={{
      background: '#fff',
      borderRadius: 12,
      padding: compact ? 14 : 20,
      boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
      marginBottom: 16,
      borderLeft: `4px solid ${color}`,
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginBottom: compact ? 8 : 12,
      }}>
        <div>
          <div style={{ fontSize: 11, color, fontWeight: 600, marginBottom: 2, textTransform: 'uppercase' }}>
            {roleLabel}
          </div>
          {!compact && (
            <h2 style={{ fontSize: 18, fontWeight: 700, color: '#262626', margin: '4px 0 0' }}>
              {location.code || location.barcode}
            </h2>
          )}
        </div>
        <span style={{
          display: 'inline-block', padding: '4px 10px', fontSize: 12, fontWeight: 600,
          color: '#fff', background: location.status === 'active' ? '#389e0d' : '#fa8c16',
          borderRadius: 4, whiteSpace: 'nowrap',
        }}>
          {location.status}
        </span>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        <DetailRow label={t('transfer.locationCode')} value={location.code || '—'} />
        <DetailRow label={t('transfer.locationBarcode')} value={location.barcode || '—'} />
        <DetailRow label={t('transfer.locationType')} value={location.location_type || '—'} />
      </div>
    </div>
  )
}

// ── Detail row helper ───────────────────────────────────────────────────

function DetailRow({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <span style={{ fontSize: 13, color: '#8c8c8c' }}>{label}</span>
      <span style={{
        fontSize: highlight ? 16 : 14,
        fontWeight: highlight ? 700 : 500,
        color: highlight ? '#1677ff' : '#262626',
        textAlign: 'right',
        maxWidth: '60%',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
      }}>
        {value}
      </span>
    </div>
  )
}
