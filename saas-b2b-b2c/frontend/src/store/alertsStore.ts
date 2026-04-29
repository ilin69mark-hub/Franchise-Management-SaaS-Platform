import { create } from 'zustand';

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
      const res = await fetch('/api/admin/alerts/unread');
      const data = await res.json();
      get().setAlerts(data);
    } finally {
      set({ isLoading: false });
    }
  },
  markAsRead: async (id: string) => {
    await fetch(`/api/admin/alerts/${id}/read`, { method: 'PATCH' });
    const alerts = get().alerts.map(a => a.id === id ? { ...a, read: true } : a);
    get().setAlerts(alerts);
  },
  updateSettings: async (settings: Partial<AlertSettings>) => {
    await fetch('/api/admin/alert-settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings),
    });
    get().loadAlerts();
  },
}));