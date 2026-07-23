// Zone management page — standalone list with warehouse filter.
// Create/edit modal with zone type selector and status management.
// Uses GET /api/v1/zones?warehouse_id=X, POST /api/v1/warehouses/{id}/zones,
// PUT /api/v1/zones/{id}. i18n: zh-CN + en.

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
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  AppstoreOutlined,
  FilterOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  Warehouse,
  Zone,
  ListResponse,
  CreateZoneRequest,
} from '@/api/types'

// ── Status & type tag colors ──────────────────────────────────────────────

const zoneTypeColors: Record<string, string> = {
  receiving: 'blue',
  storage: 'purple',
  picking: 'cyan',
  shipping: 'magenta',
  returns: 'red',
  staging: 'gold',
}

const zoneStatusColors: Record<string, string> = {
  active: 'green',
  inactive: 'orange',
  full: 'red',
}

const PAGE_SIZE = 10

export default function ZonesPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ────────────────────────────────────────────────────

  const zoneTypeLabels: Record<string, string> = useMemo(() => ({
    receiving: t('zone.typeReceiving'),
    storage: t('zone.typeStorage'),
    picking: t('zone.typePicking'),
    shipping: t('zone.typeShipping'),
    returns: t('zone.typeReturns'),
    staging: t('zone.typeStaging'),
  }), [t])

  const zoneStatusLabels: Record<string, string> = useMemo(() => ({
    active: t('zone.statusActive'),
    inactive: t('zone.statusInactive'),
    full: t('zone.statusFull'),
  }), [t])

  // ── State ───────────────────────────────────────────────────────────────
  const [zones, setZones] = useState<Zone[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // Warehouse filter
  const [warehouses, setWarehouses] = useState<Warehouse[]>([])
  const [selectedWarehouse, setSelectedWarehouse] = useState<string>('')

  // Create zone modal
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm] = Form.useForm<CreateZoneRequest>()
  const [createLoading, setCreateLoading] = useState(false)

  // Edit zone modal
  const [editOpen, setEditOpen] = useState(false)
  const [editZone, setEditZone] = useState<Zone | null>(null)
  const [editForm] = Form.useForm()
  const [editLoading, setEditLoading] = useState(false)

  // ── Data fetching ───────────────────────────────────────────────────────

  const fetchWarehouses = useCallback(async () => {
    try {
      const { data } = await client.get<ListResponse<Warehouse>>('/warehouses', {
        params: { page: 1, page_size: 999 },
      })
      setWarehouses(data.data)
    } catch {
      // Warehouse filter is optional; silently fail
    }
  }, [])

  const fetchZones = useCallback(async (p: number, wh: string) => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = { page: p, page_size: PAGE_SIZE }
      if (wh) params.warehouse_id = wh

      const { data } = await client.get<ListResponse<Zone>>('/zones', { params })
      setZones(data.data)
      setTotal(data.pagination.total)
    } catch {
      message.error(t('zone.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [message, t])

  useEffect(() => {
    fetchWarehouses()
  }, [fetchWarehouses])

  useEffect(() => {
    fetchZones(page, selectedWarehouse)
  }, [page, selectedWarehouse, fetchZones])

  // ── Handlers ────────────────────────────────────────────────────────────

  const handleFilterChange = (wh: string) => {
    setSelectedWarehouse(wh)
    setPage(1)
  }

  const handleRefresh = () => {
    if (page === 1) {
      fetchZones(1, selectedWarehouse)
    } else {
      setPage(1)
    }
  }

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Create zone ─────────────────────────────────────────────────────────

  const openCreate = () => {
    createForm.resetFields()
    setCreateOpen(true)
  }

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      if (!selectedWarehouse) {
        message.error(t('zone.pleaseSelectWarehouse'))
        return
      }
      setCreateLoading(true)
      await client.post(`/warehouses/${selectedWarehouse}/zones`, values)
      message.success(t('zone.zoneCreated'))
      setCreateOpen(false)
      fetchZones(page, selectedWarehouse)
    } catch {
      if (createOpen) message.error(t('zone.createFailed'))
    } finally {
      setCreateLoading(false)
    }
  }

  // ── Edit zone ───────────────────────────────────────────────────────────

  const openEdit = (zone: Zone) => {
    setEditZone(zone)
    editForm.setFieldsValue({
      name: zone.name,
      zone_type: zone.zone_type,
      status: zone.status,
    })
    setEditOpen(true)
  }

  const handleEdit = async () => {
    if (!editZone) return
    try {
      const values = await editForm.validateFields()
      setEditLoading(true)
      const payload: Record<string, string> = {}
      if (values.name && values.name !== editZone.name) payload.name = values.name
      if (values.zone_type && values.zone_type !== editZone.zone_type) payload.zone_type = values.zone_type
      if (values.status && values.status !== editZone.status) payload.status = values.status

      if (Object.keys(payload).length > 0) {
        await client.put(`/zones/${editZone.id}`, payload)
        message.success(t('zone.zoneUpdated'))
      }
      setEditOpen(false)
      fetchZones(page, selectedWarehouse)
    } catch {
      message.error(t('zone.updateFailed'))
    } finally {
      setEditLoading(false)
    }
  }

  // ── Stats ───────────────────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = { active: 0, inactive: 0, full: 0 }
    zones.forEach((z) => { counts[z.status] = (counts[z.status] ?? 0) + 1 })
    return { total, ...counts }
  }, [zones, total])

  // ── Table columns ───────────────────────────────────────────────────────

  const columns: ColumnsType<Zone> = useMemo(() => [
    {
      title: t('zone.zoneCode'),
      dataIndex: 'code',
      key: 'code',
      width: 160,
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
      render: (tt: string) => (
        <Tag color={zoneTypeColors[tt] ?? 'default'}>{zoneTypeLabels[tt] ?? tt}</Tag>
      ),
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={zoneStatusColors[status] ?? 'default'}>
          {zoneStatusLabels[status] ?? status}
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
      width: 100,
      render: (_: unknown, record: Zone) => (
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
  ], [t, zoneTypeLabels, zoneStatusLabels])

  // ── Render ──────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('zone.title')}</Typography.Title>
        <Typography.Text type="secondary">{t('zone.subtitle')}</Typography.Text>
      </div>

      {/* ── Stat cards ──────────────────────────────────────────────────── */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic title={t('common.total')} value={stats.total}
              prefix={<AppstoreOutlined />}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic title={t('zone.statusActive')} value={stats.active}
              valueStyle={{ color: '#52c41a' }}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic title={t('zone.statusInactive')} value={stats.inactive}
              valueStyle={{ color: '#fa8c16' }}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic title={t('zone.statusFull')} value={stats.full}
              valueStyle={{ color: '#ff4d4f' }}
              formatter={(v) => Number(v).toLocaleString()} />
          </Card>
        </Col>
      </Row>

      {/* ── Table card ──────────────────────────────────────────────────── */}
      <Card
        title={
          <Space>
            <AppstoreOutlined />
            <span>{t('zone.allZones')}</span>
          </Space>
        }
        extra={
          <Space>
            <FilterOutlined />
            <Select
              placeholder={t('zone.filterByWarehouse')}
              allowClear
              style={{ width: 200 }}
              value={selectedWarehouse || undefined}
              onChange={(v) => handleFilterChange(v ?? '')}
              options={warehouses.map((w) => ({
                label: `${w.name} (${w.code})`,
                value: w.id,
              }))}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}
              disabled={!selectedWarehouse}>
              {t('zone.newZone')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('refresh')}
            </Button>
          </Space>
        }
      >
        <Table<Zone>
          columns={columns}
          dataSource={zones}
          rowKey="id"
          loading={loading}
          onChange={handleTableChange}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total,
            showSizeChanger: false,
            showTotal: (t2, range) => `${range[0]}-${range[1]} of ${t2}`,
          }}
          locale={{
            emptyText: (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t('zone.noZones')} />
            ),
          }}
        />
      </Card>

      {/* ── Create Zone Modal ───────────────────────────────────────────── */}
      <Modal
        title={t('zone.createZone')}
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        confirmLoading={createLoading}
        destroyOnClose
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
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

      {/* ── Edit Zone Modal ─────────────────────────────────────────────── */}
      <Modal
        title={t('zone.editZone')}
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleEdit}
        confirmLoading={editLoading}
        destroyOnClose
      >
        {editZone && (
          <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
            <Form.Item label={t('zone.zoneCode')}>
              <Input value={editZone.code} disabled />
            </Form.Item>
            <Form.Item
              name="name"
              label={t('zone.zoneName')}
              rules={[{ required: true, message: t('zone.pleaseEnterName') }]}
            >
              <Input placeholder={t('zone.namePlaceholder')} />
            </Form.Item>
            <Form.Item name="zone_type" label={t('zone.zoneType')}>
              <Select placeholder={t('zone.pleaseSelectType')}>
                <Select.Option value="receiving">{t('zone.typeReceiving')}</Select.Option>
                <Select.Option value="storage">{t('zone.typeStorage')}</Select.Option>
                <Select.Option value="picking">{t('zone.typePicking')}</Select.Option>
                <Select.Option value="shipping">{t('zone.typeShipping')}</Select.Option>
                <Select.Option value="returns">{t('zone.typeReturns')}</Select.Option>
                <Select.Option value="staging">{t('zone.typeStaging')}</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="status" label={t('common.status')}>
              <Select>
                <Select.Option value="active">{t('zone.statusActive')}</Select.Option>
                <Select.Option value="inactive">{t('zone.statusInactive')}</Select.Option>
                <Select.Option value="full">{t('zone.statusFull')}</Select.Option>
              </Select>
            </Form.Item>
          </Form>
        )}
      </Modal>
    </div>
  )
}
