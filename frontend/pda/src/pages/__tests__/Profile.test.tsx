// Smoke test: PDA Profile page renders operator info and logout button.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ProfilePage from '@/pages/Profile';

// ── Mock useAuth hook ──────────────────────────────────────────────────

const mockClearTokens = vi.fn();

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    clearTokens: mockClearTokens,
  }),
}));

describe('ProfilePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the profile page without crashing', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>,
    );

    expect(screen.getByText('profile.operator')).toBeInTheDocument();
    expect(screen.getByText('profile.warehouseOps')).toBeInTheDocument();
  });

  it('displays operator details', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>,
    );

    expect(screen.getByText('profile.operatorId')).toBeInTheDocument();
    expect(screen.getByText('op-001')).toBeInTheDocument();
    expect(screen.getByText('Demo Warehouse')).toBeInTheDocument();
  });

  it('renders the sign-out button', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>,
    );

    expect(screen.getByText('profile.signOut')).toBeInTheDocument();
  });
});
