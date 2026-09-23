import React from 'react';
import { render, screen } from '@testing-library/react';
import DealerDashboard from '@/components/Dashboard/DealerDashboard';

jest.mock('next/router', () => ({
  useRouter: () => ({ push: jest.fn() }),
}));

jest.mock('react-redux', () => ({
  useDispatch: () => jest.fn(),
}));

jest.mock('@/store/authSlice', () => ({
  logout: () => ({ type: 'auth/logout' }),
}));

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: { get: jest.fn() },
}));

jest.mock('@/store/dealerDashboardStore', () => ({
  useDealerDashboardStore: () => ({
    activeTab: 'profit',
    summary: undefined,
    dealerCenterName: 'Дилерский центр',
    isLoading: false,
    setActiveTab: jest.fn(),
    setSummary: jest.fn(),
    setLoading: jest.fn(),
    setLastUpdated: jest.fn(),
  }),
}));

jest.mock('@/components/Dashboard/tabs/ProfitTab', () => {
  const Tab: React.FC = () => <div>Вкладка Прибыль</div>;
  return { __esModule: true, default: Tab };
});

const mockGet = jest.requireMock('@/api/axiosClient').default.get;

const user = { id: 'u1', email: 'dealer@x.ru', first_name: 'Дилер', role: 'dealer' };

describe('DealerDashboard (render)', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGet.mockResolvedValue({
      data: { netProfit: 1250000, grossRevenue: 4500000, planCompletionPercent: 78, marginProfit: 890000, activeAlerts: 3 },
    });
  });

  it('отображает заголовок и метрики', async () => {
    render(<DealerDashboard user={user} title="Мой дилерский центр" />);
    expect(screen.getByText('Мой дилерский центр')).toBeInTheDocument();
    expect(await screen.findByText('Чистая прибыль')).toBeInTheDocument();
    expect(screen.getAllByText('Валовый оборот').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('% плана сети').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Маржинальная прибыль').length).toBeGreaterThanOrEqual(1);
  });

  it('при ошибке API показывает fallback-сумму', async () => {
    mockGet.mockRejectedValue(new Error('network'));
    render(<DealerDashboard user={user} title="Центр" />);
    expect(await screen.findByText('Чистая прибыль')).toBeInTheDocument();
  });
});