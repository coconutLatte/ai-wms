// Wave management page with list, filtering, detail view, status transitions,
// order management, and wave creation.
// Uses GET/POST/PUT/DELETE /api/v1/waves endpoints.
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
  List,
  Popover,
} from 'antd'
import {
  ThunderboltOutlined,
  FilterOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  PlusOutlined,
  SendOutlined,
  UnorderedListOutlined,
  AppstoreAddOutlined,
  DeleteOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  Wave,
  ListResponse,
  CreateWaveRequest,
  AddWaveOrdersRequest,
  RemoveWaveOrdersRequest,
} from '@/api/types'

// ── Status / Type tag colors ──────────────────────────────────────────────────

const waveStatusColors: Record<string, string> = {
  created: 'default',
  released: 'blue',
  in_progress: 'processing',
  completed: 'success',
}

const waveTypeColors: Record<string, string> = {
  single_order: 'green',
  batch: 'blue',
  zone: 'orange',
  carrier: 'purple',
}

// ── Valid status transitions (from domain wave state machine) ─────────────────

const statusTransitions: Record<string, string[]> = {
  created: ['released'],
  released: ['in_progress'],
  in_progress: ['completed'],
}

// ── Status icons ──────────────────────────────────────────────────────────────

function statusIcon(s: string) {
  if (s === 'completed') return <CheckCircleOutlined />
  if (s === 'in_progress') return <SyncOutlined spin />
  return null
}

// ── Helper ────────────────────────────────────────────────────────────────────

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// ── Constants ─────────────────────────────────────────────────────────────────

const PAGE_SIZE = 10

// ── Main component ────────────────────────────────────────────────────────────

