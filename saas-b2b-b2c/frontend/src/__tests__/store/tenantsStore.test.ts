import { renderHook, act } from '@testing-library/react';
import { useTenantsStore } from '../../store/tenantsStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApi = apiClient as jest.Mocked<typeof apiClient>;

describe('useTenantsStore', () => {
  beforeEach(() => {
    useTenantsStore.setState({
      tenants: [],
      selectedTenant: null,
      tenantsLoading: false,
      cardLoading: false,
      showCreateModal: false,
      showSuspendModal: false,
      showTerminateModal: false,
      showImpersonateModal: false,
      selectedTenantForAction: null,
      showCard: false,
      filters: { status: '', tariff: '', hasDebt: false, search: '' },
      onboarding: [],
    });
    jest.clearAllMocks();
  });

  describe('setters', () => {
    it('setTenants updates tenants array', () => {
      const { result } = renderHook(() => useTenantsStore());
      const mockTenants = [
        { id: '1', name: 'Tenant 1', tariff: 'Premium', status: 'active' } as any,
      ];
      act(() => result.current.setTenants(mockTenants));
      expect(result.current.tenants).toEqual(mockTenants);
    });

    it('setSelectedTenant updates selectedTenant', () => {
      const { result } = renderHook(() => useTenantsStore());
      const tenant = { id: '1', name: 'Tenant 1' } as any;
      act(() => result.current.setSelectedTenant(tenant));
      expect(result.current.selectedTenant).toEqual(tenant);
    });

    it('setTenantsLoading updates loading state', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setTenantsLoading(true));
      expect(result.current.tenantsLoading).toBe(true);
    });

    it('setCardLoading updates loading state', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setCardLoading(true));
      expect(result.current.cardLoading).toBe(true);
    });

    it('setShowCreateModal toggles modal', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setShowCreateModal(true));
      expect(result.current.showCreateModal).toBe(true);
      act(() => result.current.setShowCreateModal(false));
      expect(result.current.showCreateModal).toBe(false);
    });

    it('setShowSuspendModal toggles modal', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setShowSuspendModal(true));
      expect(result.current.showSuspendModal).toBe(true);
    });

    it('setShowTerminateModal toggles modal', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setShowTerminateModal(true));
      expect(result.current.showTerminateModal).toBe(true);
    });

    it('setShowImpersonateModal toggles modal', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setShowImpersonateModal(true));
      expect(result.current.showImpersonateModal).toBe(true);
    });

    it('setSelectedTenantForAction updates tenant for action', () => {
      const { result } = renderHook(() => useTenantsStore());
      const tenant = { id: '1', name: 'Test' } as any;
      act(() => result.current.setSelectedTenantForAction(tenant));
      expect(result.current.selectedTenantForAction).toEqual(tenant);
    });

    it('setShowCard toggles card visibility', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setShowCard(true));
      expect(result.current.showCard).toBe(true);
    });

    it('setOnboarding updates onboarding array', () => {
      const { result } = renderHook(() => useTenantsStore());
      const mockOnboarding = [{ id: '1', name: 'New Tenant', step: 'account_created' }] as any;
      act(() => result.current.setOnboarding(mockOnboarding));
      expect(result.current.onboarding).toEqual(mockOnboarding);
    });
  });

  describe('setFilters', () => {
    it('updates single filter', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setFilters({ status: 'active' }));
      expect(result.current.filters.status).toBe('active');
    });

    it('updates multiple filters', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setFilters({ status: 'suspended', tariff: 'Premium' }));
      expect(result.current.filters.status).toBe('suspended');
      expect(result.current.filters.tariff).toBe('Premium');
    });

    it('preserves existing filters when updating', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setFilters({ status: 'active' }));
      act(() => result.current.setFilters({ search: 'test' }));
      expect(result.current.filters.status).toBe('active');
      expect(result.current.filters.search).toBe('test');
    });

    it('clears filter when empty string', () => {
      const { result } = renderHook(() => useTenantsStore());
      act(() => result.current.setFilters({ status: 'active' }));
      act(() => result.current.setFilters({ status: '' }));
      expect(result.current.filters.status).toBe('');
    });
  });

  describe('fetchTenants', () => {
    it('fetches tenants successfully', async () => {
      const mockData = [
        {
          id: '1',
          name: 'Tenant 1',
          plan_name: 'Premium',
          status: 'active',
          dealer_count: 5,
          user_count: 20,
          mrr: 5000,
          payment_status: 'paid',
        },
      ];
      mockApi.get.mockResolvedValue({ data: mockData });

      const { result } = renderHook(() => useTenantsStore());
      
      await act(async () => {
        await result.current.fetchTenants();
      });

      expect(result.current.tenants).toHaveLength(1);
      expect(result.current.tenants[0].name).toBe('Tenant 1');
      expect(result.current.tenantsLoading).toBe(false);
    });

    it('handles API error', async () => {
      mockApi.get.mockRejectedValue(new Error('API Error'));

      const { result } = renderHook(() => useTenantsStore());
      
      await act(async () => {
        await result.current.fetchTenants();
      });

      expect(result.current.tenants).toEqual([]);
      expect(result.current.tenantsLoading).toBe(false);
    });

    it('handles empty response', async () => {
      mockApi.get.mockResolvedValue({ data: null });

      const { result } = renderHook(() => useTenantsStore());
      
      await act(async () => {
        await result.current.fetchTenants();
      });

      expect(result.current.tenants).toEqual([]);
      expect(result.current.tenantsLoading).toBe(false);
    });

    it('sets loading state during fetch', async () => {
      let resolvePromise: (value: any) => void;
      const promise = new Promise((resolve) => {
        resolvePromise = resolve;
      });
      mockApi.get.mockReturnValue(promise);

      const { result } = renderHook(() => useTenantsStore());
      
      act(() => {
        result.current.fetchTenants();
      });

      expect(result.current.tenantsLoading).toBe(true);

      await act(async () => {
        resolvePromise!({ data: [] });
      });

      expect(result.current.tenantsLoading).toBe(false);
    });

    it('maps API fields correctly', async () => {
      const mockData = [
        {
          id: '123',
          name: 'Test Tenant',
          legal_entity: 'OOO Test',
          inn: '1234567890',
          plan_name: 'Business',
          dealer_count: 10,
          user_count: 50,
          max_users: 100,
          mrr: 10000,
          payment_status: 'pending',
          paid_until: '2024-12-31',
          status: 'suspended',
          contact_name: 'John Doe',
          contact_email: 'john@test.com',
          created_at: '2024-01-01',
        },
      ];
      mockApi.get.mockResolvedValue({ data: mockData });

      const { result } = renderHook(() => useTenantsStore());
      
      await act(async () => {
        await result.current.fetchTenants();
      });

      const tenant = result.current.tenants[0];
      expect(tenant.id).toBe('123');
      expect(tenant.name).toBe('Test Tenant');
      expect(tenant.legalEntity).toBe('OOO Test');
      expect(tenant.inn).toBe('1234567890');
      expect(tenant.tariff).toBe('Business');
      expect(tenant.dealers).toBe(10);
      expect(tenant.users).toBe(50);
      expect(tenant.licenseLimit).toBe(100);
      expect(tenant.mrr).toBe(10000);
      expect(tenant.paymentStatus).toBe('pending');
      expect(tenant.nextPaymentDate).toBe('2024-12-31');
      expect(tenant.status).toBe('suspended');
      expect(tenant.contactName).toBe('John Doe');
      expect(tenant.contactEmail).toBe('john@test.com');
      expect(tenant.contractStartDate).toBe('2024-01-01');
    });

    it('applies defaults for missing fields', async () => {
      const mockData = [{ id: '1' }];
      mockApi.get.mockResolvedValue({ data: mockData });

      const { result } = renderHook(() => useTenantsStore());
      
      await act(async () => {
        await result.current.fetchTenants();
      });

      const tenant = result.current.tenants[0];
      expect(tenant.name).toBe('');
      expect(tenant.tariff).toBe('Start');
      expect(tenant.status).toBe('active');
      expect(tenant.paymentStatus).toBe('paid');
    });
  });

  describe('initial state', () => {
    it('has correct initial values', () => {
      const { result } = renderHook(() => useTenantsStore());
      
      expect(result.current.tenants).toEqual([]);
      expect(result.current.selectedTenant).toBeNull();
      expect(result.current.tenantsLoading).toBe(false);
      expect(result.current.cardLoading).toBe(false);
      expect(result.current.showCreateModal).toBe(false);
      expect(result.current.showSuspendModal).toBe(false);
      expect(result.current.showTerminateModal).toBe(false);
      expect(result.current.showImpersonateModal).toBe(false);
      expect(result.current.selectedTenantForAction).toBeNull();
      expect(result.current.showCard).toBe(false);
      expect(result.current.onboarding).toEqual([]);
      expect(result.current.filters).toEqual({
        status: '',
        tariff: '',
        hasDebt: false,
        search: '',
      });
    });
  });
});