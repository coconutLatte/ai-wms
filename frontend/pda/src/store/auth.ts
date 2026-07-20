// Auth state managed with Zustand.
// Persisted to localStorage so the operator stays logged in across page reloads.

import { create } from 'zustand'

const STORAGE_KEY = 'wms_pda_auth'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean

  setTokens: (access: string, refresh: string) => void
  clearTokens: () => void
}

function loadFromStorage(): Pick<AuthState, 'accessToken' | 'refreshToken'> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      return {
        accessToken: parsed.accessToken ?? null,
        refreshToken: parsed.refreshToken ?? null,
      }
    }
  } catch {
    // Corrupt storage — ignore
  }
  return { accessToken: null, refreshToken: null }
}

function saveToStorage(accessToken: string | null, refreshToken: string | null) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ accessToken, refreshToken }))
  } catch {
    // Storage full or unavailable — ignore
  }
}

export const useAuthStore = create<AuthState>((set) => {
  const stored = loadFromStorage()

  return {
    accessToken: stored.accessToken,
    refreshToken: stored.refreshToken,
    isAuthenticated: !!stored.accessToken,

    setTokens: (access: string, refresh: string) => {
      saveToStorage(access, refresh)
      set({ accessToken: access, refreshToken: refresh, isAuthenticated: true })
    },

    clearTokens: () => {
      saveToStorage(null, null)
      set({ accessToken: null, refreshToken: null, isAuthenticated: false })
    },
  }
})
