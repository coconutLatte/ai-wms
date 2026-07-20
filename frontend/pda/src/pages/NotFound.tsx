// 404 page for PDA — mobile-friendly not-found screen.

import { useNavigate } from 'react-router-dom'

export default function NotFoundPage() {
  const navigate = useNavigate()

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        padding: 32,
        textAlign: 'center',
      }}
    >
      <div style={{ fontSize: 72, marginBottom: 16 }}>📦</div>
      <h1 style={{ fontSize: 48, fontWeight: 800, color: '#1677ff', marginBottom: 8 }}>
        404
      </h1>
      <p style={{ fontSize: 16, color: '#8c8c8c', marginBottom: 24 }}>
        Page not found
      </p>
      <button
        onClick={() => navigate('/tasks', { replace: true })}
        style={{
          padding: '12px 32px',
          fontSize: 15,
          fontWeight: 600,
          color: '#fff',
          background: '#1677ff',
          border: 'none',
          borderRadius: 8,
          cursor: 'pointer',
        }}
      >
        Go to Tasks
      </button>
    </div>
  )
}
