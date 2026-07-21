// Warehouse management page with full CRUD for warehouses, zones, and locations.
// Supports listing, creating, and editing warehouses with nested zone/location management.
// All UI text is translated via react-i18next.

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
  InputNumber,
  Breadcrumb,
  Card,
  App,
  Descriptions,
  Empty,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  ShopOutlined,
  EnvironmentOutlined,
  AppstoreOutlined,
  HomeOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  Warehouse,
  Zone,
  Location,
  ListResponse,
  CreateWarehouseRequest,
  UpdateWarehouseRequest,
  CreateZoneRequest,
  CreateLocationRequest,
} from '@/api/types'

// ── Status & type tag colors ────────────────────────────────────────────

const warehouseStatusColors: Record<string, string> = {
  active: 'green',
  inactive: 'orange',
  archived: 'red',
}

const zoneTypeColors: Record<string, string> = {
  receiving: 'blue',
  storage: 'purple',
  picking: 'cyan',
  shipping: 'magenta',
  returns: 'red',
  staging: 'gold',
}

const locationStatusColors: Record<string, string> = {
  empty: 'default',
  occupied: 'green',
  reserved: 'blue',
  blocked: 'red',
}

// ── View state ──────────────────────────────────────────────────────────

type ViewState =
  | { type: 'warehouses' }
  | { type: 'zones'; warehouse: Warehouse }
  | { type: 'locations'; warehouse: Warehouse; zone: Zone }

type ModalState =
  | { type: 'none' }
  | { type: 'create-warehouse' }
  | { type: 'edit-warehouse'; warehouse: Warehouse }
  | { type: 'create-zone' }
  | { type: 'create-location' }

const PAGE_SIZE = 10

