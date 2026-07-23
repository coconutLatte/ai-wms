// Inventory Transactions — global inventory transaction history with filters.
// Shows every inventory change (receipt, putaway, pick, ship, transfer, adjustment, return)
// with type badges, delta_qty, resulting_qty, and date. Supports type/date/SKU/warehouse filters.

import { useState, useEffect, useCallback } from 'react'
import { Table, Tag, Input, Select, DatePicker, Space, Typography, Modal, Descriptions } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getTransactions, type TransactionQueryParams } from '@/api/inventory'
import type { InventoryTransaction, PaginationMeta } from '@/api/types'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker
const { Text } = Typography

// ── Type badge colors ──────────────────────────────────────────────────────────

const typeColors: Record<string, string> = {
  receipt: 'green',
  putaway: 'blue',
  pick: 'orange',
  ship: 'purple',
  transfer: 'cyan',
  adjustment: 'gold',
  return: 'magenta',
  reserve: 'geekblue',
  unreserve: 'lime',
}

export default function InventoryTransactionsPage() {
  const { t } = useTranslation()

  const [transactions, setTransactions] = useState<InventoryTransaction[]>([])
  const [pagination, setPagination] = useState<PaginationMeta>({
    total: 0, page: 1, page_size: 20, total_pages: 0,
  })
  const [loading, setLoading] = useState(false)

  // Filters
  const [filterType, setFilterType] = useState<string | undefined>(undefined)
  const [filterSku, setFilterSku] = useState('')
  const [filterWarehouse, setFilterWarehouse] = useState('')
  const [dateRange, setDateRange] = useState<[string, string] | null>(null)

  // Detail modal
  const [detailTx, setDetailTx] = useState<InventoryTransaction | null>(null)

  const fetchTransactions = useCallback(async (page = 1, pageSize = 20) => {
    setLoading(true)
    try {
      const params: TransactionQueryParams = { page, page_size: pageSize }
      if (filterType) params.type = filterType
      if (filterSku) params.sku_id = filterSku
      if (filterWarehouse) params.warehouse_id = filterWarehouse
      if (dateRange) {
        params.date_from = dateRange[0]
        params.date_to = dateRange[1]
      }
      const resp = await getTransactions(params)
      setTransactions(resp.data)
      setPagination(resp.pagination)
    } catch {
      // Error handled by interceptor
    } finally {
      setLoading(false)
    }
  }, [filterType, filterSku, filterWarehouse, dateRange])

  useEffect(() => {
    fetchTransactions(1, pagination.page_size)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleTableChange = (pag: { current?: number; pageSize?: number }) => {
    const page = pag.current ?? pagination.page
    const pageSize = pag.pageSize ?? pagination.page_size
    fetchTransactions(page, pageSize)
  }

  const handleRefresh = () => fetchTransactions(pagination.page, pagination.page_size)

  const handleSearch = () => fetchTransactions(1, pagination.page_size)

  const columns = [
    {
      title: t('inventoryTxn.type'),
      dataIndex: 'type',
      key: 'type',
      width: 110,
      render: (type: string) => (
        <Tag color={typeColors[type] || 'default'}>{t(`inventoryTxn.${type}`, type)}</Tag>
      ),
    },
    {
      title: t('inventoryTxn.deltaQty'),
      dataIndex: 'delta_qty',
      key: 'delta_qty',
      width: 100,
      render: (v: number) => (
        <Text style={{ color: v > 0 ? '#52c41a' : v < 0 ? '#ff4d4f' : undefined }} strong>
          {v > 0 ? '+' : ''}{v}
        </Text>
      ),
    },
    {
      title: t('inventoryTxn.resultingQty'),
      dataIndex: 'resulting_qty',
      key: 'resulting_qty',
      width: 110,
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: t('inventoryTxn.skuId'),
      dataIndex: 'sku_id',
      key: 'sku_id',
      width: 300,
      ellipsis: true,
      render: (v: string) => <Text code style={{ fontSize: 12 }}>{v}</Text>,
    },
    {
      title: t('inventoryTxn.referenceType'),
      dataIndex: 'reference_type',
      key: 'reference_type',
      width: 120,
      render: (v: string) => v || '-',
    },
    {
      title: t('inventoryTxn.createdBy'),
      dataIndex: 'created_by',
      key: 'created_by',
      width: 100,
      render: (v: string) => v || '-',
    },
    {
      title: t('inventoryTxn.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 80,
      render: (_: unknown, record: InventoryTransaction) => (
        <a onClick={() => setDetailTx(record)}>{t('common.view')}</a>
      ),
    },
  ]

  const handleDateChange = (_: unknown, dateStrings: [string, string]) => {
    if (dateStrings[0] && dateStrings[1]) {
      setDateRange([`${dateStrings[0]}T00:00:00Z`, `${dateStrings[1]}T23:59:59Z`])
    } else {
      setDateRange(null)
    }
  }

  return (
    <div style={{ padding: 24 }}>
      <Typography.Title level={4}>{t('inventoryTxn.title')}</Typography.Title>
      <Typography.Paragraph type="secondary">{t('inventoryTxn.subtitle')}</Typography.Paragraph>

      {/* Filters */}
      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          allowClear
          placeholder={t('inventoryTxn.filterType')}
          value={filterType}
          onChange={setFilterType}
          style={{ width: 140 }}
          options={[
            'receipt', 'putaway', 'pick', 'ship', 'transfer', 'adjustment', 'return', 'reserve', 'unreserve',
          ].map((v) => ({ value: v, label: t(`inventoryTxn.${v}`, v) }))}
        />
        <Input
          placeholder={t('inventoryTxn.filterSku')}
          value={filterSku}
          onChange={(e) => setFilterSku(e.target.value)}
          style={{ width: 300 }}
          allowClear
        />
        <Input
          placeholder={t('inventoryTxn.filterWarehouse')}
          value={filterWarehouse}
          onChange={(e) => setFilterWarehouse(e.target.value)}
          style={{ width: 300 }}
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
        dataSource={transactions}
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
        title={t('inventoryTxn.detailTitle')}
        open={!!detailTx}
        onCancel={() => setDetailTx(null)}
        footer={null}
        width={560}
      >
        {detailTx && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID">{detailTx.id}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.type')}>
              <Tag color={typeColors[detailTx.type] || 'default'}>
                {t(`inventoryTxn.${detailTx.type}`, detailTx.type)}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.inventoryId')}>{detailTx.inventory_id}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.skuId')}>{detailTx.sku_id}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.locationId')}>{detailTx.location_id}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.deltaQty')}>
              <Text style={{ color: detailTx.delta_qty > 0 ? '#52c41a' : detailTx.delta_qty < 0 ? '#ff4d4f' : undefined }} strong>
                {detailTx.delta_qty > 0 ? '+' : ''}{detailTx.delta_qty}
              </Text>
            </Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.resultingQty')}>{detailTx.resulting_qty}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.referenceType')}>{detailTx.reference_type || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.referenceId')}>{detailTx.reference_id || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.createdBy')}>{detailTx.created_by || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('inventoryTxn.createdAt')}>
              {dayjs(detailTx.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}
