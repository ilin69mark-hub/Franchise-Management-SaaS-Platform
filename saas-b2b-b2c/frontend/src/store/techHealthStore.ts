import { create } from 'zustand';

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
}

export const useTechHealthStore = create<TechHealthState>((set) => ({
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
}));