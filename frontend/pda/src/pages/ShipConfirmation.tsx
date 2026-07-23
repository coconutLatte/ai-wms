// ShipConfirmation page — ship confirmation workflow for warehouse operators.
// Operator scans or selects a confirmed outbound order, verifies picked line items,
// enters carrier and tracking information, and confirms the shipment.
// All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import client from '@/api/client'
import BarcodeScanner from '@/components/BarcodeScanner'
import type { Order, OrderStatus, ListResponse, Shipment } from '@/api/types'

// ── Step enum for workflow state ───────────────────────────────────────

type Step = 'order_list' | 'verify_items' | 'enter_shipping' | 'confirm'

// ── Status helpers ──────────────────────────────────────────────────────

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
    draft: '#8c8c8c',
    confirmed: '#1677ff',
    processing: '#fa8c16',
    partial: '#722ed1',
    completed: '#389e0d',
    cancelled: '#cf1322',
  }
  return map[status] || '#8c8c8c'
}

interface OrderSummary {
  id: string
  order_no: string
  order_type: string
  warehouse_id: string
  status: OrderStatus
  priority: string
  external_ref?: string
  created_at: string
}

export default function ShipConfirmationPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  // ── State ────────────────────────────────────────────────────────────

  const [step, setStep] = useState<Step>('order_list')
  const [searchOrderNo, setSearchOrderNo] = useState('')
  const [selectedOrderId, setSelectedOrderId] = useState<string | null>(null)
  const [carrier, setCarrier] = useState('')
  const [trackingNo, setTrackingNo] = useState('')
  const [carrierService, setCarrierService] = useState('')
  const [notes, setNotes] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [createdShipment, setCreatedShipment] = useState<Shipment | null>(null)

  // ── Fetch confirmed outbound orders ──────────────────────────────────

  const {
    data: ordersList,
    isLoading: isLoadingOrders,
    isError: isOrdersError,
    error: ordersError,
    refetch: refetchOrders,
  } = useQuery<ListResponse<OrderSummary>>({
    queryKey: ['shipConfirmOrders'],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<OrderSummary>>('/orders', {
        params: {
          order_type: 'outbound',
          status: 'confirmed,processing,partial',
          page_size: 50,
        },
      })
      return data
    },
    retry: false,
  })

  const candidateOrders = (ordersList?.data ?? []).filter(
    (o) => o.status === 'confirmed' || o.status === 'processing' || o.status === 'partial'
  )

  // ── Lookup order by scanned order_no ─────────────────────────────────

  const {
    data: scannedOrder,
    isLoading: isScanningOrder,
    isError: isScanOrderError,
    error: scanOrderError,
  } = useQuery<ListResponse<OrderSummary>>({
    queryKey: ['shipConfirmOrderScan', searchOrderNo],
    queryFn: async () => {
      const { data } = await client.get<ListResponse<OrderSummary>>('/orders', {
        params: { order_no: searchOrderNo, page_size: 1 },
      })
      return data
    },
    enabled: searchOrderNo.length > 0,
    retry: false,
  })

  const scannedResult = scannedOrder?.data?.[0] ?? null
  const isScannedValid =
    scannedResult &&
    scannedResult.order_type === 'outbound' &&
    ['confirmed', 'processing', 'partial'].includes(scannedResult.status)

  // ── Fetch selected order details ─────────────────────────────────────

  const {
    data: orderDetail,
    isLoading: isLoadingDetail,
    isError: isDetailError,
    error: detailError,
  } = useQuery<Order>({
    queryKey: ['shipConfirmOrderDetail', selectedOrderId],
    queryFn: async () => {
      const { data } = await client.get<Order>(`/orders/${selectedOrderId}`)
      return data
    },
    enabled: !!selectedOrderId,
    retry: false,
  })

  // ── Shipment creation mutation ───────────────────────────────────────

  const createShipmentMutation = useMutation({
    mutationFn: async (payload: {
      order_id: string
      warehouse_id: string
      carrier: string
      tracking_no?: string
      carrier_service?: string
      notes?: string
    }) => {
      const { data } = await client.post<Shipment>('/shipments', payload)
      return data
    },
  })

  // ── Handle order selection from list ─────────────────────────────────

  const handleSelectOrder = (order: OrderSummary) => {
    setSelectedOrderId(order.id)
    setStep('verify_items')
  }

  // ── Handle scanned order (from barcode input) ────────────────────────

  const handleOrderScan = useCallback((barcode: string) => {
    setSearchOrderNo(barcode)
  }, [])

  // When scanned order is validated, offer to select it
  const handleAcceptScannedOrder = useCallback(() => {
    if (scannedResult) {
      setSelectedOrderId(scannedResult.id)
      setStep('verify_items')
    }
  }, [scannedResult])

  // ── Verify items finished → enter shipping info ──────────────────────

  const handleProceedToShipping = useCallback(() => {
    setStep('enter_shipping')
    // Default carrier to empty — operator fills in
    setCarrier('')
    setTrackingNo('')
    setCarrierService('')
    setNotes('')
  }, [])

  // ── Confirm shipment ─────────────────────────────────────────────────

  const handleConfirmShipment = async () => {
    if (!orderDetail || !carrier.trim()) return

    setIsSubmitting(true)
    try {
      const shipment = await createShipmentMutation.mutateAsync({
        order_id: orderDetail.id,
        warehouse_id: orderDetail.warehouse_id,
        carrier: carrier.trim(),
        tracking_no: trackingNo.trim() || undefined,
        carrier_service: carrierService.trim() || undefined,
        notes: notes.trim() || undefined,
      })
      setCreatedShipment(shipment)
      setStep('confirm')
    } catch {
      // Error handled by mutation state
    } finally {
      setIsSubmitting(false)
    }
  }

  // ── Reset entirely ───────────────────────────────────────────────────

  const handleResetAll = useCallback(() => {
    setStep('order_list')
    setSelectedOrderId(null)
    setSearchOrderNo('')
    setCarrier('')
    setTrackingNo('')
    setCarrierService('')
    setNotes('')
    setIsSubmitting(false)
    setCreatedShipment(null)
    queryClient.removeQueries({ queryKey: ['shipConfirmOrderDetail'] })
  }, [queryClient])

  // ── Back one step ────────────────────────────────────────────────────

  const handleBack = useCallback(() => {
    if (step === 'verify_items') {
      setStep('order_list')
      setSelectedOrderId(null)
    } else if (step === 'enter_shipping') {
      setStep('verify_items')
      setCarrier('')
      setTrackingNo('')
      setCarrierService('')
      setNotes('')
    } else if (step === 'confirm') {
      setStep('enter_shipping')
      setCreatedShipment(null)
    }
  }, [step])

  // ── Compute verification stats for the order ─────────────────────────

  const pickedLines = orderDetail?.lines?.filter(
    (l) => l.status === 'fulfilled' || l.status === 'partial'
  ) ?? []
  const pendingLines = orderDetail?.lines?.filter(
    (l) => l.status === 'pending' || l.status === 'allocated'
  ) ?? []
  const allPicked = pendingLines.length === 0 && pickedLines.length > 0
  const somePicked = pickedLines.length > 0

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      {/* ── Step indicator ────────────────────────────────────────────── */}
      {step !== 'order_list' && (
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
            flexWrap: 'wrap',
          }}
        >
          <StepBadge
            num={1}
            done={step !== 'verify_items'}
            active={step === 'verify_items'}
          />
          <span
            style={{
              color: step === 'verify_items' ? '#1677ff' : '#389e0d',
              fontWeight: step === 'verify_items' ? 600 : 400,
            }}
          >
            {t('shipConfirm.verifyItems')}
          </span>
          <span style={{ color: '#d9d9d9' }}>→</span>
          <StepBadge
            num={2}
            done={step === 'confirm' && !!createdShipment}
            active={step === 'enter_shipping'}
            pending={step === 'verify_items'}
          />
          <span
            style={{
              color: step === 'enter_shipping' ? '#1677ff' : step === 'confirm' ? '#389e0d' : '#8c8c8c',
              fontWeight: step === 'enter_shipping' ? 600 : 400,
            }}
          >
            {t('shipConfirm.enterShipping')}
          </span>
          <span style={{ color: '#d9d9d9' }}>→</span>
          <StepBadge
            num={3}
            done={false}
            active={step === 'confirm'}
            pending={step !== 'confirm'}
          />
          <span
            style={{
              color: step === 'confirm' ? '#1677ff' : '#8c8c8c',
              fontWeight: step === 'confirm' ? 600 : 400,
            }}
          >
            {t('shipConfirm.confirmShipment')}
          </span>
        </div>
      )}

      {/* ── Back button (when not on order_list) ───────────────────────── */}
      {step !== 'order_list' && (
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
      {createShipmentMutation.isError && (
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
          {t('shipConfirm.createError', {
            error: (createShipmentMutation.error as Error)?.message,
          })}
        </div>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 1: Order list / scan */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'order_list' && (
        <>
          <div style={{ marginBottom: 16 }}>
            <h3
              style={{
                fontSize: 15,
                fontWeight: 600,
                color: '#595959',
                marginBottom: 4,
              }}
            >
              {t('shipConfirm.title')}
            </h3>
            <p style={{ fontSize: 13, color: '#8c8c8c', margin: 0 }}>
              {t('shipConfirm.description')}
            </p>
          </div>

          {/* Barcode scanner */}
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
                fontSize: 14,
                fontWeight: 600,
                color: '#262626',
                marginBottom: 12,
              }}
            >
              {t('shipConfirm.scanOrder')}
            </h3>
            <BarcodeScanner
              onScan={handleOrderScan}
              placeholder={t('shipConfirm.scanOrderPlaceholder')}
            />

            {/* Scanning lookup in progress */}
            {isScanningOrder && (
              <div
                style={{
                  marginTop: 12,
                  padding: '12px',
                  textAlign: 'center',
                  color: '#1677ff',
                  fontSize: 14,
                }}
              >
                {'\u{1F50D}'} {t('orderLookup.searching')}
              </div>
            )}

            {/* Scan error */}
            {isScanOrderError && (
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
                {t('shipConfirm.lookupFailed', {
                  error: (scanOrderError as Error)?.message,
                })}
              </div>
            )}

            {/* Scanned order not found */}
            {!isScanningOrder && searchOrderNo && !scannedResult && (
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
                {t('shipConfirm.orderNotFound', { orderNo: searchOrderNo })}
              </div>
            )}

            {/* Scanned order is not an outbound order or wrong status */}
            {!isScanningOrder && scannedResult && !isScannedValid && (
              <div
                style={{
                  marginTop: 12,
                  padding: '14px 18px',
                  background: '#fffbe6',
                  border: '1px solid #ffe58f',
                  borderRadius: 10,
                  fontSize: 14,
                  color: '#d48806',
                }}
              >
                {scannedResult.order_type !== 'outbound'
                  ? t('shipConfirm.notOutboundOrder')
                  : t('shipConfirm.orderNotReady', {
                      status: t(getOrderStatusLabelKey(scannedResult.status)),
                    })}
              </div>
            )}

            {/* Scanned order is valid */}
            {!isScanningOrder && scannedResult && isScannedValid && (
              <div style={{ marginTop: 12 }}>
                <div
                  style={{
                    padding: '14px 18px',
                    background: '#f6ffed',
                    border: '1px solid #b7eb8f',
                    borderRadius: 10,
                  }}
                >
                  <div
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: '#389e0d',
                      marginBottom: 8,
                    }}
                  >
                    {t('shipConfirm.orderFound')}
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
                    <DetailRow
                      label={t('orderLookup.statusConfirmed')}
                      value={scannedResult.order_no}
                    />
                    <DetailRow
                      label={t('shipConfirm.statusLabel')}
                      value={t(getOrderStatusLabelKey(scannedResult.status))}
                    />
                    {scannedResult.external_ref && (
                      <DetailRow
                        label={t('orderLookup.externalRef')}
                        value={scannedResult.external_ref}
                      />
                    )}
                  </div>
                </div>

                <button
                  onClick={handleAcceptScannedOrder}
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
                  {t('shipConfirm.selectThisOrder')}
                </button>
              </div>
            )}
          </div>

          {/* Order list */}
          <div>
            <div
              style={{
                fontSize: 14,
                fontWeight: 600,
                color: '#595959',
                marginBottom: 4,
              }}
            >
              {t('shipConfirm.readyOrders')}
            </div>
            <p style={{ fontSize: 12, color: '#8c8c8c', marginBottom: 12 }}>
              {t('shipConfirm.readyOrdersDesc')}
            </p>

            {/* Loading */}
            {isLoadingOrders && (
              <div className="pda-empty">
                <span className="empty-icon">{'⏳'}</span>
                <p className="empty-text">{t('shipConfirm.loadingOrders')}</p>
              </div>
            )}

            {/* Error */}
            {isOrdersError && (
              <div className="pda-empty">
                <span className="empty-icon">{'⚠️'}</span>
                <p className="empty-text">
                  {t('shipConfirm.ordersLoadFailed', {
                    error: (ordersError as Error)?.message,
                  })}
                </p>
                <button
                  onClick={() => refetchOrders()}
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
            {!isLoadingOrders && !isOrdersError && candidateOrders.length === 0 && (
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
                <div
                  style={{
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#8c8c8c',
                    marginBottom: 8,
                  }}
                >
                  {t('shipConfirm.noReadyOrders')}
                </div>
                <div style={{ fontSize: 13, color: '#8c8c8c', marginBottom: 20 }}>
                  {t('shipConfirm.noReadyOrdersDesc')}
                </div>
              </div>
            )}

            {/* Order cards */}
            {!isLoadingOrders &&
              !isOrdersError &&
              candidateOrders.map((order) => (
                <div
                  key={order.id}
                  onClick={() => handleSelectOrder(order)}
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
                        {order.order_no}
                      </div>
                      {order.external_ref && (
                        <div style={{ fontSize: 12, color: '#8c8c8c' }}>
                          {t('orderLookup.externalRef')}: {order.external_ref}
                        </div>
                      )}
                    </div>
                    <span
                      style={{
                        display: 'inline-block',
                        padding: '3px 8px',
                        fontSize: 11,
                        fontWeight: 600,
                        color: getOrderStatusColor(order.status),
                        background: '#fff',
                        border: `1px solid ${getOrderStatusColor(order.status)}`,
                        borderRadius: 4,
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {t(getOrderStatusLabelKey(order.status))}
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
                      <span style={{ color: '#8c8c8c' }}>{t('task.type')}: </span>
                      <span style={{ fontWeight: 500, color: '#262626' }}>
                        {t('orderLookup.typeOutbound')}
                      </span>
                    </div>
                    <div>
                      <span style={{ color: '#8c8c8c' }}>{t('orderLookup.createdAt')}: </span>
                      <span style={{ fontWeight: 500, color: '#262626' }}>
                        {order.created_at?.substring(0, 10)}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
          </div>
        </>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 2: Verify picked items */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'verify_items' && (
        <>
          {/* Order summary card */}
          {orderDetail && (
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
                    {orderDetail.order_no}
                  </h2>
                  <div style={{ fontSize: 13, color: '#8c8c8c' }}>
                    {t(getOrderStatusLabelKey(orderDetail.status))}
                  </div>
                </div>
                <span
                  style={{
                    display: 'inline-block',
                    padding: '4px 10px',
                    fontSize: 12,
                    fontWeight: 600,
                    color: '#fff',
                    background: getOrderStatusColor(orderDetail.status),
                    borderRadius: 4,
                    whiteSpace: 'nowrap',
                  }}
                >
                  {t(getOrderStatusLabelKey(orderDetail.status))}
                </span>
              </div>

              {orderDetail.external_ref && (
                <DetailRow
                  label={t('orderLookup.externalRef')}
                  value={orderDetail.external_ref}
                />
              )}
            </div>
          )}

          {/* Loading detail */}
          {isLoadingDetail && (
            <div className="pda-empty">
              <span className="empty-icon">{'⏳'}</span>
              <p className="empty-text">{t('orderLookup.loadingOrder')}</p>
            </div>
          )}

          {/* Detail error */}
          {isDetailError && (
            <div
              style={{
                padding: '14px 18px',
                background: '#fff1f0',
                border: '1px solid #ffa39e',
                borderRadius: 10,
                fontSize: 14,
                color: '#cf1322',
                marginBottom: 16,
              }}
            >
              {t('shipConfirm.detailFailed', {
                error: (detailError as Error)?.message,
              })}
            </div>
          )}

          {/* Line items verification */}
          {orderDetail && !isLoadingDetail && !isDetailError && (
            <>
              {/* Picking status summary */}
              <div
                style={{
                  padding: '14px 18px',
                  background: allPicked ? '#f6ffed' : somePicked ? '#fffbe6' : '#fff7e6',
                  border: allPicked
                    ? '1px solid #b7eb8f'
                    : somePicked
                      ? '1px solid #ffe58f'
                      : '1px solid #ffd591',
                  borderRadius: 10,
                  marginBottom: 16,
                  fontSize: 13,
                }}
              >
                <div
                  style={{
                    fontWeight: 600,
                    marginBottom: 4,
                    color: allPicked ? '#389e0d' : somePicked ? '#d48806' : '#fa8c16',
                  }}
                >
                  {allPicked
                    ? t('shipConfirm.allItemsPicked')
                    : somePicked
                      ? t('shipConfirm.someItemsPicked')
                      : t('shipConfirm.noItemsPicked')}
                </div>
                <div style={{ color: '#595959', fontSize: 12 }}>
                  {t('shipConfirm.pickedSummary', {
                    picked: pickedLines.length,
                    total: orderDetail.lines?.length ?? 0,
                  })}
                </div>
              </div>

              {/* Line items table */}
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
                    fontSize: 14,
                    fontWeight: 600,
                    color: '#262626',
                    marginBottom: 12,
                  }}
                >
                  {t('orderLookup.lineItems')}
                </h3>

                {/* Table header */}
                <div
                  style={{
                    display: 'grid',
                    gridTemplateColumns: '2fr 1.5fr 1.5fr 1fr',
                    gap: 8,
                    paddingBottom: 8,
                    borderBottom: '2px solid #f0f0f0',
                    fontWeight: 600,
                    fontSize: 12,
                    color: '#8c8c8c',
                  }}
                >
                  <div>{t('orderLookup.skuCode')}</div>
                  <div>{t('orderLookup.orderedQty')}</div>
                  <div>{t('orderLookup.fulfilledQty')}</div>
                  <div style={{ textAlign: 'right' }}>{t('shipConfirm.picked')}</div>
                </div>

                {orderDetail.lines?.map((line) => {
                  const isFulfilled =
                    line.status === 'fulfilled' || line.status === 'partial'
                  return (
                    <div
                      key={line.id}
                      style={{
                        display: 'grid',
                        gridTemplateColumns: '2fr 1.5fr 1.5fr 1fr',
                        gap: 8,
                        padding: '10px 0',
                        borderBottom: '1px solid #fafafa',
                        fontSize: 13,
                        color: '#262626',
                      }}
                    >
                      <div style={{ fontFamily: 'monospace', fontSize: 12 }}>
                        {line.sku_id.substring(0, 8)}...
                      </div>
                      <div>{line.ordered_qty} {line.uom}</div>
                      <div style={{ color: isFulfilled ? '#389e0d' : '#8c8c8c' }}>
                        {line.fulfilled_qty} {line.uom}
                      </div>
                      <div style={{ textAlign: 'right' }}>
                        {isFulfilled ? (
                          <span style={{ color: '#389e0d', fontWeight: 600 }}>{'✓'}</span>
                        ) : (
                          <span style={{ color: '#d9d9d9' }}>{'—'}</span>
                        )}
                      </div>
                    </div>
                  )
                })}

                {(!orderDetail.lines || orderDetail.lines.length === 0) && (
                  <div
                    style={{
                      padding: '16px',
                      textAlign: 'center',
                      fontSize: 13,
                      color: '#8c8c8c',
                    }}
                  >
                    {t('orderLookup.noLines')}
                  </div>
                )}
              </div>

              {/* Action button */}
              <button
                onClick={handleProceedToShipping}
                style={{
                  width: '100%',
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
                {t('shipConfirm.proceedToShipping')}
              </button>

              {/* Warn if some items not picked — but still allow proceeding */}
              {!allPicked && (
                <div
                  style={{
                    marginTop: 12,
                    padding: '10px 14px',
                    background: '#fffbe6',
                    border: '1px solid #ffe58f',
                    borderRadius: 8,
                    fontSize: 12,
                    color: '#d48806',
                    textAlign: 'center',
                  }}
                >
                  {t('shipConfirm.partialPickWarning')}
                </div>
              )}
            </>
          )}
        </>
      )}

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 3: Enter shipping information */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'enter_shipping' && (
        <>
          {/* Order summary banner */}
          {orderDetail && (
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
              {t('shipConfirm.shippingForOrder', { orderNo: orderDetail.order_no })}
            </div>
          )}

          {/* Shipping info form */}
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
              {t('shipConfirm.enterShippingInfo')}
            </h3>

            {/* Carrier */}
            <div style={{ marginBottom: 14 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 13,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 6,
                }}
              >
                {t('shipConfirm.carrier')} <span style={{ color: '#ff4d4f' }}>*</span>
              </label>
              <input
                type="text"
                value={carrier}
                onChange={(e) => setCarrier(e.target.value)}
                placeholder={t('shipConfirm.carrierPlaceholder')}
                disabled={isSubmitting}
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  fontSize: 15,
                  border: carrier.trim() ? '2px solid #d9d9d9' : '2px solid #ff4d4f',
                  borderRadius: 8,
                  outline: 'none',
                  background: isSubmitting ? '#f5f5f5' : '#fff',
                  boxSizing: 'border-box',
                }}
                onFocus={(e) => {
                  e.currentTarget.style.borderColor = '#1677ff'
                }}
                onBlur={(e) => {
                  e.currentTarget.style.borderColor = carrier.trim() ? '#d9d9d9' : '#ff4d4f'
                }}
              />
            </div>

            {/* Carrier Service */}
            <div style={{ marginBottom: 14 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 13,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 6,
                }}
              >
                {t('shipConfirm.carrierService')}
              </label>
              <input
                type="text"
                value={carrierService}
                onChange={(e) => setCarrierService(e.target.value)}
                placeholder={t('shipConfirm.carrierServicePlaceholder')}
                disabled={isSubmitting}
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  fontSize: 15,
                  border: '2px solid #d9d9d9',
                  borderRadius: 8,
                  outline: 'none',
                  background: isSubmitting ? '#f5f5f5' : '#fff',
                  boxSizing: 'border-box',
                }}
                onFocus={(e) => {
                  e.currentTarget.style.borderColor = '#1677ff'
                }}
                onBlur={(e) => {
                  e.currentTarget.style.borderColor = '#d9d9d9'
                }}
              />
            </div>

            {/* Tracking Number */}
            <div style={{ marginBottom: 14 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 13,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 6,
                }}
              >
                {t('shipConfirm.trackingNo')}
              </label>
              <BarcodeScanner
                onScan={(barcode) => setTrackingNo(barcode)}
                placeholder={t('shipConfirm.trackingNoPlaceholder')}
              />
              {/* Show scanned/typed tracking number */}
              {trackingNo && (
                <div
                  style={{
                    marginTop: 8,
                    padding: '8px 14px',
                    background: '#f6ffed',
                    border: '1px solid #b7eb8f',
                    borderRadius: 8,
                    fontSize: 14,
                    fontFamily: 'monospace',
                    color: '#389e0d',
                  }}
                >
                  {t('shipConfirm.trackingNoLabel')}: {trackingNo}
                </div>
              )}
            </div>

            {/* Notes */}
            <div style={{ marginBottom: 4 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 13,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 6,
                }}
              >
                {t('orderLookup.notes')}
              </label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder={t('shipConfirm.notesPlaceholder')}
                disabled={isSubmitting}
                rows={2}
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  fontSize: 14,
                  border: '2px solid #d9d9d9',
                  borderRadius: 8,
                  outline: 'none',
                  background: isSubmitting ? '#f5f5f5' : '#fff',
                  boxSizing: 'border-box',
                  resize: 'vertical',
                  fontFamily: 'inherit',
                }}
                onFocus={(e) => {
                  e.currentTarget.style.borderColor = '#1677ff'
                }}
                onBlur={(e) => {
                  e.currentTarget.style.borderColor = '#d9d9d9'
                }}
              />
            </div>
          </div>

          {/* Confirm button */}
          <button
            onClick={handleConfirmShipment}
            disabled={isSubmitting || !carrier.trim()}
            style={{
              width: '100%',
              padding: '14px 0',
              fontSize: 15,
              fontWeight: 600,
              color: '#fff',
              background: isSubmitting || !carrier.trim() ? '#d9d9d9' : '#389e0d',
              border: 'none',
              borderRadius: 8,
              cursor: isSubmitting || !carrier.trim() ? 'not-allowed' : 'pointer',
              marginBottom: 16,
              transition: 'background 0.2s',
            }}
          >
            {isSubmitting ? t('shipConfirm.creatingShipment') : t('shipConfirm.confirmAndCreate')}
          </button>

          {/* Cancel / start over */}
          <div>
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

      {/* ════════════════════════════════════════════════════════════════ */}
      {/* STEP 4: Confirmation — shipment created */}
      {/* ════════════════════════════════════════════════════════════════ */}
      {step === 'confirm' && createdShipment && (
        <>
          {/* Success banner */}
          <div
            style={{
              padding: '20px',
              background: '#f6ffed',
              border: '1px solid #b7eb8f',
              borderRadius: 12,
              marginBottom: 16,
              textAlign: 'center',
            }}
          >
            <div style={{ fontSize: 48, marginBottom: 8 }}>{'✅'}</div>
            <div
              style={{
                fontSize: 18,
                fontWeight: 700,
                color: '#389e0d',
                marginBottom: 8,
              }}
            >
              {t('shipConfirm.shipmentCreated')}
            </div>
            <div style={{ fontSize: 14, color: '#595959' }}>
              {t('shipConfirm.shipmentCreatedDesc')}
            </div>
          </div>

          {/* Shipment details card */}
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
              {t('shipConfirm.shipmentDetails')}
            </h3>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <DetailRow
                label={t('shipConfirm.shipmentNo')}
                value={createdShipment.shipment_no}
              />
              <DetailRow
                label={t('shipConfirm.carrier')}
                value={createdShipment.carrier}
              />
              {createdShipment.tracking_no && (
                <DetailRow
                  label={t('shipConfirm.trackingNoLabel')}
                  value={createdShipment.tracking_no}
                />
              )}
              {createdShipment.carrier_service && (
                <DetailRow
                  label={t('shipConfirm.carrierService')}
                  value={createdShipment.carrier_service}
                />
              )}
              <DetailRow
                label={t('shipConfirm.statusLabel')}
                value={t(`shipConfirm.status${capitalize(createdShipment.status)}`)}
              />
              <DetailRow
                label={t('orderLookup.createdAt')}
                value={createdShipment.created_at?.substring(0, 19)?.replace('T', ' ') ?? '—'}
              />
            </div>
          </div>

          {/* Start new / done button */}
          <button
            onClick={handleResetAll}
            style={{
              width: '100%',
              padding: '14px 0',
              fontSize: 15,
              fontWeight: 600,
              color: '#fff',
              background: '#1677ff',
              border: 'none',
              borderRadius: 8,
              cursor: 'pointer',
              marginBottom: 16,
            }}
          >
            {t('shipConfirm.startNew')}
          </button>

          {/* Back to scan button */}
          <div>
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
              {t('receive.backToScan')}
            </button>
          </div>
        </>
      )}
    </div>
  )
}

// ── Step badge component ────────────────────────────────────────────────

function StepBadge({
  num,
  done,
  active,
  pending,
}: {
  num: number
  done: boolean
  active: boolean
  pending?: boolean
}) {
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: 22,
        height: 22,
        borderRadius: '50%',
        background: done || active ? '#1677ff' : pending ? '#f0f0f0' : '#f0f0f0',
        color: done || active ? '#fff' : '#8c8c8c',
        fontSize: 12,
        fontWeight: 700,
      }}
    >
      {done ? '✓' : num}
    </span>
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

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}
