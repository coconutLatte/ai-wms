// Profile page — operator info and session management.
// P3-09 will integrate with GET /api/v1/auth/me.

import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'

export default function ProfilePage() {
  const navigate = useNavigate()
  const { clearTokens } = useAuth()

  const handleLogout = () => {
    clearTokens()
    navigate('/login', { replace: true })
  }

  return (
    <div>
      {/* Operator card */}
      <div
        style={{
          background: '#fff',
          borderRadius: 12,
          padding: 24,
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
          marginBottom: 16,
          textAlign: 'center',
        }}
      >
        <div
          style={{
            width: 72,
            height: 72,
            borderRadius: '50%',
            background: 'linear-gradient(135deg, #1677ff, #69b1ff)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 12px',
            fontSize: 28,
            color: '#fff',
            fontWeight: 700,
          }}
        >
          👤
        </div>
        <h2 style={{ fontSize: 18, fontWeight: 600, color: '#262626', marginBottom: 4 }}>
          Operator
        </h2>
        <p style={{ fontSize: 13, color: '#8c8c8c' }}>
          Warehouse Operations
        </p>
      </div>

      {/* Settings list */}
      <div
        style={{
          background: '#fff',
          borderRadius: 12,
          overflow: 'hidden',
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
          marginBottom: 16,
        }}
      >
        <ProfileRow label="Operator ID" value="op-001" />
        <ProfileRow label="Warehouse" value="Demo Warehouse" />
        <ProfileRow label="Role" value="Operator" />
        <ProfileRow label="Status" value="Active" isLast />
      </div>

      {/* Session info */}
      <div
        style={{
          background: '#fff',
          borderRadius: 12,
          overflow: 'hidden',
          boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
          marginBottom: 16,
        }}
      >
        <ProfileRow label="App Version" value="0.1.0" />
        <ProfileRow label="Session Started" value={new Date().toLocaleString()} isLast />
      </div>

      {/* Logout button */}
      <button
        onClick={handleLogout}
        style={{
          width: '100%',
          padding: '14px 0',
          fontSize: 15,
          fontWeight: 600,
          color: '#cf1322',
          background: '#fff',
          border: '1px solid #ffccc7',
          borderRadius: 10,
          cursor: 'pointer',
          marginBottom: 16,
        }}
      >
        Sign Out
      </button>
    </div>
  )
}

// ── Profile row ───────────────────────────────────────────────────────

function ProfileRow({
  label,
  value,
  isLast = false,
}: {
  label: string
  value: string
  isLast?: boolean
}) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '14px 16px',
        borderBottom: isLast ? 'none' : '1px solid #f0f0f0',
      }}
    >
      <span style={{ fontSize: 14, color: '#595959' }}>{label}</span>
      <span style={{ fontSize: 14, fontWeight: 500, color: '#262626' }}>{value}</span>
    </div>
  )
}
