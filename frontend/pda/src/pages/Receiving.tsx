// Receiving page — ASN receiving workflow for warehouse operators.
// Operator scans/enters an ASN number, reviews lines, enters received quantities,
// and confirms receipt line-by-line. All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import client from '@/api/client'
import BarcodeScanner from '@/components/BarcodeScanner'
import type { ASN, ASNSummary, ASNStatus, ASNLineStatus, ListResponse } from '@/api/types'

// ── Status labels & colors ─────────────────────────────────────────────

function getStatusLabelKey(status: ASNStatus): string {
  const map: Record<ASNStatus, string> = {
    pending: 'receive.statusPending',
    arrived: 'receive.statusArrived',
    receiving: 'receive.statusReceiving',
    partial: 'receive.statusPartial',
    received: 'receive.statusReceived',
  }
  return map[status] || 'receive.statusPending'
}

function getStatusColor(status: ASNStatus): string {
  const map: Record<ASNStatus, string> = {
    pending: '#595959',
    arrived: '#1677ff',
    receiving: '#fa8c16',
    partial: '#722ed1',
    received: '#389e0d',
  }
  return map[status] || '#595959'
}

function getLineStatusLabelKey(status: ASNLineStatus): string {
  const map: Record<ASNLineStatus, string> = {
    pending: 'receive.lineStatusPending',
    partial: 'receive.lineStatusPartial',
    received: 'receive.lineStatusReceived',
  }
  return map[status] || 'receive.lineStatusPending'
}

function getLineStatusColor(status: ASNLineStatus): string {
  const map: Record<ASNLineStatus, string> = {
    pending: '#d9d9d9',
    partial: '#fa8c16',
    received: '#389e0d',
  }
  return map[status] || '#d9d9d9'
}

