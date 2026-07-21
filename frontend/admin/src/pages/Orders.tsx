// Order management page with list, filtering, detail view and status transitions.
// Replaces the P3-04 placeholder with real data from GET /api/v1/orders.

import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Typography,
  Table,
  Tag,
  Select,
  Space,
  Drawer,
  Descriptions,
  Button,
  App,
  Empty,
  Card,
  Row,
  Col,
  Statistic,
  Popconfirm,
} from 'antd'
import {
  FileTextOutlined,
  FilterOutlined,
  ReloadOutlined,
  InboxOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  EditOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  OrderSummary,
  Order,
  OrderLine,
  ListResponse,
  UpdateOrderStatusRequest,
} from '@/api/types'

// ── Status / Type / Priority tag colors ────────────────────────────────────

const orderStatusColors: Record<string, string> = {
  draft: 'default',
  confirmed: 'blue',
  processing: 'processing',
  partial: 'warning',
  completed: 'success',
  cancelled: 'error',
}

const orderTypeColors: Record<string, string> = {
  inbound: 'green',
  outbound: 'red',
  transfer: 'purple',
  return: 'orange',
}

const orderPriorityColors: Record<string, string> = {
  low: 'default',
  normal: 'blue',
  high: 'orange',
  urgent: 'red',
}

const orderLineStatusColors: Record<string, string> = {
  pending: 'default',
  allocated: 'blue',
  partial: 'warning',
  fulfilled: 'success',
  cancelled: 'error',
}

// ── Status labels (title case) ─────────────────────────────────────────────

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// ── Valid status transitions ───────────────────────────────────────────────
// Mirrors the backend state machine in domain/order.go

const statusTransitions: Record<string, string[]> = {
  draft: ['confirmed', 'cancelled'],
  confirmed: ['processing', 'cancelled'],
  processing: ['completed', 'partial', 'cancelled'],
  partial: ['completed', 'cancelled'],
}

// ── Constants ──────────────────────────────────────────────────────────────

const PAGE_SIZE = 10

// ── Main component ─────────────────────────────────────────────────────────

