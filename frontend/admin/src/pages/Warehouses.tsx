// Warehouse management page with full CRUD for warehouses, zones, and locations.
// Supports listing, creating, and editing warehouses with nested zone/location management.

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

const locationTypeLabels: Record<string, string> = {
  pallet: 'Pallet',
  shelf: 'Shelf',
  floor: 'Floor',
  conveyor: 'Conveyor',
  agv: 'AGV Dock',
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
      message.error('Failed to load warehouses')
    } finally {
      setWarehouseLoading(false)
    }
  }, [message])

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
      message.error('Failed to load zones')
    } finally {
      setZoneLoading(false)
    }
  }, [message])

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
      message.error('Failed to load locations')
    } finally {
      setLocationLoading(false)
    }
  }, [message])

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
    // Only react to view.type changes
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
      message.success('Warehouse created')
      closeModal()
      fetchWarehouses(warehousePage)
    } catch {
      // Validation error — form shows inline errors; API error shows via interceptor
      if (modal.type !== 'none') message.error('Failed to create warehouse')
    } finally {
      setModalLoading(false)
    }
  }

  const handleEditWarehouse = async () => {
    if (modal.type !== 'edit-warehouse') return
    try {
      const values = await editWarehouseForm.validateFields()
      setModalLoading(true)
      // Only send changed fields
      const payload: UpdateWarehouseRequest = {}
      if (values.name) payload.name = values.name
      if (values.address !== undefined) payload.address = values.address
      if (values.status) payload.status = values.status
      await client.put(`/warehouses/${modal.warehouse.id}`, payload)
      message.success('Warehouse updated')
      closeModal()
      fetchWarehouses(warehousePage)
    } catch {
      if (modal.type !== 'none') message.error('Failed to update warehouse')
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
      message.success('Zone created')
      closeModal()
      fetchZones(view.warehouse.id, zonePage)
    } catch {
      if (modal.type !== 'none') message.error('Failed to create zone')
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
      message.success('Location created')
      closeModal()
      fetchLocations(view.zone.id, locationPage)
    } catch {
      if (modal.type !== 'none') message.error('Failed to create location')
    } finally {
      setModalLoading(false)
    }
  }

  // ── Warehouse columns ────────────────────────────────────────────────

  const warehouseColumns: ColumnsType<Warehouse> = useMemo(() => [
    {
      title: 'Code',
      dataIndex: 'code',
      key: 'code',
      width: 140,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Address',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
      responsive: ['md'],
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={warehouseStatusColors[status] ?? 'default'}>{status}</Tag>
      ),
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
      width: 180,
      render: (_: unknown, record: Warehouse) => (
        <Space>
          <Button
            type="default"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditWarehouse(record)}
          >
            Edit
          </Button>
          <Button
            type="primary"
            size="small"
            icon={<AppstoreOutlined />}
            onClick={() => navigateToZones(record)}
          >
            Zones
          </Button>
        </Space>
      ),
    },
  ], [])

  // ── Zone columns ─────────────────────────────────────────────────────

  const zoneColumns: ColumnsType<Zone> = useMemo(() => [
    {
      title: 'Code',
      dataIndex: 'code',
      key: 'code',
      width: 140,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Type',
      dataIndex: 'zone_type',
      key: 'zone_type',
      width: 120,
      render: (t: string) => (
        <Tag color={zoneTypeColors[t] ?? 'default'}>{t}</Tag>
      ),
    },
    {
      title: 'Status',
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
      title: 'Actions',
      key: 'actions',
      width: 140,
      render: (_: unknown, record: Zone) => (
        <Button
          type="primary"
          size="small"
          icon={<EnvironmentOutlined />}
          onClick={() => navigateToLocations(record)}
        >
          Locations
        </Button>
      ),
    },
  ], [])

  // ── Location columns ─────────────────────────────────────────────────

  const locationColumns: ColumnsType<Location> = useMemo(() => [
    {
      title: 'Code',
      dataIndex: 'code',
      key: 'code',
      width: 150,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: 'Barcode',
      dataIndex: 'barcode',
      key: 'barcode',
      width: 160,
      ellipsis: true,
      render: (b: string) => b || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: 'Type',
      dataIndex: 'location_type',
      key: 'location_type',
      width: 110,
      render: (t: string) => locationTypeLabels[t] ?? t,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={locationStatusColors[status] ?? 'default'}>{status}</Tag>
      ),
    },
    {
      title: 'Capacity',
      key: 'capacity',
      width: 170,
      responsive: ['lg'],
      render: (_: unknown, record: Location) => {
        if (!record.capacity) return <Typography.Text type="secondary">Unlimited</Typography.Text>
        const c = record.capacity
        return (
          <Typography.Text>
            {c.max_qty > 0 ? `${c.max_qty} units` : ''}
            {c.max_weight > 0 ? ` · ${c.max_weight}kg` : ''}
            {c.max_volume > 0 ? ` · ${c.max_volume}m³` : ''}
          </Typography.Text>
        )
      },
    },
  ], [])

  // ── Breadcrumb items ─────────────────────────────────────────────────

  const breadcrumbItems = useMemo(() => {
    const items = [
      {
        title: (
          <Button type="link" onClick={navigateToWarehouses} style={{ padding: 0 }}>
            <HomeOutlined /> Warehouses
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
          <Button type="link" onClick={navigateToZones} style={{ padding: 0 }}>
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
  }, [view])

  // ── Render ───────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Breadcrumb items={breadcrumbItems} style={{ marginBottom: 8 }} />
        <Typography.Title level={2}>
          {view.type === 'warehouses'
            ? 'Warehouses'
            : view.type === 'zones'
              ? `Zones — ${view.warehouse.name}`
              : `Locations — ${view.zone.name}`}
        </Typography.Title>
      </div>

      {/* ── Warehouse view ───────────────────────────────────────────── */}

      {view.type === 'warehouses' && (
        <Card
          title={
            <Space>
              <ShopOutlined />
              <span>All Warehouses</span>
            </Space>
          }
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateWarehouse}>
              New Warehouse
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
            locale={{ emptyText: <Empty description="No warehouses yet" /> }}
          />
        </Card>
      )}

      {/* ── Zone view ────────────────────────────────────────────────── */}

      {view.type === 'zones' && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="Code">
                <Typography.Text code>{view.warehouse.code}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label="Address">{view.warehouse.address}</Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={warehouseStatusColors[view.warehouse.status]}>
                  {view.warehouse.status}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card
            title={
              <Space>
                <AppstoreOutlined />
                <span>Zones</span>
              </Space>
            }
            extra={
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateZone}>
                New Zone
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
              locale={{ emptyText: <Empty description="No zones yet. Create one to get started." /> }}
            />
          </Card>
        </>
      )}

      {/* ── Location view ─────────────────────────────────────────────── */}

      {view.type === 'locations' && (
        <>
          <Card size="small" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2, md: 3 }} size="small">
              <Descriptions.Item label="Warehouse">{view.warehouse.name}</Descriptions.Item>
              <Descriptions.Item label="Zone">
                <Typography.Text code>{view.zone.code}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label="Zone Type">
                <Tag color={zoneTypeColors[view.zone.zone_type]}>{view.zone.zone_type}</Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          <Card
            title={
              <Space>
                <EnvironmentOutlined />
                <span>Locations</span>
              </Space>
            }
            extra={
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateLocation}>
                New Location
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
              locale={{ emptyText: <Empty description="No locations yet. Create one to get started." /> }}
            />
          </Card>
        </>
      )}

      {/* ── Create Warehouse Modal ────────────────────────────────────── */}

      <Modal
        title="Create Warehouse"
        open={modal.type === 'create-warehouse'}
        onCancel={closeModal}
        onOk={handleCreateWarehouse}
        confirmLoading={modalLoading}
        destroyOnClose
      >
        <Form form={warehouseForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="code"
            label="Warehouse Code"
            rules={[{ required: true, message: 'Please enter a warehouse code' }]}
          >
            <Input placeholder="e.g. WH-SH-01" />
          </Form.Item>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Please enter a warehouse name' }]}
          >
            <Input placeholder="e.g. Shanghai Main Warehouse" />
          </Form.Item>
          <Form.Item name="address" label="Address">
            <Input placeholder="Physical address (optional)" />
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Edit Warehouse Modal ──────────────────────────────────────── */}

      <Modal
        title="Edit Warehouse"
        open={modal.type === 'edit-warehouse'}
        onCancel={closeModal}
        onOk={handleEditWarehouse}
        confirmLoading={modalLoading}
        destroyOnClose
      >
        <Form form={editWarehouseForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="code" label="Warehouse Code">
            <Input disabled />
          </Form.Item>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Please enter a warehouse name' }]}
          >
            <Input placeholder="e.g. Shanghai Main Warehouse" />
          </Form.Item>
          <Form.Item name="address" label="Address">
            <Input placeholder="Physical address" />
          </Form.Item>
          <Form.Item
            name="status"
            label="Status"
            rules={[{ required: true, message: 'Please select a status' }]}
          >
            <Select>
              <Select.Option value="active">Active</Select.Option>
              <Select.Option value="inactive">Inactive</Select.Option>
              <Select.Option value="archived">Archived</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Create Zone Modal ─────────────────────────────────────────── */}

      <Modal
        title="Create Zone"
        open={modal.type === 'create-zone'}
        onCancel={closeModal}
        onOk={handleCreateZone}
        confirmLoading={modalLoading}
        destroyOnClose
      >
        <Form form={zoneForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="code"
            label="Zone Code"
            rules={[{ required: true, message: 'Please enter a zone code' }]}
          >
            <Input placeholder="e.g. ZONE-RCV-01" />
          </Form.Item>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Please enter a zone name' }]}
          >
            <Input placeholder="e.g. Receiving Zone A" />
          </Form.Item>
          <Form.Item
            name="zone_type"
            label="Zone Type"
            rules={[{ required: true, message: 'Please select a zone type' }]}
          >
            <Select placeholder="Select zone type">
              <Select.Option value="receiving">Receiving</Select.Option>
              <Select.Option value="storage">Storage</Select.Option>
              <Select.Option value="picking">Picking</Select.Option>
              <Select.Option value="shipping">Shipping</Select.Option>
              <Select.Option value="returns">Returns</Select.Option>
              <Select.Option value="staging">Staging</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Create Location Modal ─────────────────────────────────────── */}

      <Modal
        title="Create Location"
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
            label="Location Code"
            rules={[{ required: true, message: 'Please enter a location code' }]}
          >
            <Input placeholder="e.g. A-01-02-03" />
          </Form.Item>
          <Form.Item name="barcode" label="Barcode">
            <Input placeholder="Barcode on physical location (optional)" />
          </Form.Item>
          <Form.Item
            name="location_type"
            label="Location Type"
            rules={[{ required: true, message: 'Please select a location type' }]}
          >
            <Select placeholder="Select location type">
              <Select.Option value="pallet">Pallet</Select.Option>
              <Select.Option value="shelf">Shelf</Select.Option>
              <Select.Option value="floor">Floor</Select.Option>
              <Select.Option value="conveyor">Conveyor</Select.Option>
              <Select.Option value="agv">AGV Dock</Select.Option>
            </Select>
          </Form.Item>

          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
            Capacity (optional — leave empty or 0 for unlimited)
          </Typography.Text>

          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) =>
              prev?.capacity !== cur?.capacity
            }
          >
            {({ getFieldValue }) => {
              const capacity = getFieldValue('capacity') ?? {}
              return (
                <Space style={{ display: 'flex' }} align="start">
                  <Form.Item
                    name={['capacity', 'max_qty']}
                    label="Max Qty"
                    style={{ marginBottom: 0 }}
                  >
                    <InputNumber min={0} placeholder="Units" style={{ width: 110 }} />
                  </Form.Item>
                  <Form.Item
                    name={['capacity', 'max_weight']}
                    label="Max Weight"
                    style={{ marginBottom: 0 }}
                  >
                    <InputNumber min={0} step={0.1} placeholder="kg" style={{ width: 130 }} />
                  </Form.Item>
                  <Form.Item
                    name={['capacity', 'max_volume']}
                    label="Max Volume"
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
