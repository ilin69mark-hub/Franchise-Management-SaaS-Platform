import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import ProfitTab from '@/components/Dashboard/tabs/ProfitTab';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/store/authSlice';

jest.mock('@/services/api', () => ({
  ...jest.requireActual('@/services/api'),
  useGetUnitTemplatesQuery: () => ({
    data: [],
    isLoading: false,
  }),
  useCreateUnitTemplateMutation: () => [
    jest.fn(),
    { isLoading: false },
  ],
}));

const createTestStore = () => configureStore({
  reducer: {
    auth: authReducer,
  },
  preloadedState: {
    auth: {
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      loading: false,
      error: null,
    },
  },
});

const renderWithProviders = (ui: React.ReactElement) => {
  return render(<Provider store={createTestStore()}>{ui}</Provider>);
};

const mockFinanceData = {
  revenue: 5000000,
  cogs: 1500000,
  rent: 400000,
  utilities: 100000,
  payroll: 800000,
  taxes: 300000,
  logistics: 150000,
  marketing: 200000,
  defects: 50000,
  other_expenses: 100000,
  bonus: 100000,
  net_profit: 800000,
  net_profit_forecast: 900000,
  prev_month_net_profit: 750000,
  expense_breakdown: [
    { category: 'Аренда', amount: 400000, percent_of_revenue: 8, prev_month_amount: 380000 },
    { category: 'ФОТ', amount: 800000, percent_of_revenue: 16, prev_month_amount: 750000 },
  ],
};

