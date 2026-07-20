// Axios HTTP client with JWT auth interceptors.
// - Attaches Authorization header from the auth store.
// - Auto-refreshes on 401 and retries the original request.
// - Shared by all API modules.

import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/store/auth'
import type { RefreshRequest, LoginResponse } from './types'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15_000,
  headers: { 'Content-Type': 'application/json' },
})

// ── Request interceptor: attach access token ───────────────────────────────

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().accessToken
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ── Response interceptor: auto-refresh on 401 ──────────────────────────────

let isRefreshing = false
let pendingRequests: Array<{
  resolve: (token: string) => void
  reject: (err: unknown) => void
}> = []

function enqueuePending(resolve: (token: string) => void, reject: (err: unknown) => void) {
  pendingRequests.push({ resolve, reject })
}

function flushPending(token: string | null, error: unknown | null) {
  pendingRequests.forEach(({ resolve, reject }) => {
    if (token) resolve(token)
    else reject(error)
  })
  pendingRequests = []
}

client.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // Only attempt refresh for 401 on non-auth endpoints
    if (error.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(error)
    }

    const { refreshToken, setTokens, clearTokens } = useAuthStore.getState()

    if (!refreshToken) {
      clearTokens()
      return Promise.reject(error)
    }

    if (isRefreshing) {
      // Another request is already refreshing — queue this one
      return new Promise((resolve, reject) => {
        enqueuePending((token: string) => {
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${token}`
          }
          resolve(client(originalRequest))
        }, reject)
      })
    }

    isRefreshing = true
    originalRequest._retry = true

    try {
      const { data } = await axios.post<LoginResponse>('/api/v1/auth/refresh', {
        refresh_token: refreshToken,
      } satisfies RefreshRequest)

      setTokens(data.access_token, data.refresh_token)
      flushPending(data.access_token, null)

      if (originalRequest.headers) {
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`
      }
      return client(originalRequest)
    } catch (refreshError) {
      flushPending(null, refreshError)
      clearTokens()
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  },
)

export default client
