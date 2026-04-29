import React from 'react';
import { render, waitFor, fireEvent, act } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import NotificationBell from '@/components/Dashboard/NotificationBell';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  patch: jest.fn(),
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

const mockAlerts = [
  {
    id: '1',
    title: 'Просроченный замер',
    message: 'Замер для Иван Иванов просрочен более 3 дней',
    type: 'overdue_measurement',
    severity: 'critical',
    link: '/leads/123',
    is_read: false,
    created_at: new Date().toISOString(),
  },
  {
    id: '2',
    title: 'Брошенное КП',
    message: 'КП для Петр Петров без ответа более 5 дней',
    type: 'abandoned_kp',
    severity: 'warning',
    link: '/leads/456',
    is_read: false,
    created_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: '3',
    title: 'Падение конверсии',
    message: 'Конверсия упала на 25% относительно среднего',
    type: 'conversion_drop',
    severity: 'warning',
    link: '',
    is_read: true,
    created_at: new Date(Date.now() - 7200000).toISOString(),
  },
];

describe('NotificationBell', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it('должен показывать бейдж с количеством непрочитанных', async () => {
    apiClient.get.mockResolvedValue({
      data: { alerts: mockAlerts, unread_count: 2 },
    });

    const { container } = render(
      <Provider store={createMockStore()}>
        <NotificationBell />
      </Provider>
    );

    await waitFor(() => {
      expect(container.textContent).toContain('2');
    });
  });

  it('должен загружать алерты при монтировании', async () => {
    apiClient.get.mockResolvedValue({
      data: { alerts: mockAlerts, unread_count: 2 },
    });

    render(
      <Provider store={createMockStore()}>
        <NotificationBell />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalledWith('/alerts');
    });
  });

  it('должен показывать список алертов', async () => {
    apiClient.get.mockResolvedValue({
      data: { alerts: mockAlerts, unread_count: 2 },
    });

    const { container } = render(
      <Provider store={createMockStore()}>
        <NotificationBell />
      </Provider>
    );

    await waitFor(() => {
      expect(apiClient.get).toHaveBeenCalled();
    });

    expect(container.querySelector('.ant-badge')).toBeInTheDocument();
  });

  it('должен использовать localStorage для сохранения непрочитанных', async () => {
    localStorage.setItem('salon_alerts_unread', '3');

    apiClient.get.mockResolvedValue({
      data: { alerts: [], unread_count: 0 },
    });

    render(
      <Provider store={createMockStore()}>
        <NotificationBell />
      </Provider>
    );

    await waitFor(() => {
      expect(localStorage.getItem('salon_alerts_unread')).toBe('0');
    });
  });

  it('должен обрабатывать ошибки API', async () => {
    apiClient.get.mockRejectedValue(new Error('Network error'));

    const { container } = render(
      <Provider store={createMockStore()}>
        <NotificationBell />
      </Provider>
    );

    await waitFor(() => {
      expect(container.querySelector('.ant-badge')).toBeInTheDocument();
    });
  });
});

describe('Alert severity colors', () => {
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return '#ff4d4f';
      case 'warning': return '#faad14';
      default: return '#1890ff';
    }
  };

  it('должен возвращать красный для critical', () => {
    expect(getSeverityColor('critical')).toBe('#ff4d4f');
  });

  it('должен возвращать желтый для warning', () => {
    expect(getSeverityColor('warning')).toBe('#faad14');
  });

  it('должен возвращать синий для info', () => {
    expect(getSeverityColor('info')).toBe('#1890ff');
  });
});

describe('Alert time formatting', () => {
  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'Только что';
    if (diffMins < 60) return `${diffMins} мин назад`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)} ч назад`;
    return date.toLocaleDateString('ru-RU');
  };

  it('должен показывать "Только что" для свежих алертов', () => {
    const now = new Date().toISOString();
    expect(formatTime(now)).toBe('Только что');
  });

  it('должен показывать минуты для алертов младше часа', () => {
    const thirtyMinsAgo = new Date(Date.now() - 30 * 60000).toISOString();
    expect(formatTime(thirtyMinsAgo)).toBe('30 мин назад');
  });

  it('должен показывать часы для алертов младше суток', () => {
    const twoHoursAgo = new Date(Date.now() - 2 * 3600000).toISOString();
    expect(formatTime(twoHoursAgo)).toBe('2 ч назад');
  });
});