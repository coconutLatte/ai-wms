// Login placeholder page.
// P3-01 will implement the full login form with JWT.

import { Card, Typography } from 'antd'
import { ShopOutlined } from '@ant-design/icons'

export default function LoginPage() {
  return (
    <div className="login-container">
      <Card className="login-card" title={<><ShopOutlined /> AI-WMS Admin</>}>
        <Typography.Paragraph style={{ textAlign: 'center', color: '#8c8c8c' }}>
          Login form will be implemented in P3-01.
        </Typography.Paragraph>
      </Card>
    </div>
  )
}
