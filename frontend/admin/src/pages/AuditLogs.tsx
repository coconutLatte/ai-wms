// Audit Logs — system audit trail viewer with filters.
// Shows all audited operations (user actions on resources) with
// action/resource/user/date filters and a detail modal showing raw JSON details.

import { useState, useEffect, useCallback } from 'react'
import { Table, Tag, Input, DatePicker, Space, Typography, Modal, Descriptions } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getAuditLogs, type AuditLogQueryParams } from '@/api/audit'
import type { AuditLog, PaginationMeta } from '@/api/types'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { Text, Paragraph } = Typography

// ── Action badge colors ───────────────────────────────────────────────────────

const actionColors: Record<string, string> = {
  create: 'green',
  update: 'blue',
  delete: 'red',
  login: 'purple',
  logout: 'geekblue',
  confirm: 'cyan',
  cancel: 'orange',
  adjust: 'gold',
}

function actionColor(action: string): string {
  for (const [key, color] of Object.entries(actionColors)) {
    if (action.includes(key)) return color
  }
  return 'default'
}

export default function AuditLogsPage() {
  const { t } = useTranslation()

  const [logs, setLogs] = useState<AuditLog[]>([])
  const [pagination, setPagination] = useState<PaginationMeta>({
    total: 0, page: 1, page_size: 20, total_pages: 0,
  })
  const [loading, setLoading] = useState(false)

  // Filters
  const [filterAction, setFilterAction] = useState('')
  const [filterUser, setFilterUser] = useState('')
  const [filterResource, setFilterResource] = useState('')
  const [dateRange, setDateRange] = useState<[string, string] | null>(null)

  // Detail modal
  const [detailLog, setDetailLog] = useState<AuditLog | null>(null)

  const fetchLogs = useCallback(async (page = 1, pageSize = 20) => {
    setLoading(true)
    try {
      const params: AuditLogQueryParams = { page, page_size: pageSize }
      if (filterAction) params.action = filterAction
      if (filterUser) params.user_id = filterUser
      if (filterResource) params.resource = filterResource
      if (dateRange) {
        params.date_from = dateRange[0]
        params.date_to = dateRange[1]
      }
      const resp = await getAuditLogs(params)
      setLogs(resp.data)
      setPagination(resp.pagination)
    } catch {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }, [filterAction, filterUser, filterResource, dateRange])

  useEffect(() => {
    fetchLogs(1, pagination.page_size)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleTableChange = (pag: { current?: number; pageSize?: number }) => {
    const page = pag.current ?? pagination.page
    const pageSize = pag.pageSize ?? pagination.page_size
    fetchLogs(page, pageSize)
  }

  const handleRefresh = () => fetchLogs(pagination.page, pagination.page_size)

  const handleSearch = () => fetchLogs(1, pagination.page_size)

  const handleDateChange = (_: unknown, dateStrings: [string, string]) => {
    if (dateStrings[0] && dateStrings[1]) {
      setDateRange([`${dateStrings[0]}T00:00:00Z`, `${dateStrings[1]}T23:59:59Z`])
    } else {
      setDateRange(null)
    }
  }

  // ── Format details JSON for display ──────────────────────────────────────────

  const formatDetails = (details: string): React.ReactNode => {
    if (!details || details === '{}') return <Text type="secondary">-</Text>
    try {
      const parsed = JSON.parse(details)
      // If it has old/new values, show them side by side
      if (parsed.old || parsed.new) {
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            {parsed.old && (
              <div>
                <Text strong>{t('auditLog.oldValues')}: </Text>
                <pre style={{
                  margin: 0, padding: '4px 8px', background: '#fff2f0',
                  borderRadius: 4, fontSize: 12, overflow: 'auto', maxHeight: 200,
                }}>
                  {JSON.stringify(parsed.old, null, 2)}
                </pre>
              </div>
            )}
            {parsed.new && (
              <div>
                <Text strong>{t('auditLog.newValues')}: </Text>
                <pre style={{
                  margin: 0, padding: '4px 8px', background: '#f6ffed',
                  borderRadius: 4, fontSize: 12, overflow: 'auto', maxHeight: 200,
                }}>
                  {JSON.stringify(parsed.new, null, 2)}
                </pre>
              </div>
            )}
          </Space>
        )
      }
      // Otherwise, show full JSON
      return (
        <pre style={{
          margin: 0, padding: '4px 8px', background: '#f5f5f5',
          borderRadius: 4, fontSize: 12, overflow: 'auto', maxHeight: 300,
        }}>
          {JSON.stringify(parsed, null, 2)}
        </pre>
      )
    } catch {
      return <Text style={{ wordBreak: 'break-all' }}>{details}</Text>
    }
  }

  // ── Columns ──────────────────────────────────────────────────────────────────

  const columns = [
    {
      title: t('auditLog.action'),
      dataIndex: 'action',
      key: 'action',
      width: 160,
      render: (action: string) => (
        <Tag color={actionColor(action)}>{action}</Tag>
      ),
    },
    {
      title: t('auditLog.resource'),
      dataIndex: 'resource',
      key: 'resource',
      width: 120,
      render: (v: string) => <Text strong>{v}</Text>,
    },
    {
      title: t('auditLog.resourceId'),
      dataIndex: 'resource_id',
      key: 'resource_id',
      width: 280,
      ellipsis: true,
      render: (v: string) => v ? <Text code style={{ fontSize: 12 }}>{v}</Text> : '-',
    },
    {
      title: t('auditLog.username'),
      dataIndex: 'username',
      key: 'username',
      width: 100,
    },
    {
      title: t('auditLog.ipAddress'),
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 130,
      render: (v: string) => <Text code>{v || '-'}</Text>,
    },
    {
      title: t('auditLog.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 80,
      render: (_: unknown, record: AuditLog) => (
        <a onClick={() => setDetailLog(record)}>{t('common.view')}</a>
      ),
    },
  ]

  return (
    <div style={{ padding: 24 }}>
      <Typography.Title level={4}>{t('auditLog.title')}</Typography.Title>
      <Typography.Paragraph type="secondary">{t('auditLog.subtitle')}</Typography.Paragraph>

      {/* Filters */}
      <Space wrap style={{ marginBottom: 16 }}>
        <Input
          placeholder={t('auditLog.filterAction')}
          value={filterAction}
          onChange={(e) => setFilterAction(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
        <Input
          placeholder={t('auditLog.filterUser')}
          value={filterUser}
          onChange={(e) => setFilterUser(e.target.value)}
          style={{ width: 300 }}
          allowClear
        />
        <Input
          placeholder={t('auditLog.filterResource')}
          value={filterResource}
          onChange={(e) => setFilterResource(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
        <RangePicker onChange={handleDateChange} />
        <Space>
          <span onClick={handleSearch} style={{ cursor: 'pointer' }}>
            <SearchOutlined style={{ fontSize: 18 }} />
          </span>
          <span onClick={handleRefresh} style={{ cursor: 'pointer' }}>
            <ReloadOutlined spin={loading} style={{ fontSize: 18 }} />
          </span>
        </Space>
      </Space>

      {/* Table */}
      <Table
        dataSource={logs}
        columns={columns}
        rowKey="id"
        loading={loading}
        size="middle"
        pagination={{
          current: pagination.page,
          pageSize: pagination.page_size,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => t('common.total') + `: ${total}`,
        }}
        onChange={handleTableChange}
      />

      {/* Detail Modal */}
      <Modal
        title={t('auditLog.detailTitle')}
        open={!!detailLog}
        onCancel={() => setDetailLog(null)}
        footer={null}
        width={640}
      >
        {detailLog && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID">{detailLog.id}</Descriptions.Item>
            <Descriptions.Item label={t('auditLog.action')}>
              <Tag color={actionColor(detailLog.action)}>{detailLog.action}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label={t('auditLog.username')}>{detailLog.username}</Descriptions.Item>
            <Descriptions.Item label={t('auditLog.resource')}>
              <Text strong>{detailLog.resource}</Text>
            </Descriptions.Item>
            <Descriptions.Item label={t('auditLog.resourceId')}>
              {detailLog.resource_id ? <Text code style={{ fontSize: 12 }}>{detailLog.resource_id}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label={t('auditLog.ipAddress')}>
              <Text code>{detailLog.ip_address || '-'}</Text>
            </Descriptions.Item>
            <Descriptions.Item label={t('auditLog.userId')}>
              <Text code style={{ fontSize: 12 }}>{detailLog.user_id}</Text>
            </Descriptions.Item>
            <Descriptions.Item label={t('auditLog.createdAt')}>
              {dayjs(detailLog.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
        {detailLog && (
          <div style={{ marginTop: 16 }}>
            <Paragraph type="secondary" style={{ marginBottom: 8 }}>
              {t('auditLog.details')}:
            </Paragraph>
            {formatDetails(detailLog.details)}
          </div>
        )}
      </Modal>
    </div>
  )
}
