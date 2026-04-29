// src/store/territoryManagerStore.ts
import { create } from 'zustand';

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
}

export const useTerritoryManagerStore = create<TerritoryManagerState>((set) => ({
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
}));