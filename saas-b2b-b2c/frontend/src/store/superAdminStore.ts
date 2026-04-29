import { create } from 'zustand';

interface SuperAdminStats {
  mrr: number;
  arr: number;
  churnRate: number;
  overduePayments: number;
  arpu: number;
}

interface SuperAdminState {
  activeSection: string;
  stats: SuperAdminStats;
  isLoading: boolean;
  alertCount: number;
  lastUpdated: Date | null;
  setActiveSection: (section: string) => void;
  setStats: (stats: SuperAdminStats) => void;
  setLoading: (loading: boolean) => void;
  setAlertCount: (count: number) => void;
  setLastUpdated: (date: Date) => void;
}

export const useSuperAdminStore = create<SuperAdminState>((set) => ({
  activeSection: 'metrics',
  stats: {
    mrr: 0,
    arr: 0,
    churnRate: 0,
    overduePayments: 0,
    arpu: 0,
  },
  isLoading: false,
  alertCount: 0,
  lastUpdated: null,
  setActiveSection: (section) => set({ activeSection: section }),
  setStats: (stats) => set({ stats }),
  setLoading: (loading) => set({ isLoading: loading }),
  setAlertCount: (count) => set({ alertCount: count }),
  setLastUpdated: (date) => set({ lastUpdated: date }),
}));