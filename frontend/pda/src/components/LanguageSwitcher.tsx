// Language switcher for PDA UI — simple toggle button in header.
// Toggles between Chinese (zh-CN) and English (en), persists choice in localStorage.

import { useTranslation } from 'react-i18next'

const languages: Record<string, string> = {
  'zh-CN': 'EN',
  en: '中文',
}

export default function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const nextLang = i18n.language === 'zh-CN' ? 'en' : 'zh-CN'

  const handleToggle = () => {
    i18n.changeLanguage(nextLang)
  }

  return (
    <button
      onClick={handleToggle}
      style={{
        background: 'rgba(255,255,255,0.15)',
        border: 'none',
        color: '#fff',
        padding: '4px 10px',
        borderRadius: 6,
        fontSize: 12,
        fontWeight: 600,
        cursor: 'pointer',
        minWidth: 36,
      }}
      title={i18n.language === 'zh-CN' ? 'Switch to English' : '切换到中文'}
    >
      {languages[i18n.language] ?? 'EN'}
    </button>
  )
}
