// PDA Login page — mobile-first login form.
// P3-09 will implement the full login form with JWT integration.

import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'

export default function LoginPage() {
  const navigate = useNavigate()
  const { setTokens } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (!username.trim() || !password.trim()) {
      setError('Please enter both username and password')
      return
    }

    setLoading(true)
    try {
      // Placeholder — P3-09 will integrate with the actual auth API.
      // For now, simulate a login to demonstrate the scaffold.
      await new Promise((resolve) => setTimeout(resolve, 800))

      // In the real implementation, this calls POST /api/v1/auth/login
      // and stores the tokens via setTokens(access_token, refresh_token).
      // For the scaffold, we set dummy tokens so navigation works.
      setTokens('demo-access-token', 'demo-refresh-token')
      navigate('/tasks', { replace: true })
    } catch {
      setError('Login failed. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="pda-login-container">
      <div className="pda-login-card">
        <div className="pda-login-title">AI-WMS</div>
        <div className="pda-login-subtitle">Warehouse Operations PDA</div>

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
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your operator ID"
              autoComplete="username"
              autoFocus
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
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter password"
              autoComplete="current-password"
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
            {loading ? 'Signing in...' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  )
}
