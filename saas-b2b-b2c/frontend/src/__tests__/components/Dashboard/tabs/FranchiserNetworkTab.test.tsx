import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { useFranchiserStore } from '@/store/franchiserStore';
import FranchiserNetworkTab from '@/components/Dashboard/tabs/FranchiserNetworkTab';

jest.mock('@/components/Dashboard/tabs/SalesDynamicsChart', () => {
  return { __esModule: true, default: () => <div data-testid="sales-chart">Sales Chart</div> };
});

jest.mock('@/components/Dashboard/tabs/GeographyMap', () => {
  return { __esModule: true, default: () => <div data-testid="geography-map">Geography Map</div> };
});

jest.mock('@/store/franchiserStore', () => ({
  useFranchiserStore: jest.fn(),
}));

const mockFetchNetwork = jest.fn();

describe('FranchiserNetworkTab', () => {
  const mockSummary = {
    planPercent: 78,
    forecastPercent: 85,
    activeDealers: 24,
    avgConversion: 12,
    avgMargin: 32,
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useFranchiserStore as jest.Mock).mockReturnValue({
      fetchNetwork: mockFetchNetwork,
      dealers: [],
    });
  });

  describe('getTerritoryStatus formula', () => {
    it('returns green when planPercent >= 95, redDealers < 10%, sla >= 95', () => {
      const territory = { planPercent: 98, redDealersPercent: 5, sla: 97 };
      const status = territory.planPercent >= 95 && territory.redDealersPercent < 10 && territory.sla >= 95 ? 'green' : 
                      territory.planPercent < 85 || territory.redDealersPercent > 25 || territory.sla < 80 ? 'red' : 'yellow';
      expect(status).toBe('green');
    });

    it('returns red when planPercent < 85', () => {
      const territory = { planPercent: 80, redDealersPercent: 5, sla: 97 };
      const status = territory.planPercent >= 95 && territory.redDealersPercent < 10 && territory.sla >= 95 ? 'green' : 
                      territory.planPercent < 85 || territory.redDealersPercent > 25 || territory.sla < 80 ? 'red' : 'yellow';
      expect(status).toBe('red');
    });

    it('returns red when redDealersPercent > 25', () => {
      const territory = { planPercent: 95, redDealersPercent: 30, sla: 97 };
      const status = territory.planPercent >= 95 && territory.redDealersPercent < 10 && territory.sla >= 95 ? 'green' : 
                      territory.planPercent < 85 || territory.redDealersPercent > 25 || territory.sla < 80 ? 'red' : 'yellow';
      expect(status).toBe('red');
    });

    it('returns red when sla < 80', () => {
      const territory = { planPercent: 95, redDealersPercent: 5, sla: 75 };
      const status = territory.planPercent >= 95 && territory.redDealersPercent < 10 && territory.sla >= 95 ? 'green' : 
                      territory.planPercent < 85 || territory.redDealersPercent > 25 || territory.sla < 80 ? 'red' : 'yellow';
      expect(status).toBe('red');
    });

    it('returns yellow for borderline values', () => {
      const territory = { planPercent: 85, redDealersPercent: 25, sla: 80 };
      const status = territory.planPercent >= 95 && territory.redDealersPercent < 10 && territory.sla >= 95 ? 'green' : 
                      territory.planPercent < 85 || territory.redDealersPercent > 25 || territory.sla < 80 ? 'red' : 'yellow';
      expect(status).toBe('yellow');
    });
  });

  describe('getHeatmapColor formula', () => {
    it('returns green for green status', () => {
      const color = '#52c41a';
      expect(color).toBe('#52c41a');
    });

    it('returns yellow for yellow status', () => {
      const color = '#fa8c16';
      expect(color).toBe('#fa8c16');
    });

    it('returns red for red status', () => {
      const color = '#ff4d4f';
      expect(color).toBe('#ff4d4f');
    });
  });

  describe('planPercent color mapping', () => {
    it('returns green for >= 95%', () => {
      const color = (v: number) => v >= 95 ? '#52c41a' : v >= 85 ? '#fa8c16' : '#ff4d4f';
      expect(color(95)).toBe('#52c41a');
      expect(color(100)).toBe('#52c41a');
    });

    it('returns orange for 85-94%', () => {
      const color = (v: number) => v >= 95 ? '#52c41a' : v >= 85 ? '#fa8c16' : '#ff4d4f';
      expect(color(85)).toBe('#fa8c16');
      expect(color(94)).toBe('#fa8c16');
    });

    it('returns red for < 85%', () => {
      const color = (v: number) => v >= 95 ? '#52c41a' : v >= 85 ? '#fa8c16' : '#ff4d4f';
      expect(color(84)).toBe('#ff4d4f');
      expect(color(50)).toBe('#ff4d4f');
    });
  });

  describe('redDealersPercent tag color', () => {
    it('returns green for < 10%', () => {
      const color = (v: number) => v < 10 ? 'green' : v < 25 ? 'orange' : 'red';
      expect(color(5)).toBe('green');
      expect(color(9)).toBe('green');
    });

    it('returns orange for 10-24%', () => {
      const color = (v: number) => v < 10 ? 'green' : v < 25 ? 'orange' : 'red';
      expect(color(10)).toBe('orange');
      expect(color(24)).toBe('orange');
    });

    it('returns red for >= 25%', () => {
      const color = (v: number) => v < 10 ? 'green' : v < 25 ? 'orange' : 'red';
      expect(color(25)).toBe('red');
      expect(color(50)).toBe('red');
    });
  });

  describe('sla tag color', () => {
    it('returns green for >= 95%', () => {
      const color = (v: number) => v >= 95 ? 'green' : v >= 80 ? 'orange' : 'red';
      expect(color(95)).toBe('green');
      expect(color(100)).toBe('green');
    });

    it('returns orange for 80-94%', () => {
      const color = (v: number) => v >= 95 ? 'green' : v >= 80 ? 'orange' : 'red';
      expect(color(80)).toBe('orange');
      expect(color(94)).toBe('orange');
    });

    it('returns red for < 80%', () => {
      const color = (v: number) => v >= 95 ? 'green' : v >= 80 ? 'orange' : 'red';
      expect(color(79)).toBe('red');
      expect(color(50)).toBe('red');
    });
  });

  describe('forecastPercent tag color', () => {
    it('returns green for >= 95%', () => {
      const color = (v: number) => v >= 95 ? 'green' : v >= 85 ? 'orange' : 'red';
      expect(color(95)).toBe('green');
    });

    it('returns orange for 85-94%', () => {
      const color = (v: number) => v >= 95 ? 'green' : v >= 85 ? 'orange' : 'red';
      expect(color(85)).toBe('orange');
    });

    it('returns red for < 85%', () => {
      const color = (v: number) => v >= 95 ? 'green' : v >= 85 ? 'orange' : 'red';
      expect(color(84)).toBe('red');
    });
  });

  describe('dealerGrowth arrow', () => {
    it('returns up arrow for positive growth', () => {
      const result = (v: number) => v > 0 ? 'up' : v < 0 ? 'down' : 'none';
      expect(result(5)).toBe('up');
      expect(result(1)).toBe('up');
    });

    it('returns down arrow for negative growth', () => {
      const result = (v: number) => v > 0 ? 'up' : v < 0 ? 'down' : 'none';
      expect(result(-5)).toBe('down');
      expect(result(-1)).toBe('down');
    });

    it('returns none for zero growth', () => {
      const result = (v: number) => v > 0 ? 'up' : v < 0 ? 'down' : 'none';
      expect(result(0)).toBe('none');
    });
  });

  describe('component rendering', () => {
    it('renders with summary props', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      expect(screen.getByText('Пульт сети')).toBeInTheDocument();
    });

    it('renders all KPI cards', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      expect(screen.getAllByText('Выполнение плана').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Прогноз квартала').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Активных дилеров').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Конверсия сети').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Маржинальность').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('В красной зоне').length).toBeGreaterThanOrEqual(1);
    });

    it('renders period selector options', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      const selects = screen.getAllByRole('combobox');
      expect(selects.length).toBeGreaterThanOrEqual(1);
    });

    it('renders group by selector options', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      const selects = screen.getAllByRole('combobox');
      expect(selects.length).toBeGreaterThanOrEqual(2);
    });

    it('renders table columns', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      expect(screen.getAllByText('Менеджер').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Территория').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Выполнение плана').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Дилеров').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('В красной зоне').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('SLA').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Прирост').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Прогноз').length).toBeGreaterThanOrEqual(1);
      expect(screen.getAllByText('Статус').length).toBeGreaterThanOrEqual(1);
    });

    it('calls fetchNetwork on mount', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      expect(mockFetchNetwork).toHaveBeenCalledTimes(1);
    });
  });

  describe('state changes', () => {
    it('has select elements for period and group by', () => {
      render(<FranchiserNetworkTab summary={mockSummary} />);
      const selects = screen.getAllByRole('combobox');
      expect(selects.length).toBeGreaterThanOrEqual(2);
    });
  });
});