export default function OrdersPage() {
  const { message } = App.useApp()

  // ── List state ───────────────────────────────────────────────────────────
  const [orders, setOrders] = useState<OrderSummary[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // ── Filters ──────────────────────────────────────────────────────────────
  const [filterType, setFilterType] = useState<string>('')
  const [filterStatus, setFilterStatus] = useState<string>('')

  // ── Detail drawer ────────────────────────────────────────────────────────
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [drawerLoading, setDrawerLoading] = useState(false)

  // ── Data fetching ────────────────────────────────────────────────────────

  const fetchOrders = useCallback(
    async (p: number, type: string, status: string) => {
      setLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: p,
          page_size: PAGE_SIZE,
        }
        if (type) params.order_type = type
        if (status) params.status = status

        const { data } = await client.get<ListResponse<OrderSummary>>('/orders', {
          params,
        })
        setOrders(data.data)
        setTotal(data.pagination.total)
      } catch {
        message.error('Failed to load orders')
      } finally {
        setLoading(false)
      }
    },
    [message],
  )

  useEffect(() => {
    fetchOrders(page, filterType, filterStatus)
  }, [page, filterType, filterStatus, fetchOrders])

  // ── Refresh ──────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    if (page === 1) {
      fetchOrders(1, filterType, filterStatus)
    } else {
      setPage(1)
    }
  }

  const handleFilterChange = (type: string, status: string) => {
    setFilterType(type)
    setFilterStatus(status)
    setPage(1)
  }

  // ── Table change handler ─────────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Detail drawer ────────────────────────────────────────────────────────

  const openDetail = async (orderId: string) => {
    setDrawerOpen(true)
    setDrawerLoading(true)
    try {
      const { data } = await client.get<Order>(`/orders/${orderId}`)
      setSelectedOrder(data)
    } catch {
      message.error('Failed to load order detail')
    } finally {
      setDrawerLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerOpen(false)
    setSelectedOrder(null)
  }

  // ── Status transition ────────────────────────────────────────────────────

  const handleStatusTransition = async (order: Order, newStatus: string) => {
    try {
      await client.put(`/orders/${order.id}/status`, {
        status: newStatus,
      } satisfies UpdateOrderStatusRequest)
      message.success(`Order status changed to ${titleCase(newStatus)}`)
      // Re-fetch detail
      const { data } = await client.get<Order>(`/orders/${order.id}`)
      setSelectedOrder(data)
      // Refresh list in background
      fetchOrders(page, filterType, filterStatus)
    } catch {
      message.error('Failed to update order status')
    }
  }

  // ── Table columns ────────────────────────────────────────────────────────

  const columns: ColumnsType<OrderSummary> = useMemo(
    () => [
      {
        title: 'Order No',
        dataIndex: 'order_no',
        key: 'order_no',
        width: 180,
        render: (no: string) => (
          <Typography.Text code style={{ cursor: 'pointer' }}>
            {no}
          </Typography.Text>
        ),
      },
      {
        title: 'Type',
        dataIndex: 'order_type',
        key: 'order_type',
        width: 110,
        render: (t: string) => (
          <Tag color={orderTypeColors[t] ?? 'default'}>{titleCase(t)}</Tag>
        ),
      },
      {
        title: 'Status',
        dataIndex: 'status',
        key: 'status',
        width: 130,
        render: (s: string) => {
          const icon =
            s === 'completed' ? (
              <CheckCircleOutlined />
            ) : s === 'processing' ? (
              <SyncOutlined spin />
            ) : s === 'cancelled' ? (
              <CloseCircleOutlined />
            ) : s === 'partial' ? (
              <ExclamationCircleOutlined />
            ) : null
          return (
            <Tag icon={icon} color={orderStatusColors[s] ?? 'default'}>
              {titleCase(s)}
            </Tag>
          )
        },
      },
      {
        title: 'Priority',
        dataIndex: 'priority',
        key: 'priority',
        width: 100,
        responsive: ['md'],
        render: (p: string) => (
          <Tag color={orderPriorityColors[p] ?? 'default'}>{titleCase(p)}</Tag>
        ),
      },
      {
        title: 'External Ref',
        dataIndex: 'external_ref',
        key: 'external_ref',
        width: 150,
        responsive: ['lg'],
        ellipsis: true,
        render: (ref: string) =>
          ref || <Typography.Text type="secondary">—</Typography.Text>,
      },
      {
        title: 'Created',
        dataIndex: 'created_at',
        key: 'created_at',
        width: 170,
        responsive: ['lg'],
        render: (v: string) => new Date(v).toLocaleString(),
      },
      {
        title: 'Actions',
        key: 'actions',
        width: 90,
        render: (_: unknown, record: OrderSummary) => (
          <Button
            type="default"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => openDetail(record.id)}
          >
            View
          </Button>
        ),
      },
    ],
    [],
  )

  // ── Order summary stats ──────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = {}
    orders.forEach((o) => {
      counts[o.status] = (counts[o.status] ?? 0) + 1
    })
    return {
      total: total,
      draft: counts['draft'] ?? 0,
      confirmed: counts['confirmed'] ?? 0,
      processing: counts['processing'] ?? 0,
      partial: counts['partial'] ?? 0,
      completed: counts['completed'] ?? 0,
      cancelled: counts['cancelled'] ?? 0,
    }
  }, [orders, total])

  // ── Line item columns for detail drawer ──────────────────────────────────

  const lineColumns: ColumnsType<OrderLine> = [
    {
      title: '#',
      dataIndex: 'line_no',
      key: 'line_no',
      width: 50,
    },
    {
      title: 'SKU ID',
      dataIndex: 'sku_id',
      key: 'sku_id',
      width: 200,
      ellipsis: true,
      render: (id: string) => (
        <Typography.Text code style={{ fontSize: 12 }}>
          {id}
        </Typography.Text>
      ),
    },
    {
      title: 'Qty',
      dataIndex: 'ordered_qty',
      key: 'ordered_qty',
      width: 90,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: 'Fulfilled',
      dataIndex: 'fulfilled_qty',
      key: 'fulfilled_qty',
      width: 90,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: 'UOM',
      dataIndex: 'uom',
      key: 'uom',
      width: 70,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (s: string) => (
        <Tag color={orderLineStatusColors[s] ?? 'default'}>{titleCase(s)}</Tag>
      ),
    },
    {
      title: 'Batch',
      dataIndex: 'batch_no',
      key: 'batch_no',
      width: 110,
      responsive: ['md'],
      render: (b: string) =>
        b || <Typography.Text type="secondary">—</Typography.Text>,
    },
  ]

  // ── Render ───────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>Orders</Typography.Title>
        <Typography.Text type="secondary">
          Manage inbound, outbound, transfer, and return orders.
        </Typography.Text>
      </div>

      {/* ── Summary stat cards ───────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Total"
              value={stats.total}
              prefix={<FileTextOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Draft"
              value={stats.draft}
              prefix={<EditOutlined />}
              valueStyle={{ color: stats.draft > 0 ? '#8c8c8c' : undefined }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Processing"
              value={stats.processing}
              prefix={<SyncOutlined spin={stats.processing > 0} />}
              valueStyle={{
                color: stats.processing > 0 ? '#1677ff' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Completed"
              value={stats.completed}
              prefix={<CheckCircleOutlined />}
              valueStyle={{
                color: stats.completed > 0 ? '#52c41a' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Partial"
              value={stats.partial}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{
                color: stats.partial > 0 ? '#faad14' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title="Cancelled"
              value={stats.cancelled}
              prefix={<CloseCircleOutlined />}
              valueStyle={{
                color: stats.cancelled > 0 ? '#ff4d4f' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Filter bar ────────────────────────────────────────────────────── */}

      <Card
        title={
          <Space>
            <InboxOutlined />
            <span>All Orders</span>
          </Space>
        }
        extra={
          <Space>
            <FilterOutlined />
            <Select
              placeholder="Order Type"
              allowClear
              style={{ width: 140 }}
              value={filterType || undefined}
              onChange={(v) => handleFilterChange(v ?? '', filterStatus)}
              options={[
                { label: 'Inbound', value: 'inbound' },
                { label: 'Outbound', value: 'outbound' },
                { label: 'Transfer', value: 'transfer' },
                { label: 'Return', value: 'return' },
              ]}
            />
            <Select
              placeholder="Status"
              allowClear
              style={{ width: 140 }}
              value={filterStatus || undefined}
              onChange={(v) => handleFilterChange(filterType, v ?? '')}
              options={[
                { label: 'Draft', value: 'draft' },
                { label: 'Confirmed', value: 'confirmed' },
                { label: 'Processing', value: 'processing' },
                { label: 'Partial', value: 'partial' },
                { label: 'Completed', value: 'completed' },
                { label: 'Cancelled', value: 'cancelled' },
              ]}
            />
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              Refresh
            </Button>
          </Space>
        }
      >
        <Table<OrderSummary>
          columns={columns}
          dataSource={orders}
          rowKey="id"
          loading={loading}
          onChange={handleTableChange}
          onRow={(record) => ({
            onClick: () => openDetail(record.id),
            style: { cursor: 'pointer' },
          })}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total: total,
            showSizeChanger: false,
            showTotal: (t, range) => `${range[0]}-${range[1]} of ${t}`,
          }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No orders found. Try adjusting your filters."
              />
            ),
          }}
        />
      </Card>

      {/* ── Detail Drawer ──────────────────────────────────────────────────── */}

      <Drawer
        title={
          selectedOrder ? (
            <Space>
              <FileTextOutlined />
              <span>Order Detail</span>
              <Typography.Text code>{selectedOrder.order_no}</Typography.Text>
            </Space>
          ) : (
            'Order Detail'
          )
        }
        placement="right"
        width={720}
        open={drawerOpen}
        onClose={closeDrawer}
        destroyOnClose
        extra={
          <Button onClick={closeDrawer}>Close</Button>
        }
      >
        {drawerLoading ? (
          <Empty description="Loading order detail…" />
        ) : selectedOrder ? (
          <>
            {/* ── Order info ──────────────────────────────────────────────────── */}

            <Descriptions
              bordered
              column={{ xs: 1, sm: 2 }}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Descriptions.Item label="Order No">
                <Typography.Text code>{selectedOrder.order_no}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label="Type">
                <Tag color={orderTypeColors[selectedOrder.order_type] ?? 'default'}>
                  {titleCase(selectedOrder.order_type)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={orderStatusColors[selectedOrder.status] ?? 'default'}>
                  {titleCase(selectedOrder.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Priority">
                <Tag color={orderPriorityColors[selectedOrder.priority] ?? 'default'}>
                  {titleCase(selectedOrder.priority)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Warehouse ID">
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedOrder.warehouse_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label="Created By">
                {selectedOrder.created_by}
              </Descriptions.Item>
              {selectedOrder.external_ref && (
                <Descriptions.Item label="External Ref">
                  {selectedOrder.external_ref}
                </Descriptions.Item>
              )}
              {selectedOrder.external_type && (
                <Descriptions.Item label="External Type">
                  <Typography.Text code>{selectedOrder.external_type}</Typography.Text>
                </Descriptions.Item>
              )}
              {selectedOrder.notes && (
                <Descriptions.Item label="Notes" span={2}>
                  {selectedOrder.notes}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="Created">
                {new Date(selectedOrder.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="Updated">
                {new Date(selectedOrder.updated_at).toLocaleString()}
              </Descriptions.Item>
              {selectedOrder.completed_at && (
                <Descriptions.Item label="Completed" span={2}>
                  {new Date(selectedOrder.completed_at).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>

            {/* ── Status transitions ──────────────────────────────────────────── */}

            {statusTransitions[selectedOrder.status] &&
              statusTransitions[selectedOrder.status].length > 0 && (
                <Card
                  size="small"
                  title="Status Actions"
                  style={{ marginBottom: 24 }}
                >
                  <Space wrap>
                    {statusTransitions[selectedOrder.status].map((target) => (
                      <Popconfirm
                        key={target}
                        title={`Move to "${titleCase(target)}"?`}
                        onConfirm={() =>
                          handleStatusTransition(selectedOrder, target)
                        }
                        okText="Yes"
                        cancelText="No"
                      >
                        <Button
                          type={
                            target === 'cancelled' ? 'default' : 'primary'
                          }
                          danger={target === 'cancelled'}
                          size="small"
                          icon={
                            target === 'completed' ? (
                              <CheckCircleOutlined />
                            ) : target === 'cancelled' ? (
                              <CloseCircleOutlined />
                            ) : undefined
                          }
                        >
                          {titleCase(target)}
                        </Button>
                      </Popconfirm>
                    ))}
                  </Space>
                </Card>
              )}

            {/* ── Line items table ────────────────────────────────────────────── */}

            <Card
              size="small"
              title={
                <Space>
                  <InboxOutlined />
                  <span>Line Items ({selectedOrder.lines?.length ?? 0})</span>
                </Space>
              }
            >
              <Table<OrderLine>
                columns={lineColumns}
                dataSource={selectedOrder.lines ?? []}
                rowKey="id"
                size="small"
                pagination={false}
                locale={{
                  emptyText: (
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      description="No line items."
                    />
                  ),
                }}
              />
            </Card>
          </>
        ) : (
          <Empty description="No order selected." />
        )}
      </Drawer>
    </div>
  )
}
