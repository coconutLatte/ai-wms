// Dashboard page — admin landing page with real-time warehouse metrics.
// Fetches aggregated data from GET /api/v1/dashboard and displays summary stat cards.

import { useState, useEffect } from 'react'
import { Row, Col, Card, Statistic, Typography, Spin, Alert } from 'antd'
import {
  ShopOutlined,
  BarcodeOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  ToolOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getDashboard } from '@/api/dashboard'
import type { AdminDashboardResponse } from '@/api/types'

export default function DashboardPage() {
  const { t } = useTranslation()
  const [data, setData] = useState<AdminDashboardResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    getDashboard()
      .then((result) => {
        if (!cancelled) {
          setData(result)
          setLoading(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err?.response?.data?.detail || err?.message || 'Failed to load dashboard')
          setLoading(false)
        }
      })

    return () => { cancelled = true }
  }, [])

  // Compute active (non-terminal) orders from the order summary.
  const activeOrders = data?.order_summary
    ? Object.entries(data.order_summary)
        .filter(([status]) => status !== 'completed' && status !== 'cancelled')
        .reduce((sum, [, count]) => sum + count, 0)
    : null

  // Compute pending tasks (not completed, not cancelled).
  const pendingTasks = data?.task_summary
    ? Object.entries(data.task_summary)
        .filter(([status]) => status !== 'completed' && status !== 'cancelled')
        .reduce((sum, [, count]) => sum + count, 0)
    : null

  // Compute completed tasks count.
  const completedTasks = data?.task_summary?.['completed'] ?? null
  const completedOrders = data?.order_summary?.['completed'] ?? null

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('dashboard.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('dashboard.welcome')}
        </Typography.Text>
      </div>

      {loading && (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      )}

      {error && (
        <Alert
          type="error"
          message={t('common.loadFailed', 'Load failed')}
          description={error}
          showIcon
          closable
          style={{ marginBottom: 16 }}
        />
      )}

      {data && !loading && (
        <>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title={t('dashboard.warehouses')}
                  value={data.warehouse_count}
                  prefix={<ShopOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title={t('dashboard.skus')}
                  value={data.sku_count}
                  prefix={<BarcodeOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title={t('dashboard.inventoryRecords')}
                  value={data.inventory_stats?.total_records ?? 0}
                  prefix={<DatabaseOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card>
                <Statistic
                  title={t('dashboard.activeOrders')}
                  value={activeOrders ?? 0}
                  prefix={<FileTextOutlined />}
                />
              </Card>
            </Col>
          </Row>

          {/* Second row: Order & Task status breakdowns */}
          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            {data.order_summary && (
              <Col xs={24} sm={12} lg={6}>
                <Card title={t('dashboard.ordersLabel', 'Orders')} size="small">
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                    {Object.entries(data.order_summary).map(([status, count]) => (
                      <Statistic
                        key={status}
                        title={status.charAt(0).toUpperCase() + status.slice(1)}
                        value={count}
                        valueStyle={{ fontSize: 18 }}
                      />
                    ))}
                  </div>
                </Card>
              </Col>
            )}

            {data.task_summary && (
              <Col xs={24} sm={12} lg={6}>
                <Card title={t('dashboard.tasksLabel', 'Tasks')} size="small">
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                    {Object.entries(data.task_summary).map(([status, count]) => (
                      <Statistic
                        key={status}
                        title={status.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())}
                        value={count}
                        valueStyle={{ fontSize: 18 }}
                      />
                    ))}
                  </div>
                </Card>
              </Col>
            )}

            <Col xs={24} sm={12} lg={6}>
              <Card size="small">
                <Statistic
                  title={t('dashboard.completedOrders', 'Completed Orders')}
                  value={completedOrders ?? 0}
                  prefix={<CheckCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card size="small">
                <Statistic
                  title={t('dashboard.pendingTasks', 'Pending Tasks')}
                  value={pendingTasks ?? 0}
                  prefix={<ToolOutlined />}
                />
              </Card>
            </Col>
          </Row>
        </>
      )}
    </div>
  )
}
