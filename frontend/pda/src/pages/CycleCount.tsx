// Cycle Count workflow page — PDA operator workflow for physical inventory counting.
// Workflow: scan location → review inventory lines → enter counted qty per line →
// submit for supervisor review. Supports camera barcode scanning and keyboard input.
// All UI text is translated via react-i18next.

import { useState, useCallback, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import BarcodeScanner from '@/components/BarcodeScanner'
import client from '@/api/client'
import type { CycleCount, CycleCountLine, CycleCountStatus } from '@/api/cycleCount'

// ── Status labels & colors ─────────────────────────────────────────────

function getStatusLabelKey(status: CycleCountStatus): string {
  const map: Record<CycleCountStatus, string> = {
    draft: 'cycleCount.statusDraft',
    in_progress: 'cycleCount.statusInProgress',
    pending_review: 'cycleCount.statusPendingReview',
    approved: 'cycleCount.statusApproved',
    adjusted: 'cycleCount.statusAdjusted',
    cancelled: 'cycleCount.statusCancelled',
  }
  return map[status] || 'cycleCount.statusDraft'
}

function getStatusColor(status: CycleCountStatus): string {
  const map: Record<CycleCountStatus, string> = {
    draft: '#bfbfbf',
    in_progress: '#1677ff',
    pending_review: '#fa8c16',
    approved: '#389e0d',
    adjusted: '#722ed1',
    cancelled: '#cf1322',
  }
  return map[status] || '#bfbfbf'
}

type WorkflowStep = 'scan' | 'counting' | 'summary'

// ── Main component ──────────────────────────────────────────────────────

export default function CycleCountPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [step, setStep] = useState<WorkflowStep>('scan')
  // Location-based cycle count: scan a location barcode
  const [locationId, setLocationId] = useState('')
  const [warehouseId, setWarehouseId] = useState('')
  const [scanError, setScanError] = useState<string | null>(null)
  const [isStarting, setIsStarting] = useState(false)

  // Active cycle count
  const [activeCountId, setActiveCountId] = useState<string | null>(null)
  const [countedQtys, setCountedQtys] = useState<Record<string, string>>({})
  const [submittingLines, setSubmittingLines] = useState<Set<string>>(new Set())
  const [submitError, setSubmitError] = useState<string | null>(null)

  // ── Lookup location by barcode ────────────────────────────────────────

  const { data: locationInfo, isFetching: isLookingUp } = useQuery<{
    id: string
    code: string
    warehouse_id: string
  }>({
    queryKey: ['locationLookup', locationId],
    queryFn: async () => {
      const { data } = await client.get('/locations', { params: { barcode: locationId } })
      return data as { id: string; code: string; warehouse_id: string }
    },
    enabled: locationId.length > 0,
    retry: false,
  })

  // ── Fetch active cycle count ──────────────────────────────────────────

  const {
    data: cycleCount,
    isLoading: isLoadingCount,
    isError: isCountError,
    error: countError,
    refetch: refetchCount,
  } = useQuery<CycleCount>({
    queryKey: ['cycleCount', activeCountId],
    queryFn: async () => {
      const { data } = await client.get<CycleCount>(`/cycle-counts/${activeCountId}`)
      return data
    },
    enabled: !!activeCountId,
    retry: false,
  })

  // ── Start cycle count mutation ────────────────────────────────────────

  const handleStartCount = useCallback(async () => {
    if (!locationInfo) return

    setIsStarting(true)
    setScanError(null)

    try {
      const { data } = await client.post<CycleCount>('/cycle-counts', {
        warehouse_id: locationInfo.warehouse_id,
        location_id: locationInfo.id,
      })
      setActiveCountId(data.id)
      setStep('counting')
    } catch (err: unknown) {
      const msg = (err as Error)?.message || String(err)
      setScanError(t('cycleCount.startFailed', { error: msg }))
    } finally {
      setIsStarting(false)
    }
  }, [locationInfo, t])

  // ── Barcode scan handler ──────────────────────────────────────────────

  const handleScan = useCallback((barcode: string) => {
    setLocationId(barcode)
    setScanError(null)
  }, [])

  // ── Submit single line ────────────────────────────────────────────────

  const handleSubmitLine = useCallback(async (lineId: string) => {
    const qtyStr = countedQtys[lineId]
    if (!qtyStr) return
    const qty = parseFloat(qtyStr)
    if (isNaN(qty) || qty < 0) return

    setSubmitError(null)
    setSubmittingLines((prev) => new Set(prev).add(lineId))

    try {
      await client.post(`/cycle-counts/${activeCountId}/lines`, {
        line_id: lineId,
        counted_qty: qty,
      })
      setCountedQtys((prev) => ({ ...prev, [lineId]: '' }))
      // Refresh to get updated statuses
      queryClient.invalidateQueries({ queryKey: ['cycleCount', activeCountId] })
    } catch (err: unknown) {
      const msg = (err as Error)?.message || String(err)
      setSubmitError(t('cycleCount.submitLineError', { error: msg }))
    } finally {
      setSubmittingLines((prev) => {
        const next = new Set(prev)
        next.delete(lineId)
        return next
      })
    }
  }, [countedQtys, activeCountId, queryClient, t])

  // ── Finalize count ────────────────────────────────────────────────────

  const [isFinalizing, setIsFinalizing] = useState(false)

  const handleFinalize = useCallback(async () => {
    if (!activeCountId) return
    setIsFinalizing(true)
    setSubmitError(null)

    try {
      await client.post(`/cycle-counts/${activeCountId}/finalize`)
      setStep('summary')
      queryClient.invalidateQueries({ queryKey: ['cycleCount', activeCountId] })
    } catch (err: unknown) {
      const msg = (err as Error)?.message || String(err)
      setSubmitError(t('cycleCount.finalizeError', { error: msg }))
    } finally {
      setIsFinalizing(false)
    }
  }, [activeCountId, queryClient, t])

  // ── Clear & reset ─────────────────────────────────────────────────────

  const handleClear = useCallback(() => {
    if (activeCountId) {
      queryClient.removeQueries({ queryKey: ['cycleCount', activeCountId] })
    }
    setLocationId('')
    setWarehouseId('')
    setScanError(null)
    setActiveCountId(null)
    setCountedQtys({})
    setSubmittingLines(new Set())
    setSubmitError(null)
    setStep('scan')
  }, [activeCountId, queryClient])

  // ── Derived values ────────────────────────────────────────────────────

  const lines = cycleCount?.lines || []
  const totalCounted = lines.filter((l) => l.status === 'counted').length
  const allCounted = lines.length > 0 && totalCounted === lines.length

  // ── Render ────────────────────────────────────────────────────────────

  return (
    <div>
      {/* ── Step 1: Scan location ──────────────────────────────────────── */}
      {step === 'scan' && (
        <>
          <h3
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: '#595959',
              marginBottom: 10,
            }}
          >
            {t('cycleCount.title')}
          </h3>
          <p
            style={{
              fontSize: 13,
              color: '#8c8c8c',
              marginBottom: 16,
              lineHeight: 1.5,
            }}
          >
            {t('cycleCount.scanDescription')}
          </p>

          <BarcodeScanner
            onScan={handleScan}
            placeholder={t('cycleCount.scanLocationPlaceholder')}
          />

          {/* Looking up location */}
          {isLookingUp && (
            <div
              style={{
                padding: '16px',
                textAlign: 'center',
                color: '#1677ff',
                fontSize: 14,
                marginTop: 16,
              }}
            >
              {'\u{1F50D}'} {t('cycleCount.lookingUpLocation')}
            </div>
          )}

          {/* Location found */}
          {locationInfo && !isLookingUp && (
            <div
              style={{
                marginTop: 16,
                background: '#fff',
                borderRadius: 12,
                padding: 20,
                boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
                <span style={{ fontSize: 28 }}>{'\u{1F4CD}'}</span>
                <div>
                  <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                    {t('cycleCount.locationFound')}
                  </div>
                  <div style={{ fontSize: 16, fontWeight: 600, color: '#262626' }}>
                    {locationInfo.code}
                  </div>
                </div>
              </div>

              <button
                onClick={handleStartCount}
                disabled={isStarting}
                style={{
                  width: '100%',
                  padding: '14px 0',
                  fontSize: 16,
                  fontWeight: 600,
                  color: '#fff',
                  background: isStarting ? '#91caff' : '#1677ff',
                  border: 'none',
                  borderRadius: 10,
                  cursor: isStarting ? 'not-allowed' : 'pointer',
                }}
              >
                {isStarting ? t('cycleCount.startingCount') : t('cycleCount.startCount')}
              </button>
            </div>
          )}

          {/* Scan error */}
          {scanError && (
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
              {scanError}
            </div>
          )}

          {/* Back to scan hub */}
          <button
            onClick={() => navigate('/scan')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              padding: '10px 0',
              marginTop: 16,
              fontSize: 14,
              color: '#1677ff',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
            }}
          >
            {t('cycleCount.backToScan')}
          </button>
        </>
      )}

      {/* ── Step 2: Counting lines ─────────────────────────────────────── */}
      {step === 'counting' && (
        <>
          {/* Back & info header */}
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
            {t('cycleCount.cancelAndReturn')}
          </button>

          {/* Count info card */}
          {cycleCount && (
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
                  <h2 style={{ fontSize: 18, fontWeight: 700, color: '#262626', marginBottom: 4 }}>
                    {cycleCount.count_no}
                  </h2>
                </div>
                <span
                  style={{
                    display: 'inline-block',
                    padding: '4px 10px',
                    fontSize: 12,
                    fontWeight: 600,
                    color: '#fff',
                    background: getStatusColor(cycleCount.status as CycleCountStatus),
                    borderRadius: 4,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {t(getStatusLabelKey(cycleCount.status as CycleCountStatus))}
                </span>
              </div>

              {/* Progress bar */}
              <div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    marginBottom: 6,
                    fontSize: 13,
                    color: '#595959',
                  }}
                >
                  <span>
                    {t('cycleCount.linesCounted', {
                      counted: totalCounted,
                      total: lines.length,
                    })}
                  </span>
                  <span style={{ fontWeight: 600 }}>
                    {lines.length > 0 ? Math.round((totalCounted / lines.length) * 100) : 0}%
                  </span>
                </div>
                <div
                  style={{
                    height: 8,
                    background: '#f0f0f0',
                    borderRadius: 4,
                    overflow: 'hidden',
                  }}
                >
                  <div
                    style={{
                      height: '100%',
                      width: `${lines.length > 0 ? (totalCounted / lines.length) * 100 : 0}%`,
                      background: allCounted ? '#389e0d' : '#1677ff',
                      borderRadius: 4,
                      transition: 'width 0.3s ease',
                    }}
                  />
                </div>
              </div>
            </div>
          )}

          {/* Loading */}
          {isLoadingCount && (
            <div className="pda-empty">
              <span className="empty-icon">{'⏳'}</span>
              <p className="empty-text">{t('cycleCount.loadingCount')}</p>
            </div>
          )}

          {/* Error */}
          {isCountError && (
            <div className="pda-empty">
              <span className="empty-icon">{'⚠️'}</span>
              <p className="empty-text">
                {t('cycleCount.loadFailed', {
                  error: (countError as Error)?.message,
                })}
              </p>
              <button
                onClick={() => refetchCount()}
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

          {/* Submit error */}
          {submitError && (
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
              {submitError}
            </div>
          )}

          {/* Line items */}
          {!isLoadingCount && !isCountError && lines.length > 0 && (
            <>
              {lines.map((line) => {
                const isCompleted = line.status === 'counted'
                const currentQty = countedQtys[line.id] || ''
                const isSubmitting = submittingLines.has(line.id)

                return (
                  <LineItem
                    key={line.id}
                    line={line}
                    isCompleted={isCompleted}
                    currentQty={currentQty}
                    isSubmitting={isSubmitting}
                    onQtyChange={(qty) =>
                      setCountedQtys((prev) => ({ ...prev, [line.id]: qty }))
                    }
                    onSubmit={() => handleSubmitLine(line.id)}
                    t={t}
                  />
                )
              })}
            </>
          )}

          {/* Finalize button */}
          {!isLoadingCount && !isCountError && lines.length > 0 && (
            <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
              <button
                onClick={handleFinalize}
                disabled={isFinalizing || !allCounted || lines.length === 0}
                style={{
                  width: '100%',
                  padding: '14px 0',
                  fontSize: 16,
                  fontWeight: 600,
                  color: '#fff',
                  background:
                    isFinalizing || !allCounted ? '#d9d9d9' : '#1677ff',
                  border: 'none',
                  borderRadius: 10,
                  cursor:
                    isFinalizing || !allCounted ? 'not-allowed' : 'pointer',
                }}
              >
                {isFinalizing
                  ? t('cycleCount.finalizing')
                  : t('cycleCount.finalizeCount')}
              </button>
              {!allCounted && lines.length > 0 && (
                <div style={{ fontSize: 12, color: '#8c8c8c', textAlign: 'center' }}>
                  {t('cycleCount.mustCountAllLines')}
                </div>
              )}
            </div>
          )}
        </>
      )}

      {/* ── Step 3: Summary (pending review) ────────────────────────────── */}
      {step === 'summary' && (
        <>
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
            {t('cycleCount.backToScan')}
          </button>

          {/* Summary card */}
          <div
            style={{
              background: '#fff',
              borderRadius: 12,
              padding: 24,
              boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              marginBottom: 16,
              textAlign: 'center',
            }}
          >
            <div style={{ fontSize: 48, marginBottom: 12 }}>{'✅'}</div>
            <h2 style={{ fontSize: 18, fontWeight: 700, color: '#262626', marginBottom: 8 }}>
              {t('cycleCount.countSubmittedTitle')}
            </h2>
            <p style={{ fontSize: 14, color: '#595959', marginBottom: 20, lineHeight: 1.6 }}>
              {t('cycleCount.countSubmittedMessage')}
            </p>

            {/* Stats grid */}
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(2, 1fr)',
                gap: 12,
                marginBottom: 20,
              }}
            >
              <SummaryTile
                label={t('cycleCount.totalLines')}
                value={cycleCount?.total_lines || lines.length}
                color="#1677ff"
              />
              <SummaryTile
                label={t('cycleCount.matchedLines')}
                value={
                  lines.filter(
                    (l) => l.variance !== undefined && l.variance === 0
                  ).length
                }
                color="#389e0d"
              />
            </div>

            <button
              onClick={handleClear}
              style={{
                width: '100%',
                padding: '14px 0',
                fontSize: 16,
                fontWeight: 600,
                color: '#fff',
                background: '#1677ff',
                border: 'none',
                borderRadius: 10,
                cursor: 'pointer',
              }}
            >
              {t('cycleCount.startNewCount')}
            </button>
          </div>

          <button
            onClick={() => navigate('/scan')}
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
            {t('cycleCount.backToScan')}
          </button>
        </>
      )}
    </div>
  )
}

