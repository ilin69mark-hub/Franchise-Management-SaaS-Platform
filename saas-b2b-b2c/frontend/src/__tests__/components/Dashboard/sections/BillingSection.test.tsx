import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import BillingSection from '@/components/Dashboard/sections/BillingSection';

jest.mock('@/store/billingStore', () => ({
  useBillingStore: () => ({
    invoices: [],
    settings: {
      notifyAfterDays: 3,
      secondNotifyAfterDays: 7,
      suspendAfterDays: 14,
      autoInvoiceDays: 3,
    },
    history: [],
    period: '30',
    isLoading: false,
    setInvoices: jest.fn(),
    setSettings: jest.fn(),
    setHistory: jest.fn(),
    setPeriod: jest.fn(),
    setLoading: jest.fn(),
  }),
}));

describe('BillingSection', () => {
  it('renders section title', async () => {
    render(<BillingSection />);
    await waitFor(() => {
      expect(screen.getByText('Биллинг')).toBeInTheDocument();
    });
  });

  it('renders invoices table', async () => {
    render(<BillingSection />);
    await waitFor(() => {
      expect(screen.getByText('Счета и платежи')).toBeInTheDocument();
    });
  });

  it('renders billing actions', async () => {
    render(<BillingSection />);
    await waitFor(() => {
      expect(screen.getByText('Выставить ручной счёт')).toBeInTheDocument();
      expect(screen.getByText('Зачислить оплату')).toBeInTheDocument();
    });
  });

  it('renders billing settings', async () => {
    render(<BillingSection />);
    await waitFor(() => {
      expect(screen.getByText('Настройки биллинга')).toBeInTheDocument();
    });
  });

  it('renders payment history', async () => {
    render(<BillingSection />);
    await waitFor(() => {
      expect(screen.getByText('История платежей')).toBeInTheDocument();
    });
  });
});