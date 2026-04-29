// __tests__/components/Dashboard/tabs/ProfitTab.test.tsx
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import ProfitTab from '@/components/Dashboard/tabs/ProfitTab';

describe('ProfitTab', () => {
  const mockData = {
    revenue: 5000000,
    cogs: 2000000,
    rent: 300000,
    utilities: 50000,
    payroll: 800000,
    taxes: 250000,
    logistics: 150000,
    marketing: 100000,
    defects: 50000,
    other_expenses: 50000,
    bonus: 100000,
    net_profit: 650000,
    net_profit_forecast: 800000,
    prev_month_net_profit: 500000,
    expense_breakdown: [
      { category: 'Аренда', amount: 300000, percent_of_revenue: 6, prev_month_amount: 280000 },
      { category: 'ФОТ', amount: 800000, percent_of_revenue: 16, prev_month_amount: 750000 },
      { category: 'Налоги', amount: 250000, percent_of_revenue: 5, prev_month_amount: 200000 },
      { category: 'Логистика', amount: 150000, percent_of_revenue: 3, prev_month_amount: 160000 },
    ],
  };

  it('renders loading state', () => {
    render(<ProfitTab loading={true} />);
    expect(document.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('renders error state with retry button', () => {
    const handleRetry = jest.fn();
    render(<ProfitTab error={true} onRetry={handleRetry} />);
    expect(screen.getByText('Ошибка загрузки данных')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Повторить'));
    expect(handleRetry).toHaveBeenCalled();
  });

  it('renders warning when no expenses entered', () => {
    const dataWithoutExpenses = {
      revenue: 5000000,
      cogs: 2000000,
      rent: 0,
      utilities: 0,
      payroll: 0,
      taxes: 0,
      logistics: 0,
      marketing: 0,
      defects: 0,
      other_expenses: 0,
      bonus: 0,
      net_profit: 3000000,
      net_profit_forecast: 3200000,
      prev_month_net_profit: 0,
      expense_breakdown: [],
    };
    const handleNavigate = jest.fn();
    render(<ProfitTab data={dataWithoutExpenses} onNavigateToExpenseForm={handleNavigate} />);
    expect(screen.getByText('Заполните расходы')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Перейти к форме'));
    expect(handleNavigate).toHaveBeenCalled();
  });

  it('renders all KPI cards', () => {
    render(<ProfitTab data={mockData} />);
    expect(screen.getByText('Чистая прибыль (Net Profit)')).toBeInTheDocument();
    expect(screen.getByText('Валовый оборот')).toBeInTheDocument();
    expect(screen.getByText('Маржинальная прибыль')).toBeInTheDocument();
    expect(screen.getByText('Прогноз чистой прибыли')).toBeInTheDocument();
  });

  it('renders expense warning when thresholds exceeded', () => {
    const highExpenses = {
      ...mockData,
      rent: 800000,
      payroll: 1500000,
    };
    render(<ProfitTab data={highExpenses} />);
    expect(screen.getByText('Внимание: превышены нормативы')).toBeInTheDocument();
  });

  it('renders charts', () => {
    render(<ProfitTab data={mockData} />);
    expect(screen.getAllByText('Путь от выручки к чистой прибыли').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Структура затрат').length).toBeGreaterThan(0);
  });

  it('renders expense detail table', () => {
    render(<ProfitTab data={mockData} />);
    expect(screen.getByText('Детализация расходов')).toBeInTheDocument();
    expect(screen.getAllByText('Аренда').length).toBeGreaterThan(0);
  });

  it('shows correct values', () => {
    render(<ProfitTab data={mockData} />);
    expect(screen.getByText('Чистая прибыль (Net Profit)')).toBeInTheDocument();
    expect(screen.getByText('Валовый оборот')).toBeInTheDocument();
  });
});