// Dashboard placeholder page with summary stat cards.
// P3-02 will implement real data.

import { Row, Col, Card, Statistic, Typography } from 'antd'
import {
  ShopOutlined,
  BarcodeOutlined,
  DatabaseOutlined,
  FileTextOutlined,
} from '@ant-design/icons'

export default function DashboardPage() {
  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Dashboard</Typography.Title>
        <Typography.Text type="secondary">
          Welcome to AI-WMS. Real-time warehouse metrics will appear here.
        </Typography.Text>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Warehouses" value="—" prefix={<ShopOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="SKUs" value="—" prefix={<BarcodeOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Inventory Records" value="—" prefix={<DatabaseOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="Active Orders" value="—" prefix={<FileTextOutlined />} />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