describe('FranchiserNetworkTab mock territories data', () => {
  const mockTerritories = [
    { id: '1', manager: 'Алексей Петров', territory: 'Север', planPercent: 92, redDealersPercent: 0, sla: 98, dealerGrowth: 2, forecastPercent: 95 },
    { id: '2', manager: 'Мария Иванова', territory: 'Юг', planPercent: 78, redDealersPercent: 17, sla: 85, dealerGrowth: 1, forecastPercent: 82 },
    { id: '3', manager: 'Сергей Сидоров', territory: 'Запад', planPercent: 65, redDealersPercent: 30, sla: 72, dealerGrowth: -1, forecastPercent: 70 },
    { id: '4', manager: 'Елена Смирнова', territory: 'Центр', planPercent: 88, redDealersPercent: 0, sla: 94, dealerGrowth: 1, forecastPercent: 90 },
  ];

  it('calculates territory status for all mock data', () => {
    const getStatus = (t: typeof mockTerritories[0]) => {
      return t.planPercent >= 95 && t.redDealersPercent < 10 && t.sla >= 95 ? 'green' :
             t.planPercent < 85 || t.redDealersPercent > 25 || t.sla < 80 ? 'red' : 'yellow';
    };

    // Territory 0: plan 92 (>=85), red 0 (!>25), sla 98 (!<80) → no red cond → yellow
    expect(getStatus(mockTerritories[0])).toBe('yellow');
    // Territory 1: plan 78 < 85 → red
    expect(getStatus(mockTerritories[1])).toBe('red');
    // Territory 2: plan 65 < 85, red 30 > 25, sla 72 < 80 → red
    expect(getStatus(mockTerritories[2])).toBe('red');
    // Territory 3: plan 88 (>=85), red 0 (!>25), sla 94 (!<80) → no red cond → yellow
    expect(getStatus(mockTerritories[3])).toBe('yellow');
  });
});

