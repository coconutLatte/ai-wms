// Test setup: loads @testing-library/jest-dom matchers and mocks
// global dependencies (i18next, window.matchMedia for Ant Design).

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

// ── Polyfill matchMedia (Ant Design uses it for responsive breakpoints) ──

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// ── Polyfill getComputedStyle (Ant Design uses it) ──

const originalGetComputedStyle = window.getComputedStyle;
window.getComputedStyle = (elt: Element, pseudoElt?: string | null) => {
  const style = originalGetComputedStyle(elt, pseudoElt);
  // Return a proxy that handles CSS custom properties gracefully
  return new Proxy(style, {
    get: (target, prop) => {
      if (typeof prop === 'string' && prop.startsWith('--')) return '';
      const value = Reflect.get(target, prop);
      return value;
    },
  });
};
