// Dashboard placeholder page with summary stat cards.
// Shows warehouse, SKU, inventory, and order counts.

import { Row, Col, Card, Statistic, Typography } from 'antd'
import {
  ShopOutlined,
  BarcodeOutlined,
  DatabaseOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

export default function DashboardPage() {
  const { t } = useTranslation()

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('dashboard.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('dashboard.welcome')}
        </Typography.Text>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title={t('dashboard.warehouses')} value="—" prefix={<ShopOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title={t('dashboard.skus')} value="—" prefix={<BarcodeOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title={t('dashboard.inventoryRecords')} value="—" prefix={<DatabaseOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title={t('dashboard.activeOrders')} value="—" prefix={<FileTextOutlined />} />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