export default function ReceivingPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ────────────────────────────────────────────────────────────

  const [asnNo, setAsnNo] = useState('')
  const [receiveQtys, setReceiveQtys] = useState<Record<string, string>>({})
  const [receivingLines, setReceivingLines] = useState<Set<string>>(new Set())

  // ── Fetch ASN list filtered by asn_no ────────────────────────────────

  const {
    data: asnList,
    isLoading: isLookingUp,
    isError: isLookupError,
    error: lookupErr,
  } = useQuery<ListResponse<ASNSummary>>({
    queryKey: ['asnLookup', asnNo],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<ASNSummary>>('/asns', {
        params: { asn_no: asnNo },
      })
      return data
    },
    enabled: asnNo.length > 0,
    retry: false,
  })

  const matchedSummary = asnList?.data?.[0] ?? null

  // ── Fetch full ASN details with lines ────────────────────────────────

  const {
    data: asn,
    isLoading: isLoadingDetail,
    isError: isDetailError,
    error: detailError,
    refetch: refetchAsn,
  } = useQuery<ASN>({
    queryKey: ['asn', matchedSummary?.id],
    queryFn: async () => {
      const { data } = await client.get<ASN>(`/asns/${matchedSummary!.id}`)
      return data
    },
    enabled: !!matchedSummary?.id,
    retry: false,
  })

  // ── Handle barcode scan / manual input ───────────────────────────────

  const handleScan = useCallback((barcode: string) => {
    setAsnNo(barcode)
    setReceiveQtys({})
    setReceivingLines(new Set())
  }, [])

  // ── Handle clear / rescan ────────────────────────────────────────────

  const handleClear = useCallback(() => {
    setAsnNo('')
    setReceiveQtys({})
    setReceivingLines(new Set())
  }, [])

  // ── Receive line mutation ────────────────────────────────────────────

  const receiveLineMutation = useMutation({
    mutationFn: async ({ lineId, qty }: { lineId: string; qty: number }) => {
      await client.post(`/asns/${asn!.id}/lines/${lineId}/receive`, {
        received_qty: qty,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['asn', asn?.id] })
    },
  })

  const handleReceiveLine = async (lineId: string) => {
    const qtyStr = receiveQtys[lineId]
    if (!qtyStr) return
    const qty = parseFloat(qtyStr)
    if (isNaN(qty) || qty <= 0) return

    setReceivingLines((prev) => new Set(prev).add(lineId))
    try {
      await receiveLineMutation.mutateAsync({ lineId, qty })
      setReceiveQtys((prev) => ({ ...prev, [lineId]: '' }))
    } catch {
      // Error handled by mutation state
    } finally {
      setReceivingLines((prev) => {
        const next = new Set(prev)
        next.delete(lineId)
        return next
      })
    }
  }

  // ── Computed values ──────────────────────────────────────────────────

  const totalLines = asn?.lines?.length ?? 0
  const receivedLines = asn?.lines?.filter((l) => l.status === 'received').length ?? 0
  const allReceived = totalLines > 0 && receivedLines === totalLines

  // ── Format date helper ───────────────────────────────────────────────

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '—'
    return new Date(dateStr).toLocaleString()
  }

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      {/* ASN lookup section */}
      {!asn && (
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
              {t('receive.title')}
            </h3>
            <BarcodeScanner
              onScan={handleScan}
              placeholder={t('receive.scanAsnPlaceholder')}
            />
          </div>

          {/* Lookup in progress */}
          {isLookingUp && (
            <div
              style={{
                padding: '16px',
                textAlign: 'center',
                color: '#1677ff',
                fontSize: 14,
              }}
            >
              {'\u{1F50D}'} {t('receive.searchingAsn')}
            </div>
          )}

          {/* Lookup error — not found */}
          {isLookupError ||
          (!isLookingUp && asnNo && !matchedSummary && !isLookingUp) ? (
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
              {lookupErr
                ? t('receive.lookupFailed', { error: (lookupErr as Error)?.message })
                : t('receive.asnNotFound', { asnNo })}
            </div>
          ) : null}
        </>
      )}

      {/* ASN detail card */}
      {asn && (
        <>
          {/* Back button */}
          <button
            onClick={() => {
              queryClient.removeQueries({ queryKey: ['asnLookup', asnNo] })
              queryClient.removeQueries({ queryKey: ['asn', asn.id] })
              handleClear()
            }}
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
            {t('receive.backToScan')}
          </button>

          {/* ASN info card */}
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
                  {asn.asn_no}
                </h2>
              </div>
              <span
                style={{
                  display: 'inline-block',
                  padding: '4px 10px',
                  fontSize: 12,
                  fontWeight: 600,
                  color: '#fff',
                  background: getStatusColor(asn.status as ASNStatus),
                  borderRadius: 4,
                  whiteSpace: 'nowrap',
                }}
              >
                {t(getStatusLabelKey(asn.status as ASNStatus))}
              </span>
            </div>

            {/* Detail rows */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <DetailRow label={t('receive.carrier')} value={asn.carrier || '—'} />
              <DetailRow label={t('receive.trackingNo')} value={asn.tracking_no || '—'} />
              <DetailRow label={t('receive.expectedAt')} value={formatDate(asn.expected_at)} />
              <DetailRow label={t('receive.arrivedAt')} value={formatDate(asn.arrived_at)} />
            </div>

            {/* Progress bar */}
            <div style={{ marginTop: 16 }}>
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
                  {t('receive.linesReceived', {
                    received: receivedLines,
                    total: totalLines,
                  })}
                </span>
                <span style={{ fontWeight: 600 }}>
                  {totalLines > 0 ? Math.round((receivedLines / totalLines) * 100) : 0}%
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
                    width: `${totalLines > 0 ? (receivedLines / totalLines) * 100 : 0}%`,
                    background: allReceived ? '#389e0d' : '#1677ff',
                    borderRadius: 4,
                    transition: 'width 0.3s ease',
                  }}
                />
              </div>
            </div>
          </div>

          {/* Fully received banner */}
          {allReceived && (
            <div
              style={{
                marginBottom: 16,
                padding: '16px 20px',
                background: '#f6ffed',
                border: '1px solid #b7eb8f',
                borderRadius: 10,
                textAlign: 'center',
                fontSize: 15,
                color: '#389e0d',
                fontWeight: 600,
              }}
            >
              {t('receive.asnReceived')}
            </div>
          )}

          {/* Error from receive line mutation */}
          {receiveLineMutation.isError && (
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
              {t('receive.receiveLineError', {
                error: (receiveLineMutation.error as Error)?.message,
              })}
            </div>
          )}

          {/* Detail loading */}
          {isLoadingDetail && (
            <div className="pda-empty">
              <span className="empty-icon">{'⏳'}</span>
              <p className="empty-text">{t('receive.searchingAsn')}</p>
            </div>
          )}

          {/* Detail error */}
          {isDetailError && (
            <div className="pda-empty">
              <span className="empty-icon">{'⚠️'}</span>
              <p className="empty-text">
                {t('receive.lookupFailed', {
                  error: (detailError as Error)?.message,
                })}
              </p>
              <button
                onClick={() => refetchAsn()}
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

          {/* Line items */}
          {!isLoadingDetail &&
            !isDetailError &&
            asn.lines &&
            asn.lines.map((line) => {
              const remainingQty = line.expected_qty - line.received_qty
              const isLineComplete = line.status === 'received'
              const currentQty = receiveQtys[line.id] || ''
              const isReceiving = receivingLines.has(line.id)

              let qtyError: string | null = null
              const parsed = parseFloat(currentQty)
              if (currentQty && !isNaN(parsed)) {
                if (parsed > remainingQty && remainingQty > 0) {
                  qtyError = t('receive.qtyExceedsExpected', {
                    total: (line.received_qty + parsed).toString(),
                    expected: line.expected_qty.toString(),
                  })
                }
              }

              return (
                <div
                  key={line.id}
                  style={{
                    background: '#fff',
                    borderRadius: 12,
                    padding: 16,
                    boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
                    marginBottom: 12,
                    opacity: isLineComplete ? 0.7 : 1,
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
                      <div
                        style={{
                          fontSize: 13,
                          color: '#8c8c8c',
                          marginBottom: 2,
                        }}
                      >
                        {t('task.sku')}:{' '}
                        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
                          {line.sku_id.substring(0, 8)}...
                        </span>
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
                        color: isLineComplete ? '#389e0d' : '#595959',
                        background: isLineComplete ? '#f6ffed' : '#f0f0f0',
                        borderRadius: 4,
                        border: `1px solid ${getLineStatusColor(line.status as ASNLineStatus)}`,
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {t(getLineStatusLabelKey(line.status as ASNLineStatus))}
                    </span>
                  </div>

                  {/* Qty info */}
                  <div
                    style={{
                      display: 'grid',
                      gridTemplateColumns: '1fr 1fr',
                      gap: 8,
                      marginBottom: isLineComplete ? 0 : 12,
                      fontSize: 13,
                    }}
                  >
                    <div>
                      <span style={{ color: '#8c8c8c' }}>
                        {t('receive.expectedQty')}:
                      </span>{' '}
                      <span style={{ fontWeight: 600, color: '#262626' }}>
                        {line.expected_qty}
                      </span>
                    </div>
                    <div>
                      <span style={{ color: '#8c8c8c' }}>
                        {t('receive.receivedQty')}:
                      </span>{' '}
                      <span
                        style={{
                          fontWeight: 600,
                          color: line.received_qty >= line.expected_qty ? '#389e0d' : '#262626',
                        }}
                      >
                        {line.received_qty}
                      </span>
                    </div>
                  </div>

                  {/* Receive input — only if not complete */}
                  {!isLineComplete && (
                    <>
                      <div
                        style={{
                          display: 'flex',
                          gap: 8,
                          alignItems: 'flex-start',
                          flexDirection: 'column',
                        }}
                      >
                        {/* Qty input row */}
                        <div style={{ display: 'flex', gap: 8, width: '100%' }}>
                          <input
                            type="number"
                            inputMode="decimal"
                            min="0"
                            max={remainingQty}
                            step="any"
                            value={currentQty}
                            onChange={(e) =>
                              setReceiveQtys((prev) => ({
                                ...prev,
                                [line.id]: e.target.value,
                              }))
                            }
                            placeholder={`${t('receive.receiveQty')} (max ${remainingQty})`}
                            disabled={isReceiving}
                            style={{
                              flex: 1,
                              padding: '10px 14px',
                              fontSize: 15,
                              border: qtyError
                                ? '2px solid #ff4d4f'
                                : '2px solid #d9d9d9',
                              borderRadius: 8,
                              outline: 'none',
                              background: isReceiving ? '#f5f5f5' : '#fff',
                            }}
                            onFocus={(e) => {
                              if (!qtyError) e.currentTarget.style.borderColor = '#1677ff'
                            }}
                            onBlur={(e) => {
                              if (!qtyError) e.currentTarget.style.borderColor = '#d9d9d9'
                            }}
                          />
                          <button
                            onClick={() => handleReceiveLine(line.id)}
                            disabled={
                              isReceiving ||
                              !currentQty ||
                              isNaN(parseFloat(currentQty)) ||
                              parseFloat(currentQty) <= 0 ||
                              !!qtyError
                            }
                            style={{
                              padding: '10px 18px',
                              fontSize: 14,
                              fontWeight: 600,
                              color: '#fff',
                              background:
                                isReceiving ||
                                !currentQty ||
                                isNaN(parseFloat(currentQty)) ||
                                parseFloat(currentQty) <= 0 ||
                                !!qtyError
                                  ? '#d9d9d9'
                                  : '#1677ff',
                              border: 'none',
                              borderRadius: 8,
                              cursor:
                                isReceiving ||
                                !currentQty ||
                                isNaN(parseFloat(currentQty)) ||
                                parseFloat(currentQty) <= 0 ||
                                !!qtyError
                                  ? 'not-allowed'
                                  : 'pointer',
                              whiteSpace: 'nowrap',
                              transition: 'background 0.2s',
                            }}
                          >
                            {isReceiving ? t('receive.receiving') : t('receive.receive')}
                          </button>
                        </div>

                        {/* Qty error message */}
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
                    </>
                  )}
                </div>
              )
            })}

          {/* Clear & rescan button at bottom */}
          {!allReceived && (
            <div style={{ marginTop: 16 }}>
              <button
                onClick={() => {
                  queryClient.removeQueries({ queryKey: ['asnLookup', asnNo] })
                  queryClient.removeQueries({ queryKey: ['asn', asn.id] })
                  handleClear()
                }}
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
                {t('receive.clearAndRescan')}
              </button>
            </div>
          )}

          {/* Scan another after all received */}
          {allReceived && (
            <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
              <button
                onClick={() => {
                  queryClient.removeQueries({ queryKey: ['asnLookup', asnNo] })
                  queryClient.removeQueries({ queryKey: ['asn', asn.id] })
                  handleClear()
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
                {t('receive.clearAndRescan')}
              </button>
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
                {t('receive.backToScan')}
              </button>
            </div>
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
