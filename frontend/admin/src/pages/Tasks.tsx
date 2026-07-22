// Task management page with list, filtering, detail view, assignment, and status transitions.
// Uses GET /api/v1/tasks for real data.
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
  Input,
  Modal,
  InputNumber,
  Form,
} from 'antd'
import {
  CarryOutOutlined,
  FilterOutlined,
  ReloadOutlined,
  InboxOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  PauseCircleOutlined,
  UserAddOutlined,
  FileTextOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  Task,
  ListResponse,
  AssignTaskRequest,
  UpdateTaskStatusRequest,
  CompleteTaskRequest,
} from '@/api/types'

// ── Status / Type / Priority tag colors ────────────────────────────────────

const taskStatusColors: Record<string, string> = {
  pending: 'default',
  assigned: 'blue',
  in_progress: 'processing',
  paused: 'warning',
  completed: 'success',
  cancelled: 'error',
  exception: 'error',
}

const taskTypeColors: Record<string, string> = {
  putaway: 'green',
  pick: 'blue',
  replenish: 'orange',
  transfer: 'purple',
  cycle_count: 'cyan',
  load: 'magenta',
  unload: 'red',
}

const taskPriorityColors: Record<string, string> = {
  low: 'default',
  normal: 'blue',
  high: 'orange',
  urgent: 'red',
}

// ── Status label helper ────────────────────────────────────────────────────

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// ── Valid status transitions (from domain task state machine) ──────────────

const statusTransitions: Record<string, string[]> = {
  pending: ['assigned', 'cancelled'],
  assigned: ['in_progress', 'cancelled'],
  in_progress: ['completed', 'paused', 'cancelled'],
  paused: ['in_progress', 'cancelled'],
  exception: ['in_progress', 'cancelled'],
}

// ── Constants ──────────────────────────────────────────────────────────────

const PAGE_SIZE = 10

// ── Main component ─────────────────────────────────────────────────────────

