// Auth hook — convenience wrapper around the auth store.
// Returns auth state and actions for components.

import { useAuthStore } from '@/store/auth'

export function useAuth() {
  const accessToken = useAuthStore((s) => s.accessToken)
  const refreshToken = useAuthStore((s) => s.refreshToken)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const setTokens = useAuthStore((s) => s.setTokens)
  const clearTokens = useAuthStore((s) => s.clearTokens)

  return { accessToken, refreshToken, isAuthenticated, setTokens, clearTokens }
}
