// User management page with list, filtering, detail view, and CRUD operations.
// Uses GET/POST/PUT /api/v1/users for real data.
// All UI text is translated via react-i18next.

import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Typography,
  Table,
  Tag,
  Space,
  Button,
  App,
  Empty,
  Card,
  Row,
  Col,
  Statistic,
  Modal,
  Form,
  Input,
  Select,
  Popconfirm,
} from 'antd'
import {
  UserOutlined,
  TeamOutlined,
  ReloadOutlined,
  PlusOutlined,
  CheckCircleOutlined,
  StopOutlined,
  LockOutlined,
  EditOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  User,
  Role,
  ListResponse,
  CreateUserRequest,
  UpdateUserRequest,
  UpdateUserStatusRequest,
} from '@/api/types'

// ── Status tag colors ────────────────────────────────────────────────────────

const userStatusColors: Record<string, string> = {
  active: 'success',
  inactive: 'default',
  locked: 'error',
}

const userStatusIcons: Record<string, React.ReactNode> = {
  active: <CheckCircleOutlined />,
  inactive: <StopOutlined />,
  locked: <LockOutlined />,
}

function titleCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// ── Constants ────────────────────────────────────────────────────────────────

const PAGE_SIZE = 10

// ── Main component ───────────────────────────────────────────────────────────

