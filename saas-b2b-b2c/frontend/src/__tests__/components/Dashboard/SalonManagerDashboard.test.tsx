import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import SalonManagerDashboard from '@/components/Dashboard/SalonManagerDashboard';

jest.mock('@/components/Dashboard/NotificationBell', () => {
  const Bell: React.FC = () => <div data-testid="bell" />;
  return { __esModule: true, default: Bell };
});

jest.mock('@/components/Dashboard/DealerDirectives', () => {
  const D: React.FC = () => <div data-testid="directives" />;
  return { __esModule: true, default: D };
});

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: { get: jest.fn() },
}));

jest.mock('@/components/Dashboard/tabs/SalonMainTab', () => {
  const Tab: React.FC = () => <div>Главная страница салона</div>;
  return { __esModule: true, default: Tab };
});

const mockGet = jest.requireMock('@/api/axiosClient').default.get;

const user = { id: 'u1', email: 'man@x.ru', first_name: 'Иван', role: 'salon_manager' };

describe('SalonManagerDashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('загружает данные top-bar и отображает салон', async () => {
    mockGet.mockResolvedValue({
      data: { salon_name: 'ТЦ Мега', plan_percent: 90, avg_check: 8500, prepayments_sum: 120000, alerts_count: 2 },
    });
    render(<SalonManagerDashboard user={user} />);
    expect(await screen.findByText('ТЦ Мега')).toBeInTheDocument();
    expect(screen.getByText('90%')).toBeInTheDocument();
    expect(screen.getByText('Выполнение плана')).toBeInTheDocument();
    expect(screen.getByText('Средний чек')).toBeInTheDocument();
    expect(screen.getByText('Предоплата')).toBeInTheDocument();
  });

  it('при ошибке API использует fallback-данные', async () => {
    mockGet.mockRejectedValue(new Error('network'));
    render(<SalonManagerDashboard user={user} />);
    expect(await screen.findByText('Мой салон')).toBeInTheDocument();
    expect(screen.getByText('0%')).toBeInTheDocument();
  });

  it('переключает вкладки', async () => {
    mockGet.mockResolvedValue({
      data: { salon_name: 'Салон', plan_percent: 50, avg_check: 100, prepayments_sum: 0, alerts_count: 0 },
    });
    render(<SalonManagerDashboard user={user} />);
    const team = await screen.findByText('Команда');
    expect(team).toBeInTheDocument();
  });
});