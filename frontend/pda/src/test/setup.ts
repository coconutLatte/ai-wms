// Test setup: loads @testing-library/jest-dom matchers and mocks
// global dependencies (i18next).

import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

// ── Mock react-i18next ────────────────────────────────────────────────

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: {
      language: 'zh-CN',
      changeLanguage: () => new Promise(() => {}),
    },
  }),
  initReactI18next: {
    type: '3rdParty',
    init: () => {},
  },
}));
