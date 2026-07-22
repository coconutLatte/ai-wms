// Role management page with list, create/edit/delete, and permission editor.
// Uses GET/POST/PUT/DELETE /api/v1/roles for real data.
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
  Checkbox,
} from 'antd'
import {
  SafetyOutlined,
  ReloadOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  Role,
  Permission,
  ListResponse,
  CreateRoleRequest,
  UpdateRoleRequest,
} from '@/api/types'

// ── Known WMS resources (for permission editor) ──────────────────────────────

const WMS_RESOURCES = [
  'warehouse',
  'sku',
  'inventory',
  'order',
  'task',
  'wave',
  'location',
  'zone',
  'user',
  'role',
  'audit_log',
  'dashboard',
]

const WMS_ACTIONS = ['read', 'create', 'update', 'delete']

// ── Main component ───────────────────────────────────────────────────────────

export default function RolesPage() {
  const { message, modal } = App.useApp()
  const { t } = useTranslation()

  // ── List state ─────────────────────────────────────────────────────────────
  const [roles, setRoles] = useState<Role[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)

  // ── Detail drawer ──────────────────────────────────────────────────────────
  const [detailRole, setDetailRole] = useState<Role | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  // ── Create/edit modal ──────────────────────────────────────────────────────
  const [formModalOpen, setFormModalOpen] = useState(false)
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create')
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [formLoading, setFormLoading] = useState(false)
  const [form] = Form.useForm()

  // ── Permission editor state ────────────────────────────────────────────────
  const [permissions, setPermissions] = useState<Permission[]>([])

  // ── Data fetching ──────────────────────────────────────────────────────────

  const fetchRoles = useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await client.get<ListResponse<Role>>('/roles')
      setRoles(data.data)
      setTotal(data.pagination.total)
    } catch {
      message.error(t('role.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [message, t])

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

  // ── Refresh ────────────────────────────────────────────────────────────────

  const handleRefresh = () => {
    fetchRoles()
  }

  // ── Permission editor helpers ──────────────────────────────────────────────

  const isActionChecked = (resource: string, action: string): boolean => {
    const perm = permissions.find((p) => p.resource === resource)
    if (!perm) return false
    return perm.actions.includes(action) || perm.actions.includes('*')
  }

  const toggleAction = (resource: string, action: string) => {
    setPermissions((prev) => {
      const existing = prev.find((p) => p.resource === resource)
      if (!existing) {
        // Add new resource permission
        return [...prev, { resource, actions: [action] }]
      }

      const hasWildcard = existing.actions.includes('*')
      const hasAction = existing.actions.includes(action)

      if (hasWildcard) {
        // Remove wildcard, add all actions except this one
        return prev
          .filter((p) => p.resource !== resource)
          .concat([
            {
              resource,
              actions: WMS_ACTIONS.filter((a) => a !== action),
            },
          ])
      }

      if (hasAction) {
        // Remove this action
        const newActions = existing.actions.filter((a) => a !== action)
        if (newActions.length === 0) {
          return prev.filter((p) => p.resource !== resource)
        }
        return prev.map((p) =>
          p.resource === resource ? { ...p, actions: newActions } : p,
        )
      }

      // Add this action
      const newActions = [...existing.actions, action]
      if (newActions.length === WMS_ACTIONS.length) {
        // All actions selected → use wildcard
        return prev.map((p) =>
          p.resource === resource ? { ...p, actions: ['*'] } : p,
        )
      }
      return prev.map((p) =>
        p.resource === resource ? { ...p, actions: newActions } : p,
      )
    })
  }

  const hasAllActions = (resource: string): boolean => {
    const perm = permissions.find((p) => p.resource === resource)
    if (!perm) return false
    return perm.actions.includes('*') || perm.actions.length === WMS_ACTIONS.length
  }

  const toggleAllActions = (resource: string) => {
    if (hasAllActions(resource)) {
      setPermissions((prev) => prev.filter((p) => p.resource !== resource))
    } else {
      setPermissions((prev) =>
        prev
          .filter((p) => p.resource !== resource)
          .concat([{ resource, actions: ['*'] }]),
      )
    }
  }

  // ── Create/edit role ───────────────────────────────────────────────────────

  const openCreateModal = () => {
    setFormMode('create')
    setEditingRole(null)
    setPermissions([])
    form.resetFields()
    setFormModalOpen(true)
  }

  const openEditModal = (role: Role) => {
    setFormMode('edit')
    setEditingRole(role)
    setPermissions(
      role.permissions
        ? role.permissions.map((p) => ({
            resource: p.resource,
            actions: [...p.actions],
          }))
        : [],
    )
    form.setFieldsValue({
      name: role.name,
      description: role.description,
    })
    setFormModalOpen(true)
  }

  const closeFormModal = () => {
    setFormModalOpen(false)
    setEditingRole(null)
    setPermissions([])
    form.resetFields()
  }

  const handleFormSubmit = async () => {
    try {
      const values = await form.validateFields()
      setFormLoading(true)

      if (formMode === 'create') {
        const payload: CreateRoleRequest = {
          name: values.name,
          description: values.description || undefined,
          permissions: permissions.length > 0 ? permissions : undefined,
        }
        await client.post('/roles', payload)
        message.success(t('role.created'))
      } else if (editingRole) {
        const payload: UpdateRoleRequest = {
          name: values.name,
          description: values.description || undefined,
          permissions,
        }
        await client.put(`/roles/${editingRole.id}`, payload)
        message.success(t('role.updated'))
      }

      closeFormModal()
      fetchRoles()
    } catch {
      if (formModalOpen) {
        message.error(formMode === 'create' ? t('role.createFailed') : t('role.updateFailed'))
      }
    } finally {
      setFormLoading(false)
    }
  }

  // ── Delete role ────────────────────────────────────────────────────────────

  const handleDelete = (role: Role) => {
    modal.confirm({
      title: t('role.confirmDeleteTitle'),
      content: t('role.confirmDeleteMessage', { name: role.name }),
      okText: t('common.yes'),
      cancelText: t('common.no'),
      okType: 'danger',
      onOk: async () => {
        try {
          await client.delete(`/roles/${role.id}`)
          message.success(t('role.deleted'))
          fetchRoles()
        } catch {
          message.error(t('role.deleteFailed'))
        }
      },
    })
  }

  // ── Detail drawer ──────────────────────────────────────────────────────────

  const openDetail = (role: Role) => {
    setDetailRole(role)
    setDetailOpen(true)
  }

  const closeDetail = () => {
    setDetailOpen(false)
    setDetailRole(null)
  }

  // ── Stats ──────────────────────────────────────────────────────────────────

  const stats = useMemo(() => ({
    total,
  }), [total])

  // ── Table columns ──────────────────────────────────────────────────────────

  const columns: ColumnsType<Role> = useMemo(
    () => [
      {
        title: t('role.name'),
        dataIndex: 'name',
        key: 'name',
        width: 160,
        render: (name: string) => (
          <Space>
            <SafetyOutlined />
            <Typography.Text strong>{name}</Typography.Text>
          </Space>
        ),
      },
      {
        title: t('role.description'),
        dataIndex: 'description',
        key: 'description',
        width: 280,
        responsive: ['md'],
        ellipsis: true,
        render: (v: string) =>
          v || <Typography.Text type="secondary">—</Typography.Text>,
      },
      {
        title: t('role.permissions'),
        dataIndex: 'permissions',
        key: 'permissions',
        width: 300,
        responsive: ['lg'],
        render: (perms: Permission[]) => {
          if (!perms || perms.length === 0) {
            return <Typography.Text type="secondary">—</Typography.Text>
          }
          return (
            <Space size={4} wrap>
              {perms.flatMap((p) =>
                p.actions.map((a) => (
                  <Tag key={`${p.resource}.${a}`} color="blue" style={{ fontSize: 12 }}>
                    {p.resource}.{a === '*' ? '*' : a}
                  </Tag>
                )),
              )}
            </Space>
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
        render: (_: unknown, record: Role) => (
          <Space size="small">
            <Button
              type="default"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => openDetail(record)}
            >
              {t('common.view')}
            </Button>
            <Button
              type="default"
              size="small"
              icon={<EditOutlined />}
              onClick={() => openEditModal(record)}
            >
              {t('common.edit')}
            </Button>
            <Popconfirm
              title={t('role.confirmDeleteTitle')}
              description={t('role.confirmDeleteMessage', { name: record.name })}
              onConfirm={() => handleDelete(record)}
              okText={t('common.yes')}
              cancelText={t('common.no')}
              okType="danger"
            >
              <Button type="default" size="small" danger icon={<DeleteOutlined />}>
                {t('common.delete')}
              </Button>
            </Popconfirm>
          </Space>
        ),
      },
    ],
    [t],
  )

  // ── Permission editor UI ───────────────────────────────────────────────────

  const permissionEditor = (
    <div style={{ border: '1px solid #f0f0f0', borderRadius: 6, padding: '12px 16px' }}>
      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
        {t('role.permissionsHint')}
      </Typography.Text>
      {WMS_RESOURCES.map((resource) => (
        <Row key={resource} align="middle" style={{ marginBottom: 8 }}>
          <Col span={4}>
            <Checkbox
              checked={hasAllActions(resource)}
              indeterminate={
                !hasAllActions(resource) && isActionChecked(resource, WMS_ACTIONS[0])
              }
              onChange={() => toggleAllActions(resource)}
            >
              <Typography.Text code style={{ fontSize: 12 }}>
                {resource}
              </Typography.Text>
            </Checkbox>
          </Col>
          <Col span={20}>
            <Checkbox.Group
              value={
                permissions
                  .find((p) => p.resource === resource)
                  ?.actions.filter((a) => a !== '*') ?? []
              }
              onChange={(checked) => {
                // Replace actions for this resource
                setPermissions((prev) => {
                  const without = prev.filter((p) => p.resource !== resource)
                  if (checked.length === 0) return without
                  if (checked.length === WMS_ACTIONS.length) {
                    return [...without, { resource, actions: ['*'] }]
                  }
                  return [...without, { resource, actions: checked as string[] }]
                })
              }}
            >
              {WMS_ACTIONS.map((action) => (
                <Checkbox key={action} value={action} style={{ fontSize: 12 }}>
                  {action}
                </Checkbox>
              ))}
            </Checkbox.Group>
          </Col>
        </Row>
      ))}
    </div>
  )

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('role.title')}</Typography.Title>
        <Typography.Text type="secondary">{t('role.subtitle')}</Typography.Text>
      </div>

      {/* ── Summary stat cards ──────────────────────────────────────────────── */}

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={8}>
          <Card size="small" loading={loading}>
            <Statistic
              title={t('role.totalRoles')}
              value={stats.total}
              prefix={<SafetyOutlined />}
              formatter={(v) => Number(v).toLocaleString()}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Table card ───────────────────────────────────────────────────────── */}

      <Card
        title={
          <Space>
            <SafetyOutlined />
            <span>{t('role.allRoles')}</span>
          </Space>
        }
        extra={
          <Space>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
              {t('role.newRole')}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              {t('common.refresh')}
            </Button>
          </Space>
        }
      >
        <Table<Role>
          columns={columns}
          dataSource={roles}
          rowKey="id"
          loading={loading}
          pagination={false}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={t('role.noRoles')}
              />
            ),
          }}
        />
      </Card>

      {/* ── Create / Edit Role Modal ────────────────────────────────────────── */}

      <Modal
        title={formMode === 'create' ? t('role.createRole') : t('role.editRole')}
        open={formModalOpen}
        onCancel={closeFormModal}
        onOk={handleFormSubmit}
        confirmLoading={formLoading}
        destroyOnClose
        width={720}
        okText={formMode === 'create' ? t('common.create') : t('common.save')}
        cancelText={t('common.cancel')}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label={t('role.name')}
            rules={[{ required: true, message: t('role.pleaseEnterName') }]}
          >
            <Input placeholder={t('role.namePlaceholder')} />
          </Form.Item>
          <Form.Item
            name="description"
            label={t('role.description')}
          >
            <Input placeholder={t('role.descriptionPlaceholder')} />
          </Form.Item>
          <Form.Item label={t('role.permissions')}>
            {permissionEditor}
          </Form.Item>
        </Form>
      </Modal>

      {/* ── Detail Drawer ───────────────────────────────────────────────────── */}

      <Modal
        title={
          <Space>
            <SafetyOutlined />
            <span>{t('role.roleDetail')}</span>
            {detailRole && <Typography.Text code>{detailRole.name}</Typography.Text>}
          </Space>
        }
        open={detailOpen}
        onCancel={closeDetail}
        footer={
          <Button onClick={closeDetail}>{t('common.close')}</Button>
        }
        width={640}
      >
        {detailRole ? (
          <>
            <Typography.Title level={5}>{t('role.name')}</Typography.Title>
            <Typography.Paragraph>{detailRole.name}</Typography.Paragraph>

            {detailRole.description && (
              <>
                <Typography.Title level={5}>{t('role.description')}</Typography.Title>
                <Typography.Paragraph>{detailRole.description}</Typography.Paragraph>
              </>
            )}

            <Typography.Title level={5}>{t('role.permissions')}</Typography.Title>
            {detailRole.permissions && detailRole.permissions.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%' }}>
                {detailRole.permissions.map((perm) => (
                  <Row key={perm.resource} gutter={16}>
                    <Col span={6}>
                      <Typography.Text code>{perm.resource}</Typography.Text>
                    </Col>
                    <Col span={18}>
                      <Space size={4} wrap>
                        {(perm.actions.includes('*') ? ['read', 'create', 'update', 'delete'] : perm.actions).map((a) => (
                          <Tag key={a} color="blue">
                            {a === '*' ? t('role.allActions') : a}
                          </Tag>
                        ))}
                      </Space>
                    </Col>
                  </Row>
                ))}
              </Space>
            ) : (
              <Typography.Text type="secondary">{t('role.noPermissions')}</Typography.Text>
            )}

            <Typography.Title level={5} style={{ marginTop: 16 }}>
              {t('common.created')}
            </Typography.Title>
            <Typography.Text>{new Date(detailRole.created_at).toLocaleString()}</Typography.Text>
          </>
        ) : (
          <Empty description={t('role.noRoleSelected')} />
        )}
      </Modal>
    </div>
  )
}
