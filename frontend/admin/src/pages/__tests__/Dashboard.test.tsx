// Smoke test: Dashboard page renders the loading/stat card layout.
// Mocks the API call to simulate an empty dashboard.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import DashboardPage from '@/pages/Dashboard';

// ── Mock dashboard API ─────────────────────────────────────────────────

const mockGetDashboard = vi.fn();

vi.mock('@/api/dashboard', () => ({
  getDashboard: (...args: unknown[]) => mockGetDashboard(...args),
}));

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // By default return a promise that never resolves → we test the loading state
    mockGetDashboard.mockReturnValue(new Promise(() => {}));
  });

  it('renders the page title and loading state without crashing', () => {
    render(
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('dashboard.title')).toBeInTheDocument();
    expect(screen.getByText('dashboard.welcome')).toBeInTheDocument();
  });

  it('renders stat cards after data loads', async () => {
    mockGetDashboard.mockResolvedValue({
      warehouse_count: 3,
      sku_count: 42,
      inventory_stats: { total_records: 1500 },
      order_summary: { completed: 10, pending: 5, cancelled: 2 },
      task_summary: { pending: 12, in_progress: 8, completed: 30, cancelled: 1 },
    });

    render(
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>,
    );

    expect(await screen.findByText('dashboard.warehouses')).toBeInTheDocument();
    expect(screen.getByText('dashboard.skus')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
  });
});
