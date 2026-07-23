// Root application component with routing for the PDA mobile app.
// Sets up React Router with auth guard, PDA layout shell, and all page routes.
// Integrates i18next for multilingual support (zh-CN default).

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import ProtectedRoute from '@/components/ProtectedRoute'
import PdaLayout from '@/layouts/PdaLayout'
import LoginPage from '@/pages/Login'
import TasksPage from '@/pages/Tasks'
import TaskDetailPage from '@/pages/TaskDetail'
import ScanPage from '@/pages/Scan'
import ReceivingPage from '@/pages/Receiving'
import PutawayPage from '@/pages/Putaway'
import PickingPage from '@/pages/Picking'
import OrderLookupPage from '@/pages/OrderLookup'
import StockInquiryPage from '@/pages/StockInquiry'
import ProfilePage from '@/pages/Profile'
import NotFoundPage from '@/pages/NotFound'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public — login does not require auth */}
        <Route path="/login" element={<LoginPage />} />

        {/* Protected — requires valid JWT, then PDA layout shell with bottom tab bar */}
        <Route element={<ProtectedRoute />}>
          <Route element={<PdaLayout />}>
            <Route path="/tasks" element={<TasksPage />} />
            <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
            <Route path="/scan" element={<ScanPage />} />
            <Route path="/receive" element={<ReceivingPage />} />
            <Route path="/putaway" element={<PutawayPage />} />
            <Route path="/pick" element={<PickingPage />} />
            <Route path="/order-lookup" element={<OrderLookupPage />} />
            <Route path="/stock-inquiry" element={<StockInquiryPage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Route>
        </Route>

        {/* Default redirect */}
        <Route path="/" element={<Navigate to="/tasks" replace />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  )
}
