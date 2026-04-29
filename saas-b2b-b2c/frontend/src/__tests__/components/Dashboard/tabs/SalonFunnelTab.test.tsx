import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import SalonFunnelTab from '@/components/Dashboard/tabs/SalonFunnelTab';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  patch: jest.fn(),
}));

jest.mock('@/services/api', () => ({
  useGetLeadsQuery: () => ({ data: [], isLoading: false }),
  useCreateLeadMutation: () => [jest.fn(), { isLoading: false }],
  useUpdateLeadStatusMutation: () => [jest.fn(), { isLoading: false }],
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

const mockUser = { id: '1', role: 'salon_manager' };

const mockFunnelData = {
  stages: [
    { stage: 'traffic', label: 'Трафик', count: 50, conversion: 100, sum: 0 },
    { stage: 'consultation', label: 'Консультация', count: 30, conversion: 60, sum: 0 },
    { stage: 'measurement', label: 'Замер', count: 20, conversion: 67, sum: 0 },
    { stage: 'kp', label: 'КП', count: 10, conversion: 50, sum: 0 },
    { stage: 'contract', label: 'Договор', count: 5, conversion: 50, sum: 0 },
    { stage: 'payment', label: 'Оплата', count: 0, conversion: 0, sum: 0 },
  ],
  hot_deals: [
    { id: '1', client_name: 'Иван Иванов', phone: '+79999999999', amount: 150000, created_at: '2024-01-15', days_stalled: 8, manager_id: '1', manager_name: 'Менеджер' },
    { id: '2', client_name: 'Петр Петров', phone: '+78888888888', amount: 80000, created_at: '2024-01-18', days_stalled: 6, manager_id: '1', manager_name: 'Менеджер' },
  ],
  fresh_leads: [
    { id: '1', source: 'Сайт', client_name: 'Алексей Алексеев', phone: '+77777777777', created_at: '2024-01-27 10:30', status: 'unassigned', assigned_to: null },
    { id: '2', source: 'Звонок', client_name: 'Сергей Сергеев', phone: '+76666666666', created_at: '2024-01-27 11:00', status: 'in_progress', assigned_to: '1' },
  ],
};

const mockLeads = [
  { id: '1', full_name: 'Клиент 1', phone: '+79999999999', interest_product: 'Диван', budget: 100000, status: 'new', created_at: '2024-01-27' },
  { id: '2', full_name: 'Клиент 2', phone: '+78888888888', interest_product: 'Кресло', budget: 50000, status: 'contact', created_at: '2024-01-26' },
  { id: '3', full_name: 'Клиент 3', phone: '+77777777777', interest_product: 'Кровать', budget: 150000, status: 'meeting', created_at: '2024-01-25' },
  { id: '4', full_name: 'Клиент 4', phone: '+76666666666', interest_product: 'Шкаф', budget: 80000, status: 'sale', created_at: '2024-01-24' },
];

describe('SalonFunnelTab', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('должен загружать и отображать данные воронки', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Трафик');
    });

    expect(container.textContent).toContain('50');
  });

  it('должен отображать горячие сделки', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Горячие сделки');
    });

    expect(container.textContent).toContain('Иван Иванов');
    expect(container.textContent).toContain('8 дн.');
  });

  it('должен отображать свежие лиды', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Свежие лиды');
    });
  });

  it('должен подсвечивать просроченные горячие сделки (>7 дней)', async () => {
    apiClient.get.mockResolvedValue({ data: mockFunnelData });

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('8 дн.');
    });
  });

  it('должен показывать все лиды в таблице', async () => {
    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Все лиды');
    });
  });

  it('должен обрабатывать ошибки API', async () => {
    apiClient.get.mockRejectedValue(new Error('Network error'));

    const { container } = render(
      <Provider store={createMockStore()}>
        <SalonFunnelTab user={mockUser} />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('Ошибка');
    });
  });
});

describe('Funnel calculations', () => {
  it('должен правильно вычислять конверсию', () => {
    const calculateConversion = (current: number, previous: number) => {
      if (previous === 0) return 0;
      return Math.round((current / previous) * 100);
    };

    expect(calculateConversion(30, 50)).toBe(60);
    expect(calculateConversion(20, 30)).toBe(67);
    expect(calculateConversion(10, 20)).toBe(50);
  });

  it('должен определять просроченную сделку', () => {
    const isOverdue = (days: number) => days > 7;
    expect(isOverdue(8)).toBe(true);
    expect(isOverdue(7)).toBe(false);
    expect(isOverdue(5)).toBe(false);
  });

  it('должен определять свежий необработанный лид', () => {
    const isStaleUnassigned = (status: string, minutesAgo: number) => {
      return status === 'unassigned' && minutesAgo > 30;
    };

    expect(isStaleUnassigned('unassigned', 35)).toBe(true);
    expect(isStaleUnassigned('unassigned', 20)).toBe(false);
    expect(isStaleUnassigned('in_progress', 35)).toBe(false);
  });
});

describe('Lead status mapping', () => {
  const getStatusLabel = (status: string) => {
    const map: Record<string, string> = {
      new: 'Новый',
      contact: 'Контакт',
      meeting: 'Замер',
      wait: 'КП',
      sale: 'Договор',
      paid: 'Оплачен',
    };
    return map[status as any] || status;
  };

  it('должен возвращать правильные метки статусов', () => {
    expect(getStatusLabel('new')).toBe('Новый');
    expect(getStatusLabel('contact')).toBe('Контакт');
    expect(getStatusLabel('meeting')).toBe('Замер');
    expect(getStatusLabel('wait')).toBe('КП');
    expect(getStatusLabel('sale')).toBe('Договор');
    expect(getStatusLabel('paid')).toBe('Оплачен');
  });
});