// ── Line item component ─────────────────────────────────────────────────

function LineItem({
  line,
  isCompleted,
  currentQty,
  isSubmitting,
  onQtyChange,
  onSubmit,
  t,
}: {
  line: CycleCountLine
  isCompleted: boolean
  currentQty: string
  isSubmitting: boolean
  onQtyChange: (qty: string) => void
  onSubmit: () => void
  t: (key: string, opts?: Record<string, unknown>) => string
}) {
  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    onSubmit()
  }

  const parsedQty = parseFloat(currentQty)
  const hasValidQty = currentQty && !isNaN(parsedQty) && parsedQty >= 0

  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 12,
        padding: 16,
        boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        marginBottom: 12,
        opacity: isCompleted ? 0.7 : 1,
      }}
    >
      {/* Line header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
        }}
      >
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 13, color: '#8c8c8c', marginBottom: 2, fontFamily: 'monospace' }}>
            SKU: {line.sku_id.substring(0, 8)}...
          </div>
          {line.batch_no && (
            <div style={{ fontSize: 12, color: '#bfbfbf' }}>
              {t('task.batch')}: {line.batch_no}
            </div>
          )}
        </div>
        <span
          style={{
            display: 'inline-block',
            padding: '3px 8px',
            fontSize: 11,
            fontWeight: 600,
            color: isCompleted ? '#389e0d' : '#bfbfbf',
            background: isCompleted ? '#f6ffed' : '#fafafa',
            borderRadius: 4,
            border: isCompleted ? '1px solid #b7eb8f' : '1px solid #f0f0f0',
            whiteSpace: 'nowrap',
          }}
        >
          {isCompleted ? t('cycleCount.counted') : t('cycleCount.pending')}
        </span>
      </div>

      {/* System qty */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginBottom: isCompleted ? 0 : 12, fontSize: 13 }}>
        <div>
          <span style={{ color: '#8c8c8c' }}>{t('cycleCount.systemQty')}:</span>{' '}
          <span style={{ fontWeight: 600, color: '#262626' }}>
            {line.system_qty}
          </span>
        </div>
        {line.counted_qty !== undefined && line.counted_qty !== null && (
          <div>
            <span style={{ color: '#8c8c8c' }}>{t('cycleCount.countedQtyShort')}:</span>{' '}
            <span
              style={{
                fontWeight: 600,
                color:
                  line.variance !== undefined && line.variance !== 0
                    ? '#fa8c16'
                    : '#389e0d',
              }}
            >
              {line.counted_qty}
            </span>
          </div>
        )}
      </div>

      {/* Variance display (if completed) */}
      {isCompleted && line.variance !== undefined && line.variance !== null && line.variance !== 0 && (
        <div
          style={{
            marginTop: 8,
            padding: '6px 10px',
            background: '#fff7e6',
            borderRadius: 6,
            fontSize: 12,
            color: '#ad6800',
          }}
        >
          {t('cycleCount.variance')}:{' '}
          <strong>
            {line.variance > 0 ? '+' : ''}
            {line.variance}
          </strong>
        </div>
      )}

      {/* Qty input — only if not completed */}
      {!isCompleted && (
        <form onSubmit={handleSubmit}>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <input
              type="number"
              inputMode="decimal"
              min="0"
              step="any"
              value={currentQty}
              onChange={(e) => onQtyChange(e.target.value)}
              placeholder={t('cycleCount.enterCountedQty')}
              disabled={isSubmitting}
              style={{
                flex: 1,
                padding: '10px 14px',
                fontSize: 15,
                border: '2px solid #d9d9d9',
                borderRadius: 8,
                outline: 'none',
                background: isSubmitting ? '#f5f5f5' : '#fff',
              }}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = '#1677ff'
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = '#d9d9d9'
              }}
            />
            <button
              type="submit"
              disabled={isSubmitting || !hasValidQty}
              style={{
                padding: '10px 18px',
                fontSize: 14,
                fontWeight: 600,
                color: '#fff',
                background:
                  isSubmitting || !hasValidQty ? '#d9d9d9' : '#1677ff',
                border: 'none',
                borderRadius: 8,
                cursor:
                  isSubmitting || !hasValidQty ? 'not-allowed' : 'pointer',
                whiteSpace: 'nowrap',
              }}
            >
              {isSubmitting ? '...' : t('cycleCount.confirm')}
            </button>
          </div>
        </form>
      )}
    </div>
  )
}

// ── Summary tile helper ─────────────────────────────────────────────────

function SummaryTile({
  label,
  value,
  color,
}: {
  label: string
  value: number
  color: string
}) {
  return (
    <div
      style={{
        background: `${color}10`,
        borderRadius: 10,
        padding: 14,
        textAlign: 'center',
      }}
    >
      <div style={{ fontSize: 24, fontWeight: 700, color, lineHeight: 1.2 }}>
        {value}
      </div>
      <div style={{ fontSize: 12, color: '#8c8c8c', marginTop: 4 }}>{label}</div>
    </div>
  )
}
