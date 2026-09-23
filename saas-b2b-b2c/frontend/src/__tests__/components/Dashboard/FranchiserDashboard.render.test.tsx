import React from 'react';
import { render, screen } from '@testing-library/react';
import FranchiserDashboard from '@/components/Dashboard/FranchiserDashboard';

jest.mock('next/router', () => ({
  useRouter: () => ({ push: jest.fn() }),
}));

jest.mock('react-redux', () => ({
  useDispatch: () => jest.fn(),
}));

jest.mock('@/store/authSlice', () => ({
  logout: () => ({ type: 'auth/logout' }),
}));

jest.mock('@/store/franchiserStore', () => ({
  useFranchiserStore: () => ({
    summary: {
      planPercent: 87,
      forecastPercent: 102,
      activeDealers: 15,
      avgConversion: 31,
      avgMargin: 28,
    },
    isLoading: false,
    alertCount: 2,
    fetchSummary: jest.fn(),
  }),
}));

jest.mock('@/components/Dashboard/tabs/FranchiserNetworkTab', () => {
  const Tab: React.FC = () => <div>Пульт сети</div>;
  return { __esModule: true, default: Tab };
});

const user = { id: 'u1', email: 'fr@x.ru', first_name: 'Франчайзер', role: 'franchiser' };

describe('FranchiserDashboard (render)', () => {
  it('отображает заголовок и метрики сети', () => {
    render(<FranchiserDashboard user={user} />);
    expect(screen.getByText('Руководитель отдела франчайзинга')).toBeInTheDocument();
    expect(screen.getByText('Выполнение плана')).toBeInTheDocument();
    expect(screen.getByText('Прогноз квартала')).toBeInTheDocument();
    expect(screen.getByText('Активных дилеров')).toBeInTheDocument();
    expect(screen.getByText('Конверсия сети')).toBeInTheDocument();
    expect(screen.getByText('Маржинальность')).toBeInTheDocument();
    expect(screen.getByText('87')).toBeInTheDocument();
    expect(screen.getByText('102')).toBeInTheDocument();
  });
});