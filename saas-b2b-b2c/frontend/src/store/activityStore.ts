import { create } from 'zustand';

interface ActivityOverview {
  dau: number;
  dauChange: number;
  wau: number;
  wauChange: number;
  mau: number;
  mauChange: number;
  stickiness: number;
  avgSessionTime: number;
  activeSessions: number;
}

interface ActivityDynamics {
  date: string;
  dau: number;
  wau: number;
  mau: number;
}

interface TenantActivityItem {
  id: string;
  name: string;
  dau: number;
  wau: number;
  mau: number;
  stickiness: number;
  activeLicensePercent: number;
}

interface FeatureAdoptionItem {
  id: string;
  name: string;
  tenantPercent: number;
  userPercent: number;
  frequency: 'daily' | 'weekly' | 'monthly';
  trend: 'up' | 'stable' | 'down';
}

interface TtvItem {
  id: string;
  name: string;
  createdAt: string;
  firstSaleAt: string;
  ttvDays: number;
}

interface ActivityState {
  overview: ActivityOverview | null;
  dynamics: ActivityDynamics[];
  byTenant: TenantActivityItem[];
  featureAdoption: FeatureAdoptionItem[];
  ttv: TtvItem[];
  period: number;
  selectedTenant: string;
  isLoading: boolean;
  setOverview: (overview: ActivityOverview) => void;
  setDynamics: (dynamics: ActivityDynamics[]) => void;
  setByTenant: (byTenant: TenantActivityItem[]) => void;
  setFeatureAdoption: (featureAdoption: FeatureAdoptionItem[]) => void;
  setTtv: (ttv: TtvItem[]) => void;
  setPeriod: (period: number) => void;
  setSelectedTenant: (tenant: string) => void;
  setLoading: (loading: boolean) => void;
}

export const useActivityStore = create<ActivityState>((set) => ({
  overview: null,
  dynamics: [],
  byTenant: [],
  featureAdoption: [],
  ttv: [],
  period: 30,
  selectedTenant: '',
  isLoading: false,
  setOverview: (overview) => set({ overview }),
  setDynamics: (dynamics) => set({ dynamics }),
  setByTenant: (byTenant) => set({ byTenant }),
  setFeatureAdoption: (featureAdoption) => set({ featureAdoption }),
  setTtv: (ttv) => set({ ttv }),
  setPeriod: (period) => set({ period }),
  setSelectedTenant: (tenant) => set({ selectedTenant: tenant }),
  setLoading: (loading) => set({ isLoading: loading }),
}));