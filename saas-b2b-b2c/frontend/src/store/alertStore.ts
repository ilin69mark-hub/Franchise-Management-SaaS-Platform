// src/store/alertStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type AlertCategory = 'plan' | 'funnel' | 'stock' | 'communications' | 'activity';
export type AlertPriority = 'critical' | 'warning' | 'info';
export type AlertStatus = 'new' | 'read' | 'acknowledged';

export interface Alert {
  id: string;
  category: AlertCategory;
  priority: AlertPriority;
  dealerName?: string;
  title: string;
  description: string;
  details?: string;
  createdAt: string;
  status: AlertStatus;
  deepLink?: string;
}

export interface AlertSettings {
  enabled: boolean;
  planThreshold: number;
  conversionDropThreshold: number;
  trafficDropThreshold: number;
  stuckDealDays: number;
  stuckDealAmount: number;
  inactiveDealerDays: number;
  nonLiquidDays: number;
  channels: {
    inApp: boolean;
    emailInstant: boolean;
    emailDigest: boolean;
    push: boolean;
  };
}

interface AlertState {
  alerts: Alert[];
  settings: AlertSettings;
  filter: 'all' | 'critical' | AlertCategory;
  isLoading: boolean;
  setAlerts: (alerts: Alert[]) => void;
  addAlert: (alert: Alert) => void;
  markAsRead: (alertId: string) => void;
  markAllAsRead: () => void;
  setFilter: (filter: 'all' | 'critical' | AlertCategory) => void;
  setSettings: (settings: Partial<AlertSettings>) => void;
  getUnreadCount: () => number;
}

const defaultSettings: AlertSettings = {
  enabled: true,
  planThreshold: 70,
  conversionDropThreshold: 15,
  trafficDropThreshold: 30,
  stuckDealDays: 7,
  stuckDealAmount: 200000,
  inactiveDealerDays: 3,
  nonLiquidDays: 90,
  channels: {
    inApp: true,
    emailInstant: false,
    emailDigest: true,
    push: false,
  },
};

export const useAlertStore = create<AlertState>()(
  persist(
    (set, get) => ({
      alerts: [],
      settings: defaultSettings,
      filter: 'all',
      isLoading: false,
      
      setAlerts: (alerts) => set({ alerts }),
      
      addAlert: (alert) => set((state) => ({
        alerts: [alert, ...state.alerts],
      })),
      
      markAsRead: (alertId) => set((state) => ({
        alerts: state.alerts.map(a => 
          a.id === alertId ? { ...a, status: 'read' as AlertStatus } : a
        ),
      })),
      
      markAllAsRead: () => set((state) => ({
        alerts: state.alerts.map(a => ({ ...a, status: 'read' as AlertStatus })),
      })),
      
      setFilter: (filter) => set({ filter }),
      
      setSettings: (newSettings) => set((state) => ({
        settings: { ...state.settings, ...newSettings },
      })),
      
      getUnreadCount: () => get().alerts.filter(a => a.status === 'new').length,
    }),
    {
      name: 'alert-storage',
      partialize: (state) => ({ alerts: state.alerts, settings: state.settings }),
    }
  )
);