// SKUs placeholder page.
// P3-03 will implement the full CRUD UI with table, search, and filters.

import { Typography, Empty } from 'antd'
import { BarcodeOutlined } from '@ant-design/icons'

export default function SKUsPage() {
  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>SKUs</Typography.Title>
      </div>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Typography.Text type="secondary">
            SKU management UI will be implemented in P3-03.
          </Typography.Text>
        }
      >
        <BarcodeOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
      </Empty>
    </div>
  )
}
