// Scanner page — barcode scanning hub for warehouse operators.
// Provides quick access to scan-and-act workflows.
// P3-10+ will implement full scanning integrations for each flow.

import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import BarcodeScanner from '@/components/BarcodeScanner'

// ── Scan action cards ─────────────────────────────────────────────────

interface ScanAction {
  key: string
  icon: string
  label: string
  description: string
}

const SCAN_ACTIONS: ScanAction[] = [
  {
    key: 'receiving',
    icon: '📥',
    label: 'Receive ASN',
    description: 'Scan inbound shipment barcode to start receiving',
  },
  {
    key: 'putaway',
    icon: '📦',
    label: 'Putaway',
    description: 'Scan item to find target location',
  },
  {
    key: 'picking',
    icon: '🛒',
    label: 'Pick Order',
    description: 'Scan order barcode to start picking',
  },
  {
    key: 'locate',
    icon: '📍',
    label: 'Locate',
    description: 'Scan location or item barcode to view details',
  },
]

export default function ScanPage() {
  const navigate = useNavigate()
  const [scanLog, setScanLog] = useState<string[]>([])

  const handleScan = useCallback((barcode: string) => {
    setScanLog((prev) => [barcode, ...prev].slice(0, 20))
    // P3-10+: route to appropriate flow based on barcode prefix or context
  }, [])

  const handleActionClick = useCallback(
    (action: ScanAction) => {
      // P3-10+: navigate to specific workflow wizards
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
              {action.label}
            </div>
            <div style={{ fontSize: 12, color: '#8c8c8c', lineHeight: 1.4 }}>
              {action.description}
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
          Quick Scan
        </h3>
        <BarcodeScanner onScan={handleScan} placeholder="Scan or type barcode..." />
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
            Recent Scans ({scanLog.length})
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
