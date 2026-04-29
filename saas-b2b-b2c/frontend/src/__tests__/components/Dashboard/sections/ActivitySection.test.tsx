import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import ActivitySection from '@/components/Dashboard/sections/ActivitySection';

jest.mock('@/store/activityStore', () => ({
  useActivityStore: () => ({
    overview: {
      dau: 45,
      dauChange: 8.5,
      wau: 120,
      wauChange: 12.3,
      mau: 320,
      mauChange: 15.2,
      stickiness: 14.1,
      avgSessionTime: 25,
      activeSessions: 12,
    },
    dynamics: [],
    byTenant: [
      { id: '1', name: 'ООО Техно', dau: 25, wau: 80, mau: 150, stickiness: 16.7, activeLicensePercent: 85 },
    ],
    featureAdoption: [
      { id: '1', name: 'Управление дилерами', tenantPercent: 95, userPercent: 78, frequency: 'daily', trend: 'up' },
    ],
    ttv: [
      { id: '1', name: 'Новая компания', createdAt: '2026-03-15', firstSaleAt: '2026-03-22', ttvDays: 7 },
    ],
    period: 30,
    selectedTenant: '',
    isLoading: false,
    setOverview: jest.fn(),
    setDynamics: jest.fn(),
    setByTenant: jest.fn(),
    setFeatureAdoption: jest.fn(),
    setTtv: jest.fn(),
    setPeriod: jest.fn(),
    setSelectedTenant: jest.fn(),
    setLoading: jest.fn(),
  }),
}));

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn().mockResolvedValue({ data: {} }),
}));

describe('ActivitySection', () => {
  it('renders section title', async () => {
    render(<ActivitySection />);
    await waitFor(() => {
      expect(screen.getByText('Активность пользователей')).toBeInTheDocument();
    });
  });

  it('renders KPI cards', async () => {
    render(<ActivitySection />);
    await waitFor(() => {
      expect(screen.getByText('DAU')).toBeInTheDocument();
      expect(screen.getByText('WAU')).toBeInTheDocument();
      expect(screen.getByText('MAU')).toBeInTheDocument();
    });
  });

  it('renders dynamics chart', async () => {
    render(<ActivitySection />);
    await waitFor(() => {
      expect(screen.getByText('Динамика активности')).toBeInTheDocument();
    });
  });

  it('renders tenant activity section', async () => {
    render(<ActivitySection />);
    await waitFor(() => {
      expect(screen.getByText('Активность по тенантам')).toBeInTheDocument();
    });
  });

  it('renders feature adoption', async () => {
    render(<ActivitySection />);
    await waitFor(() => {
      expect(screen.getByText('Feature Adoption')).toBeInTheDocument();
    });
  });

  it('renders TTV section', async () => {
    render(<ActivitySection />);
    await waitFor(() => {
      expect(screen.getByText('Time to Value')).toBeInTheDocument();
    });
  });
});