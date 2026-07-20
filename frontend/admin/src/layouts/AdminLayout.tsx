// Main admin layout: collapsible sidebar + header + content area.
// Used as the shell for all authenticated admin pages.

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
} from '@ant-design/icons'
import { useAuth } from '@/hooks/useAuth'

const { Header, Sider, Content } = Layout

// ── Sidebar menu items ─────────────────────────────────────────────

interface MenuItem {
  key: string
  icon: React.ReactNode
  label: string
  path: string
}

const menuItems: MenuItem[] = [
  { key: 'dashboard', icon: <DashboardOutlined />, label: 'Dashboard', path: '/dashboard' },
  { key: 'warehouses', icon: <ShopOutlined />, label: 'Warehouses', path: '/warehouses' },
  { key: 'skus', icon: <BarcodeOutlined />, label: 'SKUs', path: '/skus' },
  { key: 'inventory', icon: <DatabaseOutlined />, label: 'Inventory', path: '/inventory' },
  { key: 'orders', icon: <FileTextOutlined />, label: 'Orders', path: '/orders' },
  { key: 'tasks', icon: <CarryOutOutlined />, label: 'Tasks', path: '/tasks' },
]

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { clearTokens } = useAuth()
  const { token: themeToken } = theme.useToken()

  // Determine selected menu item from current path
  const selectedKey = menuItems.find(
    (item) => location.pathname.startsWith(item.path),
  )?.key ?? 'dashboard'

  const handleMenuClick = (info: { key: string }) => {
    const item = menuItems.find((m) => m.key === info.key)
    if (item) navigate(item.path)
  }

  const handleLogout = () => {
    clearTokens()
    navigate('/login')
  }

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: 'Profile' },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: 'Logout', danger: true },
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
          items={menuItems.map((item) => ({
            key: item.key,
            icon: item.icon,
            label: item.label,
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

          <Dropdown
            menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
            placement="bottomRight"
          >
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: themeToken.colorPrimary }} />
              <Typography.Text>Admin</Typography.Text>
            </div>
          </Dropdown>
        </Header>

        <Content>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
