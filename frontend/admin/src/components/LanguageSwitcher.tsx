// Language switcher for Admin UI — dropdown in the header.
// Toggles between Chinese (zh-CN) and English (en), persists choice in localStorage.

import { Dropdown } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

const languages = [
  { key: 'zh-CN', label: '中文' },
  { key: 'en', label: 'English' },
]

export default function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const handleChange = (lang: string) => {
    i18n.changeLanguage(lang)
  }

  const currentLabel = languages.find((l) => l.key === i18n.language)?.label ?? '中文'

  return (
    <Dropdown
      menu={{
        items: languages.map((lang) => ({
          key: lang.key,
          label: lang.label,
          onClick: () => handleChange(lang.key),
        })),
        selectedKeys: [i18n.language],
      }}
      placement="bottomRight"
    >
      <span
        style={{
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          gap: 4,
          fontSize: 13,
          color: '#595959',
          padding: '4px 8px',
        }}
      >
        <GlobalOutlined />
        {currentLabel}
      </span>
    </Dropdown>
  )
}
