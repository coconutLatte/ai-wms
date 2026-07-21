// Login page — username/password authentication form.
// Posts credentials to the backend, stores JWT tokens on success,
// and redirects to the dashboard.

import { useState } from 'react'
import { useNavigate, useLocation, Navigate } from 'react-router-dom'
import { Card, Form, Input, Button, Typography, Alert, App } from 'antd'
import { ShopOutlined, UserOutlined, LockOutlined } from '@ant-design/icons'
import { useAuth } from '@/hooks/useAuth'
import { login } from '@/api/auth'
import type { LoginRequest } from '@/api/types'

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()
  const location = useLocation()
  const { setTokens, isAuthenticated } = useAuth()
  const { message } = App.useApp()

  // If already authenticated, redirect away from login page.
  // Use the `from` query param if present (set by ProtectedRoute), otherwise dashboard.
  if (isAuthenticated) {
    const from = (location.state as { from?: string })?.from ?? '/dashboard'
    return <Navigate to={from} replace />
  }

  const handleSubmit = async (values: LoginRequest) => {
    setLoading(true)
    setError(null)

    try {
      const result = await login(values)
      setTokens(result.access_token, result.refresh_token)
      message.success('Welcome back!')

      // Navigate to the page the user was trying to reach, or dashboard.
      const from = (location.state as { from?: string })?.from ?? '/dashboard'
      navigate(from, { replace: true })
    } catch (err: any) {
      const detail =
        err?.response?.data?.detail ??
        err?.response?.data?.message ??
        'Login failed. Please check your credentials.'
      setError(detail)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-container">
      <Card
        className="login-card"
        title={
          <div style={{ textAlign: 'center' }}>
            <ShopOutlined style={{ fontSize: 24, color: '#1677ff', marginRight: 8 }} />
            <Typography.Text strong style={{ fontSize: 18 }}>
              AI-WMS Admin
            </Typography.Text>
          </div>
        }
      >
        {error && (
          <Alert
            message={error}
            type="error"
            showIcon
            closable
            onClose={() => setError(null)}
            style={{ marginBottom: 16 }}
          />
        )}

        <Form<LoginRequest>
          name="login"
          onFinish={handleSubmit}
          layout="vertical"
          size="large"
          autoComplete="off"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: 'Please enter your username' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="Username" autoFocus />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Please enter your password' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Password" />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" loading={loading} block>
              Sign In
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
