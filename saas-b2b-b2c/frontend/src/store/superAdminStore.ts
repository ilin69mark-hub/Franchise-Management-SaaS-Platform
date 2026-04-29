import { create } from 'zustand';
import apiClient from '@/api/axiosClient';

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
  fetchStats: () => Promise<void>;
}

const mockStats: SuperAdminStats = {
  mrr: 2500000,
  arr: 30000000,
  churnRate: 5.2,
  overduePayments: 150000,
  arpu: 15000,
};

export const useSuperAdminStore = create<SuperAdminState>((set, get) => ({
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
  fetchStats: async () => {
    const { setLoading, setStats, setLastUpdated } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/stats');
      const data = res.data as Record<string, unknown>;
      setStats({
        mrr: Number(data.mrr) || 0,
        arr: Number(data.arr) || 0,
        churnRate: Number(data.churnRate) || 0,
        overduePayments: Number(data.overduePayments) || 0,
        arpu: Number(data.arpu) || 0,
      });
      setLastUpdated(new Date());
    } catch {
      setStats(mockStats);
    } finally {
      setLoading(false);
    }
  },
}));