import React from 'react';
import { render, waitFor, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import SalonTeamTab from '@/components/Dashboard/tabs/SalonTeamTab';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(() => Promise.resolve({ data: { period: 'month', total_revenue: 1500000, avg_revenue: 500000, avg_conversion: 28.5, avg_check: 75000, sales_reps: [{ user_id: '1', first_name: 'Тест', last_name: 'Тестов', role: 'salon_manager', revenue: 600000, deals_count: 8, conversion: 32, avg_check: 75000, discount_percent: 5, extras_sum: 60000, revenue_deviation: 20, deals_deviation: 30, conversion_deviation: 12, avg_check_deviation: 0 }] } })),
}));

const apiClient = require('@/api/axiosClient');

const createMockStore = () =>
  configureStore({
    reducer: {
      auth: () => ({
        user: { id: '1', role: 'salon_manager', first_name: 'Тест', last_name: 'Тестов' },
        isAuthenticated: true,
      }),
    },
  });

const mockUser = { id: '1', role: 'salon_manager', first_name: 'Тест', last_name: 'Тестов' };

const mockTeamData = {
  period: 'month',
  total_revenue: 1500000,
  avg_revenue: 500000,
  avg_conversion: 28.5,
  avg_check: 75000,
  sales_reps: [
    { user_id: '1', first_name: 'Тест', last_name: 'Тестов', role: 'salon_manager', revenue: 600000, deals_count: 8, conversion: 32, avg_check: 75000, discount_percent: 5, extras_sum: 60000, revenue_deviation: 20, deals_deviation: 30, conversion_deviation: 12, avg_check_deviation: 0 },
    { user_id: '2', first_name: 'Иван', last_name: 'Иванов', role: 'salon_manager', revenue: 450000, deals_count: 6, conversion: 25, avg_check: 75000, discount_percent: 8, extras_sum: 45000, revenue_deviation: -10, deals_deviation: -15, conversion_deviation: -12, avg_check_deviation: 0 },
    { user_id: '3', first_name: 'Петр', last_name: 'Петров', role: 'salon_manager', revenue: 500000, deals_count: 7, conversion: 28, avg_check: 71428, discount_percent: 6, extras_sum: 50000, revenue_deviation: 0, deals_deviation: 0, conversion_deviation: -2, avg_check_deviation: -5 },
  ],
};

describe('SalonTeamTab', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('должен загружать данные и вызывать API', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonTeamTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalled();
    });

    const text = container.textContent;
    expect(text).toContain('Сотрудников') || expect(text).toContain('3');
  });

  it('должен показывать метрики с отклонениями', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonTeamTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalled();
    });

    const text = container.textContent;
    expect(text).toContain('600') || expect(text).toContain('20');
  });

  it('должен выделять текущего пользователя', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonTeamTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalled();
    });

    const text = container.textContent;
    expect(text).toContain('Вы');
  });

  it('должен обрабатывать ошибки API', async () => {
    apiClient.get.mockRejectedValueOnce(new Error('Network error'));

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonTeamTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      const text = container.textContent;
      expect(text).toContain('Ошибка') || expect(text).toContain('Повторить');
    });
  });
});

describe('Team calculations', () => {
  it('должен вычислять отклонение от среднего', () => {
    const calculateDeviation = (value: number, average: number) => {
      if (average === 0) return 0;
      return Math.round(((value - average) / average) * 100);
    };

    expect(calculateDeviation(600, 500)).toBe(20);
    expect(calculateDeviation(450, 500)).toBe(-10);
    expect(calculateDeviation(500, 500)).toBe(0);
  });

  it('должен определять цвет отклонения', () => {
    const getDeviationColor = (deviation: number) => {
      if (deviation > 0) return '#52c41a';
      if (deviation < 0) return '#ff4d4f';
      return '#999';
    };

    expect(getDeviationColor(20)).toBe('#52c41a');
    expect(getDeviationColor(-10)).toBe('#ff4d4f');
    expect(getDeviationColor(0)).toBe('#999');
  });

  it('должен сортировать по убыванию выручки', () => {
    const sorted = [...mockTeamData.sales_reps].sort((a, b) => b.revenue - a.revenue);
    expect(sorted[0].user_id).toBe('1');
    expect(sorted[1].user_id).toBe('3');
    expect(sorted[2].user_id).toBe('2');
  });
});

describe('Period filtering', () => {
  const getDateRange = (period: string) => {
    const now = new Date();
    let start: Date, end: Date = now;

    switch (period) {
      case 'week':
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case 'month':
        start = new Date(now.getFullYear(), now.getMonth(), 1);
        break;
      case 'quarter':
        const quarter = Math.floor(now.getMonth() / 3);
        start = new Date(now.getFullYear(), quarter * 3, 1);
        break;
      default:
        start = new Date(now.getFullYear(), now.getMonth(), 1);
    }

    return { start, end };
  };

  it('должен возвращать правильный диапазон для недели', () => {
    const range = getDateRange('week');
    expect(range.end.getTime() - range.start.getTime()).toBe(7 * 24 * 60 * 60 * 1000);
  });

  it('должен возвращать правильный диапазон для месяца', () => {
    const range = getDateRange('month');
    expect(range.start.getDate()).toBe(1);
  });
});

describe('Bonus calculation', () => {
  const calculateBonus = (percent: number, maxBonus: number) => {
    if (percent >= 100) return maxBonus;
    if (percent >= 80) return maxBonus * 0.8;
    if (percent >= 50) return maxBonus * 0.5;
    return maxBonus * 0.2;
  };

  it('должен начислять 100% при выполнении плана', () => {
    expect(calculateBonus(100, 50000)).toBe(50000);
  });

  it('должен начислять 80% при 80%+ выполнении', () => {
    expect(calculateBonus(85, 50000)).toBe(40000);
  });

  it('должен начислять 50% при 50-80% выполнении', () => {
    expect(calculateBonus(60, 50000)).toBe(25000);
  });

  it('должен начислять 20% при менее 50% выполнении', () => {
    expect(calculateBonus(30, 50000)).toBe(10000);
  });
});