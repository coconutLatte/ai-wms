// Smoke test: NotFound page renders without crashing.
// Uses MemoryRouter since the component calls useNavigate.

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

    // The page should display a 404 title and a back-to-dashboard button
    expect(screen.getByText('notFound.title')).toBeInTheDocument();
    expect(screen.getByText('notFound.backToDashboard')).toBeInTheDocument();
  });
});
