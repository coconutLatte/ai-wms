// ASN (Advanced Shipping Notice) management page.
// List table with status badges and filters; create ASN modal; detail drawer
// with line items; status transitions (pending→arrived→receiving→partial→received).
// Uses POST/GET/PUT /api/v1/asns endpoints from P3-16.

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
  DatePicker,
  Divider,
} from 'antd'
import {
  TruckOutlined,
  FilterOutlined,
  ReloadOutlined,
  InboxOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  PlusOutlined,
  DeleteOutlined,
  IdcardOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import client from '@/api/client'
import type {
  ASNSummary,
  ASN,
  ASNLine,
  ListResponse,
  CreateASNRequest,
  CreateASNLineRequest,
  UpdateASNStatusRequest,
} from '@/api/types'

// ── Status tag colors ──────────────────────────────────────────────────────

const asnStatusColors: Record<string, string> = {
  pending: 'default',
  arrived: 'blue',
  receiving: 'processing',
  partial: 'warning',
  received: 'success',
}

const asnLineStatusColors: Record<string, string> = {
  pending: 'default',
  partial: 'warning',
  received: 'success',
}

// ── Status labels (title case) ─────────────────────────────────────────────

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// ── Valid status transitions (matches backend CanTransitionTo) ─────────────

const statusTransitions: Record<string, string[]> = {
  pending: ['arrived'],
  arrived: ['receiving'],
  receiving: ['partial', 'received'],
  partial: ['received'],
}

// ── Constants ──────────────────────────────────────────────────────────────

const PAGE_SIZE = 10

// ── Main component ─────────────────────────────────────────────────────────

export default function ASNsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ─────────────────────────────────────────────────────

  const statusLabels: Record<string, string> = useMemo(() => ({
    pending: t('asn.pending'),
    arrived: t('asn.arrived'),
    receiving: t('asn.receiving'),
    partial: t('asn.partial'),
    received: t('asn.received'),
  }), [t])

  const lineStatusLabels: Record<string, string> = useMemo(() => ({
    pending: t('asn.pending'),
    partial: t('asn.partial'),
    received: t('asn.received'),
  }), [t])

  // ── List state ───────────────────────────────────────────────────────────
  const [asns, setAsns] = useState<ASNSummary[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // ── Filters ──────────────────────────────────────────────────────────────
  const [filterStatus, setFilterStatus] = useState<string>('')

  // ── Detail drawer ────────────────────────────────────────────────────────
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedASN, setSelectedASN] = useState<ASN | null>(null)
  const [drawerLoading, setDrawerLoading] = useState(false)

  // ── Create ASN modal ─────────────────────────────────────────────────────
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createForm] = Form.useForm<CreateASNRequest>()
  const [createLoading, setCreateLoading] = useState(false)
  const [createLines, setCreateLines] = useState<CreateASNLineRequest[]>([
    { sku_id: '', expected_qty: 1 },
  ])

  // ── Data fetching ────────────────────────────────────────────────────────

  const fetchAsns = useCallback(
    async (p: number, status: string) => {
      setLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: p,
          page_size: PAGE_SIZE,
        }
        if (status) params.status = status

        const { data } = await client.get<ListResponse<ASNSummary>>('/asns', {
          params,
        })
        setAsns(data.data)
        setTotal(data.pagination.total)
      } catch {
        message.error(t('asn.loadFailed'))
      } finally {
        setLoading(false)
      }
    },
    [message, t],
  )

  useEffect(() => {
    fetchAsns(page, filterStatus)
  }, [page, filterStatus, fetchAsns])

  // ── Refresh ──────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    if (page === 1) {
      fetchAsns(1, filterStatus)
    } else {
      setPage(1)
    }
  }

  const handleFilterChange = (status: string) => {
    setFilterStatus(status)
    setPage(1)
  }

  // ── Table change handler ─────────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Detail drawer ────────────────────────────────────────────────────────

  const openDetail = async (asnId: string) => {
    setDrawerOpen(true)
    setDrawerLoading(true)
    try {
      const { data } = await client.get<ASN>(`/asns/${asnId}`)
      setSelectedASN(data)
    } catch {
      message.error(t('asn.loadDetailFailed'))
    } finally {
      setDrawerLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerOpen(false)
    setSelectedASN(null)
  }

  // ── Status transition ────────────────────────────────────────────────────

  const handleStatusTransition = async (asn: ASN, newStatus: string) => {
    try {
      await client.put(`/asns/${asn.id}/status`, {
        status: newStatus,
      } satisfies UpdateASNStatusRequest)
      message.success(t('asn.statusUpdated', { status: statusLabels[newStatus] ?? titleCase(newStatus) }))
      const { data } = await client.get<ASN>(`/asns/${asn.id}`)
      setSelectedASN(data)
      fetchAsns(page, filterStatus)
    } catch {
      message.error(t('asn.statusUpdateFailed'))
    }
  }

  // ── Create ASN helpers ───────────────────────────────────────────────────

  const openCreateModal = () => {
    createForm.resetFields()
    setCreateLines([{ sku_id: '', expected_qty: 1 }])
    setCreateModalOpen(true)
  }

  const closeCreateModal = () => {
    setCreateModalOpen(false)
  }

  const addLine = () => {
    setCreateLines([...createLines, { sku_id: '', expected_qty: 1 }])
  }

  const removeLine = (idx: number) => {
    if (createLines.length <= 1) return
    setCreateLines(createLines.filter((_, i) => i !== idx))
  }

  const updateLine = (idx: number, field: keyof CreateASNLineRequest, value: string | number) => {
    const next = [...createLines]
    next[idx] = { ...next[idx], [field]: value }
    setCreateLines(next)
  }

  const handleCreateASN = async () => {
    try {
      const values = await createForm.validateFields()

      // Validate lines
      const validLines = createLines.filter((l) => l.sku_id.trim() && l.expected_qty > 0)
      if (validLines.length === 0) {
        message.error(t('asn.pleaseAddAtLeastOneLine'))
        return
      }

      const payload: CreateASNRequest = {
        warehouse_id: values.warehouse_id,
        order_id: values.order_id || undefined,
        carrier: values.carrier || undefined,
        tracking_no: values.tracking_no || undefined,
        expected_at: values.expected_at
          ? dayjs(values.expected_at).toISOString()
          : dayjs().toISOString(),
        lines: validLines.map((l) => ({
          sku_id: l.sku_id.trim(),
          expected_qty: l.expected_qty,
          batch_no: l.batch_no || undefined,
        })),
      }

      setCreateLoading(true)
      await client.post('/asns', payload)
      message.success(t('asn.createSuccess'))
      closeCreateModal()
      fetchAsns(page, filterStatus)
    } catch {
      if (createModalOpen) message.error(t('asn.createFailed'))
    } finally {
      setCreateLoading(false)
    }
  }

  // ── Table columns ────────────────────────────────────────────────────────

  const columns: ColumnsType<ASNSummary> = useMemo(
    () => [
      {
        title: t('asn.asnNo'),
        dataIndex: 'asn_no',
        key: 'asn_no',
        width: 200,
        render: (no: string) => (
          <Typography.Text code style={{ cursor: 'pointer' }}>
            {no}
          </Typography.Text>
        ),
      },
      {
        title: t('common.status'),
        dataIndex: 'status',
        key: 'status',
        width: 130,
        render: (s: string) => {
          const icon =
            s === 'received' ? (
              <CheckCircleOutlined />
            ) : s === 'receiving' ? (
              <SyncOutlined spin />
            ) : s === 'arrived' ? (
              <ClockCircleOutlined />
            ) : s === 'partial' ? (
              <ExclamationCircleOutlined />
            ) : null
          return (
            <Tag icon={icon} color={asnStatusColors[s] ?? 'default'}>
              {statusLabels[s] ?? titleCase(s)}
            </Tag>
          )
        },
      },
      {
        title: t('asn.carrier'),
        dataIndex: 'carrier',
        key: 'carrier',
        width: 150,
        responsive: ['md'],
        ellipsis: true,
        render: (c: string) =>
          c || <Typography.Text type="secondary">—</Typography.Text>,
      },
      {
        title: t('asn.trackingNo'),
        dataIndex: 'tracking_no',
        key: 'tracking_no',
        width: 180,
        responsive: ['lg'],
        ellipsis: true,
        render: (t2: string) =>
          t2 || <Typography.Text type="secondary">—</Typography.Text>,
      },
      {
        title: t('asn.expectedAt'),
        dataIndex: 'expected_at',
        key: 'expected_at',
        width: 170,
        responsive: ['lg'],
        render: (v: string) => new Date(v).toLocaleString(),
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
        render: (_: unknown, record: ASNSummary) => (
          <Button
            type="default"
            size="small"
            icon={<IdcardOutlined />}
            onClick={() => openDetail(record.id)}
          >
            {t('asn.view')}
          </Button>
        ),
      },
    ],
    [t, statusLabels],
  )

  // ── ASN summary stats ────────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = {}
    asns.forEach((a) => {
      counts[a.status] = (counts[a.status] ?? 0) + 1
    })
    return {
      total: total,
      pending: counts['pending'] ?? 0,
      arrived: counts['arrived'] ?? 0,
      receiving: counts['receiving'] ?? 0,
      partial: counts['partial'] ?? 0,
      received: counts['received'] ?? 0,
    }
  }, [asns, total])

  // ── Line item columns for detail drawer ──────────────────────────────────

  const lineColumns: ColumnsType<ASNLine> = [
    {
      title: t('asn.skuId'),
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
      title: t('asn.expectedQty'),
      dataIndex: 'expected_qty',
      key: 'expected_qty',
      width: 100,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: t('asn.receivedQty'),
      dataIndex: 'received_qty',
      key: 'received_qty',
      width: 100,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: t('inventory.batch'),
      dataIndex: 'batch_no',
      key: 'batch_no',
      width: 130,
      responsive: ['md'],
      render: (b: string) =>
        b || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (s: string) => (
        <Tag color={asnLineStatusColors[s] ?? 'default'}>
          {lineStatusLabels[s] ?? titleCase(s)}
        </Tag>
      ),
    },
  ]

  // ── Render ───────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('asn.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('asn.subtitle')}
        </Typography.Text>
      </div>

      {/* ── Summary stat cards ───────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('common.total')}
              value={stats.total}
              prefix={<TruckOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('asn.pending')}
              value={stats.pending}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: stats.pending > 0 ? '#8c8c8c' : undefined }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('asn.arrived')}
              value={stats.arrived}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: stats.arrived > 0 ? '#1677ff' : undefined }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('asn.receiving')}
              value={stats.receiving}
              prefix={<SyncOutlined spin={stats.receiving > 0} />}
              valueStyle={{
                color: stats.receiving > 0 ? '#1677ff' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('asn.received')}
              value={stats.received}
              prefix={<CheckCircleOutlined />}
              valueStyle={{
                color: stats.received > 0 ? '#52c41a' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('asn.partial')}
              value={stats.partial}
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{
                color: stats.partial > 0 ? '#faad14' : undefined,
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
            <TruckOutlined />
            <span>{t('asn.allAsns')}</span>
          </Space>
        }
        extra={
          <Space>
            <FilterOutlined />
            <Select
              placeholder={t('asn.filterStatus')}
              allowClear
              style={{ width: 140 }}
              value={filterStatus || undefined}
              onChange={(v) => handleFilterChange(v ?? '')}
              options={[
                { label: t('asn.pending'), value: 'pending' },
                { label: t('asn.arrived'), value: 'arrived' },
                { label: t('asn.receiving'), value: 'receiving' },
                { label: t('asn.partial'), value: 'partial' },
                { label: t('asn.received'), value: 'received' },
              ]}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              {t('asn.newAsn')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('asn.refresh')}
            </Button>
          </Space>
        }
      >
        <Table<ASNSummary>
          columns={columns}
          dataSource={asns}
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
                description={t('asn.noAsns')}
              />
            ),
          }}
        />
      </Card>

      {/* ── Detail Drawer ──────────────────────────────────────────────────── */}

      <Drawer
        title={
          selectedASN ? (
            <Space>
              <TruckOutlined />
              <span>{t('asn.asnDetail')}</span>
              <Typography.Text code>{selectedASN.asn_no}</Typography.Text>
            </Space>
          ) : (
            t('asn.asnDetail')
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
          <Empty description={t('asn.loadingDetail')} />
        ) : selectedASN ? (
          <>
            {/* ── ASN info ──────────────────────────────────────────────────── */}

            <Descriptions
              bordered
              column={{ xs: 1, sm: 2 }}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Descriptions.Item label={t('asn.asnNo')}>
                <Typography.Text code>{selectedASN.asn_no}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('common.status')}>
                <Tag color={asnStatusColors[selectedASN.status] ?? 'default'}>
                  {statusLabels[selectedASN.status] ?? titleCase(selectedASN.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('asn.warehouseId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedASN.warehouse_id}
                </Typography.Text>
              </Descriptions.Item>
              {selectedASN.order_id && (
                <Descriptions.Item label={t('asn.orderId')}>
                  <Typography.Text code style={{ fontSize: 12 }}>
                    {selectedASN.order_id}
                  </Typography.Text>
                </Descriptions.Item>
              )}
              <Descriptions.Item label={t('asn.carrier')}>
                {selectedASN.carrier || <Typography.Text type="secondary">—</Typography.Text>}
              </Descriptions.Item>
              <Descriptions.Item label={t('asn.trackingNo')}>
                {selectedASN.tracking_no || <Typography.Text type="secondary">—</Typography.Text>}
              </Descriptions.Item>
              <Descriptions.Item label={t('asn.expectedAt')}>
                {new Date(selectedASN.expected_at).toLocaleString()}
              </Descriptions.Item>
              {selectedASN.arrived_at && (
                <Descriptions.Item label={t('asn.arrivedAt')}>
                  {new Date(selectedASN.arrived_at).toLocaleString()}
                </Descriptions.Item>
              )}
              <Descriptions.Item label={t('common.created')}>
                {new Date(selectedASN.created_at).toLocaleString()}
              </Descriptions.Item>
            </Descriptions>

            {/* ── Status transitions ──────────────────────────────────────────── */}

            {statusTransitions[selectedASN.status] &&
              statusTransitions[selectedASN.status].length > 0 && (
                <Card
                  size="small"
                  title={t('asn.statusActions')}
                  style={{ marginBottom: 24 }}
                >
                  <Space wrap>
                    {statusTransitions[selectedASN.status].map((target) => (
                      <Popconfirm
                        key={target}
                        title={t('asn.moveTo', { target: statusLabels[target] ?? titleCase(target) })}
                        onConfirm={() =>
                          handleStatusTransition(selectedASN, target)
                        }
                        okText={t('common.yes')}
                        cancelText={t('common.no')}
                      >
                        <Button
                          type="primary"
                          size="small"
                          icon={
                            target === 'received' ? (
                              <CheckCircleOutlined />
                            ) : target === 'arrived' ? (
                              <TruckOutlined />
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
                  <span>{t('asn.lineItems', { count: selectedASN.lines?.length ?? 0 })}</span>
                </Space>
              }
            >
              <Table<ASNLine>
                columns={lineColumns}
                dataSource={selectedASN.lines ?? []}
                rowKey="id"
                size="small"
                pagination={false}
                locale={{
                  emptyText: (
                    <Empty
                      image={Empty.PRESENTED_IMAGE_SIMPLE}
                      description={t('asn.noLineItems')}
                    />
                  ),
                }}
              />
            </Card>
          </>
        ) : (
          <Empty description={t('asn.noAsnSelected')} />
        )}
      </Drawer>

      {/* ── Create ASN Modal ────────────────────────────────────────────────── */}

      <Modal
        title={t('asn.createAsnTitle')}
        open={createModalOpen}
        onCancel={closeCreateModal}
        onOk={handleCreateASN}
        confirmLoading={createLoading}
        destroyOnClose
        width={720}
        okText={t('asn.createAsn')}
        cancelText={t('common.cancel')}
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          {/* ── ASN header fields ──────────────────────────────────────────── */}
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="warehouse_id"
                label={t('asn.warehouseId')}
                rules={[{ required: true, message: t('asn.pleaseEnterWarehouse') }]}
              >
                <Input placeholder={t('asn.warehousePlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item
                name="order_id"
                label={t('asn.orderId')}
              >
                <Input placeholder={t('asn.orderIdPlaceholder')} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="carrier"
                label={t('asn.carrier')}
              >
                <Input placeholder={t('asn.carrierPlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item
                name="tracking_no"
                label={t('asn.trackingNo')}
              >
                <Input placeholder={t('asn.trackingNoPlaceholder')} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="expected_at"
                label={t('asn.expectedAt')}
              >
                <DatePicker
                  showTime
                  style={{ width: '100%' }}
                  placeholder={t('asn.expectedAtPlaceholder')}
                />
              </Form.Item>
            </Col>
          </Row>

          {/* ── ASN lines editor ────────────────────────────────────────────── */}
          <Divider orientation="left" plain>
            <Typography.Text type="secondary">{t('asn.lineEditorTitle')} ({createLines.length})</Typography.Text>
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
                <Col xs={16}>
                  <Form.Item
                    label={t('asn.skuId')}
                    required
                    style={{ marginBottom: 8 }}
                  >
                    <Input
                      placeholder={t('asn.skuPlaceholder')}
                      value={line.sku_id}
                      onChange={(e) => updateLine(idx, 'sku_id', e.target.value)}
                    />
                  </Form.Item>
                </Col>
                <Col xs={8}>
                  <Form.Item
                    label={t('asn.expectedQty')}
                    required
                    style={{ marginBottom: 8 }}
                  >
                    <InputNumber
                      style={{ width: '100%' }}
                      min={1}
                      placeholder={t('asn.qtyPlaceholder')}
                      value={line.expected_qty}
                      onChange={(v) => updateLine(idx, 'expected_qty', v ?? 1)}
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col xs={12}>
                  <Form.Item label={t('inventory.batch')} style={{ marginBottom: 0 }}>
                    <Input
                      placeholder={t('asn.batchPlaceholder')}
                      value={line.batch_no}
                      onChange={(e) => updateLine(idx, 'batch_no', e.target.value)}
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
