import { create } from 'zustand';

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
}

export const useSaasMetricsStore = create<SaasMetricsState>((set) => ({
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
}));