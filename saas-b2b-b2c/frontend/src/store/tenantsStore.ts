import { create } from 'zustand';
import apiClient from '@/api/axiosClient';

interface Tenant {
  id: string;
  name: string;
  legalEntity?: string;
  inn?: string;
  tariff: string;
  dealers: number;
  users: number;
  licenseLimit: number;
  mrr: number;
  paymentStatus: 'paid' | 'pending' | 'overdue';
  nextPaymentDate: string;
  status: 'active' | 'suspended' | 'terminated';
  contactName?: string;
  contactEmail?: string;
  contractStartDate?: string;
  contractEndDate?: string;
}

interface TenantPayment {
  id: string;
  date: string;
  invoiceNumber: string;
  amount: number;
  status: 'paid' | 'pending' | 'overdue';
}

interface TenantUser {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  role: string;
  lastLogin?: string;
}

interface TenantActivity {
  dau: number;
  mau: number;
  monthly: { date: string; dau: number; mau: number }[];
}

interface OnboardingTenant {
  id: string;
  name: string;
  createdAt: string;
  step: 'account_created' | 'dealers_added' | 'users_added' | 'first_sale';
  progress: number;
  daysToTtv: number;
}

interface TenantsState {
  tenants: Tenant[];
  selectedTenant: Tenant | null;
  tenantsLoading: boolean;
  cardLoading: boolean;
  showCreateModal: boolean;
  showSuspendModal: boolean;
  showTerminateModal: boolean;
  showImpersonateModal: boolean;
  selectedTenantForAction: Tenant | null;
  showCard: boolean;
  filters: {
    status: string;
    tariff: string;
    hasDebt: boolean;
    search: string;
  };
  onboarding: OnboardingTenant[];
  setTenants: (tenants: Tenant[]) => void;
  setSelectedTenant: (tenant: Tenant | null) => void;
  setTenantsLoading: (loading: boolean) => void;
  setCardLoading: (loading: boolean) => void;
  setShowCreateModal: (show: boolean) => void;
  setShowSuspendModal: (show: boolean) => void;
  setShowTerminateModal: (show: boolean) => void;
  setShowImpersonateModal: (show: boolean) => void;
  setSelectedTenantForAction: (tenant: Tenant | null) => void;
  setShowCard: (show: boolean) => void;
  setFilters: (filters: Partial<TenantsState['filters']>) => void;
  setOnboarding: (onboarding: OnboardingTenant[]) => void;
  fetchTenants: () => Promise<void>;
}

export const useTenantsStore = create<TenantsState>((set, get) => ({
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
  filters: {
    status: '',
    tariff: '',
    hasDebt: false,
    search: '',
  },
  onboarding: [],
  setTenants: (tenants) => set({ tenants }),
  setSelectedTenant: (tenant) => set({ selectedTenant: tenant }),
  setTenantsLoading: (loading) => set({ tenantsLoading: loading }),
  setCardLoading: (loading) => set({ cardLoading: loading }),
  setShowCreateModal: (show) => set({ showCreateModal: show }),
  setShowSuspendModal: (show) => set({ showSuspendModal: show }),
  setShowTerminateModal: (show) => set({ showTerminateModal: show }),
  setShowImpersonateModal: (show) => set({ showImpersonateModal: show }),
  setSelectedTenantForAction: (tenant) => set({ selectedTenantForAction: tenant }),
  setShowCard: (show) => set({ showCard: show }),
  setFilters: (filters) => set((state) => ({ filters: { ...state.filters, ...filters } })),
  setOnboarding: (onboarding) => set({ onboarding }),
  fetchTenants: async () => {
    const { setTenants, setTenantsLoading } = get();
    setTenantsLoading(true);
    try {
      const res = await apiClient.get('/admin/tenants');
      const data = res.data as Array<Record<string, unknown>> || [];
      const tenants: Tenant[] = data.map((item) => ({
        id: String(item.id),
        name: (item.name as string) || '',
        legalEntity: (item.legal_entity as string) || undefined,
        inn: (item.inn as string) || undefined,
        tariff: (item.plan_name as string) || 'Start',
        dealers: (item.dealer_count as number) || 0,
        users: (item.user_count as number) || 0,
        licenseLimit: (item.max_users as number) || 0,
        mrr: (item.mrr as number) || 0,
        paymentStatus: (item.payment_status as Tenant['paymentStatus']) || 'paid',
        nextPaymentDate: (item.paid_until as string) || '',
        status: (item.status as Tenant['status']) || 'active',
        contactName: (item.contact_name as string) || undefined,
        contactEmail: (item.contact_email as string) || undefined,
        contractStartDate: (item.created_at as string) || undefined,
        contractEndDate: undefined,
      }));
      setTenants(tenants);
    } catch {
      setTenants([]);
    } finally {
      setTenantsLoading(false);
    }
  },
}));