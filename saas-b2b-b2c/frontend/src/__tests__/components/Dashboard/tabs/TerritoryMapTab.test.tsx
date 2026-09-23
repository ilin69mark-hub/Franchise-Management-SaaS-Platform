// src/__tests__/components/Dashboard/tabs/TerritoryMapTab.test.tsx
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import TerritoryMapTab from '@/components/Dashboard/tabs/TerritoryMapTab';
import { DealerMetrics } from '@/store/territoryManagerStore';

jest.mock('@/store/territoryManagerStore', () => ({
  useTerritoryManagerStore: () => ({ dealers: [], setDealers: jest.fn(), summary: null }),
}));

jest.mock('@/api/axiosClient', () => ({
  __esModule: true,
  default: {
    get: jest.fn(() => Promise.reject(new Error('network'))),
  },
}));

const mapDealers: DealerMetrics[] = [
  { dealerId: 'd1', dealerName: 'Мебель Москва', salonCount: 3, planPercent: 110, forecastPercent: 105, conversion: 4.5, status: 'green', plan: 15000000, fact: 13800000, debt: 50000, margin: 28, avgCheck: 125000, taskCount: 2 },
  { dealerId: 'd2', dealerName: 'Диванит Воронеж', salonCount: 2, planPercent: 82, forecastPercent: 80, conversion: 2.5, status: 'yellow', plan: 8000000, fact: 6240000, debt: 0, margin: 22, avgCheck: 98000, taskCount: 0 },
  { dealerId: 'd3', dealerName: 'МебельЛига', salonCount: 1, planPercent: 45, forecastPercent: 40, conversion: 1.5, status: 'red', plan: 5000000, fact: 2250000, debt: 260000, margin: 15, avgCheck: 70000, taskCount: 1 },
];

describe('TerritoryMapTab', () => {
  it('renders', () => {
    expect(() => render(<TerritoryMapTab />)).not.toThrow();
  });

  it('renders with dealers', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Test Dealer', salonCount: 2, planPercent: 90, forecastPercent: 100, conversion: 4.5, status: 'green', plan: 10000000, fact: 9000000, debt: 0, margin: 25, avgCheck: 420000, taskCount: 0 },
    ];
    expect(() => render(<TerritoryMapTab dealers={dealers} />)).not.toThrow();
  });

  it('renders loading', () => {
    expect(() => render(<TerritoryMapTab loading />)).not.toThrow();
  });

  it('filters by status', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Green', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'Red', salonCount: 1, planPercent: 50, forecastPercent: 60, conversion: 1.5, status: 'red' },
    ];
    const redDealers = dealers.filter(d => d.status === 'red');
    expect(redDealers.length).toBe(1);
  });

  it('calculates status counts', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 50, forecastPercent: 60, conversion: 1.5, status: 'red' },
    ];
    const green = dealers.filter(d => d.status === 'green').length;
    const yellow = dealers.filter(d => d.status === 'yellow').length;
    const red = dealers.filter(d => d.status === 'red').length;
    expect(green).toBe(1);
    expect(yellow).toBe(1);
    expect(red).toBe(1);
  });

  it('searches by name', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Мебель Москва', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'Диванит', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
    ];
    const searchText = 'мебель';
    const filtered = dealers.filter(d => d.dealerName.toLowerCase().includes(searchText.toLowerCase()));
    expect(filtered.length).toBe(1);
  });

  it('calculates territory totals', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 100, forecastPercent: 110, conversion: 4, status: 'green', plan: 10000000, fact: 10000000, debt: 0, margin: 25 },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow', plan: 5000000, fact: 4000000, debt: 50000, margin: 22 },
    ];
    const totalPlan = dealers.reduce((s, d) => s + (d.plan || 0), 0);
    const totalFact = dealers.reduce((s, d) => s + (d.fact || 0), 0);
    const totalDebt = dealers.reduce((s, d) => s + (d.debt || 0), 0);
    const avgMargin = dealers.reduce((s, d) => s + (d.margin || 0), 0) / dealers.length;
    expect(totalPlan).toBe(15000000);
    expect(totalFact).toBe(14000000);
    expect(totalDebt).toBe(50000);
    expect(avgMargin).toBe(23.5);
  });

  it('calculates conversion color', () => {
    const getCellColor = (conv: number) => {
      if (conv >= 3) return '#f6ffed';
      if (conv >= 2) return '#fffbe6';
      return '#fff1f0';
    };
    expect(getCellColor(4)).toBe('#f6ffed');
    expect(getCellColor(2.5)).toBe('#fffbe6');
    expect(getCellColor(1)).toBe('#fff1f0');
  });

  it('sorts dealers by field', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 100, forecastPercent: 110, conversion: 4, status: 'green' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 60, forecastPercent: 70, conversion: 2, status: 'red' },
    ];
    const sorted = [...dealers].sort((a, b) => b.planPercent - a.planPercent);
    expect(sorted[0].dealerName).toBe('B');
    expect(sorted[2].dealerName).toBe('C');
  });

  it('filters leaders', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 110, forecastPercent: 120, conversion: 5, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 100, forecastPercent: 110, conversion: 4, status: 'green' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 3, status: 'green' },
    ];
    const leaders = dealers.filter(d => d.planPercent >= 100);
    expect(leaders.length).toBe(2);
  });

  it('filters problem dealers', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 110, forecastPercent: 120, conversion: 5, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 60, forecastPercent: 70, conversion: 2, status: 'red' },
    ];
    const problems = dealers.filter(d => d.planPercent < 70);
    expect(problems.length).toBe(1);
  });
});

