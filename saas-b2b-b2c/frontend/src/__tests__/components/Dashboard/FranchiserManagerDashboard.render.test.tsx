import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import FranchiserManagerDashboard from '@/components/Dashboard/FranchiserManagerDashboard';
import type { TerritorySummary } from '@/store/territoryManagerStore';

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

jest.mock('@/store/territoryManagerStore', () => ({
  useTerritoryManagerStore: () => mockStoreState,
}));

const mockStoreState = {
  activeTab: 'map',
  summary: null as TerritorySummary | null,
  isLoading: false,
  summaryModalOpen: false,
  setActiveTab: jest.fn(),
  setSummary: jest.fn((s: TerritorySummary) => { mockStoreState.summary = s; }),
  setLoading: jest.fn(),
  setLastUpdated: jest.fn(),
  setSummaryModalOpen: jest.fn(),
  setManager: jest.fn(),
};

jest.mock('@/components/Dashboard/tabs/TerritoryMapTab', () => {
  const Tab: React.FC = () => <div>Карта территории tab</div>;
  return { __esModule: true, default: Tab };
});

jest.mock('@/components/Dashboard/tabs/TerritoryFunnelTab', () => {
  const Tab: React.FC = () => <div>Воронка tab</div>;
  return { __esModule: true, default: Tab };
});

jest.mock('@/components/Dashboard/tabs/TerritoryPlanFactTab', () => {
  const Tab: React.FC = () => <div>План-факт tab</div>;
  return { __esModule: true, default: Tab };
});

jest.mock('@/components/Dashboard/tabs/TerritoryCommunicationsTab', () => {
  const Tab: React.FC = () => <div>Коммуникации tab</div>;
  return { __esModule: true, default: Tab };
});

jest.mock('@/components/Dashboard/tabs/TerritoryBenchmarkTab', () => {
  const Tab: React.FC = () => <div>Бенчмаркинг tab</div>;
  return { __esModule: true, default: Tab };
});

const mockGet = jest.requireMock('@/api/axiosClient').default.get;

const user = { id: 'u1', email: 'tm@x.ru', first_name: 'Иван', last_name: 'Территориалов', role: 'territory_manager', territoryName: 'Москва и область' };

describe('FranchiserManagerDashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockStoreState.summary = null;
    mockStoreState.summaryModalOpen = false;
    mockGet.mockResolvedValue({
      data: {
        planCompletionPercent: 82,
        quarterForecastPercent: 91,
        redZoneDealersCount: 2,
        avgConversion: 4.2,
        activeAlerts: 5,
      },
    });
  });

  it('отображает заголовок, метрики и вкладки', async () => {
    render(<FranchiserManagerDashboard user={user} title="Мой регион" />);
    expect(screen.getByText('Мой регион')).toBeInTheDocument();
    expect(await screen.findByText('Выполнение плана')).toBeInTheDocument();
    expect(screen.getByText('Прогноз квартала')).toBeInTheDocument();
    expect(screen.getByText('Дилеров в красной')).toBeInTheDocument();
    expect(screen.getByText('Средняя конверсия')).toBeInTheDocument();
    expect(screen.getByText('Карта территории')).toBeInTheDocument();
    expect(screen.getByText('Воронка')).toBeInTheDocument();
    expect(screen.getByText('План-факт и Прогноз')).toBeInTheDocument();
    expect(screen.getByText('Коммуникации и Задачи')).toBeInTheDocument();
    expect(screen.getByText('Бенчмаркинг')).toBeInTheDocument();
    expect(await screen.findByText('Карта территории tab')).toBeInTheDocument();
  });

  it('открывает модалку утренней сводки с данными из API', async () => {
    mockStoreState.summary = { planCompletionPercent: 82, quarterForecastPercent: 91, redZoneDealersCount: 2, avgConversion: 4.2, activeAlerts: 5 };
    mockStoreState.summaryModalOpen = true;
    render(<FranchiserManagerDashboard user={user} title="Мой регион" />);
    expect(await screen.findByText('Утренняя сводка')).toBeInTheDocument();
    expect(screen.getByText('Дилер "Мебель Москва" - план 68%')).toBeInTheDocument();
    expect(screen.getByText('Отставание от плана на 12%')).toBeInTheDocument();
    expect(screen.getByText('Новый дилер "МебельЛига"')).toBeInTheDocument();
    expect(screen.getByText('Выполнение плана:')).toBeInTheDocument();
    expect(screen.getByText('82%')).toBeInTheDocument();
    expect(screen.getByText('91%')).toBeInTheDocument();
    await waitFor(() => {
      expect(mockStoreState.setSummary).toHaveBeenCalledWith(expect.objectContaining({ planCompletionPercent: 82, quarterForecastPercent: 91 }));
    });
  });

  it('при ошибке API использует дефолтную сводку', async () => {
    mockStoreState.summaryModalOpen = true;
    mockGet.mockRejectedValue(new Error('network'));
    render(<FranchiserManagerDashboard user={user} title="Мой регион" />);
    await waitFor(() => {
      expect(mockStoreState.setSummary).toHaveBeenCalledWith(expect.objectContaining({ planCompletionPercent: 82, redZoneDealersCount: 2 }));
    });
  });
});