export default function TasksPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Localized labels ─────────────────────────────────────────────────────

  const typeLabels: Record<string, string> = useMemo(
    () => ({
      putaway: t('task.putaway'),
      pick: t('task.pick'),
      replenish: t('task.replenish'),
      transfer: t('task.transfer'),
      cycle_count: t('task.cycle_count'),
      load: t('task.load'),
      unload: t('task.unload'),
    }),
    [t],
  )

  const statusLabels: Record<string, string> = useMemo(
    () => ({
      pending: t('task.pending'),
      assigned: t('task.assigned'),
      in_progress: t('task.in_progress'),
      paused: t('task.paused'),
      completed: t('task.completed'),
      cancelled: t('task.cancelled'),
      exception: t('task.exception'),
    }),
    [t],
  )

  const priorityLabels: Record<string, string> = useMemo(
    () => ({
      low: t('task.low'),
      normal: t('task.normal'),
      high: t('task.high'),
      urgent: t('task.urgent'),
    }),
    [t],
  )

  // ── List state ───────────────────────────────────────────────────────────

  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // ── Filters ──────────────────────────────────────────────────────────────

  const [filterType, setFilterType] = useState<string>('')
  const [filterStatus, setFilterStatus] = useState<string>('')
  const [filterAssignee, setFilterAssignee] = useState<string>('')

  // ── Detail drawer ────────────────────────────────────────────────────────

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [drawerLoading, setDrawerLoading] = useState(false)

  // ── Assign modal ─────────────────────────────────────────────────────────

  const [assignModalOpen, setAssignModalOpen] = useState(false)
  const [assignWorkerId, setAssignWorkerId] = useState('')
  const [assignLoading, setAssignLoading] = useState(false)

  // ── Complete modal ───────────────────────────────────────────────────────

  const [completeModalOpen, setCompleteModalOpen] = useState(false)
  const [completeForm] = Form.useForm()
  const [completeLoading, setCompleteLoading] = useState(false)

  // ── Data fetching ────────────────────────────────────────────────────────

  const fetchTasks = useCallback(
    async (p: number, type: string, status: string, assignee: string) => {
      setLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: p,
          page_size: PAGE_SIZE,
        }
        if (type) params.task_type = type
        if (status) params.status = status
        if (assignee) params.assigned_to = assignee

        const { data } = await client.get<ListResponse<Task>>('/tasks', {
          params,
        })
        setTasks(data.data)
        setTotal(data.pagination.total)
      } catch {
        message.error(t('task.loadFailed'))
      } finally {
        setLoading(false)
      }
    },
    [message, t],
  )

  useEffect(() => {
    fetchTasks(page, filterType, filterStatus, filterAssignee)
  }, [page, filterType, filterStatus, filterAssignee, fetchTasks])

  // ── Refresh ──────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    if (page === 1) {
      fetchTasks(1, filterType, filterStatus, filterAssignee)
    } else {
      setPage(1)
    }
  }

  const handleFilterChange = (type: string, status: string, assignee: string) => {
    setFilterType(type)
    setFilterStatus(status)
    setFilterAssignee(assignee)
    setPage(1)
  }

  // ── Table change handler ─────────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Detail drawer ────────────────────────────────────────────────────────

  const openDetail = async (taskId: string) => {
    setDrawerOpen(true)
    setDrawerLoading(true)
    try {
      const { data } = await client.get<Task>(`/tasks/${taskId}`)
      setSelectedTask(data)
    } catch {
      message.error(t('task.loadDetailFailed'))
    } finally {
      setDrawerLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerOpen(false)
    setSelectedTask(null)
  }

  // ── Status transition ────────────────────────────────────────────────────

  const handleStatusTransition = async (task: Task, newStatus: string) => {
    try {
      await client.put(`/tasks/${task.id}/status`, {
        status: newStatus,
      } satisfies UpdateTaskStatusRequest)
      message.success(
        t('task.statusUpdated', {
          status: statusLabels[newStatus] ?? titleCase(newStatus),
        }),
      )
      const { data } = await client.get<Task>(`/tasks/${task.id}`)
      setSelectedTask(data)
      fetchTasks(page, filterType, filterStatus, filterAssignee)
    } catch {
      message.error(t('task.statusUpdateFailed'))
    }
  }

  // ── Assign worker ────────────────────────────────────────────────────────

  const handleAssign = async () => {
    if (!selectedTask || !assignWorkerId.trim()) return
    setAssignLoading(true)
    try {
      await client.post(`/tasks/${selectedTask.id}/assign`, {
        assigned_to: assignWorkerId.trim(),
      } satisfies AssignTaskRequest)
      message.success(
        t('task.assignSuccess', { worker: assignWorkerId.trim() }),
      )
      setAssignModalOpen(false)
      setAssignWorkerId('')
      const { data } = await client.get<Task>(`/tasks/${selectedTask.id}`)
      setSelectedTask(data)
      fetchTasks(page, filterType, filterStatus, filterAssignee)
    } catch {
      message.error(t('task.assignFailed'))
    } finally {
      setAssignLoading(false)
    }
  }

  // ── Complete task ────────────────────────────────────────────────────────

  const handleComplete = async () => {
    if (!selectedTask) return
    try {
      const values = await completeForm.validateFields()
      setCompleteLoading(true)
      await client.post(`/tasks/${selectedTask.id}/complete`, {
        actual_qty: values.actual_qty,
        batch_no: values.batch_no || undefined,
      } satisfies CompleteTaskRequest)
      message.success(
        t('task.completeSuccess', { qty: values.actual_qty }),
      )
      setCompleteModalOpen(false)
      completeForm.resetFields()
      const { data } = await client.get<Task>(`/tasks/${selectedTask.id}`)
      setSelectedTask(data)
      fetchTasks(page, filterType, filterStatus, filterAssignee)
    } catch {
      message.error(t('task.completeFailed'))
    } finally {
      setCompleteLoading(false)
    }
  }

  // ── Table columns ────────────────────────────────────────────────────────

  const columns: ColumnsType<Task> = useMemo(
    () => [
      {
        title: t('task.taskNo'),
        dataIndex: 'task_no',
        key: 'task_no',
        width: 190,
        render: (no: string) => (
          <Typography.Text code style={{ cursor: 'pointer' }}>
            {no}
          </Typography.Text>
        ),
      },
      {
        title: t('task.taskType'),
        dataIndex: 'task_type',
        key: 'task_type',
        width: 110,
        render: (tt: string) => (
          <Tag color={taskTypeColors[tt] ?? 'default'}>
            {typeLabels[tt] ?? titleCase(tt)}
          </Tag>
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
            ) : s === 'in_progress' ? (
              <SyncOutlined spin />
            ) : s === 'cancelled' ? (
              <CloseCircleOutlined />
            ) : s === 'paused' ? (
              <PauseCircleOutlined />
            ) : s === 'exception' ? (
              <ExclamationCircleOutlined />
            ) : null
          return (
            <Tag icon={icon} color={taskStatusColors[s] ?? 'default'}>
              {statusLabels[s] ?? titleCase(s)}
            </Tag>
          )
        },
      },
      {
        title: t('task.priority'),
        dataIndex: 'priority',
        key: 'priority',
        width: 100,
        responsive: ['md'],
        render: (p: string) => (
          <Tag color={taskPriorityColors[p] ?? 'default'}>
            {priorityLabels[p] ?? titleCase(p)}
          </Tag>
        ),
      },
      {
        title: t('task.skuId'),
        dataIndex: 'sku_id',
        key: 'sku_id',
        width: 140,
        responsive: ['lg'],
        ellipsis: true,
        render: (id: string) => (
          <Typography.Text code style={{ fontSize: 12 }}>
            {id}
          </Typography.Text>
        ),
      },
      {
        title: t('task.assignedTo'),
        dataIndex: 'assigned_to',
        key: 'assigned_to',
        width: 120,
        responsive: ['md'],
        render: (v: string) =>
          v || <Typography.Text type="secondary">—</Typography.Text>,
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
        render: (_: unknown, record: Task) => (
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
    [t, typeLabels, statusLabels, priorityLabels],
  )

  // ── Task summary stats ──────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = {}
    tasks.forEach((t2) => {
      counts[t2.status] = (counts[t2.status] ?? 0) + 1
    })
    return {
      total: total,
      pending: counts['pending'] ?? 0,
      assigned: counts['assigned'] ?? 0,
      in_progress: counts['in_progress'] ?? 0,
      paused: counts['paused'] ?? 0,
      completed: counts['completed'] ?? 0,
      cancelled: counts['cancelled'] ?? 0,
      exception: counts['exception'] ?? 0,
    }
  }, [tasks, total])

  // ── Render ───────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('task.title')}</Typography.Title>
        <Typography.Text type="secondary">
          {t('task.subtitle')}
        </Typography.Text>
      </div>

      {/* ── Summary stat cards ───────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('common.total')}
              value={stats.total}
              prefix={<CarryOutOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('task.pending')}
              value={stats.pending}
              prefix={<InboxOutlined />}
              valueStyle={{
                color: stats.pending > 0 ? '#8c8c8c' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('task.in_progress')}
              value={stats.in_progress}
              prefix={<SyncOutlined spin={stats.in_progress > 0} />}
              valueStyle={{
                color: stats.in_progress > 0 ? '#1677ff' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('task.completed')}
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
              title={t('task.paused')}
              value={stats.paused}
              prefix={<PauseCircleOutlined />}
              valueStyle={{
                color: stats.paused > 0 ? '#faad14' : undefined,
              }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={4}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('task.cancelled')}
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
            <CarryOutOutlined />
            <span>{t('task.allTasks')}</span>
          </Space>
        }
        extra={
          <Space wrap>
            <FilterOutlined />
            <Select
              placeholder={t('task.filterType')}
              allowClear
              style={{ width: 130 }}
              value={filterType || undefined}
              onChange={(v) =>
                handleFilterChange(v ?? '', filterStatus, filterAssignee)
              }
              options={[
                { label: t('task.putaway'), value: 'putaway' },
                { label: t('task.pick'), value: 'pick' },
                { label: t('task.replenish'), value: 'replenish' },
                { label: t('task.transfer'), value: 'transfer' },
                { label: t('task.cycle_count'), value: 'cycle_count' },
                { label: t('task.load'), value: 'load' },
                { label: t('task.unload'), value: 'unload' },
              ]}
            />
            <Select
              placeholder={t('task.filterStatus')}
              allowClear
              style={{ width: 130 }}
              value={filterStatus || undefined}
              onChange={(v) =>
                handleFilterChange(filterType, v ?? '', filterAssignee)
              }
              options={[
                { label: t('task.pending'), value: 'pending' },
                { label: t('task.assigned'), value: 'assigned' },
                { label: t('task.in_progress'), value: 'in_progress' },
                { label: t('task.paused'), value: 'paused' },
                { label: t('task.completed'), value: 'completed' },
                { label: t('task.cancelled'), value: 'cancelled' },
                { label: t('task.exception'), value: 'exception' },
              ]}
            />
            <Input
              placeholder={t('task.filterAssignee')}
              allowClear
              style={{ width: 150 }}
              value={filterAssignee}
              onChange={(e) =>
                handleFilterChange(filterType, filterStatus, e.target.value)
              }
            />
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('task.refresh')}
            </Button>
          </Space>
        }
      >
        <Table<Task>
          columns={columns}
          dataSource={tasks}
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
                description={t('task.noTasks')}
              />
            ),
          }}
        />
      </Card>

      {/* ── Detail Drawer ──────────────────────────────────────────────────── */}

      <Drawer
        title={
          selectedTask ? (
            <Space>
              <CarryOutOutlined />
              <span>{t('task.taskDetail')}</span>
              <Typography.Text code>{selectedTask.task_no}</Typography.Text>
            </Space>
          ) : (
            t('task.taskDetail')
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
          <Empty description={t('task.loadingDetail')} />
        ) : selectedTask ? (
          <>
            {/* ── Task info ──────────────────────────────────────────────────── */}

            <Descriptions
              bordered
              column={{ xs: 1, sm: 2 }}
              size="small"
              style={{ marginBottom: 24 }}
            >
              <Descriptions.Item label={t('task.taskNo')}>
                <Typography.Text code>{selectedTask.task_no}</Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('task.taskType')}>
                <Tag color={taskTypeColors[selectedTask.task_type] ?? 'default'}>
                  {typeLabels[selectedTask.task_type] ??
                    titleCase(selectedTask.task_type)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('common.status')}>
                <Tag
                  color={taskStatusColors[selectedTask.status] ?? 'default'}
                >
                  {statusLabels[selectedTask.status] ??
                    titleCase(selectedTask.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('task.priority')}>
                <Tag
                  color={taskPriorityColors[selectedTask.priority] ?? 'default'}
                >
                  {priorityLabels[selectedTask.priority] ??
                    titleCase(selectedTask.priority)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label={t('task.warehouseId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedTask.warehouse_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('task.skuId')}>
                <Typography.Text code style={{ fontSize: 12 }}>
                  {selectedTask.sku_id}
                </Typography.Text>
              </Descriptions.Item>
              <Descriptions.Item label={t('task.expectedQty')}>
                {selectedTask.expected_qty.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('task.actualQty')}>
                {selectedTask.actual_qty.toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label={t('task.uom')}>
                {selectedTask.uom}
              </Descriptions.Item>
              {selectedTask.batch_no && (
                <Descriptions.Item label={t('task.batchNo')}>
                  {selectedTask.batch_no}
                </Descriptions.Item>
              )}
              <Descriptions.Item label={t('task.assignedTo')}>
                {selectedTask.assigned_to || (
                  <Typography.Text type="secondary">—</Typography.Text>
                )}
              </Descriptions.Item>
              {selectedTask.order_id && (
                <Descriptions.Item label={t('task.orderId')}>
                  <Typography.Text code style={{ fontSize: 12 }}>
                    {selectedTask.order_id}
                  </Typography.Text>
                </Descriptions.Item>
              )}
              {selectedTask.order_line_id && (
                <Descriptions.Item label={t('task.orderLineId')}>
                  <Typography.Text code style={{ fontSize: 12 }}>
                    {selectedTask.order_line_id}
                  </Typography.Text>
                </Descriptions.Item>
              )}
              {selectedTask.from_location_id && (
                <Descriptions.Item label={t('task.fromLocation')}>
                  <Typography.Text code style={{ fontSize: 12 }}>
                    {selectedTask.from_location_id}
                  </Typography.Text>
                </Descriptions.Item>
              )}
              {selectedTask.to_location_id && (
                <Descriptions.Item label={t('task.toLocation')}>
                  <Typography.Text code style={{ fontSize: 12 }}>
                    {selectedTask.to_location_id}
                  </Typography.Text>
                </Descriptions.Item>
              )}
              {selectedTask.instructions && (
                <Descriptions.Item label={t('task.instructions')} span={2}>
                  {selectedTask.instructions}
                </Descriptions.Item>
              )}
              <Descriptions.Item label={t('common.created')}>
                {new Date(selectedTask.created_at).toLocaleString()}
              </Descriptions.Item>
              {selectedTask.started_at && (
                <Descriptions.Item label={t('task.startedAt')}>
                  {new Date(selectedTask.started_at).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedTask.completed_at && (
                <Descriptions.Item label={t('task.completedAt')}>
                  {new Date(selectedTask.completed_at).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedTask.cancelled_at && (
                <Descriptions.Item label={t('task.cancelledAt')}>
                  {new Date(selectedTask.cancelled_at).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>

            {/* ── Status transitions ──────────────────────────────────────────── */}

            {statusTransitions[selectedTask.status] &&
              statusTransitions[selectedTask.status].length > 0 && (
                <Card
                  size="small"
                  title={t('task.statusActions')}
                  style={{ marginBottom: 24 }}
                >
                  <Space wrap>
                    {statusTransitions[selectedTask.status].map((target) => (
                      <Popconfirm
                        key={target}
                        title={t('task.moveTo', {
                          target:
                            statusLabels[target] ?? titleCase(target),
                        })}
                        onConfirm={() =>
                          handleStatusTransition(selectedTask, target)
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
                            ) : target === 'paused' ? (
                              <PauseCircleOutlined />
                            ) : target === 'in_progress' ? (
                              <PlayCircleOutlined />
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

            {/* ── Assignment action ───────────────────────────────────────────── */}

            {selectedTask.status === 'pending' && (
              <Card
                size="small"
                title={t('task.assignmentActions')}
                style={{ marginBottom: 24 }}
              >
                <Space>
                  <Button
                    type="primary"
                    icon={<UserAddOutlined />}
                    onClick={() => {
                      setAssignWorkerId('')
                      setAssignModalOpen(true)
                    }}
                  >
                    {t('task.assignWorker')}
                  </Button>
                </Space>
              </Card>
            )}

            {/* ── Complete action ─────────────────────────────────────────────── */}

            {selectedTask.status === 'in_progress' && (
              <Card
                size="small"
                title={t('task.completeTaskTitle')}
                style={{ marginBottom: 24 }}
              >
                <Space>
                  <Button
                    type="primary"
                    icon={<CheckCircleOutlined />}
                    onClick={() => {
                      completeForm.resetFields()
                      completeForm.setFieldsValue({
                        actual_qty: selectedTask.expected_qty,
                      })
                      setCompleteModalOpen(true)
                    }}
                  >
                    {t('task.completeTask')}
                  </Button>
                </Space>
              </Card>
            )}
          </>
        ) : (
          <Empty description={t('task.noTaskSelected')} />
        )}
      </Drawer>

      {/* ── Assign Worker Modal ────────────────────────────────────────────── */}

      <Modal
        title={t('task.assignWorker')}
        open={assignModalOpen}
        onOk={handleAssign}
        onCancel={() => {
          setAssignModalOpen(false)
          setAssignWorkerId('')
        }}
        confirmLoading={assignLoading}
        okText={t('task.assign')}
        cancelText={t('common.cancel')}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Typography.Text>{t('task.assignWorkerPlaceholder')}</Typography.Text>
          <Input
            placeholder={t('task.assignWorkerPlaceholder')}
            value={assignWorkerId}
            onChange={(e) => setAssignWorkerId(e.target.value)}
            onPressEnter={handleAssign}
          />
        </Space>
      </Modal>

      {/* ── Complete Task Modal ────────────────────────────────────────────── */}

      <Modal
        title={t('task.completeTaskTitle')}
        open={completeModalOpen}
        onOk={handleComplete}
        onCancel={() => {
          setCompleteModalOpen(false)
          completeForm.resetFields()
        }}
        confirmLoading={completeLoading}
        okText={t('task.completeTask')}
        cancelText={t('common.cancel')}
      >
        <Form form={completeForm} layout="vertical">
          <Form.Item
            label={t('task.actualQtyLabel')}
            name="actual_qty"
            rules={[
              { required: true, message: t('task.pleaseEnterQty') },
              {
                type: 'number',
                min: 0,
                message: t('task.pleaseEnterQty'),
              },
            ]}
          >
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item label={t('task.batchNoLabel')} name="batch_no">
            <Input placeholder={t('task.batchNoLabel')} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
