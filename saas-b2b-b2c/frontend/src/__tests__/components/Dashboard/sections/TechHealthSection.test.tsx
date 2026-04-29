import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import TechHealthSection from '@/components/Dashboard/sections/TechHealthSection';

jest.mock('@/store/techHealthStore', () => ({
  useTechHealthStore: () => ({
    status: {
      uptime: 99.95,
      avgResponseTime: 45,
      p95ResponseTime: 120,
      p99ResponseTime: 350,
      errorRate: 0.12,
      activeWebsockets: 25,
    },
    performance: [],
    errors: [],
    services: [],
    securityEvents: [],
    period: '24h',
    isLoading: false,
    setStatus: jest.fn(),
    setPerformance: jest.fn(),
    setErrors: jest.fn(),
    setServices: jest.fn(),
    setSecurityEvents: jest.fn(),
    setPeriod: jest.fn(),
    setLoading: jest.fn(),
  }),
}));

describe('TechHealthSection', () => {
  it('renders section title', async () => {
    render(<TechHealthSection />);
    await waitFor(() => {
      expect(screen.getByText('Техническое здоровье')).toBeInTheDocument();
    });
  });

  it('renders status cards', async () => {
    render(<TechHealthSection />);
    await waitFor(() => {
      expect(screen.getByText('Uptime (30 дней)')).toBeInTheDocument();
    });
  });

  it('renders performance section', async () => {
    render(<TechHealthSection />);
    await waitFor(() => {
      expect(screen.getByText('Производительность')).toBeInTheDocument();
    });
  });

  it('renders errors section', async () => {
    render(<TechHealthSection />);
    await waitFor(() => {
      expect(screen.getByText('Последние ошибки')).toBeInTheDocument();
    });
  });

  it('renders services section', async () => {
    render(<TechHealthSection />);
    await waitFor(() => {
      expect(screen.getByText('Нагрузка по сервисам')).toBeInTheDocument();
    });
  });

  it('renders security events section', async () => {
    render(<TechHealthSection />);
    await waitFor(() => {
      expect(screen.getByText('Подозрительная активность')).toBeInTheDocument();
    });
  });
});