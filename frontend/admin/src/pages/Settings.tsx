// Settings page — admin-only system configuration.
// Displays and allows editing of application-level settings including site name,
// default warehouse, low-stock threshold, pagination defaults, and JWT TTL.

import { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Form,
  Input,
  InputNumber,
  Button,
  Typography,
  Spin,
  Alert,
  App,
  Descriptions,
} from 'antd'
import { SettingOutlined, SaveOutlined, UndoOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { getSettings, updateSettings } from '@/api/settings'
import type { AppConfig, UpdateAppConfigRequest } from '@/api/types'

export default function SettingsPage() {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const [config, setConfig] = useState<AppConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [form] = Form.useForm()

  const fetchConfig = useCallback(() => {
    setLoading(true)
    setError(null)

    getSettings()
      .then((result) => {
        setConfig(result)
        form.setFieldsValue(result)
        setLoading(false)
      })
      .catch((err) => {
        setError(
          err?.response?.data?.detail || err?.message || 'Failed to load settings',
        )
        setLoading(false)
      })
  }, [form])

  useEffect(() => {
    fetchConfig()
  }, [fetchConfig])

  const handleSave = useCallback(async () => {
    try {
      const values = await form.validateFields()
      setSaving(true)

      const input: UpdateAppConfigRequest = {
        site_name: values.site_name,
        default_warehouse_id: values.default_warehouse_id || '',
        low_stock_threshold: values.low_stock_threshold,
        default_page_size: values.default_page_size,
        jwt_access_ttl: values.jwt_access_ttl,
      }

      const updated = await updateSettings(input)
      setConfig(updated)
      form.setFieldsValue(updated)
      message.success(t('settings.saved'))
    } catch (err: any) {
      if (err?.errorFields) return // Validation error, Ant Design handles display.
      message.error(
        err?.response?.data?.detail ||
          err?.message ||
          t('settings.saveFailed'),
      )
    } finally {
      setSaving(false)
    }
  }, [form, message, t])

  const handleReset = useCallback(() => {
    if (config) {
      form.setFieldsValue(config)
    }
  }, [config, form])

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>
          <SettingOutlined style={{ marginRight: 8 }} />
          {t('settings.title')}
        </Typography.Title>
        <Typography.Text type="secondary">
          {t('settings.subtitle')}
        </Typography.Text>
      </div>

      {loading && (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin size="large" />
        </div>
      )}

      {error && (
        <Alert
          type="error"
          message={t('common.loadFailed', 'Load failed')}
          description={error}
          showIcon
          closable
          style={{ marginBottom: 16 }}
          onClose={() => setError(null)}
        />
      )}

      {config && !loading && (
        <>
          {/* Read-only summary card */}
          <Card
            title={t('settings.currentConfig')}
            style={{ marginBottom: 16 }}
            size="small"
          >
            <Descriptions column={{ xs: 1, sm: 2, lg: 3 }} size="small">
              <Descriptions.Item label={t('settings.siteNameLabel')}>
                {config.site_name}
              </Descriptions.Item>
              <Descriptions.Item label={t('settings.lowStockLabel')}>
                {config.low_stock_threshold}
              </Descriptions.Item>
              <Descriptions.Item label={t('settings.pageSizeLabel')}>
                {config.default_page_size}
              </Descriptions.Item>
              <Descriptions.Item label={t('settings.jwtTtlLabel')}>
                {config.jwt_access_ttl}s
              </Descriptions.Item>
              {config.updated_at && (
                <Descriptions.Item label={t('settings.lastUpdated')}>
                  {new Date(config.updated_at).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>

          {/* Edit form */}
          <Card title={t('settings.editConfig')}>
            <Form
              form={form}
              layout="vertical"
              initialValues={config}
              onFinish={handleSave}
            >
              <Form.Item
                name="site_name"
                label={t('settings.siteNameLabel')}
                rules={[
                  { required: true, message: t('settings.siteNameRequired') },
                ]}
              >
                <Input
                  placeholder={t('settings.siteNamePlaceholder')}
                  maxLength={100}
                />
              </Form.Item>

              <Form.Item
                name="default_warehouse_id"
                label={t('settings.defaultWarehouseLabel')}
                help={t('settings.defaultWarehouseHelp')}
              >
                <Input
                  placeholder={t('settings.defaultWarehousePlaceholder')}
                  maxLength={36}
                />
              </Form.Item>

              <Form.Item
                name="low_stock_threshold"
                label={t('settings.lowStockLabel')}
                rules={[
                  { required: true, message: t('settings.lowStockRequired') },
                ]}
              >
                <InputNumber
                  min={0}
                  max={999999}
                  style={{ width: 200 }}
                  placeholder={t('settings.lowStockPlaceholder')}
                />
              </Form.Item>

              <Form.Item
                name="default_page_size"
                label={t('settings.pageSizeLabel')}
                rules={[
                  { required: true, message: t('settings.pageSizeRequired') },
                ]}
                help={t('settings.pageSizeHelp')}
              >
                <InputNumber
                  min={1}
                  max={100}
                  style={{ width: 200 }}
                  placeholder={t('settings.pageSizePlaceholder')}
                />
              </Form.Item>

              <Form.Item
                name="jwt_access_ttl"
                label={t('settings.jwtTtlLabel')}
                rules={[
                  { required: true, message: t('settings.jwtTtlRequired') },
                ]}
                help={t('settings.jwtTtlHelp')}
              >
                <InputNumber
                  min={60}
                  max={86400}
                  style={{ width: 200 }}
                  addonAfter={t('settings.seconds')}
                  placeholder={t('settings.jwtTtlPlaceholder')}
                />
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  icon={<SaveOutlined />}
                  loading={saving}
                  style={{ marginRight: 8 }}
                >
                  {t('common.save')}
                </Button>
                <Button icon={<UndoOutlined />} onClick={handleReset}>
                  {t('settings.reset')}
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </>
      )}
    </div>
  )
}
