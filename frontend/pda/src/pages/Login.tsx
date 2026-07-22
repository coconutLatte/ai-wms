// PDA Login page — mobile-first login form with JWT integration and i18n.
// Authenticates warehouse operators via POST /api/v1/auth/login.

import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/hooks/useAuth'
import client from '@/api/client'
import type { LoginRequest, LoginResponse } from '@/api/types'

export default function LoginPage() {
  const navigate = useNavigate()
  const { setTokens } = useAuth()
  const { t } = useTranslation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (!username.trim() || !password.trim()) {
      setError(t('auth.pleaseEnterBoth'))
      return
    }

    setLoading(true)
    try {
      const { data } = await client.post<LoginResponse>('/auth/login', {
        username: username.trim(),
        password,
      } satisfies LoginRequest)

      setTokens(data.access_token, data.refresh_token)
      navigate('/tasks', { replace: true })
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { detail?: string } } }
      const detail = axiosErr?.response?.data?.detail
      setError(detail || t('auth.loginFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="pda-login-container">
      <div className="pda-login-card">
        <div className="pda-login-title">{t('app.title')}</div>
        <div className="pda-login-subtitle">{t('app.subtitle')}</div>

        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 16 }}>
            <label
              htmlFor="username"
              style={{
                display: 'block',
                fontSize: 14,
                fontWeight: 500,
                color: '#595959',
                marginBottom: 6,
              }}
            >
              {t('auth.username')}
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder={t('auth.usernamePlaceholder')}
              autoComplete="username"
              autoFocus
              disabled={loading}
              style={{
                width: '100%',
                padding: '12px 14px',
                fontSize: 16,
                border: '1px solid #d9d9d9',
                borderRadius: 8,
                outline: 'none',
                transition: 'border-color 0.2s',
              }}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = '#1677ff'
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = '#d9d9d9'
              }}
            />
          </div>

          <div style={{ marginBottom: 20 }}>
            <label
              htmlFor="password"
              style={{
                display: 'block',
                fontSize: 14,
                fontWeight: 500,
                color: '#595959',
                marginBottom: 6,
              }}
            >
              {t('auth.password')}
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={t('auth.passwordPlaceholder')}
              autoComplete="current-password"
              disabled={loading}
              style={{
                width: '100%',
                padding: '12px 14px',
                fontSize: 16,
                border: '1px solid #d9d9d9',
                borderRadius: 8,
                outline: 'none',
                transition: 'border-color 0.2s',
              }}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = '#1677ff'
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = '#d9d9d9'
              }}
            />
          </div>

          {error && (
            <div
              style={{
                color: '#cf1322',
                fontSize: 13,
                marginBottom: 16,
                textAlign: 'center',
                padding: '8px 12px',
                background: '#fff1f0',
                borderRadius: 6,
              }}
            >
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            style={{
              width: '100%',
              padding: '14px 0',
              fontSize: 16,
              fontWeight: 600,
              color: '#fff',
              background: loading ? '#91caff' : '#1677ff',
              border: 'none',
              borderRadius: 8,
              cursor: loading ? 'not-allowed' : 'pointer',
              transition: 'background 0.2s',
            }}
          >
            {loading ? t('auth.signingIn') : t('auth.signIn')}
          </button>
        </form>
      </div>
    </div>
  )
}
