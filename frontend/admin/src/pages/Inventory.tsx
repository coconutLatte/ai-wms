// Inventory dashboard page with summary cards, low stock alerts, and warehouse breakdown.
// Replaces the P3-02 placeholder with real data from GET /api/v1/inventory/dashboard.

import { useCallback, useEffect, useState } from 'react'
import {
  Typography,
  Row,
  Col,
  Card,
  Statistic,
  Table,
  Tag,
  Select,
  InputNumber,
  Space,
  App,
  Empty,
} from 'antd'
import {
  DatabaseOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  StopOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  ShopOutlined,
  FallOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  DashboardResponse,
  Inventory,
  WarehouseBreakdown,
} from '@/api/types'

// ── Status tag colors ────────────────────────────────────────────────────

const inventoryStatusColors: Record<string, string> = {
  available: 'green',
  quarantine: 'orange',
  damaged: 'red',
  expired: 'default',
}

// ── Format number helper ─────────────────────────────────────────────────

function fmt(n: number, decimals = 0): string {
  return n.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  })
}

// ── Main component ───────────────────────────────────────────────────────

export default function InventoryPage() {
  const { message } = App.useApp()

  const [loading, setLoading] = useState(false)
  const [dashboard, setDashboard] = useState<DashboardResponse | null>(null)

  // ── Filters ───────────────────────────────────────────────────────────
  const [warehouseId, setWarehouseId] = useState<string>('')
  const [lowStockThreshold, setLowStockThreshold] = useState(10)

  // ── Data fetching ─────────────────────────────────────────────────────

  const fetchDashboard = useCallback(async (whId: string, threshold: number) => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = {}
      if (whId) params.warehouse_id = whId
      params.low_stock_threshold = threshold

      const { data } = await client.get<DashboardResponse>('/inventory/dashboard', { params })
      setDashboard(data)
    } catch {
      message.error('Failed to load inventory dashboard')
    } finally {
      setLoading(false)
    }
  }, [message])

  useEffect(() => {
    fetchDashboard(warehouseId, lowStockThreshold)
  }, [warehouseId, lowStockThreshold, fetchDashboard])

  // ── Low stock columns ─────────────────────────────────────────────────

  const lowStockColumns: ColumnsType<Inventory> = [
    {
      title: 'Batch',
      dataIndex: 'batch_no',
      key: 'batch_no',
      width: 130,
      render: (b: string) =>
        b ? <Typography.Text code>{b}</Typography.Text> : <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: 'SKU',
      dataIndex: 'sku_id',
      key: 'sku_id',
      width: 200,
      ellipsis: true,
      render: (id: string) => <Typography.Text code style={{ fontSize: 12 }}>{id}</Typography.Text>,
    },
    {
      title: 'On Hand',
      dataIndex: 'qty',
      key: 'qty',
      width: 100,
      align: 'right',
      render: (v: number) => fmt(v),
    },
    {
      title: 'Reserved',
      dataIndex: 'reserved_qty',
      key: 'reserved_qty',
      width: 100,
      align: 'right',
      responsive: ['md'],
      render: (v: number) => fmt(v),
    },
    {
      title: 'Available',
      dataIndex: 'available_qty',
      key: 'available_qty',
      width: 100,
      align: 'right',
      render: (v: number) => (
        <Typography.Text strong type="warning">
          {fmt(v)}
        </Typography.Text>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status: string) => (
        <Tag color={inventoryStatusColors[status] ?? 'default'}>{status}</Tag>
      ),
    },
  ]

  // ── Warehouse breakdown columns ───────────────────────────────────────

  const warehouseColumns: ColumnsType<WarehouseBreakdown> = [
    {
      title: 'Warehouse',
      key: 'warehouse',
      render: (_: unknown, record: WarehouseBreakdown) => (
        <Space>
          <Typography.Text strong>{record.warehouse_name}</Typography.Text>
          <Typography.Text type="secondary" code>
            {record.warehouse_code}
          </Typography.Text>
        </Space>
      ),
    },
    {
      title: 'Records',
      dataIndex: 'record_count',
      key: 'record_count',
      width: 90,
      align: 'right',
      responsive: ['md'],
      render: (v: number) => fmt(v),
    },
    {
      title: 'On Hand',
      dataIndex: 'total_qty',
      key: 'total_qty',
      width: 110,
      align: 'right',
      render: (v: number) => fmt(v),
    },
    {
      title: 'Reserved',
      dataIndex: 'reserved_qty',
      key: 'reserved_qty',
      width: 110,
      align: 'right',
      responsive: ['md'],
      render: (v: number) => fmt(v),
    },
    {
      title: 'Available',
      dataIndex: 'available_qty',
      key: 'available_qty',
      width: 110,
      align: 'right',
      render: (v: number) => (
        <Typography.Text strong style={{ color: v > 0 ? '#52c41a' : '#ff4d4f' }}>
          {fmt(v)}
        </Typography.Text>
      ),
    },
  ]

  // ── Render ────────────────────────────────────────────────────────────

  const stats = dashboard?.stats

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Inventory Dashboard</Typography.Title>
        <Typography.Text type="secondary">Real-time inventory overview and insights.</Typography.Text>
      </div>

      {/* ── Filters bar ────────────────────────────────────────────────── */}

      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Typography.Text strong>Warehouse:</Typography.Text>
          <Select
            placeholder="All warehouses"
            allowClear
            style={{ width: 200 }}
            value={warehouseId || undefined}
            onChange={(v) => setWarehouseId(v ?? '')}
            options={[
              { label: 'All Warehouses', value: '' },
            ]}
          />
          <Typography.Text strong>Low stock threshold:</Typography.Text>
          <InputNumber
            min={1}
            max={1000}
            value={lowStockThreshold}
            onChange={(v) => setLowStockThreshold(v ?? 10)}
            addonAfter="units"
            style={{ width: 150 }}
          />
        </Space>
      </Card>

      {/* ── Summary stat cards ─────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={8} lg={4}>
          <Card loading={loading}>
            <Statistic
              title="Total Records"
              value={stats?.total_records ?? 0}
              prefix={<DatabaseOutlined />}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={4}>
          <Card loading={loading}>
            <Statistic
              title="On Hand"
              value={stats?.total_qty ?? 0}
              prefix={<ShopOutlined />}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={4}>
          <Card loading={loading}>
            <Statistic
              title="Reserved"
              value={stats?.total_reserved_qty ?? 0}
              prefix={<ClockCircleOutlined />}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={4}>
          <Card loading={loading}>
            <Statistic
              title="Available"
              value={stats?.total_available_qty ?? 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: (stats?.total_available_qty ?? 0) > 0 ? '#3f8600' : '#cf1322' }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={4}>
          <Card loading={loading}>
            <Statistic
              title="Available"
              value={stats?.available_count ?? 0}
              prefix={<CheckCircleOutlined />}
              suffix={<Typography.Text type="secondary">records</Typography.Text>}
              valueStyle={{ color: '#3f8600' }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={4}>
          <Card loading={loading}>
            <Statistic
              title="Low Stock"
              value={stats?.low_stock_count ?? 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: (stats?.low_stock_count ?? 0) > 0 ? '#faad14' : '#3f8600' }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Status breakdown ───────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Quarantine"
              value={stats?.quarantine_count ?? 0}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: (stats?.quarantine_count ?? 0) > 0 ? '#faad14' : undefined }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Damaged"
              value={stats?.damaged_count ?? 0}
              prefix={<StopOutlined />}
              valueStyle={{ color: (stats?.damaged_count ?? 0) > 0 ? '#ff4d4f' : undefined }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Expired"
              value={stats?.expired_count ?? 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: (stats?.expired_count ?? 0) > 0 ? '#8c8c8c' : undefined }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Low Stock"
              value={stats?.low_stock_count ?? 0}
              prefix={<FallOutlined />}
              valueStyle={{ color: (stats?.low_stock_count ?? 0) > 0 ? '#faad14' : undefined }}
              formatter={(v) => fmt(Number(v))}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Low stock table ────────────────────────────────────────────── */}

      <Card
        title={
          <Space>
            <WarningOutlined style={{ color: '#faad14' }} />
            <span>Low Stock Items (available ≤ {lowStockThreshold})</span>
          </Space>
        }
        style={{ marginBottom: 24 }}
        loading={loading}
      >
        <Table<Inventory>
          columns={lowStockColumns}
          dataSource={dashboard?.low_stock ?? []}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 10, showSizeChanger: false }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No low stock items. All inventory levels are healthy."
              />
            ),
          }}
        />
      </Card>

      {/* ── Warehouse breakdown ─────────────────────────────────────────── */}

      <Card
        title={
          <Space>
            <ShopOutlined />
            <span>Inventory by Warehouse</span>
          </Space>
        }
        loading={loading}
      >
        <Table<WarehouseBreakdown>
          columns={warehouseColumns}
          dataSource={dashboard?.by_warehouse ?? []}
          rowKey="warehouse_id"
          size="small"
          pagination={{ pageSize: 10, showSizeChanger: false }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No warehouse data available."
              />
            ),
          }}
          summary={() => {
            if (!dashboard?.by_warehouse?.length) return null
            const totalRecords = dashboard.by_warehouse.reduce((s, w) => s + w.record_count, 0)
            const totalQty = dashboard.by_warehouse.reduce((s, w) => s + w.total_qty, 0)
            const totalReserved = dashboard.by_warehouse.reduce((s, w) => s + w.reserved_qty, 0)
            const totalAvailable = dashboard.by_warehouse.reduce((s, w) => s + w.available_qty, 0)
            return (
              <Table.Summary.Row>
                <Table.Summary.Cell index={0}>
                  <Typography.Text strong>Total</Typography.Text>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={1} align="right">
                  <Typography.Text strong>{fmt(totalRecords)}</Typography.Text>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={2} align="right">
                  <Typography.Text strong>{fmt(totalQty)}</Typography.Text>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={3} align="right">
                  <Typography.Text strong>{fmt(totalReserved)}</Typography.Text>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={4} align="right">
                  <Typography.Text strong>{fmt(totalAvailable)}</Typography.Text>
                </Table.Summary.Cell>
              </Table.Summary.Row>
            )
          }}
        />
      </Card>
    </div>
  )
}
