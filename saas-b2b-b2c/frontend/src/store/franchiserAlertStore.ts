import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type AlertType = 'plan' | 'dealer' | 'manager' | 'escalation';
export type AlertSeverity = 'critical' | 'warning';
export type AlertStatus = 'new' | 'acknowledged' | 'resolved';

export interface FranchiserAlert {
  id: string;
  type: AlertType;
  severity: AlertSeverity;
  title: string;
  description: string;
  metric?: number;
  target?: string;
  createdAt: string;
  status: AlertStatus;
  assignedTo?: string;
  resolvedAt?: string;
}

export interface FranchiserAlertSettings {
  enabled: boolean;
  channel: {
    inApp: boolean;
    email: boolean;
    push: boolean;
  };
  thresholds: {
    criticalForecastPercent: number;
    dealerChurnPercent: number;
    managerKpiPercent: number;
  };
}

interface FranchiserAlertState {
  alerts: FranchiserAlert[];
  settings: FranchiserAlertSettings;
  isOpen: boolean;
  filter: 'all' | 'critical' | AlertType;
  
  setAlerts: (alerts: FranchiserAlert[]) => void;
  addAlert: (alert: FranchiserAlert) => void;
  markAsRead: (alertId: string) => void;
  markAllAsRead: () => void;
  assignAlert: (alertId: string, managerId: string) => void;
  resolveAlert: (alertId: string) => void;
  setOpen: (open: boolean) => void;
  setFilter: (filter: 'all' | 'critical' | AlertType) => void;
  setSettings: (settings: Partial<FranchiserAlertSettings>) => void;
  getUnreadCount: () => number;
  getCriticalCount: () => number;
}

const defaultSettings: FranchiserAlertSettings = {
  enabled: true,
  channel: {
    inApp: true,
    email: false,
    push: false,
  },
  thresholds: {
    criticalForecastPercent: 90,
    dealerChurnPercent: 5,
    managerKpiPercent: 70,
  },
};

const mockAlerts: FranchiserAlert[] = [
  {
    id: '1',
    type: 'plan',
    severity: 'critical',
    title: 'Прогноз сети опустился ниже 90%',
    description: 'Текущий прогноз закрытия квартала: 85%. Порог: 90%.',
    metric: 85,
    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
    status: 'new',
  },
  {
    id: '2',
    type: 'dealer',
    severity: 'critical',
    title: 'Дилер категории А перешёл в красную зону',
    description: 'Дилер "Мебель Плюс" (12% выручки сети) опустился до 65% плана.',
    metric: 65,
    target: 'Мебель Плюс',
    createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    status: 'new',
  },
  {
    id: '3',
    type: 'dealer',
    severity: 'warning',
    title: 'Массовый отток дилеров',
    description: 'За квартал выбыло 6 дилеров (25% от сети). Порог: 5%.',
    metric: 25,
    createdAt: new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString(),
    status: 'acknowledged',
    assignedTo: 'Алексей Петров',
  },
  {
    id: '4',
    type: 'manager',
    severity: 'warning',
    title: 'Менеджер: KPI ниже порога',
    description: 'Мария Иванова: KPI 68% два месяца подряд. Порог: 70%.',
    metric: 68,
    target: 'Мария Иванова',
    createdAt: new Date(Date.now() - 72 * 60 * 60 * 1000).toISOString(),
    status: 'new',
  },
  {
    id: '5',
    type: 'manager',
    severity: 'warning',
    title: 'Менеджер: SLA ниже 80%',
    description: 'Сергей Сидоров: SLA 72% за последние две недели.',
    metric: 72,
    target: 'Сергей Сидоров',
    createdAt: new Date(Date.now() - 96 * 60 * 60 * 1000).toISOString(),
    status: 'new',
  },
  {
    id: '6',
    type: 'escalation',
    severity: 'critical',
    title: 'Эскалация от дилера',
    description: 'Жалоба на менеджера Марию Иванову по поводу задержки поставок.',
    target: 'Мария Иванова',
    createdAt: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString(),
    status: 'new',
  },
  {
    id: '7',
    type: 'plan',
    severity: 'critical',
    title: 'Выручка сети упала > 10%',
    description: 'Апрель 2026: -12% к апрелю 2025. Сезонный спад или системная проблема?',
    metric: -12,
    createdAt: new Date(Date.now() - 168 * 60 * 60 * 1000).toISOString(),
    status: 'resolved',
    resolvedAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
  },
];

export const useFranchiserAlertStore = create<FranchiserAlertState>()(
  persist(
    (set, get) => ({
      alerts: mockAlerts,
      settings: defaultSettings,
      isOpen: false,
      filter: 'all',

      setAlerts: (alerts) => set({ alerts }),
      
      addAlert: (alert) => set((state) => ({
        alerts: [alert, ...state.alerts],
      })),

      markAsRead: (alertId) => set((state) => ({
        alerts: state.alerts.map(a => 
          a.id === alertId ? { ...a, status: 'acknowledged' as AlertStatus } : a
        ),
      })),

      markAllAsRead: () => set((state) => ({
        alerts: state.alerts.map(a => ({ ...a, status: 'acknowledged' as AlertStatus })),
      })),

      assignAlert: (alertId, managerId) => set((state) => ({
        alerts: state.alerts.map(a => 
          a.id === alertId ? { ...a, assignedTo: managerId, status: 'acknowledged' as AlertStatus } : a
        ),
      })),

      resolveAlert: (alertId) => set((state) => ({
        alerts: state.alerts.map(a => 
          a.id === alertId ? { ...a, status: 'resolved' as AlertStatus, resolvedAt: new Date().toISOString() } : a
        ),
      })),

      setOpen: (isOpen) => set({ isOpen }),
      
      setFilter: (filter) => set({ filter }),

      setSettings: (newSettings) => set((state) => ({
        settings: { ...state.settings, ...newSettings },
      })),

      getUnreadCount: () => {
        const { alerts } = get();
        return alerts.filter(a => a.status === 'new').length;
      },

      getCriticalCount: () => {
        const { alerts } = get();
        return alerts.filter(a => a.severity === 'critical' && a.status !== 'resolved').length;
      },
    }),
    {
      name: 'franchiser-alert-storage',
      partialize: (state) => ({ alerts: state.alerts, settings: state.settings }),
    }
  )
);