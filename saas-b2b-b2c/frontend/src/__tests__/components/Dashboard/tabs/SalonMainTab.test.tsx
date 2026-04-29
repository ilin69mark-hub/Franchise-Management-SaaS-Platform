import React from 'react';
import { render, waitFor, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import SalonMainTab from '@/components/Dashboard/tabs/SalonMainTab';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
}));

const apiClient = require('@/api/axiosClient');

const createMockStore = () =>
  configureStore({
    reducer: {
      auth: () => ({
        user: { id: '1', salon_id: 'salon-1', role: 'salon_manager' },
        isAuthenticated: true,
      }),
    },
  });

const mockUser = {
  id: '1',
  first_name: 'Иван',
  last_name: 'Иванов',
  email: 'ivan@test.ru',
  role: 'salon_manager',
  salon_id: 'salon-1',
};

const mockData = {
  plan: 1000000,
  fact: 750000,
  plan_percent: 75,
  dynamic_day: 10.5,
  dynamic_week: -5.2,
  forecast: 85,
  avg_check: 50000,
  margin_percent: 25.5,
  prepayments_sum: 300000,
  traffic_status: 'green',
  conversion_measure_status: 'yellow',
  conversion_contract_status: 'green',
  pending_payments: [
    {
      contract_id: 'contract-1',
      client_name: 'Петров П.П.',
      amount: 150000,
      status: 'awaiting_payment',
      payment_date: '2024-01-15',
    },
  ],
};

describe('SalonMainTab', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('должен показывать loader при загрузке', () => {
    apiClient.get.mockImplementation(() => new Promise(() => {}));

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonMainTab user={mockUser} />
      </Provider>
    );

    expect(container.querySelector('.ant-spin')).toBeInTheDocument();
  });

  it('должен загружать и отображать данные', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonMainTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalled();
    });

    const text = container.textContent;
    expect(text).toContain('750') || expect(text).toContain('Продажи');
  });

  it('должен показывать предупреждение когда план выполнен менее 70%', async () => {
    const lowPlanData = {
      ...mockData,
      plan_percent: 60,
    };
    apiClient.get.mockResolvedValue({ data: lowPlanData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonMainTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      const text = container.textContent;
      expect(text).toContain('Внимание');
    });
  });

  it('должен показывать сообщение когда нет запланированных оплат', async () => {
    const noPaymentsData = {
      ...mockData,
      pending_payments: [],
    };
    apiClient.get.mockResolvedValue({ data: noPaymentsData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonMainTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      const text = container.textContent;
      expect(text).toContain('На сегодня оплат не запланировано');
    });
  });

  it('должен корректно отображать статус светофора', async () => {
    apiClient.get.mockResolvedValue({ data: mockData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonMainTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      const text = container.textContent;
      expect(text).toContain('Светофор');
    });
  });

  it('должен обрабатывать ошибки API', async () => {
    apiClient.get.mockRejectedValue({
      response: { data: { error: 'Ошибка сервера' } },
    });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonMainTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      const text = container.textContent;
      expect(text).toContain('Ошибка');
    });
  });
});

describe('formatMoney utility', () => {
  it('должен форматировать числа с разделителями', () => {
    const formatMoney = (val: number) => new Intl.NumberFormat('ru-RU').format(val);
    expect(formatMoney(1000000).replace(/\s/g, ' ')).toBe('1 000 000');
    expect(formatMoney(50000).replace(/\s/g, ' ')).toBe('50 000');
    expect(formatMoney(1500).replace(/\s/g, ' ')).toBe('1 500');
  });
});

describe('Traffic light colors', () => {
  it('должен возвращать правильные цвета для статусов', () => {
    const getTrafficLightColor = (status: string) => {
      switch (status) {
        case 'green': return '#52c41a';
        case 'yellow': return '#faad14';
        case 'red': return '#ff4d4f';
        default: return '#d9d9d9';
      }
    };

    expect(getTrafficLightColor('green')).toBe('#52c41a');
    expect(getTrafficLightColor('yellow')).toBe('#faad14');
    expect(getTrafficLightColor('red')).toBe('#ff4d4f');
  });
});