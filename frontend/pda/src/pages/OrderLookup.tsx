// Order Lookup page — scan or type an order number to view order details.
// Displays status badge, type, priority, and a line item table
// (SKU, ordered qty, fulfilled qty) for warehouse operators.
// All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import BarcodeScanner from '@/components/BarcodeScanner'
import { fetchOrders, fetchOrder } from '@/api/order'
import type { Order, OrderSummary, OrderStatus, OrderType, OrderPriority, OrderLineStatus, ListResponse } from '@/api/types'

// ── Status labels & colors ─────────────────────────────────────────────

function getOrderStatusLabelKey(status: OrderStatus): string {
  const map: Record<OrderStatus, string> = {
    draft: 'orderLookup.statusDraft',
    confirmed: 'orderLookup.statusConfirmed',
    processing: 'orderLookup.statusProcessing',
    partial: 'orderLookup.statusPartial',
    completed: 'orderLookup.statusCompleted',
    cancelled: 'orderLookup.statusCancelled',
  }
  return map[status] || 'orderLookup.statusDraft'
}

function getOrderStatusColor(status: OrderStatus): string {
  const map: Record<OrderStatus, string> = {
    draft: '#595959',
    confirmed: '#1677ff',
    processing: '#fa8c16',
    partial: '#722ed1',
    completed: '#389e0d',
    cancelled: '#cf1322',
  }
  return map[status] || '#595959'
}

function getOrderTypeLabelKey(orderType: OrderType): string {
  const map: Record<OrderType, string> = {
    inbound: 'orderLookup.typeInbound',
    outbound: 'orderLookup.typeOutbound',
    transfer: 'orderLookup.typeTransfer',
    return: 'orderLookup.typeReturn',
  }
  return map[orderType] || 'orderLookup.typeInbound'
}

function getOrderPriorityLabelKey(priority: OrderPriority): string {
  const map: Record<OrderPriority, string> = {
    low: 'orderLookup.priorityLow',
    normal: 'orderLookup.priorityNormal',
    high: 'orderLookup.priorityHigh',
    urgent: 'orderLookup.priorityUrgent',
  }
  return map[priority] || 'orderLookup.priorityNormal'
}

function getOrderPriorityColor(priority: OrderPriority): string {
  const map: Record<OrderPriority, string> = {
    low: '#8c8c8c',
    normal: '#1677ff',
    high: '#fa8c16',
    urgent: '#cf1322',
  }
  return map[priority] || '#1677ff'
}

function getLineStatusLabelKey(status: OrderLineStatus): string {
  const map: Record<OrderLineStatus, string> = {
    pending: 'orderLookup.lineStatusPending',
    allocated: 'orderLookup.lineStatusAllocated',
    partial: 'orderLookup.lineStatusPartial',
    fulfilled: 'orderLookup.lineStatusFulfilled',
    cancelled: 'orderLookup.lineStatusCancelled',
  }
  return map[status] || 'orderLookup.lineStatusPending'
}

function getLineStatusColor(status: OrderLineStatus): string {
  const map: Record<OrderLineStatus, string> = {
    pending: '#595959',
    allocated: '#1677ff',
    partial: '#fa8c16',
    fulfilled: '#389e0d',
    cancelled: '#cf1322',
  }
  return map[status] || '#595959'
}

