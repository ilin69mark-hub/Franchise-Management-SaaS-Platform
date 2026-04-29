import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import TenantsSection from '@/components/Dashboard/sections/TenantsSection';

jest.mock('@/store/tenantsStore', () => ({
  useTenantsStore: () => ({
    tenants: [
      {
        id: '1',
        name: 'ООО Техно',
        legalEntity: 'ООО Техно',
        tariff: 'Pro',
        dealers: 10,
        users: 25,
        licenseLimit: 30,
        mrr: 40000,
        paymentStatus: 'paid',
        nextPaymentDate: '2026-05-01',
        status: 'active',
      },
    ],
    tenantsLoading: false,
    selectedTenant: null,
    showCard: false,
    filters: { status: '', tariff: '', hasDebt: false, search: '' },
    showCreateModal: false,
    showSuspendModal: false,
    showTerminateModal: false,
    showImpersonateModal: false,
    selectedTenantForAction: null,
    onboarding: [],
    setTenants: jest.fn(),
    setSelectedTenant: jest.fn(),
    setTenantsLoading: jest.fn(),
    setShowCreateModal: jest.fn(),
    setShowSuspendModal: jest.fn(),
    setShowTerminateModal: jest.fn(),
    setShowImpersonateModal: jest.fn(),
    setSelectedTenantForAction: jest.fn(),
    setShowCard: jest.fn(),
    setFilters: jest.fn(),
    setOnboarding: jest.fn(),
  }),
}));

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn().mockResolvedValue({ data: [] }),
}));

describe('TenantsSection', () => {
  it('renders section title', async () => {
    render(<TenantsSection />);
    await waitFor(() => {
      expect(screen.getByText('Тенанты')).toBeInTheDocument();
    });
  });

  it('renders add button', async () => {
    render(<TenantsSection />);
    await waitFor(() => {
      expect(screen.getByText('Добавить тенанта')).toBeInTheDocument();
    });
  });

  it('renders search input', async () => {
    render(<TenantsSection />);
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/Поиск/i)).toBeInTheDocument();
    });
  });

  it('renders status filters', async () => {
    render(<TenantsSection />);
    await waitFor(() => {
      expect(screen.getByText('Активен')).toBeInTheDocument();
    });
  });
});