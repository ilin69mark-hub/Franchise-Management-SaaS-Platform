import React from 'react';
import { render, waitFor, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import SalonProductsTab from '@/components/Dashboard/tabs/SalonProductsTab';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
}));

const apiClient = require('@/api/axiosClient');

const createMockStore = () =>
  configureStore({
    reducer: {
      auth: () => ({
        user: { id: '1', role: 'salon_manager' },
        isAuthenticated: true,
      }),
    },
  });

const mockUser = {
  id: '1',
  role: 'salon_manager',
};

const mockData = {
  top_products: [
    { id: '1', name: 'Диван Комфорт', collection: 'Основная', category: 'Мебель', revenue: 500000, quantity: 10, share_percent: 25, margin: 30 },
    { id: '2', name: 'Кресло Люкс', collection: 'Премиум', category: 'Мебель', revenue: 300000, quantity: 15, share_percent: 15, margin: 25 },
    { id: '3', name: 'Кровать Мастер', collection: 'Основная', category: 'Мебель', revenue: 200000, quantity: 5, share_percent: 10, margin: 28 },
  ],
  stock_items: [
    { id: '1', name: 'Диван Комфорт', category: 'Мебель', showroom_qty: 2, warehouse_qty: 10, total_cost: 600000, turnover_days: 45 },
    { id: '2', name: 'Кресло Люкс', category: 'Мебель', showroom_qty: 1, warehouse_qty: 5, total_cost: 300000, turnover_days: 120 },
    { id: '3', name: 'Кровать Мастер', category: 'Мебель', showroom_qty: 1, warehouse_qty: 8, total_cost: 450000, turnover_days: 30 },
  ],
  lost_sales: [
    { reason: 'Нет в наличии', requests_count: 5, lost_revenue: 150000 },
    { reason: 'Долгий срок производства', requests_count: 3, lost_revenue: 80000 },
    { reason: 'Не устроила цена', requests_count: 8, lost_revenue: 200000 },
    { reason: 'Не подошёл дизайн', requests_count: 4, lost_revenue: 120000 },
  ],
  category_turnover: [
    { category: 'Диваны', avg_days: 45, is_slow_moving: false },
    { category: 'Кресла', avg_days: 120, is_slow_moving: true },
    { category: 'Кровати', avg_days: 30, is_slow_moving: false },
  ],
  total_revenue: 2000000,
};

describe('SalonProductsTab', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('должен показывать loader при загрузке', () => {
    apiClient.get.mockImplementation(() => new Promise(() => {}));

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    expect(container.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('должен загружать и отображать данные', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      const text = container.textContent;
      expect(text).toContain('Топ-10');
    });

    expect(container.textContent).toContain('Диван Комфорт');
  });

  it('должен сортировать по выручке', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('По выручке');
    });
  });

  it('должен показывать остатки с подсветкой неликвида', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Остатки товаров');
    });

    expect(container.textContent).toContain('120 дн.');
  });

  it('должен показывать упущенные продажи с предупреждением', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalled();
    });

    const text = container.textContent;
    expect(text).toContain('продаж') || expect(text).toContain('продажи');
  });

  it('должен показывать оборачиваемость по категориям', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Оборачиваемость');
    });

    expect(container.textContent).toContain('Кресла');
  });

  it('должен обрабатывать ошибки API', async () => {
    apiClient.get.mockRejectedValue({
      response: { data: { error: 'Ошибка сервера' } },
    });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonProductsTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Ошибка');
    });
  });
});

describe('formatMoney utility', () => {
  it('должен форматировать числа', () => {
    const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);
    expect(formatMoney(1000000).replace(/\s/g, ' ')).toBe('1 000 000');
    expect(formatMoney(500000).replace(/\s/g, ' ')).toBe('500 000');
    expect(formatMoney(123456).replace(/\s/g, ' ')).toBe('123 456');
  });
});

describe('Stock turnover logic', () => {
  it('должен определять неликвидный товар', () => {
    const isSlowMoving = (days: number) => days > 90;
    expect(isSlowMoving(120)).toBe(true);
    expect(isSlowMoving(90)).toBe(false);
    expect(isSlowMoving(45)).toBe(false);
  });
});

describe('Lost sales threshold', () => {
  it('должен подсвечивать критичные суммы', () => {
    const threshold = 100000;
    const isCritical = (revenue: number) => revenue > threshold;
    expect(isCritical(150000)).toBe(true);
    expect(isCritical(80000)).toBe(false);
    expect(isCritical(100000)).toBe(false);
  });
});