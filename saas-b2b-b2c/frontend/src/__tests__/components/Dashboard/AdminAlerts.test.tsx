import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import AdminAlerts from '@/components/Dashboard/AdminAlerts';

jest.mock('@/store/alertsStore', () => ({
  useAlertStore: () => ({
    alerts: [
      {
        id: '1',
        type: 'billing',
        title: 'Просроченная оплата',
        message: 'ООО Техно просрочил оплату на 5 дней',
        timestamp: new Date().toISOString(),
        read: false,
        link: '/admin?tab=billing',
      },
    ],
    unreadCount: 1,
    settings: {
      channels: ['in-app'],
      thresholds: { errorRate: 1, uptime: 99.9 },
    },
    isLoading: false,
    loadAlerts: jest.fn().mockResolvedValue([]),
    markAsRead: jest.fn().mockResolvedValue(undefined),
    updateSettings: jest.fn().mockResolvedValue(undefined),
  }),
}));

describe('AdminAlerts', () => {
  it('renders without crashing', () => {
    render(<AdminAlerts />);
  });

  it('renders bell icon', () => {
    const { container } = render(<AdminAlerts />);
    const bell = container.querySelector('.anticon-bell');
    expect(bell).toBeInTheDocument();
  });
});