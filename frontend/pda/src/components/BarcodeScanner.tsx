// Barcode scanner component — placeholder for html5-qrcode integration.
// Supports both camera-based scanning and keyboard wedge (hardware scanner) input.
// Full implementation in P3-10.

import { useState, useCallback, useRef, useEffect, type FormEvent } from 'react'

interface BarcodeScannerProps {
  onScan: (barcode: string) => void
  placeholder?: string
}

export default function BarcodeScanner({ onScan, placeholder = 'Scan a barcode...' }: BarcodeScannerProps) {
  const [manualInput, setManualInput] = useState('')
  const [lastScanned, setLastScanned] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

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

    // Re-focus for next scan
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
          placeholder={placeholder}
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

      {/* Camera scanner placeholder — P3-10 will add html5-qrcode */}
      <div className="pda-scanner-placeholder">
        <span className="scanner-icon">📷</span>
        <p style={{ fontSize: 14, marginBottom: 4 }}>
          Camera scanner coming soon
        </p>
        <p style={{ fontSize: 12, color: '#bfbfbf' }}>
          Use the input above or scan with a hardware scanner
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
          Scanned: <strong>{lastScanned}</strong>
        </div>
      )}
    </div>
  )
}
