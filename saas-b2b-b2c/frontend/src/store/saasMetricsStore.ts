import { create } from 'zustand';
import apiClient from '@/api/axiosClient';

interface SaasOverview {
  mrr: number;
  mrrChange: number;
  arr: number;
  churnRate: number;
  overdueAmount: number;
  overdueCount: number;
  arpu: number;
}

interface MrrDynamicsMonth {
  month: string;
  newMrr: number;
  expansionMrr: number;
  churnMrr: number;
  netMrr: number;
}

interface CashflowExpected {
  date: string;
  tenant: string;
  amount: number;
}

interface CashflowOverdue {
  date: string;
  tenant: string;
  amount: number;
  daysOverdue: number;
}

interface CashflowRisk {
  date: string;
  tenant: string;
  amount: number;
}

interface CashflowData {
  expected: CashflowExpected[];
  overdue: CashflowOverdue[];
  risks: CashflowRisk[];
  totalExpected: number;
  atRisk: number;
}

interface TariffData {
  name: string;
  count: number;
  mrr: number;
  avgCheck: number;
  growth: number;
}

interface SaasMetricsState {
  activeSection: string;
  overview: SaasOverview | null;
  mrrDynamics: MrrDynamicsMonth[];
  cashflow: CashflowData | null;
  tariffs: TariffData[];
  isLoading: boolean;
  setActiveSection: (section: string) => void;
  setOverview: (overview: SaasOverview) => void;
  setMrrDynamics: (dynamics: MrrDynamicsMonth[]) => void;
  setCashflow: (cashflow: CashflowData) => void;
  setTariffs: (tariffs: TariffData[]) => void;
  setLoading: (loading: boolean) => void;
  fetchAnalytics: () => Promise<void>;
  fetchPaymentStatus: () => Promise<void>;
}

const mockOverview: SaasOverview = {
  mrr: 2500000,
  mrrChange: 12.5,
  arr: 30000000,
  churnRate: 5.2,
  overdueAmount: 150000,
  overdueCount: 3,
  arpu: 15000,
};

const mockCashflow: CashflowData = {
  expected: [
    { date: '2025-05-01', tenant: 'Сеть А', amount: 500000 },
    { date: '2025-05-15', tenant: 'Сеть Б', amount: 300000 },
  ],
  overdue: [
    { date: '2025-04-01', tenant: 'Сеть В', amount: 100000, daysOverdue: 28 },
  ],
  risks: [],
  totalExpected: 800000,
  atRisk: 100000,
};

export const useSaasMetricsStore = create<SaasMetricsState>((set, get) => ({
  activeSection: 'metrics',
  overview: null,
  mrrDynamics: [],
  cashflow: null,
  tariffs: [],
  isLoading: false,
  setActiveSection: (section) => set({ activeSection: section }),
  setOverview: (overview) => set({ overview }),
  setMrrDynamics: (dynamics) => set({ mrrDynamics: dynamics }),
  setCashflow: (cashflow) => set({ cashflow }),
  setTariffs: (tariffs) => set({ tariffs }),
  setLoading: (loading) => set({ isLoading: loading }),
  fetchAnalytics: async () => {
    const { setOverview, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/analytics');
      const data = res.data as Record<string, unknown>;
      setOverview({
        mrr: Number(data.currentMrr) || mockOverview.mrr,
        mrrChange: Number(data.mrrGrowth) || mockOverview.mrrChange,
        arr: Number(data.arr) || mockOverview.arr,
        churnRate: Number(data.churnRate) || mockOverview.churnRate,
        overdueAmount: Number(data.overdueAmount) || mockOverview.overdueAmount,
        overdueCount: Number(data.overdueCount) || mockOverview.overdueCount,
        arpu: Number(data.arpu) || mockOverview.arpu,
      });
    } catch {
      setOverview(mockOverview);
    } finally {
      setLoading(false);
    }
  },
  fetchPaymentStatus: async () => {
    const { setCashflow, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/tenants/payments');
      const data = res.data as Array<Record<string, unknown>>;
      const expected: CashflowExpected[] = [];
      const overdue: CashflowOverdue[] = [];
      let totalExpected = 0;
      let atRisk = 0;
      data.forEach((item) => {
        const amount = Number(item.amount) || 0;
        const status = item.status as string;
        if (status === 'paid') {
          expected.push({
            date: (item.due_date as string) || '',
            tenant: (item.tenant_name as string) || '',
            amount,
          });
          totalExpected += amount;
        } else if (status === 'overdue') {
          overdue.push({
            date: (item.due_date as string) || '',
            tenant: (item.tenant_name as string) || '',
            amount,
            daysOverdue: Number(item.days_overdue) || 0,
          });
          atRisk += amount;
        }
      });
      setCashflow({ expected, overdue, risks: [], totalExpected, atRisk });
    } catch {
      setCashflow(mockCashflow);
    } finally {
      setLoading(false);
    }
  },
}));