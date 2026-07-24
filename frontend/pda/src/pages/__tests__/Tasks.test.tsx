// Smoke test: PDA Tasks page renders the filter tabs and task list area.
// We test only the loading/empty states, not the full data flow.

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import TasksPage from '@/pages/Tasks';

// ── Mock @tanstack/react-query useQuery ────────────────────────────────

const mockUseQuery = vi.fn();

vi.mock('@tanstack/react-query', () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

describe('TasksPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Default: loading state
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
      refetch: vi.fn(),
      isFetching: false,
    });
  });

  it('renders filter tabs without crashing', () => {
    render(
      <MemoryRouter>
        <TasksPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('task.all')).toBeInTheDocument();
    expect(screen.getByText('task.pending')).toBeInTheDocument();
    expect(screen.getByText('task.inProgress')).toBeInTheDocument();
    expect(screen.getByText('task.done')).toBeInTheDocument();
  });

  it('shows loading text when fetching tasks', () => {
    render(
      <MemoryRouter>
        <TasksPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('task.loadingTasks')).toBeInTheDocument();
  });

  it('shows empty state when no tasks are returned', () => {
    mockUseQuery.mockReturnValue({
      data: { data: [] },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
      isFetching: false,
    });

    render(
      <MemoryRouter>
        <TasksPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('task.noTasks')).toBeInTheDocument();
    expect(screen.getByText('task.refresh')).toBeInTheDocument();
  });

  it('shows error state with retry button', () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Network error'),
      refetch: vi.fn(),
      isFetching: false,
    });

    render(
      <MemoryRouter>
        <TasksPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('task.failedToLoad')).toBeInTheDocument();
    expect(screen.getByText('task.retry')).toBeInTheDocument();
  });
});
