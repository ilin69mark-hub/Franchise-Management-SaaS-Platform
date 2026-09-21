import { create } from 'zustand';
import apiClient from '../api/axiosClient';

interface Alert {
  id: string;
  type: 'billing' | 'technical' | 'security' | 'activity';
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
  link?: string;
}

interface AlertSettings {
  channels: ('in-app' | 'email' | 'telegram')[];
  thresholds: {
    errorRate: number;
    responseTime: number;
    uptime: number;
    churnRate: number;
  };
}

interface AlertState {
  alerts: Alert[];
  unreadCount: number;
  settings: AlertSettings | null;
  isLoading: boolean;
  setAlerts: (alerts: Alert[]) => void;
  setSettings: (settings: AlertSettings) => void;
  setLoading: (loading: boolean) => void;
  loadAlerts: () => Promise<void>;
  markAsRead: (id: string) => Promise<void>;
  updateSettings: (settings: Partial<AlertSettings>) => Promise<void>;
}

export const useAlertStore = create<AlertState>((set, get) => ({
  alerts: [],
  unreadCount: 0,
  settings: null,
  isLoading: false,
  setAlerts: (alerts) => set({ alerts, unreadCount: alerts.filter(a => !a.read).length }),
  setSettings: (settings) => set({ settings }),
  setLoading: (loading) => set({ isLoading: loading }),
  loadAlerts: async () => {
    set({ isLoading: true });
    try {
      const res = await apiClient.get('/admin/alerts/unread');
      get().setAlerts(Array.isArray(res.data) ? res.data : (res.data?.alerts || []));
    } catch (e) {
      // Backend currently has no /admin/alerts routes — keep empty alerts, don't throw.
    } finally {
      set({ isLoading: false });
    }
  },
  markAsRead: async (id: string) => {
    try {
      await apiClient.patch(`/admin/alerts/${id}/read`);
    } catch (e) {
      // Backend has no /admin/alerts route — keep the local state update.
    }
    const alerts = get().alerts.map(a => a.id === id ? { ...a, read: true } : a);
    get().setAlerts(alerts);
  },
  updateSettings: async (settings: Partial<AlertSettings>) => {
    try {
      await apiClient.put('/admin/alert-settings', settings);
    } catch (e) {
      // Backend has no /admin/alert-settings route — keep local settings, don't throw.
    }
    get().loadAlerts();
  },
}));