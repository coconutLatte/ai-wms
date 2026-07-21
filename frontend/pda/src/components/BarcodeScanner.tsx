// Barcode scanner component — supports keyboard wedge (hardware scanner) input.
// Camera-based scanning is planned for a future release.
// All UI text is translated via react-i18next.

import { useState, useCallback, useRef, useEffect, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'

interface BarcodeScannerProps {
  onScan: (barcode: string) => void
  placeholder?: string
}

export default function BarcodeScanner({ onScan, placeholder }: BarcodeScannerProps) {
  const { t } = useTranslation()
  const [manualInput, setManualInput] = useState('')
  const [lastScanned, setLastScanned] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const defaultPlaceholder = placeholder ?? t('scan.scanPlaceholder')

  // Auto-focus the input field so hardware scanners (keyboard wedge) work immediately.
  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  // Hardware scanners typically end input with an Enter key.
  const handleSubmit = useCallback((e: FormEvent) => {
    e.preventDefault()
    const trimmed = manualInput.trim()
    if (!trimmed) return

    setLastScanned(trimmed)
    setManualInput('')
    onScan(trimmed)

    inputRef.current?.focus()
  }, [manualInput, onScan])

  // Also handle paste events (some scanners emit paste)
  const handlePaste = useCallback((e: React.ClipboardEvent) => {
    const pasted = e.clipboardData.getData('text').trim()
    if (pasted) {
      e.preventDefault()
      setLastScanned(pasted)
      onScan(pasted)
    }
  }, [onScan])

  return (
    <div>
      {/* Manual input — always available as keyboard wedge fallback */}
      <form onSubmit={handleSubmit} style={{ marginBottom: 12 }}>
        <input
          ref={inputRef}
          type="text"
          value={manualInput}
          onChange={(e) => setManualInput(e.target.value)}
          onPaste={handlePaste}
          placeholder={defaultPlaceholder}
          autoComplete="off"
          autoCorrect="off"
          autoCapitalize="off"
          style={{
            width: '100%',
            padding: '14px 16px',
            fontSize: 16,
            border: '2px solid #d9d9d9',
            borderRadius: 10,
            outline: 'none',
            transition: 'border-color 0.2s',
          }}
          onFocus={(e) => {
            e.currentTarget.style.borderColor = '#1677ff'
          }}
          onBlur={(e) => {
            e.currentTarget.style.borderColor = '#d9d9d9'
          }}
        />
      </form>

      {/* Camera scanner placeholder */}
      <div className="pda-scanner-placeholder">
        <span className="scanner-icon">📷</span>
        <p style={{ fontSize: 14, marginBottom: 4 }}>
          {t('scan.cameraComing')}
        </p>
        <p style={{ fontSize: 12, color: '#bfbfbf' }}>
          {t('scan.cameraHint')}
        </p>
      </div>

      {/* Last scanned feedback */}
      {lastScanned && (
        <div
          style={{
            marginTop: 12,
            padding: '12px 16px',
            background: '#f6ffed',
            border: '1px solid #b7eb8f',
            borderRadius: 8,
            fontSize: 14,
            color: '#389e0d',
          }}
        >
          {t('scan.scanned')}: <strong>{lastScanned}</strong>
        </div>
      )}
    </div>
  )
}
