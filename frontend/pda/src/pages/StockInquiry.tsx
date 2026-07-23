// Stock Inquiry page — scan a location or SKU barcode to view current
// inventory levels. Displays the scanned entity info along with inventory
// records showing qty, available, reserved, batch_no, and status.
// All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import BarcodeScanner from '@/components/BarcodeScanner'
import { inquireStock } from '@/api/stockInquiry'
import type { StockInquiryResponse, StockInquiryInventory } from '@/api/stockInquiry'

// ── Status labels & colors ─────────────────────────────────────────────

function getInventoryStatusLabelKey(status: string): string {
  const map: Record<string, string> = {
    available: 'stockInquiry.statusAvailable',
    quarantine: 'stockInquiry.statusQuarantine',
    damaged: 'stockInquiry.statusDamaged',
    expired: 'stockInquiry.statusExpired',
  }
  return map[status] || 'stockInquiry.statusAvailable'
}

function getInventoryStatusColor(status: string): string {
  const map: Record<string, string> = {
    available: '#389e0d',
    quarantine: '#fa8c16',
    damaged: '#cf1322',
    expired: '#8c8c8c',
  }
  return map[status] || '#595959'
}

// ── Main component ──────────────────────────────────────────────────────

export default function StockInquiryPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [barcode, setBarcode] = useState('')

  // ── Fetch stock inquiry result ────────────────────────────────────────

  const {
    data: result,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery<StockInquiryResponse>({
    queryKey: ['stockInquiry', barcode],
    queryFn: async () => inquireStock(barcode),
    enabled: barcode.length > 0,
    retry: false,
  })

  // ── Handlers ──────────────────────────────────────────────────────────

  const handleScan = useCallback((scanned: string) => {
    setBarcode(scanned)
  }, [])

  const handleClear = useCallback(() => {
    setBarcode('')
  }, [])

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '—'
    return new Date(dateStr).toLocaleString()
  }

  // Determine if no inventory was found (entity resolved but no records).
  const isEmptyResult = result && !result.inventory.length && result.entity_type === ''

  // ── Render ────────────────────────────────────────────────────────────

  return (
    <div>
      {/* Scan section */}
      {!result && (
        <>
          <h3
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: '#595959',
              marginBottom: 10,
            }}
          >
            {t('stockInquiry.title')}
          </h3>

          <BarcodeScanner
            onScan={handleScan}
            placeholder={t('stockInquiry.scanPlaceholder')}
          />

          {/* Loading state */}
          {isLoading && (
            <div
              style={{
                padding: '16px',
                textAlign: 'center',
                color: '#1677ff',
                fontSize: 14,
                marginTop: 16,
              }}
            >
              {'\u{1F50D}'} {t('stockInquiry.lookingUp')}
            </div>
          )}

          {/* Error state */}
          {isError && (
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
              {t('stockInquiry.lookupFailed', {
                error: (error as Error)?.message || '',
              })}
            </div>
          )}

          {/* Not-found: entity resolved but empty inventory */}
          {isEmptyResult && (
            <div
              style={{
                marginTop: 16,
                padding: '14px 18px',
                background: '#fff7e6',
                border: '1px solid #ffd591',
                borderRadius: 10,
                fontSize: 14,
                color: '#ad6800',
              }}
            >
              {t('stockInquiry.barcodeNotFound', { barcode })}
            </div>
          )}

          {/* Back button */}
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
            {t('stockInquiry.backToScan')}
          </button>
        </>
      )}

      {/* Result view */}
      {result && (
        <>
          {/* Back button */}
          <button
            onClick={() => {
              queryClient.removeQueries({ queryKey: ['stockInquiry', barcode] })
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
            {t('stockInquiry.backToScan')}
          </button>

          {/* Entity info card */}
          <div
            style={{
              background: '#fff',
              borderRadius: 12,
              padding: 20,
              boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              marginBottom: 16,
            }}
          >
            {/* Header with entity type badge */}
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: 14,
              }}
            >
              <h2
                style={{
                  fontSize: 18,
                  fontWeight: 700,
                  color: '#262626',
                  margin: 0,
                }}
              >
                {result.entity_type === 'location'
                  ? result.location?.code || ''
                  : result.sku?.name || result.sku?.code || ''}
              </h2>
              <span
                style={{
                  display: 'inline-block',
                  padding: '4px 10px',
                  fontSize: 12,
                  fontWeight: 600,
                  color: '#fff',
                  background: result.entity_type === 'location' ? '#1677ff' : '#389e0d',
                  borderRadius: 4,
                  whiteSpace: 'nowrap',
                }}
              >
                {result.entity_type === 'location'
                  ? t('stockInquiry.location')
                  : t('stockInquiry.sku')}
              </span>
            </div>

            {/* Entity details */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {result.location && (
                <>
                  <DetailRow label={t('stockInquiry.locationCode')} value={result.location.code} />
                  <DetailRow label={t('stockInquiry.locationBarcode')} value={result.location.barcode} />
                  <DetailRow label={t('stockInquiry.locationType')} value={result.location.location_type} />
                  <DetailRow label={t('stockInquiry.locationStatus')} value={result.location.status} />
                </>
              )}
              {result.sku && (
                <>
                  <DetailRow label={t('stockInquiry.skuCode')} value={result.sku.code} />
                  <DetailRow label={t('stockInquiry.skuName')} value={result.sku.name} />
                  <DetailRow label={t('stockInquiry.skuBarcode')} value={result.sku.barcode || '—'} />
                  <DetailRow label={t('stockInquiry.skuStatus')} value={result.sku.status} />
                </>
              )}
            </div>
          </div>

          {/* Aggregated totals card */}
          {result.inventory.length > 0 && (
            <div
              style={{
                background: 'linear-gradient(135deg, #e6f7ff 0%, #f0f5ff 100%)',
                borderRadius: 12,
                padding: 16,
                marginBottom: 16,
                boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
              }}
            >
              <h3
                style={{
                  fontSize: 14,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 12,
                  margin: 0,
                  marginBottom: 12,
                }}
              >
                {t('stockInquiry.inventorySummary')}
              </h3>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(3, 1fr)',
                  gap: 12,
                }}
              >
                <SummaryTile
                  label={t('stockInquiry.totalQty')}
                  value={result.total_qty}
                  color="#1677ff"
                />
                <SummaryTile
                  label={t('stockInquiry.totalReserved')}
                  value={result.total_reserved}
                  color="#fa8c16"
                />
                <SummaryTile
                  label={t('stockInquiry.totalAvailable')}
                  value={result.total_available}
                  color="#389e0d"
                />
              </div>
            </div>
          )}

          {/* Inventory records */}
          {result.inventory.length > 0 && (
            <>
              <h3
                style={{
                  fontSize: 15,
                  fontWeight: 600,
                  color: '#595959',
                  marginBottom: 10,
                }}
              >
                {t('stockInquiry.inventoryRecords')} ({result.inventory.length})
              </h3>

              {result.inventory.map((inv) => (
                <InventoryCard key={inv.id} inv={inv} t={t} formatDate={formatDate} />
              ))}
            </>
          )}

          {/* No inventory records */}
          {result.inventory.length === 0 && result.entity_type !== '' && (
            <div className="pda-empty">
              <span className="empty-icon">{'\u{1F4ED}'}</span>
              <p className="empty-text">{t('stockInquiry.noInventory')}</p>
            </div>
          )}

          {/* Clear & rescan */}
          <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
            <button
              onClick={() => {
                queryClient.removeQueries({ queryKey: ['stockInquiry', barcode] })
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
              {t('stockInquiry.clearAndRescan')}
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
              {t('stockInquiry.backToScan')}
            </button>
          </div>
        </>
      )}
    </div>
  )
}

