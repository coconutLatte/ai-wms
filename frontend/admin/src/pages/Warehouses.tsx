// Warehouses placeholder page.
// P2-08 will implement the full CRUD UI.

import { Typography, Empty } from 'antd'
import { ShopOutlined } from '@ant-design/icons'

export default function WarehousesPage() {
  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Warehouses</Typography.Title>
      </div>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Typography.Text type="secondary">
            Warehouse management UI will be implemented in P2-08.
          </Typography.Text>
        }
      >
        <ShopOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
      </Empty>
    </div>
  )
}