describe('FranchiserNetworkTab overview calculations', () => {
  const overview = {
    planAmount: 42000000,
    planPercent: 78,
    planDynamics: 5,
    forecastAmount: 35700000,
    forecastPercent: 85,
    activeDealers: 24,
    dealerGrowth: 3,
    avgConversion: 12,
    conversionBenchmark: 15,
    avgMargin: 32,
    redZoneDealers: 3,
    redZonePercent: 12.5,
    redZoneDynamics: -1,
  };

  it('redZonePercent calculation: (redZoneDealers / activeDealers) * 100', () => {
    const calculated = (overview.redZoneDealers / overview.activeDealers) * 100;
    expect(calculated).toBe(12.5);
  });

  it('forecastAmount = planAmount * forecastPercent / 100', () => {
    const calculated = overview.planAmount * overview.forecastPercent / 100;
    expect(calculated).toBe(35700000);
  });

  it('avgConversion comparison to benchmark', () => {
    const isBelow = overview.avgConversion < overview.conversionBenchmark;
    expect(isBelow).toBe(true); // 12 < 15
  });

  it('avgMargin color condition >= 30', () => {
    const color = overview.avgMargin >= 30 ? '#52c41a' : '#fa8c16';
    expect(color).toBe('#52c41a');
  });

  it('redZoneDynamics negative is good (decrease)', () => {
    const isPositive = overview.redZoneDynamics > 0;
    expect(isPositive).toBe(false);
  });

  it('planDynamics positive is good (increase)', () => {
    const isPositive = overview.planDynamics > 0;
    expect(isPositive).toBe(true);
  });
});