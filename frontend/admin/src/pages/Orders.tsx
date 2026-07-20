// Orders placeholder page.
// P3-04 will implement the full order management UI with status transitions.

import { Typography, Empty } from 'antd'
import { FileTextOutlined } from '@ant-design/icons'

export default function OrdersPage() {
  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Orders</Typography.Title>
      </div>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Typography.Text type="secondary">
            Order management UI will be implemented in P3-04.
          </Typography.Text>
        }
      >
        <FileTextOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
      </Empty>
    </div>
  )
}
