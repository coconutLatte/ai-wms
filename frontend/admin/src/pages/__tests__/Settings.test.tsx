// Smoke test: Settings page renders the configuration form.
// Mocks the settings API calls.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import SettingsPage from '@/pages/Settings';

// ── Mock settings API ──────────────────────────────────────────────────

const mockGetSettings = vi.fn();
const mockUpdateSettings = vi.fn();

vi.mock('@/api/settings', () => ({
  getSettings: (...args: unknown[]) => mockGetSettings(...args),
  updateSettings: (...args: unknown[]) => mockUpdateSettings(...args),
}));

describe('SettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetSettings.mockReturnValue(new Promise(() => {})); // loading by default
  });

  it('renders the page title and loading state without crashing', () => {
    render(
      <MemoryRouter>
        <SettingsPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('settings.title')).toBeInTheDocument();
  });

  it('renders the settings form after data loads', async () => {
    mockGetSettings.mockResolvedValue({
      site_name: 'AI-WMS Demo',
      default_warehouse: 'wh-demo',
      low_stock_threshold: 10,
      default_page_size: 20,
      jwt_ttl_minutes: 1440,
    });

    render(
      <MemoryRouter>
        <SettingsPage />
      </MemoryRouter>,
    );

    expect(await screen.findByText('common.save')).toBeInTheDocument();
    expect(screen.getByText('settings.reset')).toBeInTheDocument();
  });
});
