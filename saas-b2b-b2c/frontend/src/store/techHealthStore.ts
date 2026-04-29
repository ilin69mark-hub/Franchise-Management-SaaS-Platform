import { create } from 'zustand';
import apiClient from '@/api/axiosClient';

interface HealthStatus {
  uptime: number;
  avgResponseTime: number;
  p95ResponseTime: number;
  p99ResponseTime: number;
  errorRate: number;
  activeWebsockets: number;
}

interface PerformancePoint {
  timestamp: string;
  avgResponseTime: number;
  p95ResponseTime: number;
  p99ResponseTime: number;
  requestsPerSecond: number;
  errorRate4xx: number;
  errorRate5xx: number;
}

interface ErrorLog {
  id: string;
  timestamp: string;
  code: number;
  endpoint: string;
  tenant?: string;
  message: string;
  stackTrace?: string;
}

interface ServiceStatus {
  id: string;
  name: string;
  status: 'healthy' | 'degraded' | 'down';
  cpu: number;
  memory: number;
  instances: number;
  lastDeploy: string;
  version: string;
}

interface SecurityEvent {
  id: string;
  timestamp: string;
  type: 'failed_login' | 'unusual_ip' | 'mass_requests' | 'unauthorized_access';
  tenant?: string;
  ip: string;
  details: string;
  count: number;
}

interface TechHealthState {
  status: HealthStatus | null;
  performance: PerformancePoint[];
  errors: ErrorLog[];
  services: ServiceStatus[];
  securityEvents: SecurityEvent[];
  period: string;
  isLoading: boolean;
  setStatus: (status: HealthStatus) => void;
  setPerformance: (performance: PerformancePoint[]) => void;
  setErrors: (errors: ErrorLog[]) => void;
  setServices: (services: ServiceStatus[]) => void;
  setSecurityEvents: (events: SecurityEvent[]) => void;
  setPeriod: (period: string) => void;
  setLoading: (loading: boolean) => void;
  fetchRisks: () => Promise<void>;
  fetchEconomics: () => Promise<void>;
}

export const useTechHealthStore = create<TechHealthState>((set, get) => ({
  status: null,
  performance: [],
  errors: [],
  services: [],
  securityEvents: [],
  period: '24h',
  isLoading: false,
  setStatus: (status) => set({ status }),
  setPerformance: (performance) => set({ performance }),
  setErrors: (errors) => set({ errors }),
  setServices: (services) => set({ services }),
  setSecurityEvents: (events) => set({ securityEvents: events }),
  setPeriod: (period) => set({ period }),
  setLoading: (loading) => set({ isLoading: loading }),
  fetchRisks: async () => {
    const { setSecurityEvents, setErrors, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/risks');
      const data = res.data as Array<Record<string, unknown>> || [];
      const events: SecurityEvent[] = data.map((item, idx) => ({
        id: String(idx),
        timestamp: (item.created_at as string) || new Date().toISOString(),
        type: (item.risk_type as SecurityEvent['type']) || 'failed_login',
        tenant: (item.tenant_name as string) || undefined,
        ip: (item.ip as string) || '',
        details: (item.description as string) || '',
        count: Number(item.severity) || 1,
      }));
      setSecurityEvents(events);
    } catch {
    } finally {
      setLoading(false);
    }
  },
  fetchEconomics: async () => {
    const { setStatus, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/economics');
      const data = res.data as Record<string, unknown> || {};
      setStatus({
        uptime: 99.9,
        avgResponseTime: Number(data.avg_response_time) || 150,
        p95ResponseTime: Number(data.p95_response_time) || 300,
        p99ResponseTime: Number(data.p99_response_time) || 500,
        errorRate: Number(data.error_rate) || 0.1,
        activeWebsockets: 0,
      });
    } catch {
    } finally {
      setLoading(false);
    }
  },
}));