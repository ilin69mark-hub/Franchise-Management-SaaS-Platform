import React from 'react';
import { render, screen } from '@testing-library/react';
import FranchiserNetworkTab from '@/components/Dashboard/tabs/FranchiserNetworkTab';

jest.mock('@/components/Dashboard/tabs/SalesDynamicsChart', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="sales-chart">Sales Chart</div>,
  };
});

describe('FranchiserNetworkTab', () => {
  const mockSummary = {
    planPercent: 78,
    forecastPercent: 85,
    activeDealers: 24,
    avgConversion: 12,
    avgMargin: 32,
  };

  it('renders title', () => {
    render(<FranchiserNetworkTab summary={mockSummary} />);
    expect(screen.getByText('Пульт сети')).toBeInTheDocument();
  });

  it('renders KPI cards', () => {
    render(<FranchiserNetworkTab summary={mockSummary} />);
    const content = document.body.textContent || '';
    expect(content).toContain('Выполнение плана');
    expect(content).toContain('Прогноз квартала');
    expect(content).toContain('Активных дилеров');
    expect(content).toContain('Конверсия сети');
    expect(content).toContain('Маржинальность');
    expect(content).toContain('В красной зоне');
  });

  it('renders filters', () => {
    render(<FranchiserNetworkTab summary={mockSummary} />);
    const content = document.body.textContent || '';
    expect(content).toContain('Месяц');
    expect(content).toContain('По менеджерам');
  });

  it('renders sales dynamics chart section', () => {
    render(<FranchiserNetworkTab summary={mockSummary} />);
    expect(screen.getByText('Динамика продаж сети')).toBeInTheDocument();
  });

  it('renders heatmap table', () => {
    render(<FranchiserNetworkTab summary={mockSummary} />);
    expect(screen.getByText('Теплокарта территорий')).toBeInTheDocument();
    const content = document.body.textContent || '';
    expect(content).toContain('Алексей Петров');
    expect(content).toContain('Мария Иванова');
  });

  it('renders territory columns', () => {
    render(<FranchiserNetworkTab summary={mockSummary} />);
    const content = document.body.textContent || '';
    expect(content).toContain('Территория');
    expect(content).toContain('Дилеров');
    expect(content).toContain('В красной зоне');
    expect(content).toContain('SLA');
    expect(content).toContain('Прирост');
    expect(content).toContain('Прогноз');
  });
});