import { create } from 'zustand';

interface AdminAction {
  id: string;
  timestamp: string;
  adminName: string;
  adminEmail: string;
  action: string;
  object: string;
  details: string;
  ip: string;
}

interface Impersonation {
  id: string;
  timestamp: string;
  adminName: string;
  tenant: string;
  role: string;
  userName: string;
  duration: number;
  actionsSummary: string;
  ip: string;
}

interface ActiveSession {
  id: string;
  adminName: string;
  ip: string;
  startedAt: string;
  lastActivity: string;
}

interface UserLogin {
  id: string;
  timestamp: string;
  tenant: string;
  userName: string;
  userEmail: string;
  role: string;
  action: string;
  ip: string;
  userAgent: string;
  geo?: string;
}

interface AuditFilters {
  admin: string;
  action: string;
  dateRange: [string, string] | null;
  tenant: string;
}

interface AuditState {
  adminActions: AdminAction[];
  impersonations: Impersonation[];
  activeSessions: ActiveSession[];
  userLogins: UserLogin[];
  filters: AuditFilters;
  isLoading: boolean;
  setAdminActions: (actions: AdminAction[]) => void;
  setImpersonations: (impersonations: Impersonation[]) => void;
  setActiveSessions: (sessions: ActiveSession[]) => void;
  setUserLogins: (logins: UserLogin[]) => void;
  setFilters: (filters: Partial<AuditFilters>) => void;
  setLoading: (loading: boolean) => void;
}

export const useAuditStore = create<AuditState>((set) => ({
  adminActions: [],
  impersonations: [],
  activeSessions: [],
  userLogins: [],
  filters: { admin: '', action: '', dateRange: null, tenant: '' },
  isLoading: false,
  setAdminActions: (actions) => set({ adminActions: actions }),
  setImpersonations: (impersonations) => set({ impersonations }),
  setActiveSessions: (sessions) => set({ activeSessions: sessions }),
  setUserLogins: (logins) => set({ userLogins: logins }),
  setFilters: (filters) => set((state) => ({ filters: { ...state.filters, ...filters } })),
  setLoading: (loading) => set({ isLoading: loading }),
}));