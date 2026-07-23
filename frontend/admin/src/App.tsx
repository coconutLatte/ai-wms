// Root application component with routing.
// Sets up React Router with the admin layout shell and all page routes.
// Integrates i18next for multilingual support (zh-CN default) and Ant Design locale.

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider, App as AntApp } from 'antd'
import { useTranslation } from 'react-i18next'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import AdminLayout from '@/layouts/AdminLayout'
import ProtectedRoute from '@/components/ProtectedRoute'
import LoginPage from '@/pages/Login'
import DashboardPage from '@/pages/Dashboard'
import WarehousesPage from '@/pages/Warehouses'
import SKUsPage from '@/pages/Skus'
import InventoryPage from '@/pages/Inventory'
import OrdersPage from '@/pages/Orders'
import TasksPage from '@/pages/Tasks'
import UsersPage from '@/pages/Users'
import RolesPage from '@/pages/Roles'
import AsnsPage from '@/pages/Asns'
import ZonesPage from '@/pages/Zones'
import LocationsPage from '@/pages/Locations'
import WavesPage from '@/pages/Waves'
import InventoryTransactionsPage from '@/pages/InventoryTransactions'
import AuditLogsPage from '@/pages/AuditLogs'
import ShipmentsPage from '@/pages/Shipments'
import NotFoundPage from '@/pages/NotFound'

// ── Ant Design theme ────────────────────────────────────────────────

const themeConfig = {
  token: {
    colorPrimary: '#1677ff',
    borderRadius: 6,
    fontFamily: `-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif`,
  },
}

function AppRoutes() {
  const { i18n } = useTranslation()

  const antdLocale = i18n.language === 'en' ? enUS : zhCN

  return (
    <ConfigProvider theme={themeConfig} locale={antdLocale}>
      <AntApp>
        <BrowserRouter basename="/ai-wms">
          <Routes>
            {/* Public */}
            <Route path="/login" element={<LoginPage />} />

            {/* Protected — auth guard + admin layout shell */}
            <Route element={<ProtectedRoute />}>
              <Route element={<AdminLayout />}>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/warehouses" element={<WarehousesPage />} />
                <Route path="/skus" element={<SKUsPage />} />
                <Route path="/inventory" element={<InventoryPage />} />
                <Route path="/orders" element={<OrdersPage />} />
                <Route path="/tasks" element={<TasksPage />} />
                <Route path="/users" element={<UsersPage />} />
                <Route path="/roles" element={<RolesPage />} />
                <Route path="/asns" element={<AsnsPage />} />
                <Route path="/zones" element={<ZonesPage />} />
                <Route path="/locations" element={<LocationsPage />} />
                <Route path="/waves" element={<WavesPage />} />
                <Route path="/inventory-transactions" element={<InventoryTransactionsPage />} />
                <Route path="/audit-logs" element={<AuditLogsPage />} />
                <Route path="/shipments" element={<ShipmentsPage />} />
              </Route>
            </Route>

            {/* Default redirect */}
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  )
}

export default function App() {
  return <AppRoutes />
}
