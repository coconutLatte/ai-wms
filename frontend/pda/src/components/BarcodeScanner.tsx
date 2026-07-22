// Barcode scanner component — supports camera scanning (via @zxing/library) and manual
// keyboard wedge input for hardware scanners. Camera scanner replaces the previous
// placeholder with a live video viewfinder that auto-detects 1D/2D barcodes.
// All UI text is translated via react-i18next.

import { useState, useCallback, useRef, useEffect, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { BrowserMultiFormatReader } from '@zxing/library'

interface BarcodeScannerProps {
  onScan: (barcode: string) => void
  placeholder?: string
}

export default function BarcodeScanner({ onScan, placeholder }: BarcodeScannerProps) {
  const { t } = useTranslation()
  const [manualInput, setManualInput] = useState('')
  const [lastScanned, setLastScanned] = useState<string | null>(null)
  const [cameraActive, setCameraActive] = useState(false)
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [cameraLoading, setCameraLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const videoRef = useRef<HTMLDivElement>(null)
  const readerRef = useRef<BrowserMultiFormatReader | null>(null)

  const defaultPlaceholder = placeholder ?? t('scan.scanPlaceholder')

  // Auto-focus the input field so hardware scanners (keyboard wedge) work immediately.
  useEffect(() => {
    if (!cameraActive) {
      inputRef.current?.focus()
    }
  }, [cameraActive])

  // Clean up the camera reader on unmount
  useEffect(() => {
    const reader = readerRef.current
    return () => {
      if (reader) {
        reader.reset()
      }
    }
  }, [])

  // ── Manual / Keyboard Wedge Input ────────────────────────────────────────

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

  // ── Camera Scanner ───────────────────────────────────────────────────────

  const startCamera = useCallback(async () => {
    setCameraError(null)
    setCameraLoading(true)

    try {
      const reader = new BrowserMultiFormatReader()
      readerRef.current = reader

      // Look for a rear-facing (environment) camera on mobile devices
      const devices = await reader.listVideoInputDevices()
      let deviceId: string | null = null

      // Prefer back camera on mobile
      const backCam = devices.find((d) =>
        d.label.toLowerCase().includes('back') ||
        d.label.toLowerCase().includes('rear') ||
        d.label.toLowerCase().includes('environment'))
      if (backCam) {
        deviceId = backCam.deviceId
      } else if (devices.length > 0) {
        deviceId = devices[0].deviceId
      }

      // Only proceed if the video container exists (component is still mounted)
      if (!videoRef.current) return

      setCameraActive(true)
      setCameraLoading(false)

      // Start continuous scanning from video stream
      await reader.decodeFromVideoDevice(
        deviceId,
        videoRef.current,
        (result, err) => {
          if (result) {
            // Barcode detected!
            const text = result.getText()
            setLastScanned(text)
            onScan(text)
            // Stop camera after successful scan
            reader.reset()
            setCameraActive(false)
          }
          if (err && !(err instanceof TypeError)) {
            // Ignore TypeError (thrown when no barcode found in a frame — expected)
            // Only report unexpected errors
            if (err instanceof DOMException) {
              setCameraError(t('scan.cameraPermissionDenied'))
              reader.reset()
              setCameraActive(false)
            }
          }
        },
      )
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err)
      if (msg.includes('NotAllowedError') || msg.includes('Permission denied')) {
        setCameraError(t('scan.cameraPermissionDenied'))
      } else if (msg.includes('NotFoundError') || msg.includes('No video device')) {
        setCameraError(t('scan.cameraNotFound'))
      } else {
        setCameraError(t('scan.cameraInitFailed', { error: msg }))
      }
      setCameraLoading(false)
      setCameraActive(false)
    }
  }, [onScan, t])

  const stopCamera = useCallback(() => {
    if (readerRef.current) {
      readerRef.current.reset()
    }
    setCameraActive(false)
    setCameraError(null)
  }, [])

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div>
      {/* Camera Scanner Toggle */}
      <div style={{ marginBottom: 12 }}>
        {!cameraActive ? (
          <button
            onClick={startCamera}
            disabled={cameraLoading}
            style={{
              width: '100%',
              padding: '14px 16px',
              fontSize: 16,
              fontWeight: 600,
              color: '#fff',
              background: cameraLoading ? '#91caff' : '#1677ff',
              border: 'none',
              borderRadius: 10,
              cursor: cameraLoading ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 8,
            }}
          >
            <span style={{ fontSize: 20 }}>{'\u{1F4F7}'}</span>
            {cameraLoading ? t('scan.startingCamera') : t('scan.openCamera')}
          </button>
        ) : (
          <button
            onClick={stopCamera}
            style={{
              width: '100%',
              padding: '14px 16px',
              fontSize: 16,
              fontWeight: 600,
              color: '#595959',
              background: '#f0f0f0',
              border: 'none',
              borderRadius: 10,
              cursor: 'pointer',
            }}
          >
            {t('scan.closeCamera')}
          </button>
        )}
      </div>

      {/* Camera Viewfinder */}
      {cameraActive && (
        <div style={{ marginBottom: 12 }}>
          <div
            ref={videoRef}
            style={{
              width: '100%',
              maxWidth: 400,
              margin: '0 auto',
              aspectRatio: '1',
              borderRadius: 12,
              overflow: 'hidden',
              background: '#000',
              position: 'relative',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {/* Scanning overlay frame */}
            <div
              style={{
                position: 'absolute',
                inset: 0,
                border: '3px solid rgba(22, 119, 255, 0.5)',
                margin: '10%',
                borderRadius: 8,
                pointerEvents: 'none',
                zIndex: 10,
              }}
            />
            {/* Corner markers */}
            <svg
              style={{
                position: 'absolute',
                inset: 0,
                pointerEvents: 'none',
                zIndex: 11,
              }}
              viewBox="0 0 100 100"
            >
              {[
                'M 9,19 L 9,9 L 19,9',
                'M 81,9 L 91,9 L 91,19',
                'M 9,81 L 9,91 L 19,91',
                'M 81,91 L 91,91 L 91,81',
              ].map((d) => (
                <path key={d} d={d} stroke="#1677ff" strokeWidth="3" fill="none" />
              ))}
            </svg>
            {/* Scanning text */}
            <span
              style={{
                position: 'absolute',
                bottom: '8%',
                color: '#fff',
                fontSize: 13,
                fontWeight: 500,
                zIndex: 10,
                textShadow: '0 1px 3px rgba(0,0,0,0.8)',
                pointerEvents: 'none',
              }}
            >
              {t('scan.cameraScanning')}
            </span>
          </div>
        </div>
      )}

      {/* Camera error */}
      {cameraError && (
        <div
          style={{
            marginBottom: 12,
            padding: '10px 14px',
            background: '#fff1f0',
            border: '1px solid #ffa39e',
            borderRadius: 8,
            fontSize: 13,
            color: '#cf1322',
          }}
        >
          {cameraError}
        </div>
      )}

      {/* Manual input — always available as keyboard wedge fallback */}
      <form onSubmit={handleSubmit} style={{ marginBottom: 12 }}>
        <input
          ref={inputRef}
          type="text"
          value={manualInput}
          onChange={(e) => setManualInput(e.target.value)}
          onPaste={handlePaste}
          placeholder={cameraActive ? t('scan.manualOverride') : defaultPlaceholder}
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
