// Order management page with list, filtering, detail view and status transitions.
// Uses GET /api/v1/orders for real data.
// All UI text is translated via react-i18next.

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
  Modal,
  Form,
  Input,
  InputNumber,
  Divider,
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
  PlusOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  OrderSummary,
  Order,
  OrderLine,
  ListResponse,
  UpdateOrderStatusRequest,
  CreateOrderRequest,
  CreateOrderLineRequest,
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
  const { t } = useTranslation()

  // ── Localized labels ─────────────────────────────────────────────────────

  const typeLabels: Record<string, string> = useMemo(() => ({
    inbound: t('order.inbound'),
    outbound: t('order.outbound'),
    transfer: t('order.transfer'),
    return: t('order.return'),
  }), [t])

  const statusLabels: Record<string, string> = useMemo(() => ({
    draft: t('order.draft'),
    confirmed: t('order.confirmed'),
    processing: t('order.processing'),
    partial: t('order.partial'),
    completed: t('order.completed'),
    cancelled: t('order.cancelled'),
  }), [t])

  const priorityLabels: Record<string, string> = useMemo(() => ({
    low: t('order.low'),
    normal: t('order.normal'),
    high: t('order.high'),
    urgent: t('order.urgent'),
  }), [t])

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

  // ── Create order modal ────────────────────────────────────────────────────
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createForm] = Form.useForm<CreateOrderRequest>()
  const [createLoading, setCreateLoading] = useState(false)
  const [createLines, setCreateLines] = useState<CreateOrderLineRequest[]>([
    { sku_id: '', ordered_qty: 1, uom: 'EA', batch_no: '', notes: '' },
  ])

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
        message.error(t('order.loadFailed'))
      } finally {
        setLoading(false)
      }
    },
    [message, t],
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
      message.error(t('order.loadDetailFailed'))
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
      message.success(t('order.statusUpdated', { status: titleCase(newStatus) }))
      const { data } = await client.get<Order>(`/orders/${order.id}`)
      setSelectedOrder(data)
      fetchOrders(page, filterType, filterStatus)
    } catch {
      message.error(t('order.statusUpdateFailed'))
    }
  }

  // ── Create order helpers ──────────────────────────────────────────────────

  const openCreateModal = () => {
    createForm.resetFields()
    createForm.setFieldsValue({ priority: 'normal' })
    setCreateLines([{ sku_id: '', ordered_qty: 1, uom: 'EA', batch_no: '', notes: '' }])
    setCreateModalOpen(true)
  }

  const closeCreateModal = () => {
    setCreateModalOpen(false)
  }

  const addLine = () => {
    setCreateLines([...createLines, { sku_id: '', ordered_qty: 1, uom: 'EA', batch_no: '', notes: '' }])
  }

  const removeLine = (idx: number) => {
    if (createLines.length <= 1) return
    setCreateLines(createLines.filter((_, i) => i !== idx))
  }

  const updateLine = (idx: number, field: keyof CreateOrderLineRequest, value: string | number) => {
    const next = [...createLines]
    next[idx] = { ...next[idx], [field]: value }
    setCreateLines(next)
  }

  const handleCreateOrder = async () => {
    try {
      const values = await createForm.validateFields()

      // Validate lines
      const validLines = createLines.filter((l) => l.sku_id.trim() && l.ordered_qty > 0)
      if (validLines.length === 0) {
        message.error(t('order.pleaseAddAtLeastOneLine'))
        return
      }
      for (const line of validLines) {
        if (!line.sku_id.trim()) {
          message.error(t('order.pleaseAddAtLeastOneLine'))
          return
        }
      }

      const payload: CreateOrderRequest = {
        ...values,
        lines: validLines.map((l) => ({
          sku_id: l.sku_id.trim(),
          ordered_qty: l.ordered_qty,
          uom: l.uom || 'EA',
          batch_no: l.batch_no || undefined,
          notes: l.notes || undefined,
        })),
      }

      setCreateLoading(true)
      await client.post('/orders', payload)
      message.success(t('order.createSuccess'))
      closeCreateModal()
      fetchOrders(page, filterType, filterStatus)
    } catch {
      if (createModalOpen) message.error(t('order.createFailed'))
    } finally {
      setCreateLoading(false)
    }
  }

  // ── Table columns ────────────────────────────────────────────────────────

  const columns: ColumnsType<OrderSummary> = useMemo(
    () => [
      {
        title: t('order.orderNo'),
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
        title: t('order.orderType'),
        dataIndex: 'order_type',
        key: 'order_type',
        width: 110,
        render: (ot: string) => (
          <Tag color={orderTypeColors[ot] ?? 'default'}>{typeLabels[ot] ?? titleCase(ot)}</Tag>
        ),
      },
      {
        title: t('common.status'),
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
              {statusLabels[s] ?? titleCase(s)}
            </Tag>
          )
        },
      },
      {
        title: t('order.priority'),
        dataIndex: 'priority',
        key: 'priority',
        width: 100,
        responsive: ['md'],
        render: (p: string) => (
          <Tag color={orderPriorityColors[p] ?? 'default'}>{priorityLabels[p] ?? titleCase(p)}</Tag>
        ),
      },
      {
        title: t('order.externalRef'),
        dataIndex: 'external_ref',
        key: 'external_ref',
        width: 150,
        responsive: ['lg'],
        ellipsis: true,
        render: (ref: string) =>
          ref || <Typography.Text type="secondary">—</Typography.Text>,
      },
      {
        title: t('common.created'),
        dataIndex: 'created_at',
        key: 'created_at',
        width: 170,
        responsive: ['lg'],
        render: (v: string) => new Date(v).toLocaleString(),
      },
      {
        title: t('common.actions'),
        key: 'actions',
        width: 90,
        render: (_: unknown, record: OrderSummary) => (
          <Button
            type="default"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => openDetail(record.id)}
          >
            {t('order.view')}
          </Button>
        ),
      },
    ],
    [t, typeLabels, statusLabels, priorityLabels],
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
      title: t('order.lineNo'),
      dataIndex: 'line_no',
      key: 'line_no',
      width: 50,
    },
    {
      title: t('order.skuId'),
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
      title: t('order.qty'),
      dataIndex: 'ordered_qty',
      key: 'ordered_qty',
      width: 90,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: t('order.fulfilled'),
      dataIndex: 'fulfilled_qty',
      key: 'fulfilled_qty',
      width: 90,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: t('order.uom'),
      dataIndex: 'uom',
      key: 'uom',
      width: 70,
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (s: string) => (
        <Tag color={orderLineStatusColors[s] ?? 'default'}>{titleCase(s)}</Tag>
      ),
    },
    {
      title: t('inventory.batch'),
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
        <Typography.Title level={2}>{t('order.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('order.subtitle')}
        </Typography.Text>
      </div>

      {/* ── Summary stat cards ───────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('common.total')}
              value={stats.total}
              prefix={<FileTextOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('order.draft')}
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
              title={t('order.processing')}
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
              title={t('order.completed')}
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
              title={t('order.partial')}
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
              title={t('order.cancelled')}
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
            <span>{t('order.allOrders')}</span>
          </Space>
        }
        extra={
          <Space>
            <FilterOutlined />
            <Select
              placeholder={t('order.filterType')}
              allowClear
              style={{ width: 140 }}
              value={filterType || undefined}
              onChange={(v) => handleFilterChange(v ?? '', filterStatus)}
              options={[
                { label: t('order.inbound'), value: 'inbound' },
                { label: t('order.outbound'), value: 'outbound' },
                { label: t('order.transfer'), value: 'transfer' },
                { label: t('order.return'), value: 'return' },
              ]}
            />
            <Select
              placeholder={t('order.filterStatus')}
              allowClear
              style={{ width: 140 }}
              value={filterStatus || undefined}
              onChange={(v) => handleFilterChange(filterType, v ?? '')}
              options={[
                { label: t('order.draft'), value: 'draft' },
                { label: t('order.confirmed'), value: 'confirmed' },
                { label: t('order.processing'), value: 'processing' },
                { label: t('order.partial'), value: 'partial' },
                { label: t('order.completed'), value: 'completed' },
                { label: t('order.cancelled'), value: 'cancelled' },
              ]}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              {t('order.newOrder')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('order.refresh')}
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
            showTotal: (t2, range) => `${range[0]}-${range[1]} of ${t2}`,
          }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={t('order.noOrders')}
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
              <span>{t('order.orderDetail')}</span>
              <Typography.Text code>{selectedOrder.order_no}</Typography.Text>
            </Space>
          ) : (
            t('order.orderDetail')
          )
        }
        placement="right"
        width={720}
        open={drawerOpen}
        onClose={closeDrawer}
        destroyOnClose
        extra={
          <Button onClick={closeDrawer}>{t('common.close')}</Button>
        }
      >
        {drawerLoading ? (
          <Empty description={t('order.loadingDetail')} />
        ) : selectedOrder ? (
          <>
            {/* ── Order info ──────────────────────────────────────────────────── */}

            <Descriptions
              bordered
              column={{ xs: 1, sm: 2 }}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Descriptions.Item label={t('order.orderNo')}>
                <Typography.Text code>{selectedOrder.order_no}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('order.orderType')}>
                <Tag color={orderTypeColors[selectedOrder.order_type] ?? 'default'}>
                  {typeLabels[selectedOrder.order_type] ?? titleCase(selectedOrder.order_type)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('common.status')}>
                <Tag color={orderStatusColors[selectedOrder.status] ?? 'default'}>
                  {statusLabels[selectedOrder.status] ?? titleCase(selectedOrder.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('order.priority')}>
                <Tag color={orderPriorityColors[selectedOrder.priority] ?? 'default'}>
                  {priorityLabels[selectedOrder.priority] ?? titleCase(selectedOrder.priority)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('order.warehouseId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedOrder.warehouse_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('order.createdBy')}>
                {selectedOrder.created_by}
              </Descriptions.Item>
              {selectedOrder.external_ref && (
                <Descriptions.Item label={t('order.externalRef')}>
                  {selectedOrder.external_ref}
                </Descriptions.Item>
              )}
              {selectedOrder.external_type && (
                <Descriptions.Item label={t('order.externalType')}>
                  <Typography.Text code>{selectedOrder.external_type}</Typography.Text>
                </Descriptions.Item>
              )}
              {selectedOrder.notes && (
                <Descriptions.Item label={t('order.notes')} span={2}>
                  {selectedOrder.notes}
                </Descriptions.Item>
              )}
              <Descriptions.Item label={t('common.created')}>
                {new Date(selectedOrder.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('common.updated')}>
                {new Date(selectedOrder.updated_at).toLocaleString()}
              </Descriptions.Item>
              {selectedOrder.completed_at && (
                <Descriptions.Item label={t('order.completed')} span={2}>
                  {new Date(selectedOrder.completed_at).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>

            {/* ── Status transitions ──────────────────────────────────────────── */}

            {statusTransitions[selectedOrder.status] &&
              statusTransitions[selectedOrder.status].length > 0 && (
                <Card
                  size="small"
                  title={t('order.statusActions')}
                  style={{ marginBottom: 24 }}
                >
                  <Space wrap>
                    {statusTransitions[selectedOrder.status].map((target) => (
                      <Popconfirm
                        key={target}
                        title={t('order.moveTo', { target: statusLabels[target] ?? titleCase(target) })}
                        onConfirm={() =>
                          handleStatusTransition(selectedOrder, target)
                        }
                        okText={t('common.yes')}
                        cancelText={t('common.no')}
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
                          {statusLabels[target] ?? titleCase(target)}
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
                  <span>{t('order.lineItems', { count: selectedOrder.lines?.length ?? 0 })}</span>
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
                      description={t('order.noLineItems')}
                    />
                  ),
                }}
              />
            </Card>
          </>
        ) : (
          <Empty description={t('order.noOrderSelected')} />
        )}
      </Drawer>

      {/* ── Create Order Modal ────────────────────────────────────────────────── */}

      <Modal
        title={t('order.createOrderTitle')}
        open={createModalOpen}
        onCancel={closeCreateModal}
        onOk={handleCreateOrder}
        confirmLoading={createLoading}
        destroyOnClose
        width={720}
        okText={t('order.createOrder')}
        cancelText={t('common.cancel')}
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          {/* ── Order header fields ──────────────────────────────────────────── */}
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="order_type"
                label={t('order.orderType')}
                rules={[{ required: true, message: t('order.pleaseSelectOrderType') }]}
              >
                <Select placeholder={t('order.filterType')}>
                  <Select.Option value="inbound">{t('order.inbound')}</Select.Option>
                  <Select.Option value="outbound">{t('order.outbound')}</Select.Option>
                  <Select.Option value="transfer">{t('order.transfer')}</Select.Option>
                  <Select.Option value="return">{t('order.return')}</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item
                name="warehouse_id"
                label={t('order.warehouseId')}
                rules={[{ required: true, message: t('order.pleaseEnterWarehouse') }]}
              >
                <Input placeholder={t('order.warehousePlaceholder')} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="priority"
                label={t('order.priority')}
                initialValue="normal"
              >
                <Select>
                  <Select.Option value="low">{t('order.low')}</Select.Option>
                  <Select.Option value="normal">{t('order.normal')}</Select.Option>
                  <Select.Option value="high">{t('order.high')}</Select.Option>
                  <Select.Option value="urgent">{t('order.urgent')}</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item
                name="created_by"
                label={t('order.createdBy')}
                rules={[{ required: true, message: t('order.pleaseEnterCreatedBy') }]}
              >
                <Input placeholder={t('order.createdByPlaceholder')} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item name="external_ref" label={t('order.externalRef')}>
                <Input placeholder={t('order.externalRefPlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item name="external_type" label={t('order.externalType')}>
                <Input placeholder={t('order.externalTypePlaceholder')} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="notes" label={t('order.notes')}>
            <Input placeholder={t('order.notesPlaceholder')} />
          </Form.Item>

          {/* ── Order lines editor ────────────────────────────────────────────── */}
          <Divider orientation="left" plain>
            <Typography.Text type="secondary">{t('order.lineEditorTitle')} ({createLines.length})</Typography.Text>
          </Divider>

          {createLines.map((line, idx) => (
            <Card
              key={idx}
              size="small"
              style={{ marginBottom: 12 }}
              title={`${t('order.lineNo')} ${idx + 1}`}
              extra={
                createLines.length > 1 && (
                  <Button
                    type="text"
                    danger
                    size="small"
                    icon={<DeleteOutlined />}
                    onClick={() => removeLine(idx)}
                  >
                    {t('order.removeLine')}
                  </Button>
                )
              }
            >
              <Row gutter={12}>
                <Col xs={12}>
                  <Form.Item
                    label={t('order.skuId')}
                    required
                    style={{ marginBottom: 8 }}
                  >
                    <Input
                      placeholder={t('order.skuPlaceholder')}
                      value={line.sku_id}
                      onChange={(e) => updateLine(idx, 'sku_id', e.target.value)}
                    />
                  </Form.Item>
                </Col>
                <Col xs={6}>
                  <Form.Item
                    label={t('order.qty')}
                    required
                    style={{ marginBottom: 8 }}
                  >
                    <InputNumber
                      style={{ width: '100%' }}
                      min={1}
                      placeholder={t('order.qtyPlaceholder')}
                      value={line.ordered_qty}
                      onChange={(v) => updateLine(idx, 'ordered_qty', v ?? 1)}
                    />
                  </Form.Item>
                </Col>
                <Col xs={6}>
                  <Form.Item label={t('order.uom')} style={{ marginBottom: 8 }}>
                    <Input
                      placeholder={t('order.uomPlaceholder')}
                      value={line.uom}
                      onChange={(e) => updateLine(idx, 'uom', e.target.value)}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col xs={12}>
                  <Form.Item label={t('inventory.batch')} style={{ marginBottom: 0 }}>
                    <Input
                      placeholder={t('order.batchPlaceholder')}
                      value={line.batch_no}
                      onChange={(e) => updateLine(idx, 'batch_no', e.target.value)}
                    />
                  </Form.Item>
                </Col>
                <Col xs={12}>
                  <Form.Item label={t('order.notes')} style={{ marginBottom: 0 }}>
                    <Input
                      placeholder={t('order.lineNotesPlaceholder')}
                      value={line.notes}
                      onChange={(e) => updateLine(idx, 'notes', e.target.value)}
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          ))}

          <Button
            type="dashed"
            onClick={addLine}
            block
            icon={<PlusOutlined />}
            style={{ marginTop: 4 }}
          >
            {t('order.addLine')}
          </Button>
        </Form>
      </Modal>
    </div>
  )
}