// ── Helper components ───────────────────────────────────────────────────

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
    <div style={{ textAlign: 'center' }}>
      <div
        style={{
          fontSize: 22,
          fontWeight: 700,
          color,
          lineHeight: 1.2,
        }}
      >
        {value}
      </div>
      <div style={{ fontSize: 11, color: '#8c8c8c', marginTop: 4 }}>{label}</div>
    </div>
  )
}

function InventoryCard({
  inv,
  t,
  formatDate,
}: {
  inv: StockInquiryInventory
  t: (key: string, opts?: Record<string, unknown>) => string
  formatDate: (s?: string) => string
}) {
  const statusColor = getInventoryStatusColor(inv.status)

  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 12,
        padding: 16,
        boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        marginBottom: 12,
      }}
    >
      {/* Header: batch + status badge */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
        }}
      >
        <div>
          <div style={{ fontSize: 14, fontWeight: 600, color: '#262626' }}>
            {inv.batch_no || t('stockInquiry.noBatch')}
          </div>
          <div style={{ fontSize: 12, color: '#8c8c8c', fontFamily: 'monospace', marginTop: 2 }}>
            {t('stockInquiry.skuId')}: {inv.sku_id.substring(0, 8)}...
          </div>
        </div>
        <span
          style={{
            display: 'inline-block',
            padding: '3px 8px',
            fontSize: 11,
            fontWeight: 600,
            color: statusColor,
            background: `${statusColor}15`,
            borderRadius: 4,
            border: `1px solid ${statusColor}40`,
            whiteSpace: 'nowrap',
          }}
        >
          {t(getInventoryStatusLabelKey(inv.status))}
        </span>
      </div>

      {/* Qty grid */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr 1fr',
          gap: 8,
          fontSize: 13,
        }}
      >
        <div>
          <span style={{ color: '#8c8c8c' }}>{t('stockInquiry.qty')}:</span>{' '}
          <span style={{ fontWeight: 600, color: '#262626' }}>{inv.qty}</span>
        </div>
        <div>
          <span style={{ color: '#8c8c8c' }}>{t('stockInquiry.reserved')}:</span>{' '}
          <span style={{ fontWeight: 600, color: '#fa8c16' }}>{inv.reserved_qty}</span>
        </div>
        <div>
          <span style={{ color: '#8c8c8c' }}>{t('stockInquiry.available')}:</span>{' '}
          <span style={{ fontWeight: 600, color: '#389e0d' }}>{inv.available_qty}</span>
        </div>
      </div>

      {/* Dates */}
      <div style={{ marginTop: 10, fontSize: 12, color: '#bfbfbf' }}>
        {inv.production_date && (
          <span>
            {t('stockInquiry.productionDate')}: {formatDate(inv.production_date)}{' '}
          </span>
        )}
        {inv.expiry_date && (
          <span>
            {t('stockInquiry.expiryDate')}: {formatDate(inv.expiry_date)}{' '}
          </span>
        )}
        <div style={{ marginTop: 2 }}>
          {t('stockInquiry.receivedAt')}: {formatDate(inv.received_at)}
        </div>
      </div>
    </div>
  )
}
