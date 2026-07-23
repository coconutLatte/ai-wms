// Location management page — standalone list with zone and warehouse filters.
// Create/edit modals with type and capacity. Status transitions with state machine.
// Uses GET /api/v1/locations?zone_id=X&warehouse_id=Y, POST /api/v1/zones/{id}/locations,
// PUT /api/v1/locations/{id}, PATCH /api/v1/locations/{id}/status. i18n: zh-CN + en.

import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Typography,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Card,
  App,
  Empty,
  Row,
  Col,
  Statistic,
  InputNumber,
  Popconfirm,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  EnvironmentOutlined,
  FilterOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  StopOutlined,
  LockOutlined,
  UnlockOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  Warehouse,
  Zone,
  Location,
  ListResponse,
  CreateLocationRequest,
} from '@/api/types'

// ── Status & type colors ──────────────────────────────────────────────────

const locationStatusColors: Record<string, string> = {
  empty: 'default',
  occupied: 'green',
  reserved: 'blue',
  blocked: 'red',
}

// ── Valid status transitions (matches domain.CanTransitionTo) ─────────────

const statusTransitions: Record<string, string[]> = {
  empty: ['occupied', 'reserved', 'blocked'],
  occupied: ['empty', 'blocked'],
  reserved: ['occupied', 'empty', 'blocked'],
  blocked: ['empty'],
}

const PAGE_SIZE = 10