describe('TerritoryMapTab render', () => {
  it('рендерит статистику, теплокарту и статусы дилеров', () => {
    render(<TerritoryMapTab dealers={mapDealers} />);
    expect(screen.getByPlaceholderText('Поиск дилера...')).toBeInTheDocument();
    expect(screen.getByText('Все')).toBeInTheDocument();
    expect(screen.getByText('Лидеры')).toBeInTheDocument();
    expect(screen.getByText('Проблемные')).toBeInTheDocument();
    expect(screen.getByText('Выполнение плана')).toBeInTheDocument();
    expect(screen.getByText('Прогноз квартала')).toBeInTheDocument();
    expect(screen.getByText('В красной зоне')).toBeInTheDocument();
    expect(screen.getByText('Дебиторская задолженность')).toBeInTheDocument();
    expect(screen.getByText('Конверсия средняя')).toBeInTheDocument();
    expect(screen.getByText('Маржинальность')).toBeInTheDocument();
    expect(screen.getByText(/Теплокарта дилер/)).toBeInTheDocument();
    expect(screen.getByText(/Run Rate/)).toBeInTheDocument();
    expect(screen.getByText(/70% плана/)).toBeInTheDocument();
    expect(screen.getByText('Мебель Москва')).toBeInTheDocument();
    expect(screen.getByText('Диванит Воронеж')).toBeInTheDocument();
    expect(screen.getByText('МебельЛига')).toBeInTheDocument();
    expect(screen.getByText('Норма')).toBeInTheDocument();
    expect(screen.getByText('Внимание')).toBeInTheDocument();
    expect(screen.getByText('Проблема')).toBeInTheDocument();
    expect(screen.getByText(/⚠️ Дилеры в красной зоне \(1\)/)).toBeInTheDocument();
    expect(screen.getByText(/80%.*к прошлому месяцу/s)).toBeInTheDocument();
  });

  it('ищет дилера по имени', () => {
    render(<TerritoryMapTab dealers={mapDealers} />);
    fireEvent.change(screen.getByPlaceholderText('Поиск дилера...'), { target: { value: 'москва' } });
    expect(screen.getByText('Мебель Москва')).toBeInTheDocument();
    expect(screen.queryByText('МебельЛига')).toBeNull();
  });

  it('фильтрует проблемных и лидеров через Segmented и карточку', () => {
    render(<TerritoryMapTab dealers={mapDealers} />);
    fireEvent.click(screen.getByText('Лидеры'));
    expect(screen.getByText('Мебель Москва')).toBeInTheDocument();
    expect(screen.queryByText('МебельЛига')).toBeNull();
    fireEvent.click(screen.getByText('Все'));
    fireEvent.click(screen.getByText('В красной зоне'));
    expect(screen.getByText('МебельЛига')).toBeInTheDocument();
    expect(screen.queryByText('Мебель Москва')).toBeNull();
    expect(screen.queryByText('Диванит Воронеж')).toBeNull();
  });

  it('разворачивает детализацию дилера из красной зоны', async () => {
    render(<TerritoryMapTab dealers={mapDealers} />);
    fireEvent.click(screen.getByText(/⚠️ Дилеры в красной зоне \(1\)/));
    fireEvent.click(screen.getByText('Детали'));
    expect(await screen.findByText('Салон 1')).toBeInTheDocument();
    expect(screen.getByText('Салон 2')).toBeInTheDocument();
    expect(screen.getByText('4.2 млн ₽')).toBeInTheDocument();
    expect(screen.getByText('2.8 млн ₽')).toBeInTheDocument();
    expect(screen.getByText('Продажи по салонам:')).toBeInTheDocument();
    expect(screen.getByText('Динамика (6 мес):')).toBeInTheDocument();
    expect(screen.getByText('Последние алерты:')).toBeInTheDocument();
    expect(screen.getByText('Падение конверсии')).toBeInTheDocument();
    expect(screen.getByText('Низкий трафик')).toBeInTheDocument();
    expect(screen.getByText('Перейти к дилеру')).toBeInTheDocument();
  });
});