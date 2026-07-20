// PDA Layout — mobile-first shell with bottom tab bar.
// Wraps all authenticated PDA pages with mobile navigation.

import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'

// ── Tab definitions ───────────────────────────────────────────────────

interface TabItem {
  key: string
  icon: string
  label: string
  path: string
}

const tabs: TabItem[] = [
  { key: 'tasks', icon: '📋', label: 'Tasks', path: '/tasks' },
  { key: 'scan', icon: '📷', label: 'Scan', path: '/scan' },
  { key: 'profile', icon: '👤', label: 'Me', path: '/profile' },
]

// ── Helper: map path → tab key ────────────────────────────────────────

function getActiveTab(pathname: string): string {
  if (pathname.startsWith('/tasks')) return 'tasks'
  if (pathname.startsWith('/scan')) return 'scan'
  if (pathname.startsWith('/profile')) return 'profile'
  return 'tasks'
}

// ── Header title ──────────────────────────────────────────────────────

function getHeaderTitle(pathname: string): string {
  if (pathname.startsWith('/tasks')) return 'Tasks'
  if (pathname.startsWith('/scan')) return 'Scanner'
  if (pathname.startsWith('/profile')) return 'Profile'
  return 'AI-WMS PDA'
}

export default function PdaLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { clearTokens } = useAuth()

  const activeTab = getActiveTab(location.pathname)
  const title = getHeaderTitle(location.pathname)

  const handleTabClick = (tab: TabItem) => {
    navigate(tab.path)
  }

  const handleLogout = () => {
    clearTokens()
    navigate('/login')
  }

  return (
    <div className="pda-layout">
      {/* Header */}
      <div className="pda-header">
        <span className="pda-title">{title}</span>
        <button
          onClick={handleLogout}
          style={{
            background: 'rgba(255,255,255,0.15)',
            border: 'none',
            color: '#fff',
            padding: '6px 12px',
            borderRadius: 6,
            fontSize: 13,
            cursor: 'pointer',
          }}
        >
          Logout
        </button>
      </div>

      {/* Content */}
      <div className="pda-content">
        <Outlet />
      </div>

      {/* Bottom tab bar */}
      <div className="pda-tab-bar">
        {tabs.map((tab) => (
          <div
            key={tab.key}
            className={`pda-tab-item${activeTab === tab.key ? ' active' : ''}`}
            onClick={() => handleTabClick(tab)}
          >
            <span className="pda-tab-icon">{tab.icon}</span>
            <span>{tab.label}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
