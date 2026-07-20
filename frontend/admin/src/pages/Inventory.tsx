// Inventory placeholder page.
// P3-02 will implement the full inventory dashboard.

import { Typography, Empty } from 'antd'
import { DatabaseOutlined } from '@ant-design/icons'

export default function InventoryPage() {
  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Inventory</Typography.Title>
      </div>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Typography.Text type="secondary">
            Inventory dashboard will be implemented in P3-02.
          </Typography.Text>
        }
      >
        <DatabaseOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
      </Empty>
    </div>
  )
}
