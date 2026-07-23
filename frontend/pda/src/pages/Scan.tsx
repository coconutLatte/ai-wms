// Scanner page — barcode scanning hub for warehouse operators.
// Provides quick access to scan-and-act workflows plus real-time barcode lookup.
// Scanned barcodes are resolved via API: location barcodes → putaway task navigation,
// SKU barcodes → inventory info display. All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import BarcodeScanner from '@/components/BarcodeScanner'
import { lookupBarcode, type BarcodeResult } from '@/api/barcode'

// ── Scan action cards ─────────────────────────────────────────────────

interface ScanAction {
  key: string
  icon: string
  labelKey: string
  descriptionKey: string
}

export default function ScanPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [scanLog, setScanLog] = useState<string[]>([])
  const [barcodeResult, setBarcodeResult] = useState<BarcodeResult | null>(null)
  const [lookingUp, setLookingUp] = useState(false)
  const [lookupError, setLookupError] = useState<string | null>(null)

  const SCAN_ACTIONS: ScanAction[] = [
    {
      key: 'receiving',
      icon: '\u{1F4E5}',
      labelKey: 'scan.receiveAsn',
      descriptionKey: 'scan.receiveAsnDesc',
    },
    {
      key: 'putaway',
      icon: '\u{1F4E6}',
      labelKey: 'scan.putaway',
      descriptionKey: 'scan.putawayDesc',
    },
    {
      key: 'picking',
      icon: '\u{1F6D2}',
      labelKey: 'scan.pickOrder',
      descriptionKey: 'scan.pickOrderDesc',
    },
    {
      key: 'locate',
      icon: '\u{1F4CD}',
      labelKey: 'scan.locate',
      descriptionKey: 'scan.locateDesc',
    },
    {
      key: 'orderLookup',
      icon: '\u{1F4CB}',
      labelKey: 'scan.orderLookup',
      descriptionKey: 'scan.orderLookupDesc',
    },
  ]

  const handleScan = useCallback(async (barcode: string) => {
    setScanLog((prev) => [barcode, ...prev].slice(0, 20))
    setBarcodeResult(null)
    setLookupError(null)
    setLookingUp(true)

    try {
      const result = await lookupBarcode(barcode)
      if (result) {
        setBarcodeResult(result)
      } else {
        setLookupError(t('scan.barcodeNotFound', { barcode }))
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err)
      setLookupError(t('scan.lookupFailed', { error: msg }))
    } finally {
      setLookingUp(false)
    }
  }, [t])

  const handleActionClick = useCallback(
    (action: ScanAction) => {
      if (action.key === 'receiving') {
        navigate('/receive')
      } else if (action.key === 'putaway') {
        navigate('/putaway')
      } else if (action.key === 'picking') {
        navigate('/pick')
      } else if (action.key === 'orderLookup') {
        navigate('/order-lookup')
      } else if (action.key === 'locate') {
        navigate('/stock-inquiry')
      } else {
        navigate('/tasks')
      }
    },
    [navigate],
  )

  // Handle navigation from barcode result
  const handleResultAction = useCallback(() => {
    if (!barcodeResult) return
    if (barcodeResult.type === 'location') {
      // Navigate to tasks filtered for putaway at this location
      navigate(`/tasks`)
    } else if (barcodeResult.type === 'sku') {
      // Navigate to tasks — SKU details shown inline
      navigate(`/tasks`)
    }
  }, [barcodeResult, navigate])

  return (
    <div>
      {/* Scan actions grid */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(2, 1fr)',
          gap: 10,
          marginBottom: 16,
        }}
      >
        {SCAN_ACTIONS.map((action) => (
          <div
            key={action.key}
            onClick={() => handleActionClick(action)}
            style={{
              background: '#fff',
              borderRadius: 10,
              padding: 16,
              boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
              cursor: 'pointer',
              transition: 'box-shadow 0.2s',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.12)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.08)'
            }}
          >
            <div style={{ fontSize: 28, marginBottom: 8 }}>{action.icon}</div>
            <div style={{ fontSize: 14, fontWeight: 600, color: '#262626', marginBottom: 4 }}>
              {t(action.labelKey)}
            </div>
            <div style={{ fontSize: 12, color: '#8c8c8c', lineHeight: 1.4 }}>
              {t(action.descriptionKey)}
            </div>
          </div>
        ))}
      </div>

      {/* Barcode scanner section */}
      <div style={{ marginBottom: 16 }}>
        <h3
          style={{
            fontSize: 15,
            fontWeight: 600,
            color: '#595959',
            marginBottom: 10,
          }}
        >
          {t('scan.quickScan')}
        </h3>
        <BarcodeScanner onScan={handleScan} placeholder={t('scan.scanPlaceholder')} />
      </div>

      {/* Lookup loading indicator */}
      {lookingUp && (
        <div
          style={{
            padding: '16px',
            textAlign: 'center',
            color: '#1677ff',
            fontSize: 14,
          }}
        >
          {'\u{1F50D}'} {t('scan.lookingUp')}
        </div>
      )}

      {/* Lookup error */}
      {lookupError && (
        <div
          style={{
            marginTop: 0,
            marginBottom: 16,
            padding: '14px 18px',
            background: '#fff1f0',
            border: '1px solid #ffa39e',
            borderRadius: 10,
            fontSize: 14,
            color: '#cf1322',
          }}
        >
          {lookupError}
        </div>
      )}

      {/* Barcode result card */}
      {barcodeResult && !lookingUp && (
        <div
          style={{
            marginBottom: 16,
            background: '#fff',
            borderRadius: 12,
            boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
            overflow: 'hidden',
          }}
        >
          {/* Result header */}
          <div
            style={{
              padding: '14px 18px',
              background: barcodeResult.type === 'location' ? '#e6f7ff' : '#f6ffed',
              borderBottom: '1px solid #f0f0f0',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <span style={{ fontSize: 20 }}>
                {barcodeResult.type === 'location' ? '\u{1F4CD}' : '\u{1F4E6}'}
              </span>
              <div>
                <div style={{ fontSize: 12, color: '#8c8c8c', textTransform: 'uppercase' }}>
                  {t(barcodeResult.type === 'location' ? 'scan.locationBarcode' : 'scan.skuBarcode')}
                </div>
                <div style={{ fontSize: 15, fontWeight: 600, color: '#262626' }}>
                  {barcodeResult.type === 'location'
                    ? barcodeResult.code
                    : barcodeResult.name || barcodeResult.code}
                </div>
              </div>
            </div>
          </div>

          {/* Result body */}
          <div style={{ padding: '14px 18px' }}>
            {barcodeResult.type === 'location' && (
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.locationType')}</span>
                  <span style={{ fontSize: 13, color: '#262626' }}>
                    {barcodeResult.location_type}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.locationStatus')}</span>
                  <span
                    style={{
                      fontSize: 13,
                      color: barcodeResult.status === 'active' ? '#389e0d' : '#8c8c8c',
                    }}
                  >
                    {barcodeResult.status}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.barcode')}</span>
                  <span style={{ fontSize: 13, color: '#595959', fontFamily: 'monospace' }}>
                    {barcodeResult.barcode}
                  </span>
                </div>
                <button
                  onClick={handleResultAction}
                  style={{
                    width: '100%',
                    padding: '12px 0',
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#fff',
                    background: '#1677ff',
                    border: 'none',
                    borderRadius: 8,
                    cursor: 'pointer',
                  }}
                >
                  {t('scan.viewPutawayTasks')}
                </button>
              </div>
            )}

            {barcodeResult.type === 'sku' && (
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.skuCode')}</span>
                  <span style={{ fontSize: 13, color: '#262626', fontFamily: 'monospace' }}>
                    {barcodeResult.code}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.category')}</span>
                  <span style={{ fontSize: 13, color: '#262626' }}>
                    {barcodeResult.category || '-'}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.baseUnit')}</span>
                  <span style={{ fontSize: 13, color: '#262626' }}>
                    {barcodeResult.uom?.base_unit || '-'}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                  <span style={{ fontSize: 13, color: '#8c8c8c' }}>{t('scan.skuStatus')}</span>
                  <span
                    style={{
                      fontSize: 13,
                      color: barcodeResult.status === 'active' ? '#389e0d' : '#8c8c8c',
                    }}
                  >
                    {barcodeResult.status}
                  </span>
                </div>
                <button
                  onClick={handleResultAction}
                  style={{
                    width: '100%',
                    padding: '12px 0',
                    fontSize: 15,
                    fontWeight: 600,
                    color: '#fff',
                    background: '#389e0d',
                    border: 'none',
                    borderRadius: 8,
                    cursor: 'pointer',
                  }}
                >
                  {t('scan.viewInventoryInfo')}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Scan history */}
      {scanLog.length > 0 && (
        <div>
          <h3
            style={{
              fontSize: 15,
              fontWeight: 600,
              color: '#595959',
              marginBottom: 10,
            }}
          >
            {t('scan.recentScans', { count: scanLog.length })}
          </h3>
          <div
            style={{
              background: '#fff',
              borderRadius: 10,
              overflow: 'hidden',
              boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
            }}
          >
            {scanLog.map((barcode, i) => (
              <div
                key={`${barcode}-${i}`}
                onClick={() => handleScan(barcode)}
                style={{
                  padding: '12px 16px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  borderBottom: i < scanLog.length - 1 ? '1px solid #f0f0f0' : 'none',
                  fontSize: 14,
                  color: '#262626',
                  cursor: 'pointer',
                }}
              >
                <span style={{ fontFamily: 'monospace' }}>{barcode}</span>
                <span style={{ fontSize: 12, color: '#bfbfbf' }}>#{i + 1}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