export default function UsersPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── List state ─────────────────────────────────────────────────────────────
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [filterStatus, setFilterStatus] = useState<string>('')

  // ── Roles (for role_ids selector) ──────────────────────────────────────────
  const [roles, setRoles] = useState<Role[]>([])

  // ── Create/edit modal ──────────────────────────────────────────────────────
  const [formModalOpen, setFormModalOpen] = useState(false)
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create')
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [formLoading, setFormLoading] = useState(false)
  const [form] = Form.useForm()

  // ── Role assignment in modal ───────────────────────────────────────────────
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([])

  // ── Fetch roles for the form selector ──────────────────────────────────────

  const fetchRoles = useCallback(async () => {
    try {
      const { data } = await client.get<ListResponse<Role>>('/roles')
      setRoles(data.data)
    } catch {
      // Silently fail — roles aren't always needed
    }
  }, [])

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

  // ── Data fetching ──────────────────────────────────────────────────────────

  const fetchUsers = useCallback(
    async (p: number, status: string) => {
      setLoading(true)
      try {
        const params: Record<string, string | number> = {
          page: p,
          page_size: PAGE_SIZE,
        }
        if (status) params.status = status

        const { data } = await client.get<ListResponse<User>>('/users', {
          params,
        })
        setUsers(data.data)
        setTotal(data.pagination.total)
      } catch {
        message.error(t('user.loadFailed'))
      } finally {
        setLoading(false)
      }
    },
    [message, t],
  )

  useEffect(() => {
    fetchUsers(page, filterStatus)
  }, [page, filterStatus, fetchUsers])

  // ── Refresh ────────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    if (page === 1) {
      fetchUsers(1, filterStatus)
    } else {
      setPage(1)
    }
  }

  const handleFilterChange = (status: string) => {
    setFilterStatus(status)
    setPage(1)
  }

  // ── Table change handler ───────────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Create user ────────────────────────────────────────────────────────────

  const openCreateModal = () => {
    setFormMode('create')
    setEditingUser(null)
    setSelectedRoleIds([])
    form.resetFields()
    setFormModalOpen(true)
  }

  const openEditModal = (user: User) => {
    setFormMode('edit')
    setEditingUser(user)
    setSelectedRoleIds([...user.role_ids])
    form.setFieldsValue({
      email: user.email,
      display_name: user.display_name,
    })
    setFormModalOpen(true)
  }

  const closeFormModal = () => {
    setFormModalOpen(false)
    setEditingUser(null)
    setSelectedRoleIds([])
    form.resetFields()
  }

  const handleFormSubmit = async () => {
    try {
      const values = await form.validateFields()
      setFormLoading(true)

      if (formMode === 'create') {
        const payload: CreateUserRequest = {
          username: values.username,
          email: values.email,
          password: values.password,
          display_name: values.display_name || undefined,
          role_ids: selectedRoleIds.length > 0 ? selectedRoleIds : undefined,
        }
        await client.post('/users', payload)
        message.success(t('user.created'))
      } else if (editingUser) {
        const payload: UpdateUserRequest = {
          email: values.email,
          display_name: values.display_name || undefined,
          role_ids: selectedRoleIds,
        }
        await client.put(`/users/${editingUser.id}`, payload)
        message.success(t('user.updated'))
      }

      closeFormModal()
      fetchUsers(page, filterStatus)
    } catch {
      if (formModalOpen) {
        message.error(formMode === 'create' ? t('user.createFailed') : t('user.updateFailed'))
      }
    } finally {
      setFormLoading(false)
    }
  }

  // ── Status transition ──────────────────────────────────────────────────────

  const handleStatusChange = async (user: User, newStatus: string) => {
    try {
      await client.put(`/users/${user.id}/status`, {
        status: newStatus,
      } satisfies UpdateUserStatusRequest)
      message.success(t('user.statusUpdated', { status: t(`user.status${titleCase(newStatus)}`) }))
      fetchUsers(page, filterStatus)
    } catch {
      message.error(t('user.statusUpdateFailed'))
    }
  }

  // ── Stats ──────────────────────────────────────────────────────────────────

  const stats = useMemo(() => {
    const counts: Record<string, number> = {}
    users.forEach((u) => {
      counts[u.status] = (counts[u.status] ?? 0) + 1
    })
    return {
      total,
      active: counts['active'] ?? 0,
      inactive: counts['inactive'] ?? 0,
      locked: counts['locked'] ?? 0,
    }
  }, [users, total])

  // ── Table columns ──────────────────────────────────────────────────────────

  const columns: ColumnsType<User> = useMemo(
    () => [
      {
        title: t('user.username'),
        dataIndex: 'username',
        key: 'username',
        width: 150,
        render: (name: string) => (
          <Space>
            <UserOutlined />
            <Typography.Text strong>{name}</Typography.Text>
          </Space>
        ),
      },
      {
        title: t('user.displayName'),
        dataIndex: 'display_name',
        key: 'display_name',
        width: 160,
        responsive: ['md'],
        render: (v: string) =>
          v || <Typography.Text type="secondary">—</Typography.Text>,
      },
      {
        title: t('user.email'),
        dataIndex: 'email',
        key: 'email',
        width: 220,
        responsive: ['lg'],
        ellipsis: true,
      },
      {
        title: t('common.status'),
        dataIndex: 'status',
        key: 'status',
        width: 120,
        render: (s: string) => (
          <Tag icon={userStatusIcons[s]} color={userStatusColors[s] ?? 'default'}>
            {t(`user.status${titleCase(s)}`)}
          </Tag>
        ),
      },
      {
        title: t('user.roles'),
        dataIndex: 'role_ids',
        key: 'role_ids',
        width: 150,
        responsive: ['md'],
        render: (ids: string[]) => {
          const roleNames = ids
            .map((id) => roles.find((r) => r.id === id)?.name)
            .filter(Boolean) as string[]
          return roleNames.length > 0 ? (
            <Space size={4} wrap>
              {roleNames.map((name) => (
                <Tag key={name} color="blue">
                  {name}
                </Tag>
              ))}
            </Space>
          ) : (
            <Typography.Text type="secondary">—</Typography.Text>
          )
        },
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
        width: 200,
        render: (_: unknown, record: User) => {
          const transitions: string[] =
            record.status === 'active'
              ? ['inactive', 'locked']
              : record.status === 'inactive'
                ? ['active', 'locked']
                : record.status === 'locked'
                  ? ['active', 'inactive']
                  : []
          return (
            <Space size="small">
              <Button
                type="default"
                size="small"
                icon={<EditOutlined />}
                onClick={() => openEditModal(record)}
              >
                {t('common.edit')}
              </Button>
              {transitions.map((target) => (
                <Popconfirm
                  key={target}
                  title={t('user.confirmStatusChange', {
                    target: t(`user.status${titleCase(target)}`),
                  })}
                  onConfirm={() => handleStatusChange(record, target)}
                  okText={t('common.yes')}
                  cancelText={t('common.no')}
                >
                  <Button
                    type="default"
                    size="small"
                    danger={target === 'locked'}
                    icon={
                      target === 'active' ? (
                        <CheckCircleOutlined />
                      ) : target === 'locked' ? (
                        <LockOutlined />
                      ) : (
                        <StopOutlined />
                      )
                    }
                  >
                    {t(`user.status${titleCase(target)}`)}
                  </Button>
                </Popconfirm>
              ))}
            </Space>
          )
        },
      },
    ],
    [t, roles],
  )

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('user.title')}</Typography.Title>
        <Typography.Text type="secondary">{t('user.subtitle')}</Typography.Text>
      </div>

      {/* ── Summary stat cards ──────────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('user.totalUsers')}
              value={stats.total}
              prefix={<TeamOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('user.statusActive')}
              value={stats.active}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: stats.active > 0 ? '#52c41a' : undefined }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('user.statusInactive')}
              value={stats.inactive}
              prefix={<StopOutlined />}
              valueStyle={{ color: stats.inactive > 0 ? '#8c8c8c' : undefined }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('user.statusLocked')}
              value={stats.locked}
              prefix={<LockOutlined />}
              valueStyle={{ color: stats.locked > 0 ? '#ff4d4f' : undefined }}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Table card ───────────────────────────────────────────────────────── */}

      <Card
        title={
          <Space>
            <TeamOutlined />
            <span>{t('user.allUsers')}</span>
          </Space>
        }
        extra={
          <Space>
            <Select
              placeholder={t('common.status')}
              allowClear
              style={{ width: 140 }}
              value={filterStatus || undefined}
              onChange={(v) => handleFilterChange(v ?? '')}
              options={[
                { label: t('user.statusActive'), value: 'active' },
                { label: t('user.statusInactive'), value: 'inactive' },
                { label: t('user.statusLocked'), value: 'locked' },
              ]}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              {t('user.newUser')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('common.refresh')}
            </Button>
          </Space>
        }
      >
        <Table<User>
          columns={columns}
          dataSource={users}
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
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={t('user.noUsers')}
              />
            ),
          }}
        />
      </Card>

      {/* ── Create / Edit User Modal ────────────────────────────────────────── */}

      <Modal
        title={
          formMode === 'create' ? t('user.createUser') : t('user.editUser')
        }
        open={formModalOpen}
        onCancel={closeFormModal}
        onOk={handleFormSubmit}
        confirmLoading={formLoading}
        destroyOnClose
        width={560}
        okText={formMode === 'create' ? t('common.create') : t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          {formMode === 'create' && (
            <>
              <Form.Item
                name="username"
                label={t('user.username')}
                rules={[{ required: true, message: t('user.pleaseEnterUsername') }]}
              >
                <Input placeholder={t('user.usernamePlaceholder')} />
              </Form.Item>
              <Form.Item
                name="password"
                label={t('auth.password')}
                rules={[
                  { required: true, message: t('user.pleaseEnterPassword') },
                  { min: 6, message: t('user.passwordMinLength') },
                ]}
              >
                <Input.Password placeholder={t('user.passwordPlaceholder')} />
              </Form.Item>
            </>
          )}
          <Form.Item
            name="display_name"
            label={t('user.displayName')}
          >
            <Input placeholder={t('user.displayNamePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="email"
            label={t('user.email')}
            rules={[
              { required: true, message: t('user.pleaseEnterEmail') },
              { type: 'email', message: t('user.invalidEmail') },
            ]}
          >
            <Input placeholder={t('user.emailPlaceholder')} />
          </Form.Item>
          <Form.Item label={t('user.roles')}>
            <Select
              mode="multiple"
              placeholder={t('user.rolesPlaceholder')}
              value={selectedRoleIds}
              onChange={setSelectedRoleIds}
              options={roles.map((r) => ({
                label: r.name,
                value: r.id,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
