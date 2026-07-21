// SKU management page with full CRUD for stock keeping units.
// Supports listing, searching, creating, and editing SKUs with UOM and attributes.
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
  Card,
  App,
  Empty,
  Row,
  Col,
  Divider,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  SearchOutlined,
  DeleteOutlined,
  BarcodeOutlined,
} from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import client from '@/api/client'
import type {
  SKU,
  UOM,
  CreateSKURequest,
  UpdateSKURequest,
  ListResponse,
} from '@/api/types'

// ── Status tag colors ────────────────────────────────────────────────────

const skuStatusColors: Record<string, string> = {
  active: 'green',
  inactive: 'orange',
  discontinued: 'red',
}

// ── Modal state ──────────────────────────────────────────────────────────

type ModalState =
  | { type: 'none' }
  | { type: 'create' }
  | { type: 'edit'; sku: SKU }

const PAGE_SIZE = 10

// ── Default UOM ──────────────────────────────────────────────────────────

const defaultUOM: UOM = {
  base_unit: 'EA',
  pack_unit: '',
  pack_qty: 1,
  weight: 0,
  volume: 0,
  length: 0,
  width: 0,
  height: 0,
}

// ── Main component ───────────────────────────────────────────────────────

export default function SKUsPage() {
  const { message } = App.useApp()
  const { t } = useTranslation()

  // ── Status labels ──────────────────────────────────────────────────────
  const skuStatusLabels: Record<string, string> = useMemo(() => ({
    active: t('sku.statusActive'),
    inactive: t('sku.statusInactive'),
    discontinued: t('sku.statusDiscontinued'),
  }), [t])

  // ── List state ─────────────────────────────────────────────────────────
  const [skus, setSkus] = useState<SKU[]>([])
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  // ── Search ─────────────────────────────────────────────────────────────
  const [search, setSearch] = useState('')

  // ── Modal state ────────────────────────────────────────────────────────
  const [modal, setModal] = useState<ModalState>({ type: 'none' })
  const [modalLoading, setModalLoading] = useState(false)

  // ── Forms ──────────────────────────────────────────────────────────────
  const [createForm] = Form.useForm<CreateSKURequest>()
  const [editForm] = Form.useForm<UpdateSKURequest>()

  // ── Attributes editor state ────────────────────────────────────────────
  const [createAttrs, setCreateAttrs] = useState<Array<{ key: string; value: string }>>([])
  const [editAttrs, setEditAttrs] = useState<Array<{ key: string; value: string }>>([])

  // ── Data fetching ──────────────────────────────────────────────────────

  const fetchSKUs = useCallback(async (p: number) => {
    setLoading(true)
    try {
      const { data } = await client.get<ListResponse<SKU>>('/skus', {
        params: { page: p, page_size: PAGE_SIZE },
      })
      setSkus(data.data)
      setTotal(data.pagination.total)
    } catch {
      message.error(t('sku.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [message, t])

  useEffect(() => {
    fetchSKUs(page)
  }, [page, fetchSKUs])

  // ── Filtered data (client-side search) ─────────────────────────────────

  const filteredSkus = useMemo(() => {
    if (!search.trim()) return skus
    const q = search.toLowerCase()
    return skus.filter(
      (s) =>
        s.code.toLowerCase().includes(q) ||
        s.name.toLowerCase().includes(q) ||
        (s.barcode && s.barcode.toLowerCase().includes(q)),
    )
  }, [skus, search])

  // ── Table change handler ───────────────────────────────────────────────

  const handleTableChange = (pagination: { current?: number }) => {
    setPage(pagination.current ?? 1)
  }

  // ── Modal open/close ───────────────────────────────────────────────────

  const openCreate = () => {
    createForm.resetFields()
    createForm.setFieldsValue({ uom: defaultUOM })
    setCreateAttrs([])
    setModal({ type: 'create' })
  }

  const openEdit = (sku: SKU) => {
    editForm.setFieldsValue({
      name: sku.name,
      description: sku.description,
      barcode: sku.barcode,
      category: sku.category,
      status: sku.status,
      uom: sku.uom,
    })
    setEditAttrs(
      Object.entries(sku.attributes ?? {}).map(([k, v]) => ({ key: k, value: v })),
    )
    setModal({ type: 'edit', sku })
  }

  const closeModal = () => setModal({ type: 'none' })

  // ── Create SKU ─────────────────────────────────────────────────────────

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      const attributes: Record<string, string> = {}
      createAttrs.forEach(({ key, value }) => {
        if (key.trim()) attributes[key.trim()] = value
      })
      setModalLoading(true)
      await client.post('/skus', { ...values, attributes } satisfies CreateSKURequest)
      message.success(t('sku.skuCreated'))
      closeModal()
      fetchSKUs(page)
    } catch {
      if (modal.type !== 'none') message.error(t('sku.createFailed'))
    } finally {
      setModalLoading(false)
    }
  }

  // ── Update SKU ─────────────────────────────────────────────────────────

  const handleUpdate = async () => {
    if (modal.type !== 'edit') return
    try {
      const values = await editForm.validateFields()
      const payload: UpdateSKURequest = {}
      if (values.name !== undefined && values.name !== modal.sku.name) payload.name = values.name
      if (values.description !== undefined && values.description !== modal.sku.description)
        payload.description = values.description
      if (values.barcode !== undefined && values.barcode !== modal.sku.barcode)
        payload.barcode = values.barcode
      if (values.category !== undefined && values.category !== modal.sku.category)
        payload.category = values.category
      if (values.status !== undefined && values.status !== modal.sku.status)
        payload.status = values.status
      if (values.uom) payload.uom = values.uom

      const attributes: Record<string, string> = {}
      editAttrs.forEach(({ key, value }) => {
        if (key.trim()) attributes[key.trim()] = value
      })
      const origKeys = Object.keys(modal.sku.attributes ?? {}).sort().join(',')
      const newKeys = Object.keys(attributes).sort().join(',')
      if (origKeys !== newKeys || Object.entries(attributes).some(([k, v]) => modal.sku.attributes?.[k] !== v)) {
        payload.attributes = attributes
      }

      setModalLoading(true)
      await client.put(`/skus/${modal.sku.id}`, payload)
      message.success(t('sku.skuUpdated'))
      closeModal()
      fetchSKUs(page)
    } catch {
      message.error(t('sku.updateFailed'))
    } finally {
      setModalLoading(false)
    }
  }

  // ── Attribute editor helpers ───────────────────────────────────────────

  const addCreateAttr = () => setCreateAttrs([...createAttrs, { key: '', value: '' }])
  const removeCreateAttr = (idx: number) =>
    setCreateAttrs(createAttrs.filter((_, i) => i !== idx))
  const updateCreateAttr = (idx: number, field: 'key' | 'value', val: string) => {
    const next = [...createAttrs]
    next[idx] = { ...next[idx], [field]: val }
    setCreateAttrs(next)
  }

  const addEditAttr = () => setEditAttrs([...editAttrs, { key: '', value: '' }])
  const removeEditAttr = (idx: number) =>
    setEditAttrs(editAttrs.filter((_, i) => i !== idx))
  const updateEditAttr = (idx: number, field: 'key' | 'value', val: string) => {
    const next = [...editAttrs]
    next[idx] = { ...next[idx], [field]: val }
    setEditAttrs(next)
  }

  // ── SKU columns ────────────────────────────────────────────────────────

  const columns: ColumnsType<SKU> = useMemo(() => [
    {
      title: t('sku.skuCode'),
      dataIndex: 'code',
      key: 'code',
      width: 150,
      render: (code: string) => <Typography.Text code>{code}</Typography.Text>,
    },
    {
      title: t('sku.skuName'),
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: t('sku.barcode'),
      dataIndex: 'barcode',
      key: 'barcode',
      width: 160,
      ellipsis: true,
      responsive: ['md'],
      render: (b: string) => b || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: t('sku.category'),
      dataIndex: 'category',
      key: 'category',
      width: 120,
      responsive: ['lg'],
      render: (c: string) => c || <Typography.Text type="secondary">—</Typography.Text>,
    },
    {
      title: t('sku.baseUnit'),
      key: 'base_unit',
      width: 90,
      responsive: ['lg'],
      render: (_: unknown, record: SKU) => record.uom?.base_unit || '—',
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => (
        <Tag color={skuStatusColors[status] ?? 'default'}>{skuStatusLabels[status] ?? status}</Tag>
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
      render: (_: unknown, record: SKU) => (
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
  ], [t, skuStatusLabels])

  // ── UOM sub-form (shared between create and edit) ──────────────────────

  const uomFields = (prefix: (string | number)[]) => (
    <>
      <Row gutter={16}>
        <Col xs={12} sm={8}>
          <Form.Item
            name={[...prefix, 'base_unit']}
            label={t('sku.baseUnit')}
            rules={[{ required: true, message: t('sku.baseUnit') }]}
          >
            <Select placeholder={t('sku.baseUnit')}>
              <Select.Option value="EA">{t('sku.baseUnitEA')}</Select.Option>
              <Select.Option value="KG">{t('sku.baseUnitKG')}</Select.Option>
              <Select.Option value="M">{t('sku.baseUnitM')}</Select.Option>
              <Select.Option value="L">{t('sku.baseUnitL')}</Select.Option>
              <Select.Option value="BOX">{t('sku.baseUnitBOX')}</Select.Option>
              <Select.Option value="CASE">{t('sku.baseUnitCASE')}</Select.Option>
              <Select.Option value="PAL">{t('sku.baseUnitPAL')}</Select.Option>
              <Select.Option value="CS">{t('sku.baseUnitCS')}</Select.Option>
            </Select>
          </Form.Item>
        </Col>
        <Col xs={12} sm={8}>
          <Form.Item name={[...prefix, 'pack_unit']} label={t('sku.packUnit')}>
            <Input placeholder="e.g. CASE" />
          </Form.Item>
        </Col>
        <Col xs={12} sm={8}>
          <Form.Item name={[...prefix, 'pack_qty']} label={t('sku.packQty')}>
            <InputNumber min={1} style={{ width: '100%' }} placeholder="1" />
          </Form.Item>
        </Col>
      </Row>
      <Row gutter={16}>
        <Col xs={12} sm={6}>
          <Form.Item name={[...prefix, 'weight']} label={t('sku.weight')}>
            <InputNumber min={0} step={0.001} style={{ width: '100%' }} placeholder="0" />
          </Form.Item>
        </Col>
        <Col xs={12} sm={6}>
          <Form.Item name={[...prefix, 'volume']} label={t('sku.volume')}>
            <InputNumber min={0} step={0.0001} style={{ width: '100%' }} placeholder="0" />
          </Form.Item>
        </Col>
        <Col xs={12} sm={4}>
          <Form.Item name={[...prefix, 'length']} label={t('sku.length')}>
            <InputNumber min={0} step={0.1} style={{ width: '100%' }} placeholder="0" />
          </Form.Item>
        </Col>
        <Col xs={12} sm={4}>
          <Form.Item name={[...prefix, 'width']} label={t('sku.width')}>
            <InputNumber min={0} step={0.1} style={{ width: '100%' }} placeholder="0" />
          </Form.Item>
        </Col>
        <Col xs={12} sm={4}>
          <Form.Item name={[...prefix, 'height']} label={t('sku.height')}>
            <InputNumber min={0} step={0.1} style={{ width: '100%' }} placeholder="0" />
          </Form.Item>
        </Col>
      </Row>
    </>
  )

  // ── Attributes editor (shared) ─────────────────────────────────────────

  const attributesEditor = (
    attrs: Array<{ key: string; value: string }>,
    onAdd: () => void,
    onRemove: (idx: number) => void,
    onChange: (idx: number, field: 'key' | 'value', val: string) => void,
  ) => (
    <div>
      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
        {t('sku.customAttributes')}
      </Typography.Text>
      {attrs.map((attr, idx) => (
        <Row key={idx} gutter={8} style={{ marginBottom: 8 }}>
          <Col xs={10}>
            <Input
              placeholder={t('sku.attrKeyPlaceholder')}
              value={attr.key}
              onChange={(e) => onChange(idx, 'key', e.target.value)}
            />
          </Col>
          <Col xs={10}>
            <Input
              placeholder={t('sku.attrValuePlaceholder')}
              value={attr.value}
              onChange={(e) => onChange(idx, 'value', e.target.value)}
            />
          </Col>
          <Col xs={4}>
            <Button
              icon={<DeleteOutlined />}
              onClick={() => onRemove(idx)}
              danger
              size="small"
              style={{ width: '100%' }}
            />
          </Col>
        </Row>
      ))}
      <Button type="dashed" onClick={onAdd} block icon={<PlusOutlined />}>
        {t('sku.addAttribute')}
      </Button>
    </div>
  )

  // ── Render ─────────────────────────────────────────────────────────────

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('sku.title')}</Typography.Title>
      </div>

      <Card
        title={
          <Space>
            <BarcodeOutlined />
            <span>{t('sku.allSkus')}</span>
          </Space>
        }
        extra={
          <Space>
            <Input
              placeholder={t('sku.searchPlaceholder')}
              prefix={<SearchOutlined />}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              allowClear
              style={{ width: 260 }}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              {t('sku.newSku')}
            </Button>
          </Space>
        }
      >
        <Table<SKU>
          columns={columns}
          dataSource={filteredSkus}
          rowKey="id"
          loading={loading}
          onChange={handleTableChange}
          pagination={{
            current: page,
            pageSize: PAGE_SIZE,
            total: total,
            showSizeChanger: false,
            showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
          }}
          locale={{ emptyText: <Empty description={t('sku.noSkus')} /> }}
        />
      </Card>

      {/* ── Create SKU Modal ───────────────────────────────────────────── */}

      <Modal
        title={t('sku.createSku')}
        open={modal.type === 'create'}
        onCancel={closeModal}
        onOk={handleCreate}
        confirmLoading={modalLoading}
        destroyOnClose
        width={680}
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item
                name="code"
                label={t('sku.skuCode')}
                rules={[{ required: true, message: t('sku.pleaseEnterCode') }]}
              >
                <Input placeholder={t('sku.codePlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item
                name="barcode"
                label={t('sku.barcode')}
              >
                <Input placeholder={t('sku.barcodePlaceholder')} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="name"
            label={t('sku.skuName')}
            rules={[{ required: true, message: t('sku.pleaseEnterName') }]}
          >
            <Input placeholder={t('sku.namePlaceholder')} />
          </Form.Item>
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item name="category" label={t('sku.category')}>
                <Input placeholder={t('sku.categoryPlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item name="description" label={t('sku.description')}>
                <Input placeholder={t('sku.descriptionPlaceholder')} />
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left" plain>
            <Typography.Text type="secondary">{t('sku.uom')}</Typography.Text>
          </Divider>
          {uomFields(['uom'])}

          <Divider orientation="left" plain>
            <Typography.Text type="secondary">{t('sku.attributes')}</Typography.Text>
          </Divider>
          {attributesEditor(createAttrs, addCreateAttr, removeCreateAttr, updateCreateAttr)}
        </Form>
      </Modal>

      {/* ── Edit SKU Modal ─────────────────────────────────────────────── */}

      <Modal
        title={
          modal.type === 'edit' ? (
            <span>
              {t('sku.editSku')} — <Typography.Text code>{modal.sku.code}</Typography.Text>
            </span>
          ) : (
            t('sku.editSku')
          )
        }
        open={modal.type === 'edit'}
        onCancel={closeModal}
        onOk={handleUpdate}
        confirmLoading={modalLoading}
        destroyOnClose
        width={680}
      >
        <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item name="name" label={t('sku.skuName')}>
                <Input placeholder={t('sku.namePlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item name="barcode" label={t('sku.barcode')}>
                <Input placeholder={t('sku.barcodePlaceholder')} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col xs={12}>
              <Form.Item name="category" label={t('sku.category')}>
                <Input placeholder={t('sku.categoryPlaceholder')} />
              </Form.Item>
            </Col>
            <Col xs={12}>
              <Form.Item name="description" label={t('sku.description')}>
                <Input placeholder={t('sku.descriptionPlaceholderEdit')} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="status"
            label={t('common.status')}
          >
            <Select>
              <Select.Option value="active">{t('sku.statusActive')}</Select.Option>
              <Select.Option value="inactive">{t('sku.statusInactive')}</Select.Option>
              <Select.Option value="discontinued">{t('sku.statusDiscontinued')}</Select.Option>
            </Select>
          </Form.Item>

          <Divider orientation="left" plain>
            <Typography.Text type="secondary">{t('sku.uom')}</Typography.Text>
          </Divider>
          {uomFields(['uom'])}

          <Divider orientation="left" plain>
            <Typography.Text type="secondary">{t('sku.attributes')}</Typography.Text>
          </Divider>
          {attributesEditor(editAttrs, addEditAttr, removeEditAttr, updateEditAttr)}
        </Form>
      </Modal>
    </div>
  )
}
