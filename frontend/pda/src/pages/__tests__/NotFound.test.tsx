// Smoke test: PDA NotFound page renders without crashing.

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import NotFoundPage from '@/pages/NotFound';

describe('NotFoundPage', () => {
  it('renders the 404 page without crashing', () => {
    render(
      <MemoryRouter>
        <NotFoundPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('notFound.title')).toBeInTheDocument();
    expect(screen.getByText('notFound.message')).toBeInTheDocument();
    expect(screen.getByText('notFound.goToTasks')).toBeInTheDocument();
  });
});