export default function OrderLookupPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ────────────────────────────────────────────────────────────

  const [orderNo, setOrderNo] = useState('')

  // ── Fetch order summary list filtered by order_no ────────────────────

  const {
    data: orderList,
    isLoading: isLookingUp,
    isError: isLookupError,
    error: lookupErr,
  } = useQuery<ListResponse<OrderSummary>>({
    queryKey: ['orderLookup', orderNo],
    queryFn: async () => fetchOrders({ order_no: orderNo }),
    enabled: orderNo.length > 0,
    retry: false,
  })

  const matchedSummary = orderList?.data?.[0] ?? null

  // ── Fetch full order details with lines ──────────────────────────────

  const {
    data: order,
    isLoading: isLoadingDetail,
    isError: isDetailError,
    error: detailError,
    refetch: refetchOrder,
  } = useQuery<Order>({
    queryKey: ['order', matchedSummary?.id],
    queryFn: async () => fetchOrder(matchedSummary!.id),
    enabled: !!matchedSummary?.id,
    retry: false,
  })

  // ── Handle barcode scan / manual input ───────────────────────────────

  const handleScan = useCallback((barcode: string) => {
    setOrderNo(barcode)
  }, [])

  // ── Handle clear / rescan ────────────────────────────────────────────

  const handleClear = useCallback(() => {
    setOrderNo('')
  }, [])

  // ── Format date helper ───────────────────────────────────────────────

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '—'
    return new Date(dateStr).toLocaleString()
  }

  // ── Computed values ──────────────────────────────────────────────────

  const totalLines = order?.lines?.length ?? 0
  const fulfilledLines = order?.lines?.filter((l) => l.status === 'fulfilled').length ?? 0

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      {/* Order lookup section */}
      {!order && (
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
              {t('orderLookup.title')}
            </h3>
            <BarcodeScanner
              onScan={handleScan}
              placeholder={t('orderLookup.scanPlaceholder')}
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
              {'\u{1F50D}'} {t('orderLookup.searching')}
            </div>
          )}

          {/* Lookup error — not found */}
          {isLookupError ||
          (!isLookingUp && orderNo && !matchedSummary && !isLookingUp) ? (
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
                ? t('orderLookup.lookupFailed', { error: (lookupErr as Error)?.message })
                : t('orderLookup.orderNotFound', { orderNo })}
            </div>
          ) : null}

          {/* Back button */}
          <button
            onClick={() => navigate('/scan')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 4,
              padding: '10px 0',
              fontSize: 14,
              color: '#1677ff',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
            }}
          >
            {t('orderLookup.backToScan')}
          </button>
        </>
      )}

      {/* Order detail card */}
      {order && (
        <>
          {/* Back button */}
          <button
            onClick={() => {
              queryClient.removeQueries({ queryKey: ['orderLookup', orderNo] })
              queryClient.removeQueries({ queryKey: ['order', order.id] })
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
            {t('orderLookup.backToScan')}
          </button>

          {/* Order info card */}
          <div
            style={{
              background: '#fff',
              borderRadius: 12,
              padding: 20,
              boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              marginBottom: 16,
            }}
          >
            {/* Header: order number + status badge */}
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
                  {order.order_no}
                </h2>
              </div>
              <span
                style={{
                  display: 'inline-block',
                  padding: '4px 10px',
                  fontSize: 12,
                  fontWeight: 600,
                  color: '#fff',
                  background: getOrderStatusColor(order.status as OrderStatus),
                  borderRadius: 4,
                  whiteSpace: 'nowrap',
                }}
              >
                {t(getOrderStatusLabelKey(order.status as OrderStatus))}
              </span>
            </div>

            {/* Type + Priority badges */}
            <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
              <span
                style={{
                  display: 'inline-block',
                  padding: '4px 10px',
                  fontSize: 12,
                  fontWeight: 500,
                  color: '#1677ff',
                  background: '#e6f7ff',
                  borderRadius: 4,
                  border: '1px solid #91d5ff',
                }}
              >
                {t(getOrderTypeLabelKey(order.order_type as OrderType))}
              </span>
              <span
                style={{
                  display: 'inline-block',
                  padding: '4px 10px',
                  fontSize: 12,
                  fontWeight: 500,
                  color: getOrderPriorityColor(order.priority as OrderPriority),
                  background: '#fff7e6',
                  borderRadius: 4,
                  border: `1px solid ${getOrderPriorityColor(order.priority as OrderPriority)}`,
                }}
              >
                {t(getOrderPriorityLabelKey(order.priority as OrderPriority))}
              </span>
            </div>

            {/* Detail rows */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <DetailRow label={t('orderLookup.externalRef')} value={order.external_ref || '—'} />
              <DetailRow label={t('orderLookup.externalType')} value={order.external_type || '—'} />
              <DetailRow label={t('orderLookup.notes')} value={order.notes || '—'} />
              <DetailRow label={t('orderLookup.createdAt')} value={formatDate(order.created_at)} />
              <DetailRow label={t('orderLookup.updatedAt')} value={formatDate(order.updated_at)} />
              {order.completed_at && (
                <DetailRow label={t('orderLookup.completedAt')} value={formatDate(order.completed_at)} />
              )}
              <DetailRow label={t('orderLookup.createdBy')} value={order.created_by} />
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
                  {t('orderLookup.linesFulfilled', {
                    fulfilled: fulfilledLines,
                    total: totalLines,
                  })}
                </span>
                <span style={{ fontWeight: 600 }}>
                  {totalLines > 0 ? Math.round((fulfilledLines / totalLines) * 100) : 0}%
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
                    width: `${totalLines > 0 ? (fulfilledLines / totalLines) * 100 : 0}%`,
                    background: fulfilledLines === totalLines && totalLines > 0 ? '#389e0d' : '#1677ff',
                    borderRadius: 4,
                    transition: 'width 0.3s ease',
                  }}
                />
              </div>
            </div>
          </div>

          {/* Fully fulfilled banner */}
          {fulfilledLines === totalLines && totalLines > 0 && (
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
              {t('orderLookup.orderFulfilled')}
            </div>
          )}

          {/* Detail loading */}
          {isLoadingDetail && (
            <div className="pda-empty">
              <span className="empty-icon">{'⏳'}</span>
              <p className="empty-text">{t('orderLookup.loadingOrder')}</p>
            </div>
          )}

          {/* Detail error */}
          {isDetailError && (
            <div className="pda-empty">
              <span className="empty-icon">{'⚠️'}</span>
              <p className="empty-text">
                {t('orderLookup.lookupFailed', {
                  error: (detailError as Error)?.message,
                })}
              </p>
              <button
                onClick={() => refetchOrder()}
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

          {/* Line items table */}
          {!isLoadingDetail && !isDetailError && order.lines && order.lines.length > 0 && (
            <>
              <h3
                style={{
                  fontSize: 15,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 10,
                }}
              >
                {t('orderLookup.lineItems')} ({order.lines.length})
              </h3>

              {order.lines.map((line) => {
                const isLineComplete = line.status === 'fulfilled'

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
                            fontSize: 14,
                            fontWeight: 600,
                            color: '#262626',
                            marginBottom: 2,
                          }}
                        >
                          {t('orderLookup.lineNo')} {line.line_no}
                        </div>
                        <div style={{ fontSize: 12, color: '#8c8c8c', fontFamily: 'monospace' }}>
                          {t('task.sku')}: {line.sku_id.substring(0, 8)}...
                        </div>
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
                          border: `1px solid ${getLineStatusColor(line.status as OrderLineStatus)}`,
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {t(getLineStatusLabelKey(line.status as OrderLineStatus))}
                      </span>
                    </div>

                    {/* Qty info */}
                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: '1fr 1fr 1fr',
                        gap: 8,
                        fontSize: 13,
                      }}
                    >
                      <div>
                        <span style={{ color: '#8c8c8c' }}>
                          {t('orderLookup.orderedQty')}:
                        </span>{' '}
                        <span style={{ fontWeight: 600, color: '#262626' }}>
                          {line.ordered_qty}
                        </span>
                      </div>
                      <div>
                        <span style={{ color: '#8c8c8c' }}>
                          {t('orderLookup.fulfilledQty')}:
                        </span>{' '}
                        <span
                          style={{
                            fontWeight: 600,
                            color: line.fulfilled_qty >= line.ordered_qty ? '#389e0d' : '#262626',
                          }}
                        >
                          {line.fulfilled_qty}
                        </span>
                      </div>
                      <div>
                        <span style={{ color: '#8c8c8c' }}>
                          {t('orderLookup.uom')}:
                        </span>{' '}
                        <span style={{ fontWeight: 500, color: '#262626' }}>
                          {line.uom}
                        </span>
                      </div>
                    </div>

                    {/* Batch info if present */}
                    {line.batch_no && (
                      <div style={{ marginTop: 8, fontSize: 12, color: '#bfbfbf' }}>
                        {t('task.batch')}: {line.batch_no}
                      </div>
                    )}

                    {/* Notes if present */}
                    {line.notes && (
                      <div style={{ marginTop: 4, fontSize: 12, color: '#8c8c8c', fontStyle: 'italic' }}>
                        {line.notes}
                      </div>
                    )}

                    {/* Progress bar per line */}
                    <div style={{ marginTop: 10 }}>
                      <div
                        style={{
                          height: 4,
                          background: '#f0f0f0',
                          borderRadius: 2,
                          overflow: 'hidden',
                        }}
                      >
                        <div
                          style={{
                            height: '100%',
                            width: `${line.ordered_qty > 0 ? Math.min(100, (line.fulfilled_qty / line.ordered_qty) * 100) : 0}%`,
                            background: isLineComplete ? '#389e0d' : '#1677ff',
                            borderRadius: 2,
                            transition: 'width 0.3s ease',
                          }}
                        />
                      </div>
                    </div>
                  </div>
                )
              })}
            </>
          )}

          {/* Line items empty state */}
          {!isLoadingDetail && !isDetailError && (!order.lines || order.lines.length === 0) && (
            <div className="pda-empty">
              <span className="empty-icon">{'\u{1F4ED}'}</span>
              <p className="empty-text">{t('orderLookup.noLines')}</p>
            </div>
          )}

          {/* Clear & rescan button */}
          <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
            <button
              onClick={() => {
                queryClient.removeQueries({ queryKey: ['orderLookup', orderNo] })
                queryClient.removeQueries({ queryKey: ['order', order.id] })
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
              {t('orderLookup.clearAndRescan')}
            </button>
            <button
              onClick={() => navigate('/scan')}
              style={{
                width: '100%',
                padding: '14px 0',
                fontSize: 15,
                fontWeight: 500,
                color: '#595959',
                background: 'transparent',
                border: '1px solid #d9d9d9',
                borderRadius: 10,
                cursor: 'pointer',
              }}
            >
              {t('orderLookup.backToScan')}
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
