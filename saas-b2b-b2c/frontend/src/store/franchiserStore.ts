import { create } from 'zustand';
import { persist } from 'zustand/middleware';

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
}

const mockSummary: FranchiserSummary = {
  planPercent: 78,
  forecastPercent: 85,
  activeDealers: 24,
  avgConversion: 12,
  avgMargin: 32,
};

const mockDealers: DealerMetrics[] = [
  { dealerId: '1', dealerName: 'Дилер Север', salonCount: 3, planPercent: 92, forecastPercent: 95, conversion: 14, margin: 35, status: 'green', plan: 5000000, fact: 4600000 },
  { dealerId: '2', dealerName: 'Дилер Юг', salonCount: 2, planPercent: 65, forecastPercent: 70, conversion: 10, margin: 28, status: 'yellow', plan: 3000000, fact: 1950000 },
  { dealerId: '3', dealerName: 'Дилер Запад', salonCount: 4, planPercent: 45, forecastPercent: 55, conversion: 8, margin: 22, status: 'red', plan: 8000000, fact: 3600000 },
];

const mockAlerts: FranchiseAlert[] = [
  { id: '1', type: 'critical', message: 'Дилер Юг: падение конверсии >20%', dealerId: '2', createdAt: new Date().toISOString() },
  { id: '2', type: 'warning', message: 'Дилер Запад: выполнение плана <50%', dealerId: '3', createdAt: new Date().toISOString() },
];

export const useFranchiserStore = create<FranchiserState>()(
  persist(
    (set, get) => ({
      summary: mockSummary,
      dealers: mockDealers,
      alerts: mockAlerts,
      alertCount: mockAlerts.filter(a => a.type === 'critical').length,
      activeTab: 'network',
      isLoading: false,

      setSummary: (summary) => set({ summary }),
      setDealers: (dealers) => set({ dealers }),
      setAlerts: (alerts) => set({ alerts, alertCount: alerts.filter(a => a.type === 'critical').length }),
      setAlertCount: (alertCount) => set({ alertCount }),
      setActiveTab: (activeTab) => set({ activeTab }),
      setLoading: (isLoading) => set({ isLoading }),

      fetchSummary: async () => {
        set({ isLoading: true });
        await new Promise(resolve => setTimeout(resolve, 500));
        const { alerts } = get();
        set({ 
          isLoading: false,
          alertCount: alerts.filter(a => a.type === 'critical').length 
        });
      },
    }),
    {
      name: 'franchiser-storage',
      partialize: (state) => ({ summary: state.summary, dealers: state.dealers, alerts: state.alerts }),
    }
  )
);