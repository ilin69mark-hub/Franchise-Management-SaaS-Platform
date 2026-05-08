// src/store/territoryManagerStore.ts
import { create } from 'zustand';
import apiClient from '@/api/axiosClient';

export interface TerritorySummary {
  planCompletionPercent: number;
  quarterForecastPercent: number;
  redZoneDealersCount: number;
  avgConversion: number;
  activeAlerts: number;
}

export interface DealerMetrics {
  dealerId: string;
  dealerName: string;
  salonCount: number;
  planPercent: number;
  forecastPercent: number;
  conversion: number;
  status: 'green' | 'yellow' | 'red';
  plan?: number;
  fact?: number;
  debt?: number;
  margin?: number;
  avgCheck?: number;
  taskCount?: number;
}

export interface TerritoryManager {
  id: string;
  name: string;
  territoryName: string;
  region: string;
}

interface TerritoryManagerState {
  activeTab: string;
  manager: TerritoryManager | null;
  summary: TerritorySummary | null;
  dealers: DealerMetrics[];
  isLoading: boolean;
  lastUpdated: Date | null;
  summaryModalOpen: boolean;
  setActiveTab: (tab: string) => void;
  setManager: (manager: TerritoryManager) => void;
  setSummary: (summary: TerritorySummary) => void;
  setDealers: (dealers: DealerMetrics[]) => void;
  setLoading: (loading: boolean) => void;
  setLastUpdated: (date: Date) => void;
  setSummaryModalOpen: (open: boolean) => void;
  fetchSummary: () => Promise<void>;
  fetchDealers: () => Promise<void>;
  fetchFunnel: (period?: string) => Promise<any>;
  fetchPlanFact: (period?: string) => Promise<any>;
  fetchBenchmarks: () => Promise<any>;
}

export const useTerritoryManagerStore = create<TerritoryManagerState>((set, get) => ({
  activeTab: 'map',
  manager: null,
  summary: null,
  dealers: [],
  isLoading: false,
  lastUpdated: null,
  summaryModalOpen: false,
  setActiveTab: (tab) => set({ activeTab: tab }),
  setManager: (manager) => set({ manager }),
  setSummary: (summary) => set({ summary }),
  setDealers: (dealers) => set({ dealers }),
  setLoading: (loading) => set({ isLoading: loading }),
  setLastUpdated: (date) => set({ lastUpdated: date }),
  setSummaryModalOpen: (open) => set({ summaryModalOpen: open }),

  fetchSummary: async () => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get('/territory/summary');
      const data = res.data;
      set({
        summary: {
          planCompletionPercent: data.planCompletionPercent || 0,
          quarterForecastPercent: data.quarterForecastPercent || 0,
          redZoneDealersCount: data.redZoneDealersCount || 0,
          avgConversion: data.avgConversion || 0,
          activeAlerts: data.activeAlerts || 0,
        },
      });
    } catch (e) {
      console.error('Error fetching territory summary:', e);
      set({
        summary: {
          planCompletionPercent: 82,
          quarterForecastPercent: 91,
          redZoneDealersCount: 2,
          avgConversion: 4.2,
          activeAlerts: 5,
        },
      });
    } finally {
      set({ isLoading: false, lastUpdated: new Date() });
    }
  },

  fetchDealers: async () => {
    try {
      const res = await apiClient.get('/territory/planfact?period=month');
      const data = res.data;
      
      const dealers: DealerMetrics[] = (data.dealers || []).map((d: any) => ({
        dealerId: d.id,
        dealerName: d.dealer_name || '',
        salonCount: d.salon_count || 0,
        planPercent: d.plan_percent || 0,
        forecastPercent: d.plan_percent >= 80 ? 100 : d.plan_percent >= 50 ? 70 : 30,
        conversion: d.conversion || 0,
        status: d.plan_percent >= 80 ? 'green' : d.plan_percent >= 50 ? 'yellow' : 'red',
        plan: d.plan || 0,
        fact: d.fact || 0,
      }));
      
      set({ dealers });
    } catch (e) {
      console.error('Error fetching dealers:', e);
    }
  },

  fetchFunnel: async (period = 'month') => {
    try {
      const res = await apiClient.get(`/territory/funnel?period=${period}`);
      return res.data;
    } catch (e) {
      console.error('Error fetching funnel:', e);
      return null;
    }
  },

  fetchPlanFact: async (period = 'month') => {
    try {
      const res = await apiClient.get(`/territory/planfact?period=${period}`);
      return res.data;
    } catch (e) {
      console.error('Error fetching planfact:', e);
      return null;
    }
  },

  fetchBenchmarks: async () => {
    try {
      const res = await apiClient.get('/territory/benchmarks');
      return res.data;
    } catch (e) {
      console.error('Error fetching benchmarks:', e);
      return null;
    }
  },
}));