import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import SaasMetricsSection from '@/components/Dashboard/sections/SaasMetricsSection';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn().mockResolvedValue({
    data: {
      mrr: 2500000,
      mrrChange: 12.5,
      arr: 30000000,
      churnRate: 5.2,
      overdueAmount: 150000,
      overdueCount: 3,
      arpu: 15000,
    },
  }),
}));

jest.mock('@/store/saasMetricsStore', () => ({
  useSaasMetricsStore: () => ({
    overview: {
      mrr: 2500000,
      mrrChange: 12.5,
      arr: 30000000,
      churnRate: 5.2,
      overdueAmount: 150000,
      overdueCount: 3,
      arpu: 15000,
    },
    mrrDynamics: [
      { month: '2025-05', newMrr: 80000, expansionMrr: 40000, churnMrr: 15000, netMrr: 2100000 },
      { month: '2025-06', newMrr: 90000, expansionMrr: 45000, churnMrr: 20000, netMrr: 2210000 },
    ],
    cashflow: {
      expected: [
        { date: '2026-04-30', tenant: 'ООО Техно', amount: 50000 },
      ],
      overdue: [
        { date: '2026-04-15', tenant: 'ООО Долг', amount: 25000, daysOverdue: 14 },
      ],
      risks: [],
      totalExpected: 50000,
      atRisk: 25000,
    },
    tariffs: [
      { name: 'Start', count: 12, mrr: 120000, avgCheck: 10000, growth: 15 },
      { name: 'Pro', count: 8, mrr: 320000, avgCheck: 40000, growth: 8 },
    ],
    isLoading: false,
    setOverview: jest.fn(),
    setMrrDynamics: jest.fn(),
    setCashflow: jest.fn(),
    setTariffs: jest.fn(),
    setLoading: jest.fn(),
  }),
}));

describe('SaasMetricsSection', () => {
  it('renders section title', async () => {
    render(<SaasMetricsSection />);
    await waitFor(() => {
      expect(screen.getByText('SaaS-метрики')).toBeInTheDocument();
    });
  });

  it('renders KPI cards', async () => {
    render(<SaasMetricsSection />);
    await waitFor(() => {
      expect(screen.getByText('MRR (ежемесячный доход)')).toBeInTheDocument();
      expect(screen.getByText('ARR (годовой доход)')).toBeInTheDocument();
    });
  });

  it('renders dynamic MRR section', async () => {
    render(<SaasMetricsSection />);
    await waitFor(() => {
      expect(screen.getByText('Динамика MRR (12 месяцев)')).toBeInTheDocument();
    });
  });

  it('renders cashflow section', async () => {
    render(<SaasMetricsSection />);
    await waitFor(() => {
      expect(screen.getByText('Прогноз Cash Flow')).toBeInTheDocument();
    });
  });

  it('renders tariff table', async () => {
    render(<SaasMetricsSection />);
    await waitFor(() => {
      expect(screen.getByText('Тепловая карта тарифов')).toBeInTheDocument();
    });
  });
});