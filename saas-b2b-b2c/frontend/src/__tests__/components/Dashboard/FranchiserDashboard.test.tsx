import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import FranchiserDashboard from '@/components/Dashboard/FranchiserDashboard';

jest.mock('@/store/franchiserStore', () => ({
  useFranchiserStore: () => ({
    summary: {
      planPercent: 78,
      forecastPercent: 85,
      activeDealers: 24,
      avgConversion: 12,
      avgMargin: 32,
    },
    isLoading: false,
    alertCount: 2,
    fetchSummary: jest.fn(),
  }),
}));

jest.mock('@/store/authSlice', () => ({
  logout: jest.fn(),
}));

jest.mock('next/router', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}));

const mockUser = {
  id: '1',
  email: 'test@example.com',
  first_name: 'Иван',
  last_name: 'Петров',
  role: 'franchiser',
};

describe('FranchiserDashboard', () => {
  it('renders header with correct title', () => {
    render(<FranchiserDashboard user={mockUser} title="Тест" />);
    expect(screen.getByText('Руководитель отдела франчайзинга')).toBeInTheDocument();
  });

  it('renders KPIs in header', () => {
    render(<FranchiserDashboard user={mockUser} />);
    const content = document.body.textContent || '';
    expect(content).toContain('Выполнение плана');
    expect(content).toContain('Прогноз квартала');
    expect(content).toContain('Активных дилеров');
    expect(content).toContain('Конверсия');
    expect(content).toContain('Маржинальность');
  });

  it('renders tabs', () => {
    render(<FranchiserDashboard user={mockUser} />);
    const content = document.body.textContent || '';
    expect(content).toContain('Пульт сети');
    expect(content).toContain('Моя команда');
    expect(content).toContain('Здоровье сети');
    expect(content).toContain('Отчёт для B2B');
  });

  it('renders avatar', () => {
    render(<FranchiserDashboard user={mockUser} />);
    const avatar = document.querySelector('.ant-avatar');
    expect(avatar).toBeInTheDocument();
  });

  it('has date display', () => {
    render(<FranchiserDashboard user={mockUser} />);
    const dateText = document.body.textContent;
    expect(dateText).toMatch(/\d{4}/);
  });
});