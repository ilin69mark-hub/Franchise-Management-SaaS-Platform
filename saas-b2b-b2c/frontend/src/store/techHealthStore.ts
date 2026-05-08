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
  errorAlert: boolean;
  pollingInterval: number | null;
  setStatus: (status: HealthStatus) => void;
  setPerformance: (performance: PerformancePoint[]) => void;
  setErrors: (errors: ErrorLog[]) => void;
  setServices: (services: ServiceStatus[]) => void;
  setSecurityEvents: (events: SecurityEvent[]) => void;
  setPeriod: (period: string) => void;
  setLoading: (loading: boolean) => void;
  setErrorAlert: (alert: boolean) => void;
  fetchRisks: () => Promise<void>;
  fetchEconomics: () => Promise<void>;
  fetchServiceStatus: () => Promise<void>;
  fetchResponseTimes: () => Promise<void>;
  fetchErrorRates: () => Promise<void>;
  startPolling: (intervalMs?: number) => void;
  stopPolling: () => void;
}

export const useTechHealthStore = create<TechHealthState>((set, get) => ({
  status: null,
  performance: [],
  errors: [],
  services: [],
  securityEvents: [],
  period: '24h',
  isLoading: false,
  errorAlert: false,
  pollingInterval: null,
  setStatus: (status) => set({ status }),
  setPerformance: (performance) => set({ performance }),
  setErrors: (errors) => set({ errors }),
  setServices: (services) => set({ services }),
  setSecurityEvents: (events) => set({ securityEvents: events }),
  setPeriod: (period) => set({ period }),
  setLoading: (loading) => set({ isLoading: loading }),
  setErrorAlert: (alert) => set({ errorAlert: alert }),
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
  fetchServiceStatus: async () => {
    const { setServices, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/services/status');
      const data = res.data as Array<Record<string, unknown>> || [];
      const services: ServiceStatus[] = data.map((item) => ({
        id: String(item.id || ''),
        name: String(item.name || 'Unknown'),
        status: (item.status as ServiceStatus['status']) || 'healthy',
        cpu: Number(item.cpu) || 0,
        memory: Number(item.memory) || 0,
        instances: Number(item.instances) || 1,
        lastDeploy: String(item.last_deploy || ''),
        version: String(item.version || '1.0.0'),
      }));
      setServices(services);
    } catch {
      setServices([]);
    } finally {
      setLoading(false);
    }
  },
  fetchResponseTimes: async () => {
    const { setPerformance, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/metrics/response-times');
      const data = res.data as Array<Record<string, unknown>> || [];
      const performance: PerformancePoint[] = data.map((item) => ({
        timestamp: String(item.timestamp || new Date().toISOString()),
        avgResponseTime: Number(item.avg_response_time) || 0,
        p95ResponseTime: Number(item.p95_response_time) || 0,
        p99ResponseTime: Number(item.p99_response_time) || 0,
        requestsPerSecond: Number(item.requests_per_second) || 0,
        errorRate4xx: Number(item.error_rate_4xx) || 0,
        errorRate5xx: Number(item.error_rate_5xx) || 0,
      }));
      setPerformance(performance);
    } catch {
      setPerformance([]);
    } finally {
      setLoading(false);
    }
  },
  fetchErrorRates: async () => {
    const { setErrors, setErrorAlert, setLoading } = get();
    setLoading(true);
    try {
      const res = await apiClient.get('/admin/metrics/error-rates');
      const data = res.data as { errors?: Array<Record<string, unknown>>; total_requests?: number } || {};
      const errorList = data.errors || [];
      const errors: ErrorLog[] = errorList.map((item, idx) => ({
        id: String(item.id || idx),
        timestamp: String(item.timestamp || new Date().toISOString()),
        code: Number(item.code) || 500,
        endpoint: String(item.endpoint || ''),
        tenant: item.tenant ? String(item.tenant) : undefined,
        message: String(item.message || 'Unknown error'),
        stackTrace: item.stack_trace ? String(item.stack_trace) : undefined,
      }));
      setErrors(errors);
      const totalErrors = errors.length;
      const totalRequests = Number(data.total_requests) || 1000;
      const errorRate = totalRequests > 0 ? (totalErrors / totalRequests) * 100 : 0;
      setErrorAlert(errorRate > 5);
    } catch {
      setErrors([]);
      setErrorAlert(false);
    } finally {
      setLoading(false);
    }
  },
  startPolling: (intervalMs = 30000) => {
    const { pollingInterval, fetchServiceStatus } = get();
    if (pollingInterval) return;
    const interval = window.setInterval(() => {
      fetchServiceStatus();
    }, intervalMs);
    set({ pollingInterval: interval as number });
  },
  stopPolling: () => {
    const { pollingInterval } = get();
    if (pollingInterval) {
      clearInterval(pollingInterval);
      set({ pollingInterval: null });
    }
  },
}));