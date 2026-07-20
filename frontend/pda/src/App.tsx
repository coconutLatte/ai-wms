// Root application component with routing for the PDA mobile app.
// Sets up React Router with the PDA layout shell and all page routes.

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import PdaLayout from '@/layouts/PdaLayout'
import LoginPage from '@/pages/Login'
import TasksPage from '@/pages/Tasks'
import TaskDetailPage from '@/pages/TaskDetail'
import ScanPage from '@/pages/Scan'
import ProfilePage from '@/pages/Profile'
import NotFoundPage from '@/pages/NotFound'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public */}
        <Route path="/login" element={<LoginPage />} />

        {/* Protected — PDA layout shell with bottom tab bar */}
        <Route element={<PdaLayout />}>
          <Route path="/tasks" element={<TasksPage />} />
          <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
          <Route path="/scan" element={<ScanPage />} />
          <Route path="/profile" element={<ProfilePage />} />
        </Route>

        {/* Default redirect */}
        <Route path="/" element={<Navigate to="/tasks" replace />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  )
}