export default function WavesPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ────────────────────────────────────────────────────────

  const typeLabels: Record<string, string> = useMemo(
    () => ({
      single_order: t('wave.single_order'),
      batch: t('wave.batch'),
      zone: t('wave.zone'),
      carrier: t('wave.carrier'),
    }),
    [t],
  )

  const statusLabels: Record<string, string> = useMemo(
    () => ({
      created: t('wave.created'),
      released: t('wave.released'),
      in_progress: t('wave.in_progress'),
      completed: t('wave.completed'),
    }),
    [t],
  )

  // ── List state ──────────────────────────────────────────────────────────────

  const [waves, setWaves] = useState<Wave[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // ── Filters ─────────────────────────────────────────────────────────────────

  const [filterType, setFilterType] = useState<string>('')
  const [filterStatus, setFilterStatus] = useState<string>('')

  // ── Detail drawer ───────────────────────────────────────────────────────────

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedWave, setSelectedWave] = useState<Wave | null>(null)
  const [drawerLoading, setDrawerLoading] = useState(false)

  // ── Create modal ────────────────────────────────────────────────────────────

  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createWaveType, setCreateWaveType] = useState('batch')
  const [createWarehouseId, setCreateWarehouseId] = useState('')
  const [createOrderIds, setCreateOrderIds] = useState('')
  const [createLoading, setCreateLoading] = useState(false)

  // ── Add/Remove orders modals ────────────────────────────────────────────────

  const [addOrdersModalOpen, setAddOrdersModalOpen] = useState(false)
  const [addOrderIds, setAddOrderIds] = useState('')
  const [addOrdersLoading, setAddOrdersLoading] = useState(false)

  const [removeOrdersModalOpen, setRemoveOrdersModalOpen] = useState(false)
  const [removeOrderIds, setRemoveOrderIds] = useState('')
  const [removeOrdersLoading, setRemoveOrdersLoading] = useState(false)

  // ── Data fetching ───────────────────────────────────────────────────────────

  const fetchWaves = useCallback(
    async (p: number, wtype: string, status: string) => {
      setLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: p,
          page_size: PAGE_SIZE,
        }
        if (wtype) params.wave_type = wtype
        if (status) params.status = status

        const { data } = await client.get<ListResponse<Wave>>('/waves', {
          params,
        })
        setWaves(data.data)
        setTotal(data.pagination.total)
      } catch {
        message.error(t('wave.loadFailed'))
      } finally {
        setLoading(false)
      }
    },
    [message, t],
  )

  useEffect(() => {
    fetchWaves(page, filterType, filterStatus)
  }, [page, filterType, filterStatus, fetchWaves])

  // ── Refresh ─────────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    if (page === 1) {
      fetchWaves(1, filterType, filterStatus)
    } else {
      setPage(1)
    }
  }

  const handleFilterChange = (wtype: string, status: string) => {
    setFilterType(wtype)
    setFilterStatus(status)
    setPage(1)
  }

  // ── Table change handler ────────────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Detail drawer ───────────────────────────────────────────────────────────

  const openDetail = async (waveId: string) => {
    setDrawerOpen(true)
    setDrawerLoading(true)
    try {
      const { data } = await client.get<Wave>(`/waves/${waveId}`)
      setSelectedWave(data)
    } catch {
      message.error(t('wave.loadDetailFailed'))
    } finally {
      setDrawerLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerOpen(false)
    setSelectedWave(null)
  }

  // ── Status transition ───────────────────────────────────────────────────────

  const handleStatusTransition = async (wave: Wave, newStatus: string) => {
    try {
      await client.put(`/waves/${wave.id}/status`, {
        status: newStatus,
      })
      message.success(
        t('wave.statusUpdated', {
          status: statusLabels[newStatus] ?? titleCase(newStatus),
        }),
      )
      const { data } = await client.get<Wave>(`/waves/${wave.id}`)
      setSelectedWave(data)
      fetchWaves(page, filterType, filterStatus)
    } catch {
      message.error(t('wave.statusUpdateFailed'))
    }
  }

  // ── Release wave ────────────────────────────────────────────────────────────

  const handleRelease = async (wave: Wave) => {
    try {
      await client.post(`/waves/${wave.id}/release`)
      message.success(t('wave.releaseSuccess', { name: wave.wave_no }))
      const { data } = await client.get<Wave>(`/waves/${wave.id}`)
      setSelectedWave(data)
      fetchWaves(page, filterType, filterStatus)
    } catch {
      message.error(t('wave.releaseFailed'))
    }
  }

  // ── Create wave ─────────────────────────────────────────────────────────────

  const handleCreate = async () => {
    if (!createWarehouseId.trim()) {
      message.warning(t('wave.pleaseEnterWarehouse'))
      return
    }
    if (!createWaveType) {
      message.warning(t('wave.pleaseSelectWaveType'))
      return
    }
    setCreateLoading(true)
    try {
      const orderIds = createOrderIds
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean)

      await client.post('/waves', {
        wave_type: createWaveType,
        warehouse_id: createWarehouseId.trim(),
        order_ids: orderIds.length > 0 ? orderIds : undefined,
      } satisfies CreateWaveRequest)

      message.success(t('wave.createSuccess'))
      setCreateModalOpen(false)
      setCreateWarehouseId('')
      setCreateOrderIds('')
      setCreateWaveType('batch')
      fetchWaves(1, filterType, filterStatus)
    } catch {
      message.error(t('wave.createFailed'))
    } finally {
      setCreateLoading(false)
    }
  }

  // ── Add orders ──────────────────────────────────────────────────────────────

  const handleAddOrders = async () => {
    if (!selectedWave || !addOrderIds.trim()) return
    setAddOrdersLoading(true)
    try {
      const orderIds = addOrderIds
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean)

      await client.post(`/waves/${selectedWave.id}/orders`, {
        order_ids: orderIds,
      } satisfies AddWaveOrdersRequest)

      message.success(t('wave.addSuccess'))
      setAddOrdersModalOpen(false)
      setAddOrderIds('')
      const { data } = await client.get<Wave>(`/waves/${selectedWave.id}`)
      setSelectedWave(data)
      fetchWaves(page, filterType, filterStatus)
    } catch {
      message.error(t('wave.addFailed'))
    } finally {
      setAddOrdersLoading(false)
    }
  }

  // ── Remove orders ───────────────────────────────────────────────────────────

  const handleRemoveOrders = async () => {
    if (!selectedWave || !removeOrderIds.trim()) return
    setRemoveOrdersLoading(true)
    try {
      const orderIds = removeOrderIds
        .split('\n')
        .map((s) => s.trim())
        .filter(Boolean)

      await client.delete(`/waves/${selectedWave.id}/orders`, {
        data: { order_ids: orderIds } satisfies RemoveWaveOrdersRequest,
      })

      message.success(t('wave.removeSuccess'))
      setRemoveOrdersModalOpen(false)
      setRemoveOrderIds('')
      const { data } = await client.get<Wave>(`/waves/${selectedWave.id}`)
      setSelectedWave(data)
      fetchWaves(page, filterType, filterStatus)
    } catch {
      message.error(t('wave.removeFailed'))
    } finally {
      setRemoveOrdersLoading(false)
    }
  }

  // ── Table columns ───────────────────────────────────────────────────────────

  const columns: ColumnsType<Wave> = useMemo(
    () => [
      {
        title: t('wave.waveNo'),
        dataIndex: 'wave_no',
        key: 'wave_no',
        width: 220,
        render: (no: string) => (
          <Typography.Text code style={{ cursor: 'pointer' }}>
            {no}
          </Typography.Text>
        ),
      },
      {
        title: t('wave.waveType'),
        dataIndex: 'wave_type',
        key: 'wave_type',
        width: 120,
        render: (wt: string) => (
          <Tag color={waveTypeColors[wt] ?? 'default'}>
            {typeLabels[wt] ?? titleCase(wt)}
          </Tag>
        ),
      },
      {
        title: t('common.status'),
        dataIndex: 'status',
        key: 'status',
        width: 130,
        render: (s: string) => (
          <Tag icon={statusIcon(s)} color={waveStatusColors[s] ?? 'default'}>
            {statusLabels[s] ?? titleCase(s)}
          </Tag>
        ),
      },
      {
        title: t('wave.totalOrders'),
        dataIndex: 'total_orders',
        key: 'total_orders',
        width: 100,
        align: 'right',
        render: (v: number) => v.toLocaleString(),
      },
      {
        title: t('wave.totalQty'),
        dataIndex: 'total_qty',
        key: 'total_qty',
        width: 100,
        align: 'right',
        responsive: ['md'],
        render: (v: number) => v.toLocaleString(),
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
        render: (_: unknown, record: Wave) => (
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
    [t, typeLabels, statusLabels],
  )

  // ── Wave summary stats ──────────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = {}
    waves.forEach((w) => {
      counts[w.status] = (counts[w.status] ?? 0) + 1
    })
    return {
      total: total,
      created: counts['created'] ?? 0,
      released: counts['released'] ?? 0,
      in_progress: counts['in_progress'] ?? 0,
      completed: counts['completed'] ?? 0,
    }
  }, [waves, total])

  // ── Render ──────────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('wave.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('wave.subtitle')}
        </Typography.Text>
      </div>

      {/* ── Summary stat cards ──────────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('common.total')}
              value={stats.total}
              prefix={<ThunderboltOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('wave.created')}
              value={stats.created}
              prefix={<UnorderedListOutlined />}
              valueStyle={{
                color: stats.created > 0 ? '#8c8c8c' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('wave.in_progress')}
              value={stats.in_progress}
              prefix={<SyncOutlined spin={stats.in_progress > 0} />}
              valueStyle={{
                color: stats.in_progress > 0 ? '#1677ff' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('wave.completed')}
              value={stats.completed}
              prefix={<CheckCircleOutlined />}
              valueStyle={{
                color: stats.completed > 0 ? '#52c41a' : undefined,
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
            <ThunderboltOutlined />
            <span>{t('wave.allWaves')}</span>
          </Space>
        }
        extra={
          <Space wrap>
            <FilterOutlined />
            <Select
              placeholder={t('wave.filterType')}
              allowClear
              style={{ width: 130 }}
              value={filterType || undefined}
              onChange={(v) =>
                handleFilterChange(v ?? '', filterStatus)
              }
              options={[
                { label: t('wave.single_order'), value: 'single_order' },
                { label: t('wave.batch'), value: 'batch' },
                { label: t('wave.zone'), value: 'zone' },
                { label: t('wave.carrier'), value: 'carrier' },
              ]}
            />
            <Select
              placeholder={t('wave.filterStatus')}
              allowClear
              style={{ width: 130 }}
              value={filterStatus || undefined}
              onChange={(v) =>
                handleFilterChange(filterType, v ?? '')
              }
              options={[
                { label: t('wave.created'), value: 'created' },
                { label: t('wave.released'), value: 'released' },
                { label: t('wave.in_progress'), value: 'in_progress' },
                { label: t('wave.completed'), value: 'completed' },
              ]}
            />
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setCreateWarehouseId('')
                setCreateOrderIds('')
                setCreateWaveType('batch')
                setCreateModalOpen(true)
              }}
            >
              {t('wave.newWave')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('wave.refresh')}
            </Button>
          </Space>
        }
      >
        <Table<Wave>
          columns={columns}
          dataSource={waves}
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
                description={t('wave.noWaves')}
              />
            ),
          }}
        />
      </Card>

      {/* ── Detail Drawer ─────────────────────────────────────────────────────── */}

      <Drawer
        title={
          selectedWave ? (
            <Space>
              <ThunderboltOutlined />
              <span>{t('wave.waveDetail')}</span>
              <Typography.Text code>{selectedWave.wave_no}</Typography.Text>
            </Space>
          ) : (
            t('wave.waveDetail')
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
          <Empty description={t('wave.loadingDetail')} />
        ) : selectedWave ? (
          <>
            {/* ── Wave info ─────────────────────────────────────────────────────── */}

            <Descriptions
              bordered
              column={{ xs: 1, sm: 2 }}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Descriptions.Item label={t('wave.waveNo')}>
                <Typography.Text code>{selectedWave.wave_no}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('wave.waveType')}>
                <Tag color={waveTypeColors[selectedWave.wave_type] ?? 'default'}>
                  {typeLabels[selectedWave.wave_type] ??
                    titleCase(selectedWave.wave_type)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('common.status')}>
                <Tag
                  icon={statusIcon(selectedWave.status)}
                  color={waveStatusColors[selectedWave.status] ?? 'default'}
                >
                  {statusLabels[selectedWave.status] ??
                    titleCase(selectedWave.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('wave.warehouseId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedWave.warehouse_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('wave.totalOrders')}>
                {selectedWave.total_orders.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('wave.totalLines')}>
                {selectedWave.total_lines.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('wave.totalQty')}>
                {selectedWave.total_qty.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('common.created')}>
                {new Date(selectedWave.created_at).toLocaleString()}
              </Descriptions.Item>
              {selectedWave.released_at && (
                <Descriptions.Item label={t('wave.released')}>
                  {new Date(selectedWave.released_at).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedWave.completed_at && (
                <Descriptions.Item label={t('wave.completed')}>
                  {new Date(selectedWave.completed_at).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>

            {/* ── Status transitions ─────────────────────────────────────────────── */}

            {statusTransitions[selectedWave.status] &&
              statusTransitions[selectedWave.status].length > 0 && (
                <Card
                  size="small"
                  title={t('wave.statusActions')}
                  style={{ marginBottom: 24 }}
                >
                  <Space wrap>
                    {statusTransitions[selectedWave.status].map((target) => {
                      if (target === 'released') {
                        return (
                          <Popconfirm
                            key={target}
                            title={t('wave.releaseConfirmTitle')}
                            description={t('wave.releaseConfirmMsg', {
                              name: selectedWave.wave_no,
                            })}
                            onConfirm={() => handleRelease(selectedWave)}
                            okText={t('common.yes')}
                            cancelText={t('common.no')}
                          >
                            <Button
                              type="primary"
                              size="small"
                              icon={<SendOutlined />}
                            >
                              {t('wave.releaseWave')}
                            </Button>
                          </Popconfirm>
                        )
                      }
                      return (
                        <Popconfirm
                          key={target}
                          title={t('wave.moveTo', {
                            target:
                              statusLabels[target] ?? titleCase(target),
                          })}
                          onConfirm={() =>
                            handleStatusTransition(selectedWave, target)
                          }
                          okText={t('common.yes')}
                          cancelText={t('common.no')}
                        >
                          <Button
                            type="primary"
                            size="small"
                            icon={
                              target === 'completed' ? (
                                <CheckCircleOutlined />
                              ) : target === 'in_progress' ? (
                                <SyncOutlined />
                              ) : undefined
                            }
                          >
                            {statusLabels[target] ?? titleCase(target)}
                          </Button>
                        </Popconfirm>
                      )
                    })}
                  </Space>
                </Card>
              )}

            {/* ── Order management (only for created waves) ──────────────────────── */}

            <Card
              size="small"
              title={t('wave.orderList', {
                count: selectedWave.order_ids.length,
              })}
              style={{ marginBottom: 24 }}
              extra={
                selectedWave.status === 'created' ? (
                  <Space>
                    <Button
                      size="small"
                      icon={<AppstoreAddOutlined />}
                      onClick={() => {
                        setAddOrderIds('')
                        setAddOrdersModalOpen(true)
                      }}
                    >
                      {t('wave.addOrders')}
                    </Button>
                    <Button
                      size="small"
                      danger
                      icon={<DeleteOutlined />}
                      disabled={selectedWave.order_ids.length === 0}
                      onClick={() => {
                        setRemoveOrderIds('')
                        setRemoveOrdersModalOpen(true)
                      }}
                    >
                      {t('wave.removeOrders')}
                    </Button>
                  </Space>
                ) : (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    {t('wave.cannotModify')}
                  </Typography.Text>
                )
              }
            >
              {selectedWave.order_ids.length === 0 ? (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description={t('wave.noOrders')}
                />
              ) : (
                <List
                  size="small"
                  dataSource={selectedWave.order_ids}
                  renderItem={(id: string) => (
                    <List.Item>
                      <Typography.Text code style={{ fontSize: 12 }}>
                        {id}
                      </Typography.Text>
                    </List.Item>
                  )}
                />
              )}
            </Card>

            {/* ── Task list ──────────────────────────────────────────────────────── */}

            <Card
              size="small"
              title={t('wave.taskList', {
                count: selectedWave.task_ids.length,
              })}
              style={{ marginBottom: 24 }}
            >
              {selectedWave.task_ids.length === 0 ? (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description={t('wave.noTasks')}
                />
              ) : (
                <List
                  size="small"
                  dataSource={selectedWave.task_ids}
                  renderItem={(id: string) => (
                    <List.Item>
                      <Typography.Text code style={{ fontSize: 12 }}>
                        {id}
                      </Typography.Text>
                    </List.Item>
                  )}
                />
              )}
            </Card>
          </>
        ) : (
          <Empty description={t('wave.noWaveSelected')} />
        )}
      </Drawer>

      {/* ── Create Wave Modal ─────────────────────────────────────────────────── */}

      <Modal
        title={t('wave.createWaveTitle')}
        open={createModalOpen}
        onOk={handleCreate}
        onCancel={() => {
          setCreateModalOpen(false)
          setCreateWarehouseId('')
          setCreateOrderIds('')
          setCreateWaveType('batch')
        }}
        confirmLoading={createLoading}
        okText={t('wave.createWave')}
        cancelText={t('common.cancel')}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <Typography.Text strong>{t('wave.waveType')}</Typography.Text>
            <Select
              style={{ width: '100%', marginTop: 4 }}
              value={createWaveType}
              onChange={(v) => setCreateWaveType(v)}
              options={[
                { label: t('wave.single_order'), value: 'single_order' },
                { label: t('wave.batch'), value: 'batch' },
                { label: t('wave.zone'), value: 'zone' },
                { label: t('wave.carrier'), value: 'carrier' },
              ]}
            />
          </div>
          <div>
            <Typography.Text strong>
              {t('wave.warehouseId')}
            </Typography.Text>
            <Input
              style={{ marginTop: 4 }}
              placeholder={t('wave.warehousePlaceholder')}
              value={createWarehouseId}
              onChange={(e) => setCreateWarehouseId(e.target.value)}
            />
          </div>
          <div>
            <Typography.Text strong>
              {t('wave.orderIds')}
              <Typography.Text type="secondary">
                {' '}
                ({t('common.cancel').toLowerCase()} —{' '}
                {t('wave.orderIdsHint')})
              </Typography.Text>
            </Typography.Text>
            <Input.TextArea
              style={{ marginTop: 4 }}
              rows={4}
              placeholder={t('wave.orderIdsPlaceholder')}
              value={createOrderIds}
              onChange={(e) => setCreateOrderIds(e.target.value)}
            />
          </div>
        </Space>
      </Modal>

      {/* ── Add Orders Modal ──────────────────────────────────────────────────── */}

      <Modal
        title={t('wave.addOrdersTitle')}
        open={addOrdersModalOpen}
        onOk={handleAddOrders}
        onCancel={() => {
          setAddOrdersModalOpen(false)
          setAddOrderIds('')
        }}
        confirmLoading={addOrdersLoading}
        okText={t('wave.addOrders')}
        cancelText={t('common.cancel')}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Typography.Text type="secondary">
            {t('wave.orderIdsHint')}
          </Typography.Text>
          <Input.TextArea
            rows={6}
            placeholder={t('wave.orderIdsPlaceholder')}
            value={addOrderIds}
            onChange={(e) => setAddOrderIds(e.target.value)}
          />
        </Space>
      </Modal>

      {/* ── Remove Orders Modal ───────────────────────────────────────────────── */}

      <Modal
        title={t('wave.removeOrdersTitle')}
        open={removeOrdersModalOpen}
        onOk={handleRemoveOrders}
        onCancel={() => {
          setRemoveOrdersModalOpen(false)
          setRemoveOrderIds('')
        }}
        confirmLoading={removeOrdersLoading}
        okText={t('wave.removeOrders')}
        cancelText={t('common.cancel')}
        okButtonProps={{ danger: true }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Typography.Text type="secondary">
            {t('wave.orderIdsHint')}
          </Typography.Text>
          <Input.TextArea
            rows={6}
            placeholder={t('wave.orderIdsPlaceholder')}
            value={removeOrderIds}
            onChange={(e) => setRemoveOrderIds(e.target.value)}
          />
        </Space>
      </Modal>
    </div>
  )
}
