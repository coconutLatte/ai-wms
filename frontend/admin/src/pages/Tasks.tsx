// Tasks placeholder page with i18n.

import { Typography, Empty } from 'antd'
import { CarryOutOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

export default function TasksPage() {
  const { t } = useTranslation()

  return (
    <div>
      <div className="page-header">
        <Typography.Title level={2}>{t('task.title')}</Typography.Title>
      </div>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <Typography.Text type="secondary">
            {t('task.placeholder')}
          </Typography.Text>
        }
      >
        <CarryOutOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
      </Empty>
    </div>
  )
}
