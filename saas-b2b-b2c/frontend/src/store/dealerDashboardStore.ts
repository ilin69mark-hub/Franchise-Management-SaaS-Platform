// src/store/dealerDashboardStore.ts
import { create } from 'zustand';

interface DashboardSummary {
  netProfit: number;
  grossRevenue: number;
  planCompletionPercent: number;
  marginProfit: number;
  activeAlerts: number;
}

interface DealerDashboardState {
  activeTab: string;
  summary: DashboardSummary | null;
  dealerCenterName: string;
  isLoading: boolean;
  lastUpdated: Date | null;
  setActiveTab: (tab: string) => void;
  setSummary: (summary: DashboardSummary) => void;
  setDealerCenterName: (name: string) => void;
  setLoading: (loading: boolean) => void;
  setLastUpdated: (date: Date) => void;
}

export const useDealerDashboardStore = create<DealerDashboardState>((set) => ({
  activeTab: 'profit',
  summary: null,
  dealerCenterName: 'Дилерский центр',
  isLoading: false,
  lastUpdated: null,
  setActiveTab: (tab) => set({ activeTab: tab }),
  setSummary: (summary) => set({ summary }),
  setDealerCenterName: (name) => set({ dealerCenterName: name }),
  setLoading: (loading) => set({ isLoading: loading }),
  setLastUpdated: (date) => set({ lastUpdated: date }),
}));