export default function LocationsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ───────────────────────────────────────────────────

  const locTypeLabels: Record<string, string> = useMemo(() => ({
    pallet: t('location.typePallet'),
    shelf: t('location.typeShelf'),
    floor: t('location.typeFloor'),
    conveyor: t('location.typeConveyor'),
    agv: t('location.typeAgv'),
  }), [t])

  const locStatusLabels: Record<string, string> = useMemo(() => ({
    empty: t('location.statusEmpty'),
    occupied: t('location.statusOccupied'),
    reserved: t('location.statusReserved'),
    blocked: t('location.statusBlocked'),
  }), [t])

  // ── State ──────────────────────────────────────────────────────────────
  const [locations, setLocations] = useState<Location[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // Filters
  const [warehouses, setWarehouses] = useState<Warehouse[]>([])
  const [zones, setZones] = useState<Zone[]>([])
  const [selectedWarehouse, setSelectedWarehouse] = useState<string>('')
  const [selectedZone, setSelectedZone] = useState<string>('')

  // Create location modal
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm] = Form.useForm<CreateLocationRequest>()
  const [createLoading, setCreateLoading] = useState(false)

  // Edit location modal
  const [editOpen, setEditOpen] = useState(false)
  const [editLocation, setEditLocation] = useState<Location | null>(null)
  const [editForm] = Form.useForm()
  const [editLoading, setEditLoading] = useState(false)

  // ── Data fetching ──────────────────────────────────────────────────────

  const fetchWarehouses = useCallback(async () => {
    try {
      const { data } = await client.get<ListResponse<Warehouse>>('/warehouses', {
        params: { page: 1, page_size: 999 },
      })
      setWarehouses(data.data)
    } catch { /* optional */ }
  }, [])

  // Fetch zones for the selected warehouse (for the create-location dropdown)
  const fetchZones = useCallback(async (whId: string) => {
    if (!whId) { setZones([]); return }
    try {
      const { data } = await client.get<ListResponse<Zone>>(`/warehouses/${whId}/zones`, {
        params: { page: 1, page_size: 999 },
      })
      setZones(data.data)
    } catch { /* optional */ }
  }, [])

  const fetchLocations = useCallback(async (p: number, wh: string, zn: string) => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = { page: p, page_size: PAGE_SIZE }
      if (wh) params.warehouse_id = wh
      if (zn) params.zone_id = zn

      const { data } = await client.get<ListResponse<Location>>('/locations', { params })
      setLocations(data.data)
      setTotal(data.pagination.total)
    } catch {
      message.error(t('location.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [message, t])

  useEffect(() => { fetchWarehouses() }, [fetchWarehouses])

  // When warehouse filter changes, fetch zones for that warehouse
  useEffect(() => {
    if (selectedWarehouse) {
      fetchZones(selectedWarehouse)
    } else {
      setZones([])
    }
  }, [selectedWarehouse, fetchZones])

  useEffect(() => {
    fetchLocations(page, selectedWarehouse, selectedZone)
  }, [page, selectedWarehouse, selectedZone, fetchLocations])

  // ── Filter handlers ────────────────────────────────────────────────────

  const handleWarehouseFilter = (wh: string) => {
    setSelectedWarehouse(wh)
    if (!wh) setSelectedZone('')
    setPage(1)
  }

  const handleZoneFilter = (zn: string) => {
    setSelectedZone(zn)
    setPage(1)
  }

  const handleRefresh = () => {
    if (page === 1) fetchLocations(1, selectedWarehouse, selectedZone)
    else setPage(1)
  }

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Create location ────────────────────────────────────────────────────

  const openCreate = () => {
    createForm.resetFields()
    setCreateOpen(true)
  }

  const handleCreate = async () => {
    if (!selectedZone) {
      message.error(t('location.filterByZone'))
      return
    }
    try {
      const values = await createForm.validateFields()
      setCreateLoading(true)
      await client.post(`/zones/${selectedZone}/locations`, values)
      message.success(t('location.locationCreated'))
      setCreateOpen(false)
      fetchLocations(page, selectedWarehouse, selectedZone)
    } catch {
      if (createOpen) message.error(t('location.createFailed'))
    } finally {
      setCreateLoading(false)
    }
  }

  // ── Edit location ──────────────────────────────────────────────────────

  const openEdit = (loc: Location) => {
    setEditLocation(loc)
    editForm.setFieldsValue({
      code: loc.code,
      barcode: loc.barcode,
      location_type: loc.location_type,
      'capacity.max_qty': loc.capacity?.max_qty ?? 0,
      'capacity.max_weight': loc.capacity?.max_weight ?? 0,
      'capacity.max_volume': loc.capacity?.max_volume ?? 0,
    })
    setEditOpen(true)
  }

  const handleEdit = async () => {
    if (!editLocation) return
    try {
      const values = await editForm.validateFields()
      setEditLoading(true)
      const payload: Record<string, unknown> = {}

      if (values.code && values.code !== editLocation.code) payload.code = values.code
      if (values.barcode !== undefined && values.barcode !== editLocation.barcode) payload.barcode = values.barcode || ''
      if (values.location_type && values.location_type !== editLocation.location_type) payload.location_type = values.location_type

      const maxQty = values['capacity.max_qty'] ?? 0
      const maxWeight = values['capacity.max_weight'] ?? 0
      const maxVolume = values['capacity.max_volume'] ?? 0
      const curCap = editLocation.capacity
      if (maxQty !== (curCap?.max_qty ?? 0) || maxWeight !== (curCap?.max_weight ?? 0) || maxVolume !== (curCap?.max_volume ?? 0)) {
        payload.capacity = { max_qty: maxQty, max_weight: maxWeight, max_volume: maxVolume }
      }

      if (Object.keys(payload).length > 0) {
        await client.put(`/locations/${editLocation.id}`, payload)
        message.success(t('location.locationUpdated'))
      }
      setEditOpen(false)
      fetchLocations(page, selectedWarehouse, selectedZone)
    } catch {
      message.error(t('location.updateFailed'))
    } finally {
      setEditLoading(false)
    }
  }

  // ── Status transition ──────────────────────────────────────────────────

  const handleStatusTransition = async (loc: Location, newStatus: string) => {
    try {
      await client.patch(`/locations/${loc.id}/status`, { status: newStatus })
      message.success(t('location.statusUpdated', { status: locStatusLabels[newStatus] ?? newStatus }))
      fetchLocations(page, selectedWarehouse, selectedZone)
    } catch {
      message.error(t('location.statusUpdateFailed'))
    }
  }

  // ── Stats ──────────────────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = { empty: 0, occupied: 0, reserved: 0, blocked: 0 }
    locations.forEach((l) => { counts[l.status] = (counts[l.status] ?? 0) + 1 })
    return {
      total,
      empty: counts.empty ?? 0,
      occupied: counts.occupied ?? 0,
      reserved: counts.reserved ?? 0,
      blocked: counts.blocked ?? 0,
    }
  }, [locations, total])

  // ── Table columns ──────────────────────────────────────────────────────

  const columns: ColumnsType<Location> = useMemo(() => [
    {
      title: t('location.locationCode'),
      dataIndex: 'code',
      key: 'code',
      width: 160,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: t('location.barcode'),
      dataIndex: 'barcode',
      key: 'barcode',
      width: 160,
      ellipsis: true,
      responsive: ['md'],
      render: (b: string) => b || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: t('location.locationType'),
      dataIndex: 'location_type',
      key: 'location_type',
      width: 120,
      render: (lt: string) => locTypeLabels[lt] ?? lt,
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status: string) => (
        <Tag color={locationStatusColors[status] ?? 'default'}>
          {locStatusLabels[status] ?? status}
        </Tag>
      ),
    },
    {
      title: t('location.capacity'),
      key: 'capacity',
      width: 180,
      responsive: ['lg'],
      render: (_: unknown, record: Location) => {
        if (!record.capacity) return <Typography.Text type="secondary">{t('location.unlimited')}</Typography.Text>
        const c = record.capacity
        return (
          <Typography.Text>
            {c.max_qty > 0 ? `${c.max_qty} ${t('location.units')}` : ''}
            {c.max_weight > 0 ? ` · ${c.max_weight}kg` : ''}
            {c.max_volume > 0 ? ` · ${c.max_volume}m³` : ''}
          </Typography.Text>
        )
      },
    },
    {
      title: t('location.statusActions'),
      key: 'status_actions',
      width: 180,
      render: (_: unknown, record: Location) => {
        const transitions = statusTransitions[record.status]
        if (!transitions || transitions.length === 0) return null
        return (
          <Space size="small" wrap>
            {transitions.map((target) => (
              <Popconfirm
                key={target}
                title={t('location.moveTo', { target: locStatusLabels[target] ?? target })}
                onConfirm={() => handleStatusTransition(record, target)}
                okText={t('common.yes')}
                cancelText={t('common.no')}
              >
                <Button size="small"
                  type={target === 'blocked' ? 'default' : 'primary'}
                  danger={target === 'blocked'}
                  icon={
                    target === 'occupied' ? <CheckCircleOutlined /> :
                    target === 'blocked' ? <StopOutlined /> :
                    target === 'empty' ? <UnlockOutlined /> :
                    <LockOutlined />
                  }
                >
                  {locStatusLabels[target] ?? target}
                </Button>
              </Popconfirm>
            ))}
          </Space>
        )
      },
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 90,
      fixed: 'right' as const,
      render: (_: unknown, record: Location) => (
        <Button
          type="default"
          size="small"
          icon={<EditOutlined />}
          onClick={() => openEdit(record)}
        >
          {t('common.edit')}
        </Button>
      ),
    },
  ], [t, locTypeLabels, locStatusLabels])

  // ── Render ──────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('location.title')}</Typography.Title>
        <Typography.Text type="secondary">{t('location.subtitle')}</Typography.Text>
      </div>

      {/* ── Stat cards ────────────────────────────────────────────────────── */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={4}>
          <Card size="small" loading={loading}>
            <Statistic title={t('common.total')} value={stats.total}
              prefix={<EnvironmentOutlined />}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={4}>
          <Card size="small" loading={loading}>
            <Statistic title={t('location.statusEmpty')} value={stats.empty}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={4}>
          <Card size="small" loading={loading}>
            <Statistic title={t('location.statusOccupied')} value={stats.occupied}
              valueStyle={{ color: '#52c41a' }}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={4}>
          <Card size="small" loading={loading}>
            <Statistic title={t('location.statusReserved')} value={stats.reserved}
              valueStyle={{ color: '#1677ff' }}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={4}>
          <Card size="small" loading={loading}>
            <Statistic title={t('location.statusBlocked')} value={stats.blocked}
              valueStyle={{ color: '#ff4d4f' }}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
      </Row>

      {/* ── Table card ────────────────────────────────────────────────────── */}
      <Card
        title={
          <Space>
            <EnvironmentOutlined />
            <span>{t('location.allLocations')}</span>
          </Space>
        }
        extra={
          <Space wrap>
            <FilterOutlined />
            <Select
              placeholder={t('location.filterByWarehouse')}
              allowClear
              style={{ width: 180 }}
              value={selectedWarehouse || undefined}
              onChange={(v) => handleWarehouseFilter(v ?? '')}
              options={warehouses.map((w) => ({
                label: `${w.name} (${w.code})`,
                value: w.id,
              }))}
            />
            <Select
              placeholder={t('location.filterByZone')}
              allowClear
              style={{ width: 180 }}
              value={selectedZone || undefined}
              onChange={(v) => handleZoneFilter(v ?? '')}
              options={zones.map((z) => ({
                label: `${z.name} (${z.code})`,
                value: z.id,
              }))}
              disabled={!selectedWarehouse}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}
              disabled={!selectedZone}>
              {t('location.newLocation')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('refresh')}
            </Button>
          </Space>
        }
      >
        <Table<Location>
          columns={columns}
          dataSource={locations}
          rowKey="id"
          loading={loading}
          onChange={handleTableChange}
          scroll={{ x: 1100 }}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total,
            showSizeChanger: false,
            showTotal: (t2, range) => `${range[0]}-${range[1]} of ${t2}`,
          }}
          locale={{
            emptyText: (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('location.noLocations')} />
            ),
          }}
        />
      </Card>

      {/* ── Create Location Modal ─────────────────────────────────────────── */}
      <Modal
        title={t('location.createLocation')}
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        confirmLoading={createLoading}
        destroyOnClose
        width={560}
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="code"
                label={t('location.locationCode')}
                rules={[{ required: true, message: t('location.pleaseEnterCode') }]}
              >
                <Input placeholder={t('location.codePlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item name="barcode" label={t('location.barcode')}>
                <Input placeholder={t('location.barcodePlaceholder')} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="location_type"
            label={t('location.locationType')}
            rules={[{ required: true, message: t('location.pleaseSelectType') }]}
          >
            <Select placeholder={t('location.pleaseSelectType')}>
              <Select.Option value="pallet">{t('location.typePallet')}</Select.Option>
              <Select.Option value="shelf">{t('location.typeShelf')}</Select.Option>
              <Select.Option value="floor">{t('location.typeFloor')}</Select.Option>
              <Select.Option value="conveyor">{t('location.typeConveyor')}</Select.Option>
              <Select.Option value="agv">{t('location.typeAgv')}</Select.Option>
            </Select>
          </Form.Item>

          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
            {t('location.capacityHint')}
          </Typography.Text>

          <Space style={{ display: 'flex' }} align="start">
            <Form.Item name={['capacity', 'max_qty']} label={t('location.maxQty')} style={{ marginBottom: 0 }}>
              <InputNumber min={0} placeholder={t('location.units')} style={{ width: 110 }} />
            </Form.Item>
            <Form.Item name={['capacity', 'max_weight']} label={t('location.maxWeight')} style={{ marginBottom: 0 }}>
              <InputNumber min={0} step={0.1} placeholder="kg" style={{ width: 130 }} />
            </Form.Item>
            <Form.Item name={['capacity', 'max_volume']} label={t('location.maxVolume')} style={{ marginBottom: 0 }}>
              <InputNumber min={0} step={0.01} placeholder="m³" style={{ width: 130 }} />
            </Form.Item>
          </Space>
        </Form>
      </Modal>

      {/* ── Edit Location Modal ──────────────────────────────────────────── */}
      <Modal
        title={t('location.editLocation')}
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleEdit}
        confirmLoading={editLoading}
        destroyOnClose
        width={560}
      >
        {editLocation && (
          <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
            <Row gutter={16}>
              <Col xs={12}>
                <Form.Item
                  name="code"
                  label={t('location.locationCode')}
                  rules={[{ required: true, message: t('location.pleaseEnterCode') }]}
                >
                  <Input placeholder={t('location.codePlaceholder')} />
                </Form.Item>
              </Col>
              <Col xs={12}>
                <Form.Item name="barcode" label={t('location.barcode')}>
                  <Input placeholder={t('location.barcodePlaceholder')} />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="location_type" label={t('location.locationType')}>
              <Select placeholder={t('location.pleaseSelectType')}>
                <Select.Option value="pallet">{t('location.typePallet')}</Select.Option>
                <Select.Option value="shelf">{t('location.typeShelf')}</Select.Option>
                <Select.Option value="floor">{t('location.typeFloor')}</Select.Option>
                <Select.Option value="conveyor">{t('location.typeConveyor')}</Select.Option>
                <Select.Option value="agv">{t('location.typeAgv')}</Select.Option>
              </Select>
            </Form.Item>

            <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
              {t('location.capacityHint')}
            </Typography.Text>

            <Space style={{ display: 'flex' }} align="start">
              <Form.Item name={['capacity', 'max_qty']} label={t('location.maxQty')} style={{ marginBottom: 0 }}>
                <InputNumber min={0} placeholder={t('location.units')} style={{ width: 110 }} />
              </Form.Item>
              <Form.Item name={['capacity', 'max_weight']} label={t('location.maxWeight')} style={{ marginBottom: 0 }}>
                <InputNumber min={0} step={0.1} placeholder="kg" style={{ width: 130 }} />
              </Form.Item>
              <Form.Item name={['capacity', 'max_volume']} label={t('location.maxVolume')} style={{ marginBottom: 0 }}>
                <InputNumber min={0} step={0.01} placeholder="m³" style={{ width: 130 }} />
              </Form.Item>
            </Space>
          </Form>
        )}
      </Modal>
    </div>
  )
}
