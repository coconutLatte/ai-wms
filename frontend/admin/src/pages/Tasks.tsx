// Tasks placeholder page.
// P3-05 will implement the full task monitoring UI.

import { Typography, Empty } from 'antd'
import { CarryOutOutlined } from '@ant-design/icons'

export default function TasksPage() {
  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Tasks</Typography.Title>
      </div>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Typography.Text type="secondary">
            Task monitoring UI will be implemented in P3-05.
          </Typography.Text>
        }
      >
        <CarryOutOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
      </Empty>
    </div>
  )
}
