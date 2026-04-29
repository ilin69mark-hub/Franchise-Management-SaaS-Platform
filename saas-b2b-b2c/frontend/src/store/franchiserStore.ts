import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import apiClient from '@/api/axiosClient';

export interface FranchiserSummary {
  planPercent: number;
  forecastPercent: number;
  activeDealers: number;
  avgConversion: number;
  avgMargin: number;
}

export interface DealerMetrics {
  dealerId: string;
  dealerName: string;
  salonCount: number;
  planPercent: number;
  forecastPercent: number;
  conversion: number;
  margin: number;
  status: 'green' | 'yellow' | 'red';
  plan?: number;
  fact?: number;
  debt?: number;
  avgCheck?: number;
}

export interface FranchiseAlert {
  id: string;
  type: 'critical' | 'warning' | 'info';
  message: string;
  dealerId?: string;
  createdAt: string;
}

interface FranchiserState {
  summary: FranchiserSummary;
  dealers: DealerMetrics[];
  alerts: FranchiseAlert[];
  alertCount: number;
  activeTab: string;
  isLoading: boolean;
  setSummary: (summary: FranchiserSummary) => void;
  setDealers: (dealers: DealerMetrics[]) => void;
  setAlerts: (alerts: FranchiseAlert[]) => void;
  setAlertCount: (count: number) => void;
  setActiveTab: (tab: string) => void;
  setLoading: (loading: boolean) => void;
  fetchSummary: () => Promise<void>;
  fetchNetwork: () => Promise<void>;
  fetchHealth: () => Promise<any>;
  fetchTeam: () => Promise<any>;
}

const defaultSummary: FranchiserSummary = {
  planPercent: 78,
  forecastPercent: 85,
  activeDealers: 24,
  avgConversion: 12,
  avgMargin: 32,
};

export const useFranchiserStore = create<FranchiserState>()(
  persist(
    (set) => ({
      summary: defaultSummary,
      dealers: [],
      alerts: [],
      alertCount: 0,
      activeTab: 'network',
      isLoading: false,

      setSummary: (summary) => set({ summary }),
      setDealers: (dealers) => set({ dealers }),
      setAlerts: (alerts) => set({ alerts, alertCount: alerts.filter((a: FranchiseAlert) => a.type === 'critical').length }),
      setAlertCount: (alertCount) => set({ alertCount }),
      setActiveTab: (activeTab) => set({ activeTab }),
      setLoading: (isLoading) => set({ isLoading }),

      fetchSummary: async () => {
        set({ isLoading: true });
        try {
          const res = await apiClient.get('/franchiser/summary');
          const data = res.data;
          set({
            summary: {
              planPercent: data.planPercent || 0,
              forecastPercent: data.forecastPercent || 0,
              activeDealers: data.activeDealers || 0,
              avgConversion: data.avgConversion || 0,
              avgMargin: data.avgMargin || 0,
            },
            isLoading: false,
          });
        } catch (e) {
          console.error('Error fetching franchiser summary:', e);
          set({ isLoading: false });
        }
      },

      fetchNetwork: async () => {
        try {
          const res = await apiClient.get('/franchiser/network?period=month');
          const data = res.data;
          
          const dealers: DealerMetrics[] = (data.network_data || []).map((d: any) => ({
            dealerId: d.id,
            dealerName: d.name || '',
            salonCount: d.salon_count || 0,
            planPercent: d.plan_percent || 0,
            forecastPercent: d.plan_percent >= 80 ? 100 : d.plan_percent >= 50 ? 70 : 30,
            conversion: d.conversion || 0,
            margin: d.margin || 0,
            status: d.forecast || 'green',
            plan: d.plan || 0,
            fact: d.fact || 0,
          }));
          
          set({ dealers });
        } catch (e) {
          console.error('Error fetching network:', e);
        }
      },

      fetchHealth: async () => {
        try {
          const res = await apiClient.get('/franchiser/health');
          return res.data;
        } catch (e) {
          console.error('Error fetching health:', e);
          return null;
        }
      },

      fetchTeam: async () => {
        try {
          const res = await apiClient.get('/franchiser/team');
          return res.data;
        } catch (e) {
          console.error('Error fetching team:', e);
          return null;
        }
      },
    }),
    {
      name: 'franchiser-storage',
      partialize: (state) => ({ summary: state.summary, dealers: state.dealers, alerts: state.alerts }),
    }
  )
);