// PDA Layout — mobile-first shell with bottom tab bar.
// Wraps all authenticated PDA pages with mobile navigation.
// Includes language switcher in the header.

import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/hooks/useAuth'
import LanguageSwitcher from '@/components/LanguageSwitcher'

// ── Tab definitions ───────────────────────────────────────────────────

interface TabItem {
  key: string
  icon: string
  labelKey: string
  path: string
}

const tabs: TabItem[] = [
  { key: 'tasks', icon: '\u{1F4CB}', labelKey: 'tabs.tasks', path: '/tasks' },
  { key: 'scan', icon: '\u{1F4F7}', labelKey: 'tabs.scan', path: '/scan' },
  { key: 'profile', icon: '\u{1F464}', labelKey: 'tabs.profile', path: '/profile' },
]

// ── Helper: map path → tab key ────────────────────────────────────────

function getActiveTab(pathname: string): string {
  if (pathname.startsWith('/tasks')) return 'tasks'
  if (pathname.startsWith('/scan')) return 'scan'
  if (pathname.startsWith('/receive')) return 'scan'
  if (pathname.startsWith('/putaway')) return 'scan'
  if (pathname.startsWith('/pick')) return 'scan'
  if (pathname.startsWith('/order-lookup')) return 'scan'
  if (pathname.startsWith('/stock-inquiry')) return 'scan'
  if (pathname.startsWith('/cycle-count')) return 'scan'
  if (pathname.startsWith('/ship-confirm')) return 'scan'
  if (pathname.startsWith('/replenish')) return 'scan'
  if (pathname.startsWith('/profile')) return 'profile'
  return 'tasks'
}

// ── Header title ──────────────────────────────────────────────────────

function getHeaderTitle(pathname: string): string {
  if (pathname.startsWith('/tasks')) return 'app.headerTasks'
  if (pathname.startsWith('/scan')) return 'app.headerScan'
  if (pathname.startsWith('/receive')) return 'receive.title'
  if (pathname.startsWith('/putaway')) return 'putaway.title'
  if (pathname.startsWith('/pick')) return 'picking.title'
  if (pathname.startsWith('/order-lookup')) return 'orderLookup.title'
  if (pathname.startsWith('/stock-inquiry')) return 'stockInquiry.title'
  if (pathname.startsWith('/cycle-count')) return 'cycleCount.title'
  if (pathname.startsWith('/ship-confirm')) return 'shipConfirm.title'
  if (pathname.startsWith('/replenish')) return 'replenish.title'
  if (pathname.startsWith('/profile')) return 'app.headerProfile'
  return 'app.headerHome'
}

export default function PdaLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { clearTokens } = useAuth()
  const { t } = useTranslation()

  const activeTab = getActiveTab(location.pathname)
  const title = t(getHeaderTitle(location.pathname))

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
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <a
            href="/ai-wms/"
            style={{
              background: 'rgba(255,255,255,0.15)',
              border: 'none',
              color: '#fff',
              padding: '6px 12px',
              borderRadius: 6,
              fontSize: 13,
              cursor: 'pointer',
              textDecoration: 'none',
            }}
          >
            {t('nav.admin')}
          </a>
          <LanguageSwitcher />
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
            {t('common.logout')}
          </button>
        </div>
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
            <span>{t(tab.labelKey)}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
