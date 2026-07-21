// Auth guard component — wraps protected routes.
// Redirects to /login if the user is not authenticated.
// Renders child routes via <Outlet /> if authenticated.

import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'

export default function ProtectedRoute() {
  const { isAuthenticated } = useAuth()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
