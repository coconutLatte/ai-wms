// Shipment management page with list, filtering, detail view, status transitions,
// create shipment from confirmed outbound orders, and tracking updates.
// Uses GET/POST/PUT /api/v1/shipments endpoints.
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
  Input,
  DatePicker,
} from 'antd'
import {
  SendOutlined,
  FilterOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  PlusOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  StopOutlined,
  EditOutlined,
  TruckOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import client from '@/api/client'
import type {
  Shipment,
  ListResponse,
  CreateShipmentRequest,
  UpdateTrackingRequest,
} from '@/api/types'

// ── Status tag colors ──────────────────────────────────────────────────────────

const shipmentStatusColors: Record<string, string> = {
  pending: 'default',
  in_transit: 'processing',
  delivered: 'success',
  cancelled: 'error',
}

// ── Status icons ──────────────────────────────────────────────────────────────

function statusIcon(s: string) {
  if (s === 'pending') return <ClockCircleOutlined />
  if (s === 'in_transit') return <SyncOutlined spin />
  if (s === 'delivered') return <CheckCircleOutlined />
  if (s === 'cancelled') return <StopOutlined />
  return null
}

// ── Helper ────────────────────────────────────────────────────────────────────

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// ── Constants ─────────────────────────────────────────────────────────────────

const PAGE_SIZE = 10

// ── Main component ────────────────────────────────────────────────────────────

export default function ShipmentsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ────────────────────────────────────────────────────────

  const statusLabels: Record<string, string> = useMemo(
    () => ({
      pending: t('shipment.pending'),
      in_transit: t('shipment.in_transit'),
      delivered: t('shipment.delivered'),
      cancelled: t('shipment.cancelled'),
    }),
    [t],
  )

  // ── List state ──────────────────────────────────────────────────────────────

  const [shipments, setShipments] = useState<Shipment[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // ── Filters ─────────────────────────────────────────────────────────────────

  const [filterStatus, setFilterStatus] = useState<string>('')
  const [filterCarrier, setFilterCarrier] = useState<string>('')

  // ── Detail drawer ───────────────────────────────────────────────────────────

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedShipment, setSelectedShipment] = useState<Shipment | null>(null)
  const [drawerLoading, setDrawerLoading] = useState(false)

  // ── Create modal ────────────────────────────────────────────────────────────

  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createOrderId, setCreateOrderId] = useState('')
  const [createWarehouseId, setCreateWarehouseId] = useState('')
  const [createCarrier, setCreateCarrier] = useState('')
  const [createTrackingNo, setCreateTrackingNo] = useState('')
  const [createCarrierService, setCreateCarrierService] = useState('')
  const [createEstimatedDelivery, setCreateEstimatedDelivery] = useState<string | undefined>()
  const [createNotes, setCreateNotes] = useState('')
  const [createLoading, setCreateLoading] = useState(false)

  // ── Update tracking modal ───────────────────────────────────────────────────

  const [trackingModalOpen, setTrackingModalOpen] = useState(false)
  const [trackingCarrier, setTrackingCarrier] = useState('')
  const [trackingNo, setTrackingNo] = useState('')
  const [trackingCarrierService, setTrackingCarrierService] = useState('')
  const [trackingLoading, setTrackingLoading] = useState(false)

  // ── Data fetching ───────────────────────────────────────────────────────────

  const fetchShipments = useCallback(
    async (p: number, status: string, carrier: string) => {
      setLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: p,
          page_size: PAGE_SIZE,
        }
        if (status) params.status = status
        if (carrier) params.carrier = carrier

        const { data } = await client.get<ListResponse<Shipment>>('/shipments', {
          params,
        })
        setShipments(data.data)
        setTotal(data.pagination.total)
      } catch {
        message.error(t('shipment.loadFailed'))
      } finally {
        setLoading(false)
      }
    },
    [message, t],
  )

  useEffect(() => {
    fetchShipments(page, filterStatus, filterCarrier)
  }, [page, filterStatus, filterCarrier, fetchShipments])

  // ── Refresh ─────────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    if (page === 1) {
      fetchShipments(1, filterStatus, filterCarrier)
    } else {
      setPage(1)
    }
  }

  const handleFilterChange = (status: string, carrier: string) => {
    setFilterStatus(status)
    setFilterCarrier(carrier)
    setPage(1)
  }

  // ── Table change handler ────────────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Detail drawer ───────────────────────────────────────────────────────────

  const openDetail = async (shipmentId: string) => {
    setDrawerOpen(true)
    setDrawerLoading(true)
    try {
      const { data } = await client.get<Shipment>(`/shipments/${shipmentId}`)
      setSelectedShipment(data)
    } catch {
      message.error(t('shipment.loadDetailFailed'))
    } finally {
      setDrawerLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerOpen(false)
    setSelectedShipment(null)
  }

  // ── Status transition ───────────────────────────────────────────────────────

  const handleStatusTransition = async (shipment: Shipment, newStatus: string) => {
    try {
      let endpoint: string
      if (newStatus === 'delivered') {
        // Use the dedicated deliver endpoint
        await client.put(`/shipments/${shipment.id}/deliver`)
      } else {
        await client.put(`/shipments/${shipment.id}/status`, {
          status: newStatus,
        })
      }
      message.success(
        t('shipment.statusUpdated', {
          status: statusLabels[newStatus] ?? titleCase(newStatus),
        }),
      )
      const { data } = await client.get<Shipment>(`/shipments/${shipment.id}`)
      setSelectedShipment(data)
      fetchShipments(page, filterStatus, filterCarrier)
    } catch {
      message.error(t('shipment.statusUpdateFailed'))
    }
  }

  // ── Create shipment ─────────────────────────────────────────────────────────

  const handleCreate = async () => {
    if (!createOrderId.trim()) {
      message.warning(t('shipment.pleaseEnterOrder'))
      return
    }
    if (!createWarehouseId.trim()) {
      message.warning(t('shipment.pleaseEnterWarehouse'))
      return
    }
    if (!createCarrier.trim()) {
      message.warning(t('shipment.pleaseEnterCarrier'))
      return
    }
    setCreateLoading(true)
    try {
      const payload: CreateShipmentRequest = {
        order_id: createOrderId.trim(),
        warehouse_id: createWarehouseId.trim(),
        carrier: createCarrier.trim(),
      }
      if (createTrackingNo.trim()) payload.tracking_no = createTrackingNo.trim()
      if (createCarrierService.trim()) payload.carrier_service = createCarrierService.trim()
      if (createEstimatedDelivery) payload.estimated_delivery = createEstimatedDelivery
      if (createNotes.trim()) payload.notes = createNotes.trim()

      await client.post('/shipments', payload)
      message.success(t('shipment.createSuccess'))
      setCreateModalOpen(false)
      resetCreateForm()
      fetchShipments(1, filterStatus, filterCarrier)
    } catch {
      message.error(t('shipment.createFailed'))
    } finally {
      setCreateLoading(false)
    }
  }

  const resetCreateForm = () => {
    setCreateOrderId('')
    setCreateWarehouseId('')
    setCreateCarrier('')
    setCreateTrackingNo('')
    setCreateCarrierService('')
    setCreateEstimatedDelivery(undefined)
    setCreateNotes('')
  }

  // ── Update tracking ─────────────────────────────────────────────────────────

  const openTrackingModal = () => {
    if (!selectedShipment) return
    setTrackingCarrier(selectedShipment.carrier)
    setTrackingNo(selectedShipment.tracking_no ?? '')
    setTrackingCarrierService(selectedShipment.carrier_service ?? '')
    setTrackingModalOpen(true)
  }

  const handleUpdateTracking = async () => {
    if (!selectedShipment) return
    setTrackingLoading(true)
    try {
      const payload: UpdateTrackingRequest = {}
      if (trackingCarrier.trim()) payload.carrier = trackingCarrier.trim()
      if (trackingNo.trim()) payload.tracking_no = trackingNo.trim()
      if (trackingCarrierService.trim()) payload.carrier_service = trackingCarrierService.trim()

      await client.put(`/shipments/${selectedShipment.id}/tracking`, payload)
      message.success(t('shipment.updateTrackingSuccess'))
      setTrackingModalOpen(false)
      const { data } = await client.get<Shipment>(`/shipments/${selectedShipment.id}`)
      setSelectedShipment(data)
      fetchShipments(page, filterStatus, filterCarrier)
    } catch {
      message.error(t('shipment.updateTrackingFailed'))
    } finally {
      setTrackingLoading(false)
    }
  }

  // ── Table columns ───────────────────────────────────────────────────────────

  const columns: ColumnsType<Shipment> = useMemo(
    () => [
      {
        title: t('shipment.shipmentNo'),
        dataIndex: 'shipment_no',
        key: 'shipment_no',
        width: 220,
        render: (no: string) => (
          <Typography.Text code style={{ cursor: 'pointer' }}>
            {no}
          </Typography.Text>
        ),
      },
      {
        title: t('shipment.orderId'),
        dataIndex: 'order_id',
        key: 'order_id',
        width: 160,
        ellipsis: true,
        responsive: ['md'],
        render: (id: string) => (
          <Typography.Text code style={{ fontSize: 12 }}>
            {id.substring(0, 8)}...
          </Typography.Text>
        ),
      },
      {
        title: t('common.status'),
        dataIndex: 'status',
        key: 'status',
        width: 130,
        render: (s: string) => (
          <Tag icon={statusIcon(s)} color={shipmentStatusColors[s] ?? 'default'}>
            {statusLabels[s] ?? titleCase(s)}
          </Tag>
        ),
      },
      {
        title: t('shipment.carrier'),
        dataIndex: 'carrier',
        key: 'carrier',
        width: 120,
        responsive: ['sm'],
      },
      {
        title: t('shipment.trackingNo'),
        dataIndex: 'tracking_no',
        key: 'tracking_no',
        width: 170,
        responsive: ['lg'],
        render: (v: string | undefined) =>
          v ? <Typography.Text code>{v}</Typography.Text> : '-',
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
        render: (_: unknown, record: Shipment) => (
          <Button
            type="default"
            size="small"
            icon={<FileTextOutlined />}
            onClick={(e) => {
              e.stopPropagation()
              openDetail(record.id)
            }}
          >
            {t('common.view')}
          </Button>
        ),
      },
    ],
    [t, statusLabels],
  )

  // ── Shipment summary stats ──────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = {}
    shipments.forEach((s) => {
      counts[s.status] = (counts[s.status] ?? 0) + 1
    })
    return {
      total: total,
      pending: counts['pending'] ?? 0,
      inTransit: counts['in_transit'] ?? 0,
      delivered: counts['delivered'] ?? 0,
    }
  }, [shipments, total])

  // ── Available carrier options for filter (from current data) ────────────────

  const carrierOptions = useMemo(() => {
    const seen = new Set<string>()
    shipments.forEach((s) => {
      if (s.carrier) seen.add(s.carrier)
    })
    return Array.from(seen).map((c) => ({ label: c, value: c }))
  }, [shipments])

  // ── Is the selected shipment in a terminal state? ───────────────────────────

  const isTerminal =
    selectedShipment?.status === 'delivered' || selectedShipment?.status === 'cancelled'

  // ── Render ──────────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('shipment.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('shipment.subtitle')}
        </Typography.Text>
      </div>

      {/* ── Summary stat cards ──────────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('common.total')}
              value={stats.total}
              prefix={<SendOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('shipment.pending')}
              value={stats.pending}
              prefix={<ClockCircleOutlined />}
              valueStyle={{
                color: stats.pending > 0 ? '#8c8c8c' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('shipment.in_transit')}
              value={stats.inTransit}
              prefix={<SyncOutlined spin={stats.inTransit > 0} />}
              valueStyle={{
                color: stats.inTransit > 0 ? '#1677ff' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('shipment.delivered')}
              value={stats.delivered}
              prefix={<CheckCircleOutlined />}
              valueStyle={{
                color: stats.delivered > 0 ? '#52c41a' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Filter bar ───────────────────────────────────────────────────────── */}

      <Card
        title={
          <Space>
            <SendOutlined />
            <span>{t('shipment.allShipments')}</span>
          </Space>
        }
        extra={
          <Space wrap>
            <FilterOutlined />
            <Select
              placeholder={t('shipment.filterStatus')}
              allowClear
              style={{ width: 130 }}
              value={filterStatus || undefined}
              onChange={(v) => handleFilterChange(v ?? '', filterCarrier)}
              options={[
                { label: t('shipment.pending'), value: 'pending' },
                { label: t('shipment.in_transit'), value: 'in_transit' },
                { label: t('shipment.delivered'), value: 'delivered' },
                { label: t('shipment.cancelled'), value: 'cancelled' },
              ]}
            />
            <Select
              placeholder={t('shipment.filterCarrier')}
              allowClear
              style={{ width: 150 }}
              value={filterCarrier || undefined}
              onChange={(v) => handleFilterChange(filterStatus, v ?? '')}
              options={carrierOptions}
            />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                resetCreateForm()
                setCreateModalOpen(true)
              }}
            >
              {t('shipment.newShipment')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('shipment.refresh')}
            </Button>
          </Space>
        }
      >
        <Table<Shipment>
          columns={columns}
          dataSource={shipments}
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
                description={t('shipment.noShipments')}
              />
            ),
          }}
        />
      </Card>

      {/* ── Detail Drawer ─────────────────────────────────────────────────────── */}

      <Drawer
        title={
          selectedShipment ? (
            <Space>
              <SendOutlined />
              <span>{t('shipment.shipmentDetail')}</span>
              <Typography.Text code>{selectedShipment.shipment_no}</Typography.Text>
            </Space>
          ) : (
            t('shipment.shipmentDetail')
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
          <Empty description={t('shipment.loadingDetail')} />
        ) : selectedShipment ? (
          <>
            {/* ── Shipment info ─────────────────────────────────────────────────── */}

            <Descriptions
              bordered
              column={{ xs: 1, sm: 2 }}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Descriptions.Item label={t('shipment.shipmentNo')}>
                <Typography.Text code>{selectedShipment.shipment_no}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('common.status')}>
                <Tag
                  icon={statusIcon(selectedShipment.status)}
                  color={shipmentStatusColors[selectedShipment.status] ?? 'default'}
                >
                  {statusLabels[selectedShipment.status] ??
                    titleCase(selectedShipment.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('shipment.orderId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedShipment.order_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('shipment.warehouseId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedShipment.warehouse_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('common.created')}>
                {new Date(selectedShipment.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('common.updated')}>
                {new Date(selectedShipment.updated_at).toLocaleString()}
              </Descriptions.Item>
              {selectedShipment.shipped_at && (
                <Descriptions.Item label={t('shipment.shippedAt')}>
                  {new Date(selectedShipment.shipped_at).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedShipment.delivered_at && (
                <Descriptions.Item label={t('shipment.deliveredAt')}>
                  {new Date(selectedShipment.delivered_at).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedShipment.estimated_delivery && (
                <Descriptions.Item label={t('shipment.estimatedDelivery')}>
                  {new Date(selectedShipment.estimated_delivery).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedShipment.actual_delivery && (
                <Descriptions.Item label={t('shipment.actualDelivery')}>
                  {new Date(selectedShipment.actual_delivery).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedShipment.notes && (
                <Descriptions.Item label={t('shipment.notes')} span={2}>
                  {selectedShipment.notes}
                </Descriptions.Item>
              )}
            </Descriptions>

            {/* ── Tracking Info ──────────────────────────────────────────────────── */}

            <Card
              size="small"
              title={t('shipment.trackingTitle')}
              style={{ marginBottom: 24 }}
              extra={
                !isTerminal && (
                  <Button
                    size="small"
                    icon={<EditOutlined />}
                    onClick={openTrackingModal}
                  >
                    {t('shipment.updateTracking')}
                  </Button>
                )
              }
            >
              <Descriptions column={{ xs: 1, sm: 2 }} size="small">
                <Descriptions.Item label={t('shipment.carrier')}>
                  {selectedShipment.carrier}
                </Descriptions.Item>
                <Descriptions.Item label={t('shipment.trackingNo')}>
                  {selectedShipment.tracking_no ? (
                    <Typography.Text code>{selectedShipment.tracking_no}</Typography.Text>
                  ) : (
                    '-'
                  )}
                </Descriptions.Item>
                <Descriptions.Item label={t('shipment.carrierService')}>
                  {selectedShipment.carrier_service || '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {/* ── Status transitions ─────────────────────────────────────────────── */}

            {!isTerminal && (
              <Card
                size="small"
                title={t('shipment.statusActions')}
                style={{ marginBottom: 24 }}
              >
                <Space wrap>
                  {selectedShipment.status === 'pending' && (
                    <>
                      <Popconfirm
                        title={t('shipment.inTransitConfirmTitle')}
                        description={t('shipment.inTransitConfirmMsg', {
                          name: selectedShipment.shipment_no,
                        })}
                        onConfirm={() =>
                          handleStatusTransition(selectedShipment, 'in_transit')
                        }
                        okText={t('common.yes')}
                        cancelText={t('common.no')}
                      >
                        <Button
                          type="primary"
                          size="small"
                          icon={<TruckOutlined />}
                        >
                          {t('shipment.in_transit')}
                        </Button>
                      </Popconfirm>
                      <Popconfirm
                        title={t('shipment.cancelConfirmTitle')}
                        description={t('shipment.cancelConfirmMsg', {
                          name: selectedShipment.shipment_no,
                        })}
                        onConfirm={() =>
                          handleStatusTransition(selectedShipment, 'cancelled')
                        }
                        okText={t('common.yes')}
                        cancelText={t('common.no')}
                        okButtonProps={{ danger: true }}
                      >
                        <Button
                          danger
                          size="small"
                          icon={<StopOutlined />}
                        >
                          {t('shipment.cancelled')}
                        </Button>
                      </Popconfirm>
                    </>
                  )}
                  {selectedShipment.status === 'in_transit' && (
                    <>
                      <Popconfirm
                        title={t('shipment.deliverConfirmTitle')}
                        description={t('shipment.deliverConfirmMsg', {
                          name: selectedShipment.shipment_no,
                        })}
                        onConfirm={() =>
                          handleStatusTransition(selectedShipment, 'delivered')
                        }
                        okText={t('common.yes')}
                        cancelText={t('common.no')}
                      >
                        <Button
                          type="primary"
                          size="small"
                          icon={<CheckCircleOutlined />}
                        >
                          {t('shipment.delivered')}
                        </Button>
                      </Popconfirm>
                      <Popconfirm
                        title={t('shipment.cancelConfirmTitle')}
                        description={t('shipment.cancelConfirmMsg', {
                          name: selectedShipment.shipment_no,
                        })}
                        onConfirm={() =>
                          handleStatusTransition(selectedShipment, 'cancelled')
                        }
                        okText={t('common.yes')}
                        cancelText={t('common.no')}
                        okButtonProps={{ danger: true }}
                      >
                        <Button
                          danger
                          size="small"
                          icon={<StopOutlined />}
                        >
                          {t('shipment.cancelled')}
                        </Button>
                      </Popconfirm>
                    </>
                  )}
                </Space>
              </Card>
            )}
          </>
        ) : (
          <Empty description={t('shipment.noShipmentSelected')} />
        )}
      </Drawer>

      {/* ── Create Shipment Modal ─────────────────────────────────────────────── */}

      <Modal
        title={t('shipment.createShipmentTitle')}
        open={createModalOpen}
        onOk={handleCreate}
        onCancel={() => {
          setCreateModalOpen(false)
          resetCreateForm()
        }}
        confirmLoading={createLoading}
        okText={t('shipment.createShipment')}
        cancelText={t('common.cancel')}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <Typography.Text strong>{t('shipment.orderId')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.orderIdPlaceholder')}
              value={createOrderId}
              onChange={(e) => setCreateOrderId(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.warehouseId')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.warehousePlaceholder')}
              value={createWarehouseId}
              onChange={(e) => setCreateWarehouseId(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.carrier')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.carrierPlaceholder')}
              value={createCarrier}
              onChange={(e) => setCreateCarrier(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.trackingNo')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.trackingNoPlaceholder')}
              value={createTrackingNo}
              onChange={(e) => setCreateTrackingNo(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.carrierService')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.carrierServicePlaceholder')}
              value={createCarrierService}
              onChange={(e) => setCreateCarrierService(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.estimatedDelivery')}</Typography.Text>
            <DatePicker
              showTime
              style={{ width: '100%', marginTop: 4 }}
              placeholder={t('shipment.estimatedDeliveryPlaceholder')}
              value={createEstimatedDelivery ? dayjs(createEstimatedDelivery) : null}
              onChange={(d) => setCreateEstimatedDelivery(d ? d.toISOString() : undefined)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.notes')}</Typography.Text>
            <Input.TextArea
              style={{ marginTop: 4 }}
              rows={3}
              placeholder={t('shipment.notesPlaceholder')}
              value={createNotes}
              onChange={(e) => setCreateNotes(e.target.value)}
            />
          </div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            {t('shipment.onlyConfirmedOutbound')}
          </Typography.Text>
        </Space>
      </Modal>

      {/* ── Update Tracking Modal ─────────────────────────────────────────────── */}

      <Modal
        title={t('shipment.updateTracking')}
        open={trackingModalOpen}
        onOk={handleUpdateTracking}
        onCancel={() => setTrackingModalOpen(false)}
        confirmLoading={trackingLoading}
        okText={t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <Typography.Text strong>{t('shipment.carrier')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.carrierPlaceholder')}
              value={trackingCarrier}
              onChange={(e) => setTrackingCarrier(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.trackingNo')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.trackingNoPlaceholder')}
              value={trackingNo}
              onChange={(e) => setTrackingNo(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>{t('shipment.carrierService')}</Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('shipment.carrierServicePlaceholder')}
              value={trackingCarrierService}
              onChange={(e) => setTrackingCarrierService(e.target.value)}
            />
          </div>
        </Space>
      </Modal>
    </div>
  )
}