export default function WarehousesPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ──────────────────────────────────────────────────

  const locationTypeLabels: Record<string, string> = useMemo(() => ({
    pallet: t('location.typePallet'),
    shelf: t('location.typeShelf'),
    floor: t('location.typeFloor'),
    conveyor: t('location.typeConveyor'),
    agv: t('location.typeAgv'),
  }), [t])

  const zoneTypeLabels: Record<string, string> = useMemo(() => ({
    receiving: t('zone.typeReceiving'),
    storage: t('zone.typeStorage'),
    picking: t('zone.typePicking'),
    shipping: t('zone.typeShipping'),
    returns: t('zone.typeReturns'),
    staging: t('zone.typeStaging'),
  }), [t])

  const warehouseStatusLabels: Record<string, string> = useMemo(() => ({
    active: t('warehouse.statusActive'),
    inactive: t('warehouse.statusInactive'),
    archived: t('warehouse.statusArchived'),
  }), [t])

  // ── View navigation ──────────────────────────────────────────────────
  const [view, setView] = useState<ViewState>({ type: 'warehouses' })

  // ── Warehouse state ──────────────────────────────────────────────────
  const [warehouses, setWarehouses] = useState<Warehouse[]>([])
  const [warehouseLoading, setWarehouseLoading] = useState(false)
  const [warehousePage, setWarehousePage] = useState(1)
  const [warehouseTotal, setWarehouseTotal] = useState(0)

  // ── Zone state ───────────────────────────────────────────────────────
  const [zones, setZones] = useState<Zone[]>([])
  const [zoneLoading, setZoneLoading] = useState(false)
  const [zonePage, setZonePage] = useState(1)
  const [zoneTotal, setZoneTotal] = useState(0)

  // ── Location state ───────────────────────────────────────────────────
  const [locations, setLocations] = useState<Location[]>([])
  const [locationLoading, setLocationLoading] = useState(false)
  const [locationPage, setLocationPage] = useState(1)
  const [locationTotal, setLocationTotal] = useState(0)

  // ── Modal state ──────────────────────────────────────────────────────
  const [modal, setModal] = useState<ModalState>({ type: 'none' })
  const [modalLoading, setModalLoading] = useState(false)

  // ── Forms ────────────────────────────────────────────────────────────
  const [warehouseForm] = Form.useForm<CreateWarehouseRequest>()
  const [editWarehouseForm] = Form.useForm<UpdateWarehouseRequest & { code: string }>()
  const [zoneForm] = Form.useForm<CreateZoneRequest>()
  const [locationForm] = Form.useForm<CreateLocationRequest>()

  // ── Data fetching ────────────────────────────────────────────────────

  const fetchWarehouses = useCallback(async (page: number) => {
    setWarehouseLoading(true)
    try {
      const { data } = await client.get<ListResponse<Warehouse>>('/warehouses', {
        params: { page, page_size: PAGE_SIZE },
      })
      setWarehouses(data.data)
      setWarehouseTotal(data.pagination.total)
    } catch {
      message.error(t('warehouse.loadFailed'))
    } finally {
      setWarehouseLoading(false)
    }
  }, [message, t])

  const fetchZones = useCallback(async (warehouseId: string, page: number) => {
    setZoneLoading(true)
    try {
      const { data } = await client.get<ListResponse<Zone>>(
        `/warehouses/${warehouseId}/zones`,
        { params: { page, page_size: PAGE_SIZE } },
      )
      setZones(data.data)
      setZoneTotal(data.pagination.total)
    } catch {
      message.error(t('zone.loadFailed'))
    } finally {
      setZoneLoading(false)
    }
  }, [message, t])

  const fetchLocations = useCallback(async (zoneId: string, page: number) => {
    setLocationLoading(true)
    try {
      const { data } = await client.get<ListResponse<Location>>(
        `/zones/${zoneId}/locations`,
        { params: { page, page_size: PAGE_SIZE } },
      )
      setLocations(data.data)
      setLocationTotal(data.pagination.total)
    } catch {
      message.error(t('location.loadFailed'))
    } finally {
      setLocationLoading(false)
    }
  }, [message, t])

  // ── Initial load & view changes ──────────────────────────────────────

  useEffect(() => {
    fetchWarehouses(warehousePage)
  }, [warehousePage, fetchWarehouses])

  useEffect(() => {
    if (view.type === 'warehouses') {
      setWarehousePage(1)
      fetchWarehouses(1)
    } else if (view.type === 'zones') {
      setZonePage(1)
      fetchZones(view.warehouse.id, 1)
    } else if (view.type === 'locations') {
      setLocationPage(1)
      fetchLocations(view.zone.id, 1)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view])

  // ── Handlers ─────────────────────────────────────────────────────────

  const navigateToZones = (warehouse: Warehouse) => {
    setView({ type: 'zones', warehouse })
  }

  const navigateToLocations = (zone: Zone) => {
    const wh = (view as { type: 'zones'; warehouse: Warehouse }).warehouse
    if (!wh) return
    setView({ type: 'locations', warehouse: wh, zone })
  }

  const navigateToWarehouses = () => {
    setView({ type: 'warehouses' })
  }

  const handleWarehouseTableChange = (pagination: { current?: number }) => {
    const page = pagination.current ?? 1
    setWarehousePage(page)
    fetchWarehouses(page)
  }

  const handleZoneTableChange = (pagination: { current?: number }) => {
    if (view.type !== 'zones') return
    const page = pagination.current ?? 1
    setZonePage(page)
    fetchZones(view.warehouse.id, page)
  }

  const handleLocationTableChange = (pagination: { current?: number }) => {
    if (view.type !== 'locations') return
    const page = pagination.current ?? 1
    setLocationPage(page)
    fetchLocations(view.zone.id, page)
  }

  // ── Modal actions ────────────────────────────────────────────────────

  const openCreateWarehouse = () => {
    warehouseForm.resetFields()
    setModal({ type: 'create-warehouse' })
  }

  const openEditWarehouse = (warehouse: Warehouse) => {
    editWarehouseForm.setFieldsValue({
      name: warehouse.name,
      address: warehouse.address,
      status: warehouse.status,
      code: warehouse.code,
    })
    setModal({ type: 'edit-warehouse', warehouse })
  }

  const openCreateZone = () => {
    zoneForm.resetFields()
    setModal({ type: 'create-zone' })
  }

  const openCreateLocation = () => {
    locationForm.resetFields()
    setModal({ type: 'create-location' })
  }

  const closeModal = () => setModal({ type: 'none' })

  const handleCreateWarehouse = async () => {
    try {
      const values = await warehouseForm.validateFields()
      setModalLoading(true)
      await client.post('/warehouses', values)
      message.success(t('warehouse.created'))
      closeModal()
      fetchWarehouses(warehousePage)
    } catch {
      if (modal.type !== 'none') message.error(t('warehouse.createFailed'))
    } finally {
      setModalLoading(false)
    }
  }

  const handleEditWarehouse = async () => {
    if (modal.type !== 'edit-warehouse') return
    try {
      const values = await editWarehouseForm.validateFields()
      setModalLoading(true)
      const payload: UpdateWarehouseRequest = {}
      if (values.name) payload.name = values.name
      if (values.address !== undefined) payload.address = values.address
      if (values.status) payload.status = values.status
      await client.put(`/warehouses/${modal.warehouse.id}`, payload)
      message.success(t('warehouse.updated'))
      closeModal()
      fetchWarehouses(warehousePage)
    } catch {
      message.error(t('warehouse.updateFailed'))
    } finally {
      setModalLoading(false)
    }
  }

  const handleCreateZone = async () => {
    if (view.type !== 'zones') return
    try {
      const values = await zoneForm.validateFields()
      setModalLoading(true)
      await client.post(`/warehouses/${view.warehouse.id}/zones`, values)
      message.success(t('zone.zoneCreated'))
      closeModal()
      fetchZones(view.warehouse.id, zonePage)
    } catch {
      if (modal.type !== 'none') message.error(t('zone.createFailed'))
    } finally {
      setModalLoading(false)
    }
  }

  const handleCreateLocation = async () => {
    if (view.type !== 'locations') return
    try {
      const values = await locationForm.validateFields()
      setModalLoading(true)
      await client.post(`/zones/${view.zone.id}/locations`, values)
      message.success(t('location.locationCreated'))
      closeModal()
      fetchLocations(view.zone.id, locationPage)
    } catch {
      if (modal.type !== 'none') message.error(t('location.createFailed'))
    } finally {
      setModalLoading(false)
    }
  }

  // ── Warehouse columns ────────────────────────────────────────────────

  const warehouseColumns: ColumnsType<Warehouse> = useMemo(() => [
    {
      title: t('warehouse.warehouseCode'),
      dataIndex: 'code',
      key: 'code',
      width: 140,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: t('warehouse.warehouseName'),
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: t('warehouse.address'),
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
      responsive: ['md'],
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={warehouseStatusColors[status] ?? 'default'}>
          {warehouseStatusLabels[status] ?? status}
        </Tag>
      ),
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
      width: 180,
      render: (_: unknown, record: Warehouse) => (
        <Space>
          <Button
            type="default"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditWarehouse(record)}
          >
            {t('common.edit')}
          </Button>
          <Button
            type="primary"
            size="small"
            icon={<AppstoreOutlined />}
            onClick={() => navigateToZones(record)}
          >
            {t('warehouse.zones')}
          </Button>
        </Space>
      ),
    },
  ], [t, warehouseStatusLabels])

  // ── Zone columns ─────────────────────────────────────────────────────

  const zoneColumns: ColumnsType<Zone> = useMemo(() => [
    {
      title: t('zone.zoneCode'),
      dataIndex: 'code',
      key: 'code',
      width: 140,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: t('zone.zoneName'),
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: t('zone.zoneType'),
      dataIndex: 'zone_type',
      key: 'zone_type',
      width: 120,
      render: (t2: string) => (
        <Tag color={zoneTypeColors[t2] ?? 'default'}>{zoneTypeLabels[t2] ?? t2}</Tag>
      ),
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : status === 'inactive' ? 'orange' : 'red'}>
          {status}
        </Tag>
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 140,
      render: (_: unknown, record: Zone) => (
        <Button
          type="primary"
          size="small"
          icon={<EnvironmentOutlined />}
          onClick={() => navigateToLocations(record)}
        >
          {t('warehouse.locations')}
        </Button>
      ),
    },
  ], [t, zoneTypeLabels])

  // ── Location columns ─────────────────────────────────────────────────

  const locationColumns: ColumnsType<Location> = useMemo(() => [
    {
      title: t('location.locationCode'),
      dataIndex: 'code',
      key: 'code',
      width: 150,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: t('location.barcode'),
      dataIndex: 'barcode',
      key: 'barcode',
      width: 160,
      ellipsis: true,
      render: (b: string) => b || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: t('location.locationType'),
      dataIndex: 'location_type',
      key: 'location_type',
      width: 110,
      render: (lt: string) => locationTypeLabels[lt] ?? lt,
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={locationStatusColors[status] ?? 'default'}>{status}</Tag>
      ),
    },
    {
      title: t('location.capacity'),
      key: 'capacity',
      width: 170,
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
  ], [t, locationTypeLabels])

  // ── Breadcrumb items ─────────────────────────────────────────────────

  const breadcrumbItems = useMemo(() => {
    const items = [
      {
        title: (
          <Button type="link" onClick={navigateToWarehouses} style={{ padding: 0 }}>
            <HomeOutlined /> {t('warehouse.title')}
          </Button>
        ),
      },
    ]
    if (view.type === 'zones') {
      items.push({
        title: (
          <span>
            <ShopOutlined /> {view.warehouse.name}
          </span>
        ),
      })
    } else if (view.type === 'locations') {
      items.push({
        title: (
          <Button type="link" onClick={() => navigateToZones(view.warehouse)} style={{ padding: 0 }}>
            <ShopOutlined /> {view.warehouse.name}
          </Button>
        ),
      })
      items.push({
        title: (
          <span>
            <AppstoreOutlined /> {view.zone.name}
          </span>
        ),
      })
    }
    return items
  }, [view, t])

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Breadcrumb items={breadcrumbItems} style={{ marginBottom: 8 }} />
        <Typography.Title level={2}>
          {view.type === 'warehouses'
            ? t('warehouse.title')
            : view.type === 'zones'
              ? `${t('warehouse.zones')} — ${view.warehouse.name}`
              : `${t('warehouse.locations')} — ${view.zone.name}`}
        </Typography.Title>
      </div>

      {/* ── Warehouse view ───────────────────────────────────────────── */}

      {view.type === 'warehouses' && (
        <Card
          title={
            <Space>
              <ShopOutlined />
              <span>{t('warehouse.allWarehouses')}</span>
            </Space>
          }
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateWarehouse}>
              {t('warehouse.newWarehouse')}
            </Button>
          }
        >
          <Table<Warehouse>
            columns={warehouseColumns}
            dataSource={warehouses}
            rowKey="id"
            loading={warehouseLoading}
            onChange={handleWarehouseTableChange}
            pagination={{
              current: warehousePage,
              pageSize: PAGE_SIZE,
              total: warehouseTotal,
              showSizeChanger: false,
              showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
            }}
            locale={{ emptyText: <Empty description={t('warehouse.noWarehouses')} /> }}
          />
        </Card>
      )}

      {/* ── Zone view ────────────────────────────────────────────────── */}

      {view.type === 'zones' && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label={t('common.code')}>
                <Typography.Text code>{view.warehouse.code}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('warehouse.address')}>{view.warehouse.address}</Descriptions.Item>
              <Descriptions.Item label={t('common.status')}>
                <Tag color={warehouseStatusColors[view.warehouse.status]}>
                  {warehouseStatusLabels[view.warehouse.status] ?? view.warehouse.status}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card
            title={
              <Space>
                <AppstoreOutlined />
                <span>{t('zone.title')}</span>
              </Space>
            }
            extra={
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateZone}>
                {t('zone.newZone')}
              </Button>
            }
          >
            <Table<Zone>
              columns={zoneColumns}
              dataSource={zones}
              rowKey="id"
              loading={zoneLoading}
              onChange={handleZoneTableChange}
              pagination={{
                current: zonePage,
                pageSize: PAGE_SIZE,
                total: zoneTotal,
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              locale={{ emptyText: <Empty description={t('zone.noZones')} /> }}
            />
          </Card>
        </>
      )}

      {/* ── Location view ─────────────────────────────────────────────── */}

      {view.type === 'locations' && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label={t('warehouse.title')}>{view.warehouse.name}</Descriptions.Item>
              <Descriptions.Item label={t('warehouse.zones')}>
                <Typography.Text code>{view.zone.code}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('zone.zoneType')}>
                <Tag color={zoneTypeColors[view.zone.zone_type]}>{zoneTypeLabels[view.zone.zone_type] ?? view.zone.zone_type}</Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card
            title={
              <Space>
                <EnvironmentOutlined />
                <span>{t('location.title')}</span>
              </Space>
            }
            extra={
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateLocation}>
                {t('location.newLocation')}
              </Button>
            }
          >
            <Table<Location>
              columns={locationColumns}
              dataSource={locations}
              rowKey="id"
              loading={locationLoading}
              onChange={handleLocationTableChange}
              pagination={{
                current: locationPage,
                pageSize: PAGE_SIZE,
                total: locationTotal,
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              locale={{ emptyText: <Empty description={t('location.noLocations')} /> }}
            />
          </Card>
        </>
      )}

      {/* ── Create Warehouse Modal ────────────────────────────────────── */}

      <Modal
        title={t('warehouse.createWarehouse')}
        open={modal.type === 'create-warehouse'}
        onCancel={closeModal}
        onOk={handleCreateWarehouse}
        confirmLoading={modalLoading}
        destroyOnClose
      >
        <Form form={warehouseForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="code"
            label={t('warehouse.warehouseCode')}
            rules={[{ required: true, message: t('warehouse.pleaseEnterCode') }]}
          >
            <Input placeholder={t('warehouse.codePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="name"
            label={t('warehouse.warehouseName')}
            rules={[{ required: true, message: t('warehouse.pleaseEnterName') }]}
          >
            <Input placeholder={t('warehouse.namePlaceholder')} />
          </Form.Item>
          <Form.Item name="address" label={t('warehouse.address')}>
            <Input placeholder={t('warehouse.addressPlaceholder')} />
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Edit Warehouse Modal ──────────────────────────────────────── */}

      <Modal
        title={t('warehouse.editWarehouse')}
        open={modal.type === 'edit-warehouse'}
        onCancel={closeModal}
        onOk={handleEditWarehouse}
        confirmLoading={modalLoading}
        destroyOnClose
      >
        <Form form={editWarehouseForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="code" label={t('warehouse.warehouseCode')}>
            <Input disabled />
          </Form.Item>
          <Form.Item
            name="name"
            label={t('warehouse.warehouseName')}
            rules={[{ required: true, message: t('warehouse.pleaseEnterName') }]}
          >
            <Input placeholder={t('warehouse.namePlaceholder')} />
          </Form.Item>
          <Form.Item name="address" label={t('warehouse.address')}>
            <Input placeholder={t('warehouse.addressPlaceholderEdit')} />
          </Form.Item>
          <Form.Item
            name="status"
            label={t('common.status')}
            rules={[{ required: true, message: t('warehouse.pleaseSelectStatus') }]}
          >
            <Select>
              <Select.Option value="active">{t('warehouse.statusActive')}</Select.Option>
              <Select.Option value="inactive">{t('warehouse.statusInactive')}</Select.Option>
              <Select.Option value="archived">{t('warehouse.statusArchived')}</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Create Zone Modal ─────────────────────────────────────────── */}

      <Modal
        title={t('zone.createZone')}
        open={modal.type === 'create-zone'}
        onCancel={closeModal}
        onOk={handleCreateZone}
        confirmLoading={modalLoading}
        destroyOnClose
      >
        <Form form={zoneForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="code"
            label={t('zone.zoneCode')}
            rules={[{ required: true, message: t('zone.pleaseEnterCode') }]}
          >
            <Input placeholder={t('zone.codePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="name"
            label={t('zone.zoneName')}
            rules={[{ required: true, message: t('zone.pleaseEnterName') }]}
          >
            <Input placeholder={t('zone.namePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="zone_type"
            label={t('zone.zoneType')}
            rules={[{ required: true, message: t('zone.pleaseSelectType') }]}
          >
            <Select placeholder={t('zone.pleaseSelectType')}>
              <Select.Option value="receiving">{t('zone.typeReceiving')}</Select.Option>
              <Select.Option value="storage">{t('zone.typeStorage')}</Select.Option>
              <Select.Option value="picking">{t('zone.typePicking')}</Select.Option>
              <Select.Option value="shipping">{t('zone.typeShipping')}</Select.Option>
              <Select.Option value="returns">{t('zone.typeReturns')}</Select.Option>
              <Select.Option value="staging">{t('zone.typeStaging')}</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Create Location Modal ─────────────────────────────────────── */}

      <Modal
        title={t('location.createLocation')}
        open={modal.type === 'create-location'}
        onCancel={closeModal}
        onOk={handleCreateLocation}
        confirmLoading={modalLoading}
        destroyOnClose
        width={520}
      >
        <Form form={locationForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="code"
            label={t('location.locationCode')}
            rules={[{ required: true, message: t('location.pleaseEnterCode') }]}
          >
            <Input placeholder={t('location.codePlaceholder')} />
          </Form.Item>
          <Form.Item name="barcode" label={t('location.barcode')}>
            <Input placeholder={t('location.barcodePlaceholder')} />
          </Form.Item>
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

          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) =>
              prev?.capacity !== cur?.capacity
            }
          >
            {({ getFieldValue: _gfv }) => {
              void _gfv('capacity')
              return (
                <Space style={{ display: 'flex' }} align="start">
                  <Form.Item
                    name={['capacity', 'max_qty']}
                    label={t('location.maxQty')}
                    style={{ marginBottom: 0 }}
                  >
                    <InputNumber min={0} placeholder={t('location.units')} style={{ width: 110 }} />
                  </Form.Item>
                  <Form.Item
                    name={['capacity', 'max_weight']}
                    label={t('location.maxWeight')}
                    style={{ marginBottom: 0 }}
                  >
                    <InputNumber min={0} step={0.1} placeholder="kg" style={{ width: 130 }} />
                  </Form.Item>
                  <Form.Item
                    name={['capacity', 'max_volume']}
                    label={t('location.maxVolume')}
                    style={{ marginBottom: 0 }}
                  >
                    <InputNumber min={0} step={0.01} placeholder="m³" style={{ width: 130 }} />
                  </Form.Item>
                </Space>
              )
            }}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
