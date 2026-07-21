// 404 page for PDA — mobile-friendly not-found screen with i18n.

import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export default function NotFoundPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()

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
        {t('notFound.title')}
      </h1>
      <p style={{ fontSize: 16, color: '#8c8c8c', marginBottom: 24 }}>
        {t('notFound.message')}
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
        {t('notFound.goToTasks')}
      </button>
    </div>
  )
}
