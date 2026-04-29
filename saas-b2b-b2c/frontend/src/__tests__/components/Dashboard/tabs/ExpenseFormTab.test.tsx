// __tests__/components/Dashboard/tabs/ExpenseFormTab.test.tsx
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import ExpenseFormTab from '@/components/Dashboard/tabs/ExpenseFormTab';

global.fetch = jest.fn(() => Promise.resolve({ ok: true, json: () => Promise.resolve({}) })) as jest.Mock;

describe('ExpenseFormTab', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear().mockResolvedValue({
      ok: true,
      json: async () => ({}),
    });
  });

  it('renders without crashing', async () => {
    render(<ExpenseFormTab />);
    await waitFor(() => {
      expect(screen.getByText(/Месяц:/)).toBeInTheDocument();
    });
  });

  it('renders form fields', async () => {
    render(<ExpenseFormTab />);
    await waitFor(() => {
      expect(screen.getByText(/Аренда помещения/)).toBeInTheDocument();
    });
  });

  it('renders action buttons', async () => {
    render(<ExpenseFormTab />);
    await waitFor(() => {
      expect(screen.getByText(/Импорт из выписки/)).toBeInTheDocument();
    });
  });

  it('renders tax selectors', async () => {
    render(<ExpenseFormTab />);
    await waitFor(() => {
      expect(screen.getByText(/УСН/)).toBeInTheDocument();
    });
  });

  it('renders save action', async () => {
    render(<ExpenseFormTab />);
    await waitFor(() => {
      expect(screen.getByText(/Сохранить/)).toBeInTheDocument();
    });
  });
});