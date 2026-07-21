// Scanner page — barcode scanning hub for warehouse operators.
// Provides quick access to scan-and-act workflows.
// All UI text is translated via react-i18next.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import BarcodeScanner from '@/components/BarcodeScanner'

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
  ]

  const handleScan = useCallback((barcode: string) => {
    setScanLog((prev) => [barcode, ...prev].slice(0, 20))
  }, [])

  const handleActionClick = useCallback(
    (_action: ScanAction) => {
      navigate(`/tasks`)
    },
    [navigate],
  )

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
                style={{
                  padding: '12px 16px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  borderBottom: i < scanLog.length - 1 ? '1px solid #f0f0f0' : 'none',
                  fontSize: 14,
                  color: '#262626',
                }}
              >
                <span>{barcode}</span>
                <span style={{ fontSize: 12, color: '#bfbfbf' }}>#{i + 1}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
