// Replenishment page — replenishment workflow for warehouse operators.
// Operator scans a pick location (destination) that needs stock, views current
// inventory at that location, selects a SKU to replenish. The system finds the
// best reserve stock source via FEFO/FIFO. Operator confirms source/destination/qty
// and the system creates, starts, and completes a replenish task.
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
  location_type: string
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
  received_at?: string
  expiry_date?: string
  production_date?: string
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

function getStockLevelColor(qty: number, threshold = 5): string {
  if (qty <= 0) return '#cf1322'
  if (qty <= threshold) return '#fa8c16'
  return '#389e0d'
}

function getStockLevelLabel(qty: number, threshold = 5): 'low' | 'ok' | 'out' {
  if (qty <= 0) return 'out'
  if (qty <= threshold) return 'low'
  return 'ok'
}

// ── Main component ───────────────────────────────────────────────────────

export default function ReplenishmentPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ──────────────────────────────────────────────────────────────

  const [locationBarcode, setLocationBarcode] = useState('')
  const [selectedSkuId, setSelectedSkuId] = useState<string | null>(null)
  const [selectedSourceInv, setSelectedSourceInv] = useState<InventoryRecord | null>(null)
  const [replenishQty, setReplenishQty] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [completedMessage, setCompletedMessage] = useState('')

  // ── Step 1: Lookup pick location by barcode ────────────────────────────

  const {
    data: locationData,
    isLoading: isLookingUpLocation,
    isError: isLocationError,
    error: locationError,
  } = useQuery<{ data?: LocationData[] } & LocationData>({
    queryKey: ['replenLocation', locationBarcode],
    queryFn: async () => {
      const { data } = await client.get('/locations', {
        params: { barcode: locationBarcode },
      })
      return data
    },
    enabled: locationBarcode.length > 0,
    retry: false,
  })

  const location = (() => {
    if (!locationData) return null
    if ('data' in locationData && Array.isArray(locationData.data) && locationData.data.length === 1) {
      return locationData.data[0]
    }
    if (locationData.id && locationData.barcode) return locationData as LocationData
    return null
  })()

  // ── Step 2: Inventory at pick location ─────────────────────────────────

  const {
    data: pickInventoryData,
    isLoading: isLoadingPickInventory,
    isError: isPickInventoryError,
    error: pickInventoryError,
    refetch: refetchPickInventory,
  } = useQuery<ListResponse<InventoryRecord>>({
    queryKey: ['replenPickInventory', location?.id],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<InventoryRecord>>('/inventory', {
        params: { location_id: location!.id, status: 'available', page_size: 50 },
      })
      return data
    },
    enabled: !!location,
    retry: false,
  })

  const pickInventory = pickInventoryData?.data ?? []

  // ── Step 3: Reserve stock for selected SKU ─────────────────────────────

  const {
    data: reserveInventoryData,
    isLoading: isLoadingReserveInventory,
    isError: isReserveInventoryError,
    error: reserveInventoryError,
  } = useQuery<ListResponse<InventoryRecord>>({
    queryKey: ['replenReserveInventory', selectedSkuId, location?.id],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<InventoryRecord>>('/inventory', {
        params: {
          sku_id: selectedSkuId!,
          status: 'available',
          page_size: 20,
        },
      })
      return data
    },
    enabled: !!selectedSkuId && !!location,
    retry: false,
  })

  // Filter out inventory at the pick location (we need reserve locations only)
  const reserveInventory = (reserveInventoryData?.data ?? []).filter(
    (inv) => inv.location_id !== location?.id,
  )

  // ── Handlers ───────────────────────────────────────────────────────────

  const handleScanLocation = useCallback((barcode: string) => {
    setLocationBarcode(barcode)
    setSelectedSkuId(null)
    setSelectedSourceInv(null)
    setReplenishQty('')
    setIsProcessing(false)
    setCompletedMessage('')
  }, [])

  const handleClear = useCallback(() => {
    setLocationBarcode('')
    setSelectedSkuId(null)
    setSelectedSourceInv(null)
    setReplenishQty('')
    setIsProcessing(false)
    setCompletedMessage('')
    queryClient.removeQueries({ queryKey: ['replenLocation', locationBarcode] })
    queryClient.removeQueries({ queryKey: ['replenPickInventory', location?.id] })
  }, [locationBarcode, location?.id, queryClient])

  const handleSelectSku = useCallback((skuId: string) => {
    setSelectedSkuId(skuId)
    setSelectedSourceInv(null)
    setReplenishQty('')
  }, [])

  const handleBackToSkuList = useCallback(() => {
    setSelectedSkuId(null)
    setSelectedSourceInv(null)
    setReplenishQty('')
  }, [])

  const handleScanAnother = useCallback(() => {
    queryClient.removeQueries({ queryKey: ['replenLocation', locationBarcode] })
    queryClient.removeQueries({ queryKey: ['replenPickInventory', location?.id] })
    setLocationBarcode('')
    setSelectedSkuId(null)
    setSelectedSourceInv(null)
    setReplenishQty('')
    setIsProcessing(false)
    setCompletedMessage('')
  }, [locationBarcode, location?.id, queryClient])

  // ── Mutations: create → start → complete replenish task ────────────────

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
        task_type: 'replenish',
        warehouse_id: input.warehouseId,
        sku_id: input.skuId,
        from_location_id: input.fromLocationId,
        to_location_id: input.toLocationId,
        expected_qty: input.qty,
        uom: 'EA',
        batch_no: input.batchNo || undefined,
        priority: 'normal',
        instructions: t('replenish.taskInstructions', {
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

  const handleConfirmReplenish = async () => {
    if (!selectedSourceInv || !location) return
    const qty = parseFloat(replenishQty)
    if (isNaN(qty) || qty <= 0) return

    setIsProcessing(true)
    try {
      // 1. Create replenish task
      const task = await createTaskMutation.mutateAsync({
        fromLocationId: selectedSourceInv.location_id,
        toLocationId: location.id,
        skuId: selectedSkuId!,
        qty,
        warehouseId: selectedSourceInv.warehouse_id,
        batchNo: selectedSourceInv.batch_no || '',
      })

      // 2. Start task (pending → in_progress)
      await startTaskMutation.mutateAsync(task.id)

      // 3. Complete task with inventory transfer
      await completeTaskMutation.mutateAsync({
        taskId: task.id,
        qty,
        toLocationId: location.id,
      })

      // Success
      setCompletedMessage(
        t('replenish.completedMessage', {
          qty,
          location: location.code || location.barcode,
        }),
      )
      queryClient.invalidateQueries({ queryKey: ['replenPickInventory', location.id] })
    } catch {
      // Error handled by mutation state
    } finally {
      setIsProcessing(false)
    }
  }

  // ── Replenish qty validation ───────────────────────────────────────────

  let qtyError: string | null = null
  if (selectedSourceInv && replenishQty) {
    const parsed = parseFloat(replenishQty)
    if (!isNaN(parsed)) {
      const available = selectedSourceInv.available_qty ?? (selectedSourceInv.qty - (selectedSourceInv.reserved_qty || 0))
      if (parsed > available) {
        qtyError = t('replenish.qtyExceedsAvailable', {
          requested: replenishQty,
          available: available.toString(),
        })
      } else if (parsed <= 0) {
        qtyError = t('replenish.qtyMustBePositive')
      }
    }
  }

  // Determine if mutating
  const anyMutationError =
    createTaskMutation.isError || startTaskMutation.isError || completeTaskMutation.isError
  const mutationErrorMessage =
    (createTaskMutation.error as Error)?.message ||
    (startTaskMutation.error as Error)?.message ||
    (completeTaskMutation.error as Error)?.message

  // ── Render ─────────────────────────────────────────────────────────────

  return (
    <div>
      {/* Step 0: Scan location — shown when no location scanned */}
      {!location && (
        <>
          <h3
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: '#595959',
              marginBottom: 10,
            }}
          >
            {t('replenish.title')}
          </h3>
          <p
            style={{
              fontSize: 13,
              color: '#8c8c8c',
              marginBottom: 16,
              lineHeight: 1.5,
            }}
          >
            {t('replenish.description')}
          </p>

          <BarcodeScanner
            onScan={handleScanLocation}
            placeholder={t('replenish.scanLocationPlaceholder')}
          />

          {/* Lookup in progress */}
          {isLookingUpLocation && (
            <div
              style={{
                padding: '16px',
                textAlign: 'center',
                color: '#1677ff',
                fontSize: 14,
                marginTop: 16,
              }}
            >
              {'🔍'} {t('replenish.lookingUpLocation')}
            </div>
          )}

          {/* Location not found or error */}
          {(isLocationError ||
            (!isLookingUpLocation && locationBarcode && !location)) && (
            <div
              style={{
                marginTop: 16,
                padding: '14px 18px',
                background: '#fff1f0',
                border: '1px solid #ffa39e',
                borderRadius: 10,
                fontSize: 14,
                color: '#cf1322',
              }}
            >
              {locationError
                ? t('replenish.lookupFailed', { error: (locationError as Error)?.message })
                : t('replenish.locationNotFound', { barcode: locationBarcode })}
            </div>
          )}
        </>
      )}

      {/* Location confirmed → inventory view */}
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
            {t('replenish.backToScan')}
          </button>

          {/* Location info card */}
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
                  {t('replenish.pickLocation')}
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
              <DetailRow label={t('replenish.locationCode')} value={location.code || '—'} />
              <DetailRow label={t('replenish.locationBarcode')} value={location.barcode || '—'} />
              <DetailRow label={t('replenish.locationType')} value={location.location_type || '—'} />
            </div>
          </div>

          {/* Error: mutation error */}
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
              {t('replenish.processError', { error: mutationErrorMessage || '' })}
            </div>
          )}

          {/* Step 1: Pick location inventory — SKU list */}
          {!selectedSkuId && (
            <>
              {/* Loading state */}
              {isLoadingPickInventory && (
                <div className="pda-empty">
                  <span className="empty-icon">{'⏳'}</span>
                  <p className="empty-text">{t('replenish.loadingInventory')}</p>
                </div>
              )}

              {/* Error state */}
              {isPickInventoryError && !isLoadingPickInventory && (
                <div className="pda-empty">
                  <span className="empty-icon">{'⚠️'}</span>
                  <p className="empty-text">
                    {t('replenish.inventoryLoadFailed', {
                      error: (pickInventoryError as Error)?.message,
                    })}
                  </p>
                  <button
                    onClick={() => refetchPickInventory()}
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
              {!isLoadingPickInventory && !isPickInventoryError && pickInventory.length === 0 && (
                <div
                  style={{
                    padding: '32px 16px',
                    textAlign: 'center',
                    background: '#fff',
                    borderRadius: 12,
                    boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                  }}
                >
                  <div style={{ fontSize: 40, marginBottom: 12 }}>{'📦'}</div>
                  <div style={{ fontSize: 15, fontWeight: 600, color: '#595959', marginBottom: 8 }}>
                    {t('replenish.noInventory')}
                  </div>
                  <div style={{ fontSize: 13, color: '#8c8c8c', marginBottom: 20 }}>
                    {t('replenish.noInventoryDesc')}
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
                    {t('replenish.scanAnother')}
                  </button>
                </div>
              )}

              {/* Inventory list — SKU cards at pick location */}
              {!isLoadingPickInventory && !isPickInventoryError && pickInventory.length > 0 && (
                <>
                  <div
                    style={{
                      fontSize: 14,
                      fontWeight: 600,
                      color: '#595959',
                      marginBottom: 10,
                    }}
                  >
                    {t('replenish.inventoryAtLocation', { count: pickInventory.length })}
                  </div>

                  {pickInventory.map((inv) => {
                    const available = inv.available_qty ?? (inv.qty - (inv.reserved_qty || 0))
                    const level = getStockLevelLabel(available)
                    const stockColor = getStockLevelColor(available)

                    return (
                      <div
                        key={inv.id}
                        onClick={() => handleSelectSku(inv.sku_id)}
                        style={{
                          background: '#fff',
                          borderRadius: 12,
                          padding: 16,
                          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                          marginBottom: 12,
                          cursor: 'pointer',
                          border: '2px solid transparent',
                          transition: 'border-color 0.2s, box-shadow 0.2s',
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
                                fontSize: 14,
                                fontWeight: 600,
                                color: '#262626',
                                marginBottom: 2,
                                fontFamily: 'monospace',
                              }}
                            >
                              {t('replenish.skuId')}: {inv.sku_id.substring(0, 8)}...
                            </div>
                            {inv.batch_no && (
                              <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                                {t('task.batch')}: {inv.batch_no}
                              </div>
                            )}
                          </div>
                          <div style={{ textAlign: 'right' }}>
                            <span
                              style={{
                                display: 'inline-block',
                                padding: '3px 8px',
                                fontSize: 11,
                                fontWeight: 600,
                                color: stockColor,
                                background: `${stockColor}15`,
                                borderRadius: 4,
                                border: `1px solid ${stockColor}40`,
                                marginBottom: 4,
                              }}
                            >
                              {level === 'out'
                                ? t('replenish.stockOut')
                                : level === 'low'
                                  ? t('replenish.stockLow')
                                  : t('replenish.stockOk')}
                            </span>
                            {(level === 'low' || level === 'out') && (
                              <div style={{ fontSize: 11, color: '#fa8c16' }}>
                                {t('replenish.needsReplenish')}
                              </div>
                            )}
                          </div>
                        </div>

                        <div
                          style={{
                            display: 'grid',
                            gridTemplateColumns: '1fr 1fr 1fr',
                            gap: 6,
                            fontSize: 13,
                            color: '#595959',
                          }}
                        >
                          <div>
                            <span style={{ color: '#8c8c8c' }}>{t('replenish.qty')}: </span>
                            <span style={{ fontWeight: 600, color: '#262626' }}>{inv.qty}</span>
                          </div>
                          <div>
                            <span style={{ color: '#8c8c8c' }}>{t('replenish.reserved')}: </span>
                            <span style={{ fontWeight: 600, color: '#fa8c16' }}>{inv.reserved_qty || 0}</span>
                          </div>
                          <div>
                            <span style={{ color: '#8c8c8c' }}>{t('replenish.available')}: </span>
                            <span
                              style={{ fontWeight: 600, color: stockColor }}
                            >
                              {available}
                            </span>
                          </div>
                        </div>
                      </div>
                    )
                  })}

                  {/* Scan another button */}
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
                      {t('replenish.scanAnother')}
                    </button>
                  </div>
                </>
              )}
            </>
          )}

          {/* Step 2: SKU selected → show reserve stock options */}
          {selectedSkuId && !completedMessage && (
            <>
              {/* Back to SKU list */}
              <button
                onClick={handleBackToSkuList}
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
                {t('replenish.backToSkuList')}
              </button>

              {/* Selected SKU info card */}
              <div
                style={{
                  background: '#e6f7ff',
                  borderRadius: 10,
                  padding: 16,
                  marginBottom: 16,
                  border: '1px solid #91d5ff',
                }}
              >
                <div style={{ fontSize: 13, color: '#1677ff', marginBottom: 4 }}>
                  {t('replenish.replenishForSku')}
                </div>
                <div style={{ fontSize: 14, fontWeight: 700, color: '#262626', fontFamily: 'monospace' }}>
                  {selectedSkuId}
                </div>
              </div>

              {/* Loading reserve stock */}
              {isLoadingReserveInventory && (
                <div className="pda-empty">
                  <span className="empty-icon">{'⏳'}</span>
                  <p className="empty-text">{t('replenish.findingReserveStock')}</p>
                </div>
              )}

              {/* Error loading reserve stock */}
              {isReserveInventoryError && (
                <div className="pda-empty">
                  <span className="empty-icon">{'⚠️'}</span>
                  <p className="empty-text">
                    {t('replenish.reserveStockLoadFailed', {
                      error: (reserveInventoryError as Error)?.message,
                    })}
                  </p>
                </div>
              )}

              {/* No reserve stock found */}
              {!isLoadingReserveInventory &&
                !isReserveInventoryError &&
                reserveInventory.length === 0 && (
                  <div
                    style={{
                      padding: '32px 16px',
                      textAlign: 'center',
                      background: '#fff',
                      borderRadius: 12,
                      boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                    }}
                  >
                    <div style={{ fontSize: 40, marginBottom: 12 }}>{'🔍'}</div>
                    <div style={{ fontSize: 15, fontWeight: 600, color: '#595959', marginBottom: 8 }}>
                      {t('replenish.noReserveStock')}
                    </div>
                    <div style={{ fontSize: 13, color: '#8c8c8c' }}>
                      {t('replenish.noReserveStockDesc')}
                    </div>
                  </div>
                )}

              {/* Reserve stock list — source location options */}
              {!isLoadingReserveInventory &&
                !isReserveInventoryError &&
                reserveInventory.length > 0 &&
                !selectedSourceInv && (
                  <>
                    <div
                      style={{
                        fontSize: 14,
                        fontWeight: 600,
                        color: '#595959',
                        marginBottom: 10,
                      }}
                    >
                      {t('replenish.selectSourceLocation', { count: reserveInventory.length })}
                    </div>

                    {reserveInventory.map((inv) => {
                      const available = inv.available_qty ?? (inv.qty - (inv.reserved_qty || 0))
                      return (
                        <div
                          key={inv.id}
                          onClick={() => {
                            setSelectedSourceInv(inv)
                            setReplenishQty('')
                          }}
                          style={{
                            background: '#fff',
                            borderRadius: 12,
                            padding: 16,
                            boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                            marginBottom: 12,
                            cursor: 'pointer',
                            border: '2px solid transparent',
                            transition: 'border-color 0.2s, box-shadow 0.2s',
                          }}
                          onMouseEnter={(e) => {
                            e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.12)'
                            e.currentTarget.style.borderColor = '#389e0d'
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
                              <div style={{ fontSize: 14, fontWeight: 600, color: '#262626' }}>
                                {t('replenish.sourceLocation')}:{' '}
                                <span style={{ fontFamily: 'monospace' }}>
                                  {inv.location_id.substring(0, 8)}...
                                </span>
                              </div>
                              {inv.batch_no && (
                                <div style={{ fontSize: 12, color: '#8c8c8c', marginTop: 2 }}>
                                  {t('task.batch')}: {inv.batch_no}
                                </div>
                              )}
                            </div>
                            <span
                              style={{
                                display: 'inline-block',
                                padding: '3px 8px',
                                fontSize: 11,
                                fontWeight: 600,
                                color: getInventoryStatusColor(inv.status),
                                background: `${getInventoryStatusColor(inv.status)}15`,
                                borderRadius: 4,
                                border: `1px solid ${getInventoryStatusColor(inv.status)}40`,
                              }}
                            >
                              {inv.status}
                            </span>
                          </div>

                          <div
                            style={{
                              display: 'grid',
                              gridTemplateColumns: '1fr 1fr 1fr',
                              gap: 6,
                              fontSize: 13,
                              color: '#595959',
                            }}
                          >
                            <div>
                              <span style={{ color: '#8c8c8c' }}>{t('replenish.qty')}: </span>
                              <span style={{ fontWeight: 600, color: '#262626' }}>{inv.qty}</span>
                            </div>
                            <div>
                              <span style={{ color: '#8c8c8c' }}>{t('replenish.reserved')}: </span>
                              <span style={{ fontWeight: 600, color: '#fa8c16' }}>{inv.reserved_qty || 0}</span>
                            </div>
                            <div>
                              <span style={{ color: '#8c8c8c' }}>{t('replenish.available')}: </span>
                              <span style={{ fontWeight: 600, color: '#389e0d' }}>{available}</span>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </>
                )}

              {/* Step 3: Source selected → enter qty + confirm */}
              {selectedSourceInv && (
                <>
                  {/* Transfer summary card */}
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
                        marginBottom: 16,
                      }}
                    >
                      {t('replenish.confirmReplenish')}
                    </h3>

                    {/* Source → Destination */}
                    <div
                      style={{
                        padding: '14px 16px',
                        background: '#f6ffed',
                        border: '1px solid #b7eb8f',
                        borderRadius: 8,
                        marginBottom: 12,
                      }}
                    >
                      <div style={{ marginBottom: 8 }}>
                        <strong>{t('replenish.transferSummary')}</strong>
                      </div>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
                        <DetailRow
                          label={t('replenish.fromReserve')}
                          value={`${selectedSourceInv.location_id.substring(0, 8)}...`}
                        />
                        <DetailRow
                          label={t('replenish.toPick')}
                          value={`${location.code} (${location.barcode})`}
                        />
                        <DetailRow label={t('task.sku')} value={`${selectedSkuId.substring(0, 8)}...`} />
                        {selectedSourceInv.batch_no && (
                          <DetailRow label={t('task.batch')} value={selectedSourceInv.batch_no} />
                        )}
                      </div>
                    </div>

                    {/* Available indicator */}
                    <div
                      style={{
                        padding: '8px 14px',
                        background: '#e6f7ff',
                        borderRadius: 6,
                        fontSize: 13,
                        color: '#1677ff',
                        textAlign: 'center',
                        marginBottom: 12,
                      }}
                    >
                      {t('replenish.availableAtSource', {
                        qty: selectedSourceInv.available_qty ?? (selectedSourceInv.qty - (selectedSourceInv.reserved_qty || 0)),
                      })}
                    </div>

                    {/* Qty input */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                      <div style={{ display: 'flex', gap: 8, width: '100%' }}>
                        <input
                          type="number"
                          inputMode="decimal"
                          min="0"
                          step="any"
                          value={replenishQty}
                          onChange={(e) => setReplenishQty(e.target.value)}
                          placeholder={t('replenish.enterReplenishQty')}
                          disabled={isProcessing}
                          style={{
                            flex: 1,
                            padding: '10px 14px',
                            fontSize: 15,
                            border: qtyError
                              ? '2px solid #ff4d4f'
                              : '2px solid #d9d9d9',
                            borderRadius: 8,
                            outline: 'none',
                            background: isProcessing ? '#f5f5f5' : '#fff',
                          }}
                          onFocus={(e) => {
                            if (!qtyError) e.currentTarget.style.borderColor = '#1677ff'
                          }}
                          onBlur={(e) => {
                            if (!qtyError) e.currentTarget.style.borderColor = '#d9d9d9'
                          }}
                        />
                        <button
                          onClick={handleConfirmReplenish}
                          disabled={
                            isProcessing ||
                            !replenishQty ||
                            isNaN(parseFloat(replenishQty)) ||
                            parseFloat(replenishQty) <= 0 ||
                            !!qtyError
                          }
                          style={{
                            padding: '10px 18px',
                            fontSize: 14,
                            fontWeight: 600,
                            color: '#fff',
                            background:
                              isProcessing || !replenishQty || isNaN(parseFloat(replenishQty)) ||
                              parseFloat(replenishQty) <= 0 || !!qtyError
                                ? '#d9d9d9'
                                : '#389e0d',
                            border: 'none',
                            borderRadius: 8,
                            cursor:
                              isProcessing || !replenishQty || isNaN(parseFloat(replenishQty)) ||
                              parseFloat(replenishQty) <= 0 || !!qtyError
                                ? 'not-allowed'
                                : 'pointer',
                            whiteSpace: 'nowrap',
                            transition: 'background 0.2s',
                          }}
                        >
                          {isProcessing ? t('replenish.processing') : t('replenish.confirm')}
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

          {/* Step 4: Completion message */}
          {completedMessage && (
            <div
              style={{
                padding: '32px 16px',
                textAlign: 'center',
                background: '#fff',
                borderRadius: 12,
                boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              }}
            >
              <div style={{ fontSize: 48, marginBottom: 16 }}>{'✅'}</div>
              <div style={{ fontSize: 16, fontWeight: 700, color: '#389e0d', marginBottom: 8 }}>
                {t('replenish.replenishComplete')}
              </div>
              <div style={{ fontSize: 14, color: '#595959', marginBottom: 24 }}>
                {completedMessage}
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <button
                  onClick={() => {
                    setSelectedSkuId(null)
                    setSelectedSourceInv(null)
                    setReplenishQty('')
                    setIsProcessing(false)
                    setCompletedMessage('')
                    refetchPickInventory()
                  }}
                  style={{
                    width: '100%',
                    padding: '14px 0',
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#fff',
                    background: '#1677ff',
                    border: 'none',
                    borderRadius: 10,
                    cursor: 'pointer',
                  }}
                >
                  {t('replenish.replenishAnotherSku')}
                </button>
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
                  {t('replenish.scanAnotherLocation')}
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

// ── Detail row helper ───────────────────────────────────────────────────

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