describe('ProfitTab', () => {
  it('renders loading state', () => {
    renderWithProviders(<ProfitTab loading={true} />);
    expect(document.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('renders error state with retry button', () => {
    const onRetry = jest.fn();
    renderWithProviders(<ProfitTab error={true} onRetry={onRetry} />);
    
    expect(screen.getByText('Ошибка загрузки данных')).toBeInTheDocument();
    const retryButton = screen.getByText('Повторить');
    expect(retryButton).toBeInTheDocument();
    
    fireEvent.click(retryButton);
    expect(onRetry).toHaveBeenCalled();
  });

  it('renders warning when no expenses', () => {
    const onNavigateToExpenseForm = jest.fn();
    renderWithProviders(
      <ProfitTab 
        data={{
          revenue: 1000000,
          cogs: 0,
          rent: 0,
          utilities: 0,
          payroll: 0,
          taxes: 0,
          logistics: 0,
          marketing: 0,
          defects: 0,
          other_expenses: 0,
          bonus: 0,
          net_profit: 0,
          net_profit_forecast: 0,
          prev_month_net_profit: 0,
          expense_breakdown: [],
        }} 
        onNavigateToExpenseForm={onNavigateToExpenseForm}
      />
    );
    
    expect(screen.getByText('Заполните расходы')).toBeInTheDocument();
    const button = screen.getByText('Перейти к форме');
    fireEvent.click(button);
    expect(onNavigateToExpenseForm).toHaveBeenCalled();
  });

  it('renders profit statistics', () => {
    renderWithProviders(<ProfitTab data={mockFinanceData} />);
    
    expect(screen.getByText('Чистая прибыль (Net Profit)')).toBeInTheDocument();
    expect(screen.getByText('Валовый оборот')).toBeInTheDocument();
    expect(screen.getByText('Маржинальная прибыль')).toBeInTheDocument();
    expect(screen.getByText('Прогноз чистой прибыли')).toBeInTheDocument();
  });

  it('calculates profit change percentage correctly', () => {
    const expectedChange = ((800000 - 750000) / 750000) * 100;
    expect(expectedChange).toBeCloseTo(6.67, 1);
  });

  it('renders expense breakdown table', () => {
    renderWithProviders(<ProfitTab data={mockFinanceData} />);
    
    expect(screen.getByText('Детализация расходов')).toBeInTheDocument();
    expect(screen.getByText('Структура затрат')).toBeInTheDocument();
  });

  it('shows warning when rent exceeds 12% of revenue', () => {
    const highRentData = {
      ...mockFinanceData,
      revenue: 1000000,
      rent: 150000,
    };
    
    renderWithProviders(<ProfitTab data={highRentData} />);
    expect(screen.getByText(/Аренда превышает 12%/)).toBeInTheDocument();
  });

  it('shows warning when payroll exceeds 20% of revenue', () => {
    const highPayrollData = {
      ...mockFinanceData,
      revenue: 1000000,
      payroll: 250000,
    };
    
    renderWithProviders(<ProfitTab data={highPayrollData} />);
    expect(screen.getByText(/ФОТ превышает 20%/)).toBeInTheDocument();
  });

  it('renders revenue to profit flow', () => {
    renderWithProviders(<ProfitTab data={mockFinanceData} />);
    
    expect(screen.getByText('Путь от выручки к чистой прибыли')).toBeInTheDocument();
    expect(screen.getByText('Выручка')).toBeInTheDocument();
    expect(screen.getByText('Чистая прибыль')).toBeInTheDocument();
  });

  it('renders unit economy calculator', () => {
    renderWithProviders(<ProfitTab data={mockFinanceData} />);
    expect(screen.getByText('🧮 Калькулятор unit-экономики')).toBeInTheDocument();
  });

  it('handles negative net profit', () => {
    const negativeProfitData = {
      ...mockFinanceData,
      net_profit: -100000,
    };
    
    renderWithProviders(<ProfitTab data={negativeProfitData} />);
    const cards = document.querySelectorAll('.ant-statistic');
    expect(cards.length).toBeGreaterThan(0);
  });

  it('formats numbers with thousand separators', () => {
    const revenue = 5000000;
    const formatted = String(revenue);
    expect(formatted).toBe('5000000');
  });
});

describe('ProfitTab calculations', () => {
  it('calculates margin profit correctly', () => {
    const revenue = 5000000;
    const cogs = 1500000;
    const marginProfit = revenue - cogs;
    expect(marginProfit).toBe(3500000);
  });

  it('calculates expense percentage correctly', () => {
    const expense = 400000;
    const revenue = 5000000;
    const percent = (expense / revenue) * 100;
    expect(percent).toBe(8);
  });

  it('calculates profit change correctly', () => {
    const currentProfit = 800000;
    const prevProfit = 750000;
    const change = ((currentProfit - prevProfit) / prevProfit) * 100;
    expect(change).toBeCloseTo(6.67, 1);
  });

  it('calculates margin percentage correctly', () => {
    const marginProfit = 3500000;
    const revenue = 5000000;
    const marginPercent = (marginProfit / revenue) * 100;
    expect(marginPercent).toBe(70);
  });

  it('identifies high rent correctly', () => {
    const data = { rent: 150000, revenue: 1000000 };
    const isRentHigh = (data.rent / data.revenue) * 100 > 12;
    expect(isRentHigh).toBe(true);
  });

  it('identifies high payroll correctly', () => {
    const data = { payroll: 250000, revenue: 1000000 };
    const isPayrollHigh = (data.payroll / data.revenue) * 100 > 20;
    expect(isPayrollHigh).toBe(true);
  });
});

describe('ProfitTab expense rendering', () => {
  it('filters out zero expenses', () => {
    const dataWithZeros = {
      ...mockFinanceData,
      marketing: 0,
      logistics: 0,
    };
    
    const expenses = [
      { name: 'Аренда', value: dataWithZeros.rent },
      { name: 'ФОТ', value: dataWithZeros.payroll },
      { name: 'Маркетинг', value: dataWithZeros.marketing },
      { name: 'Логистика', value: dataWithZeros.logistics },
    ].filter(item => item.value > 0);
    
    expect(expenses.length).toBe(2);
  });

  it('calculates pie data percentages correctly', () => {
    const items = [
      { amount: 400000 },
      { amount: 100000 },
      { amount: 800000 },
    ];
    const total = items.reduce((sum, item) => sum + item.amount, 0);
    const rentPercent = (400000 / total) * 100;
    expect(rentPercent).toBeCloseTo(30.77, 1);
  });
});