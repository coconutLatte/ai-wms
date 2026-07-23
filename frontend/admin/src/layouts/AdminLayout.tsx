// Main admin layout: collapsible sidebar + header + content area.
// Used as the shell for all authenticated admin pages.
// Includes language switcher and user menu in the header.

import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Button, Typography, Dropdown, Avatar, theme } from 'antd'
import {
  DashboardOutlined,
  ShopOutlined,
  BarcodeOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  CarryOutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  TruckOutlined,
  AppstoreOutlined,
  EnvironmentOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/hooks/useAuth'
import LanguageSwitcher from '@/components/LanguageSwitcher'

const { Header, Sider, Content } = Layout

// ── Sidebar menu items (keys used for i18n lookup) ──────────────────

interface MenuItem {
  key: string
  icon: React.ReactNode
  labelKey: string
  path: string
}

const menuKeys: MenuItem[] = [
  { key: 'dashboard', icon: <DashboardOutlined />, labelKey: 'nav.dashboard', path: '/dashboard' },
  { key: 'warehouses', icon: <ShopOutlined />, labelKey: 'nav.warehouses', path: '/warehouses' },
  { key: 'skus', icon: <BarcodeOutlined />, labelKey: 'nav.skus', path: '/skus' },
  { key: 'inventory', icon: <DatabaseOutlined />, labelKey: 'nav.inventory', path: '/inventory' },
  { key: 'orders', icon: <FileTextOutlined />, labelKey: 'nav.orders', path: '/orders' },
  { key: 'tasks', icon: <CarryOutOutlined />, labelKey: 'nav.tasks', path: '/tasks' },
  { key: 'users', icon: <TeamOutlined />, labelKey: 'nav.users', path: '/users' },
  { key: 'roles', icon: <SafetyOutlined />, labelKey: 'nav.roles', path: '/roles' },
  { key: 'asns', icon: <TruckOutlined />, labelKey: 'nav.asns', path: '/asns' },
  { key: 'zones', icon: <AppstoreOutlined />, labelKey: 'nav.zones', path: '/zones' },
  { key: 'locations', icon: <EnvironmentOutlined />, labelKey: 'nav.locations', path: '/locations' },
]

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { clearTokens } = useAuth()
  const { token: themeToken } = theme.useToken()
  const { t } = useTranslation()

  // Determine selected menu item from current path
  const selectedKey = menuKeys.find(
    (item) => location.pathname.startsWith(item.path),
  )?.key ?? 'dashboard'

  const handleMenuClick = (info: { key: string }) => {
    const item = menuKeys.find((m) => m.key === info.key)
    if (item) navigate(item.path)
  }

  const handleLogout = () => {
    clearTokens()
    navigate('/login')
  }

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: t('nav.profile') },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: t('nav.logout'), danger: true },
  ]

  const handleUserMenuClick = (info: { key: string }) => {
    if (info.key === 'logout') handleLogout()
  }

  return (
    <Layout className="admin-layout">
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        breakpoint="lg"
        onBreakpoint={(broken) => setCollapsed(broken)}
        style={{ background: themeToken.colorBgContainer }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
          }}
        >
          <Typography.Title
            level={4}
            style={{ margin: 0, color: themeToken.colorPrimary }}
          >
            {collapsed ? 'WMS' : 'AI-WMS'}
          </Typography.Title>
        </div>

        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          onClick={handleMenuClick}
          items={menuKeys.map((item) => ({
            key: item.key,
            icon: item.icon,
            label: t(item.labelKey),
          }))}
          style={{ border: 'none' }}
        />
      </Sider>

      <Layout>
        <Header style={{ background: themeToken.colorBgContainer }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />

          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <LanguageSwitcher />

            <Button
              type="link"
              href="/ai-wms/pda/"
              target="_self"
              style={{ padding: '4px 8px' }}
            >
              {t('nav.pda')}
            </Button>

            <Dropdown
              menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
              placement="bottomRight"
            >
              <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                <Avatar icon={<UserOutlined />} style={{ backgroundColor: themeToken.colorPrimary }} />
                <Typography.Text>Admin</Typography.Text>
              </div>
            </Dropdown>
          </div>
        </Header>

        <Content>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
