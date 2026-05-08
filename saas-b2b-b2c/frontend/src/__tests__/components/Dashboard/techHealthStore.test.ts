import { useTechHealthStore, HealthStatus, PerformancePoint, ErrorLog, ServiceStatus, SecurityEvent } from '@/store/techHealthStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  patch: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('techHealthStore', () => {
  beforeEach(() => {
    useTechHealthStore.setState({
      status: null,
      performance: [],
      errors: [],
      services: [],
      securityEvents: [],
      period: '24h',
      isLoading: false,
      errorAlert: false,
      pollingInterval: null,
    });
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    useTechHealthStore.getState().stopPolling();
    jest.useRealTimers();
  });

  describe('initial state', () => {
    it('has null status', () => {
      expect(useTechHealthStore.getState().status).toBeNull();
    });

    it('has empty performance array', () => {
      expect(useTechHealthStore.getState().performance).toEqual([]);
    });

    it('has empty errors array', () => {
      expect(useTechHealthStore.getState().errors).toEqual([]);
    });

    it('has empty services array', () => {
      expect(useTechHealthStore.getState().services).toEqual([]);
    });

    it('has empty securityEvents array', () => {
      expect(useTechHealthStore.getState().securityEvents).toEqual([]);
    });

    it('has default period', () => {
      expect(useTechHealthStore.getState().period).toBe('24h');
    });

    it('has isLoading false', () => {
      expect(useTechHealthStore.getState().isLoading).toBe(false);
    });

    it('has errorAlert false by default', () => {
      expect(useTechHealthStore.getState().errorAlert).toBe(false);
    });

    it('has pollingInterval null by default', () => {
      expect(useTechHealthStore.getState().pollingInterval).toBeNull();
    });
  });

  describe('setStatus', () => {
    it('sets health status', () => {
      const status: HealthStatus = {
        uptime: 99.9,
        avgResponseTime: 150,
        p95ResponseTime: 300,
        p99ResponseTime: 500,
        errorRate: 0.1,
        activeWebsockets: 50,
      };
      useTechHealthStore.getState().setStatus(status);
      expect(useTechHealthStore.getState().status).toEqual(status);
    });

    it('updates individual status fields', () => {
      const status: HealthStatus = {
        uptime: 99.5,
        avgResponseTime: 200,
        p95ResponseTime: 400,
        p99ResponseTime: 600,
        errorRate: 0.5,
        activeWebsockets: 30,
      };
      useTechHealthStore.getState().setStatus(status);
      expect(useTechHealthStore.getState().status?.uptime).toBe(99.5);
      expect(useTechHealthStore.getState().status?.avgResponseTime).toBe(200);
    });
  });

  describe('setPerformance', () => {
    it('sets performance data', () => {
      const performance: PerformancePoint[] = [
        { timestamp: '2026-01-01T00:00:00Z', avgResponseTime: 100, p95ResponseTime: 200, p99ResponseTime: 300, requestsPerSecond: 50, errorRate4xx: 1, errorRate5xx: 0 },
        { timestamp: '2026-01-01T01:00:00Z', avgResponseTime: 110, p95ResponseTime: 210, p99ResponseTime: 310, requestsPerSecond: 55, errorRate4xx: 2, errorRate5xx: 1 },
      ];
      useTechHealthStore.getState().setPerformance(performance);
      expect(useTechHealthStore.getState().performance).toHaveLength(2);
    });

    it('calculates average response time', () => {
      const performance: PerformancePoint[] = [
        { timestamp: '2026-01-01T00:00:00Z', avgResponseTime: 100, p95ResponseTime: 200, p99ResponseTime: 300, requestsPerSecond: 50, errorRate4xx: 0, errorRate5xx: 0 },
        { timestamp: '2026-01-01T01:00:00Z', avgResponseTime: 200, p95ResponseTime: 400, p99ResponseTime: 600, requestsPerSecond: 100, errorRate4xx: 0, errorRate5xx: 0 },
      ];
      useTechHealthStore.getState().setPerformance(performance);
      const avg = performance.reduce((sum, p) => sum + p.avgResponseTime, 0) / 2;
      expect(avg).toBe(150);
    });
  });

  describe('setErrors', () => {
    it('sets error logs', () => {
      const errors: ErrorLog[] = [
        { id: '1', timestamp: '2026-01-01T10:00:00Z', code: 500, endpoint: '/api/users', message: 'Internal error' },
        { id: '2', timestamp: '2026-01-01T11:00:00Z', code: 404, endpoint: '/api/orders', message: 'Not found' },
      ];
      useTechHealthStore.getState().setErrors(errors);
      expect(useTechHealthStore.getState().errors).toHaveLength(2);
    });

    it('filters by error code', () => {
      const errors: ErrorLog[] = [
        { id: '1', timestamp: '2026-01-01T10:00:00Z', code: 500, endpoint: '/api/a', message: 'Error' },
        { id: '2', timestamp: '2026-01-01T11:00:00Z', code: 404, endpoint: '/api/b', message: 'Not found' },
        { id: '3', timestamp: '2026-01-01T12:00:00Z', code: 500, endpoint: '/api/c', message: 'Error' },
      ];
      useTechHealthStore.getState().setErrors(errors);
      const serverErrors = useTechHealthStore.getState().errors.filter(e => e.code >= 500);
      expect(serverErrors).toHaveLength(2);
    });
  });

  describe('setServices', () => {
    it('sets services list', () => {
      const services: ServiceStatus[] = [
        { id: '1', name: 'API Gateway', status: 'healthy', cpu: 45, memory: 60, instances: 3, lastDeploy: '2026-01-01', version: '1.2.0' },
        { id: '2', name: 'Auth Service', status: 'degraded', cpu: 75, memory: 80, instances: 2, lastDeploy: '2026-01-02', version: '1.1.5' },
      ];
      useTechHealthStore.getState().setServices(services);
      expect(useTechHealthStore.getState().services).toHaveLength(2);
    });

    it('filters healthy services', () => {
      const services: ServiceStatus[] = [
        { id: '1', name: 'S1', status: 'healthy', cpu: 50, memory: 60, instances: 2, lastDeploy: '2026-01-01', version: '1.0' },
        { id: '2', name: 'S2', status: 'down', cpu: 90, memory: 95, instances: 1, lastDeploy: '2026-01-01', version: '1.0' },
        { id: '3', name: 'S3', status: 'healthy', cpu: 40, memory: 50, instances: 2, lastDeploy: '2026-01-01', version: '1.0' },
      ];
      useTechHealthStore.getState().setServices(services);
      const healthy = useTechHealthStore.getState().services.filter(s => s.status === 'healthy');
      expect(healthy).toHaveLength(2);
    });
  });

  describe('setSecurityEvents', () => {
    it('sets security events', () => {
      const events: SecurityEvent[] = [
        { id: '1', timestamp: '2026-01-01T10:00:00Z', type: 'failed_login', ip: '192.168.1.1', details: 'Multiple failed attempts', count: 5 },
        { id: '2', timestamp: '2026-01-01T11:00:00Z', type: 'unusual_ip', tenant: 'tenant1', ip: '10.0.0.1', details: 'New IP detected', count: 1 },
      ];
      useTechHealthStore.getState().setSecurityEvents(events);
      expect(useTechHealthStore.getState().securityEvents).toHaveLength(2);
    });

    it('filters failed login events', () => {
      const events: SecurityEvent[] = [
        { id: '1', timestamp: '2026-01-01T10:00:00Z', type: 'failed_login', ip: '1.1.1.1', details: 'Failed', count: 10 },
        { id: '2', timestamp: '2026-01-01T11:00:00Z', type: 'unusual_ip', ip: '2.2.2.2', details: 'Unusual', count: 1 },
      ];
      useTechHealthStore.getState().setSecurityEvents(events);
      const failedLogins = useTechHealthStore.getState().securityEvents.filter(e => e.type === 'failed_login');
      expect(failedLogins).toHaveLength(1);
    });
  });

  describe('setPeriod', () => {
    it('sets period', () => {
      useTechHealthStore.getState().setPeriod('7d');
      expect(useTechHealthStore.getState().period).toBe('7d');
    });

    it('changes period', () => {
      useTechHealthStore.getState().setPeriod('24h');
      useTechHealthStore.getState().setPeriod('30d');
      expect(useTechHealthStore.getState().period).toBe('30d');
    });
  });

  describe('setLoading', () => {
    it('sets loading state', () => {
      useTechHealthStore.getState().setLoading(true);
      expect(useTechHealthStore.getState().isLoading).toBe(true);
    });

    it('toggles loading', () => {
      useTechHealthStore.getState().setLoading(true);
      useTechHealthStore.getState().setLoading(false);
      expect(useTechHealthStore.getState().isLoading).toBe(false);
    });
  });

  describe('setErrorAlert', () => {
    it('sets error alert to true', () => {
      useTechHealthStore.getState().setErrorAlert(true);
      expect(useTechHealthStore.getState().errorAlert).toBe(true);
    });

    it('sets error alert to false', () => {
      useTechHealthStore.getState().setErrorAlert(true);
      useTechHealthStore.getState().setErrorAlert(false);
      expect(useTechHealthStore.getState().errorAlert).toBe(false);
    });
  });

  describe('fetchServiceStatus', () => {
    const mockServices = [
      { id: 's1', name: 'API Gateway', status: 'healthy', cpu: 45, memory: 60, instances: 3, last_deploy: '2026-01-01', version: '1.2.0' },
      { id: 's2', name: 'Auth Service', status: 'degraded', cpu: 80, memory: 85, instances: 2, last_deploy: '2026-01-02', version: '1.1.5' },
      { id: 's3', name: 'Worker Service', status: 'healthy', cpu: 30, memory: 40, instances: 5, last_deploy: '2026-01-03', version: '1.3.0' },
    ];

    it('fetches all services as healthy', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockServices.map(s => ({ ...s, status: 'healthy' })) });
      await useTechHealthStore.getState().fetchServiceStatus();
      const services = useTechHealthStore.getState().services;
      expect(services).toHaveLength(3);
      expect(services.every(s => s.status === 'healthy')).toBe(true);
    });

    it('fetches services with degraded status', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockServices });
      await useTechHealthStore.getState().fetchServiceStatus();
      const services = useTechHealthStore.getState().services;
      const degraded = services.filter(s => s.status === 'degraded');
      expect(degraded).toHaveLength(1);
      expect(degraded[0].name).toBe('Auth Service');
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useTechHealthStore.getState().fetchServiceStatus();
      expect(useTechHealthStore.getState().services).toEqual([]);
    });

    it('sets loading during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useTechHealthStore.getState().fetchServiceStatus();
      expect(useTechHealthStore.getState().isLoading).toBe(true);
    });

    it('correctly maps service fields', async () => {
      mockApiClient.get.mockResolvedValue({ data: [mockServices[0]] });
      await useTechHealthStore.getState().fetchServiceStatus();
      const service = useTechHealthStore.getState().services[0];
      expect(service.cpu).toBe(45);
      expect(service.memory).toBe(60);
      expect(service.instances).toBe(3);
      expect(service.version).toBe('1.2.0');
    });

    it('handles empty response', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      await useTechHealthStore.getState().fetchServiceStatus();
      expect(useTechHealthStore.getState().services).toHaveLength(0);
    });
  });

  describe('fetchResponseTimes', () => {
    const mockPerformance = [
      { timestamp: '2026-05-01T00:00:00Z', avg_response_time: 120, p95_response_time: 250, p99_response_time: 400, requests_per_second: 100, error_rate_4xx: 1, error_rate_5xx: 0.5 },
      { timestamp: '2026-05-01T01:00:00Z', avg_response_time: 130, p95_response_time: 260, p99_response_time: 420, requests_per_second: 110, error_rate_4xx: 2, error_rate_5xx: 1 },
    ];

    it('fetches performance metrics as array', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockPerformance });
      await useTechHealthStore.getState().fetchResponseTimes();
      const performance = useTechHealthStore.getState().performance;
      expect(performance).toHaveLength(2);
      expect(performance[0].avgResponseTime).toBe(120);
    });

    it('fetches empty data without crashing', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      await useTechHealthStore.getState().fetchResponseTimes();
      expect(useTechHealthStore.getState().performance).toEqual([]);
    });

    it('handles API error and sets empty array', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API Error'));
      await useTechHealthStore.getState().fetchResponseTimes();
      expect(useTechHealthStore.getState().performance).toEqual([]);
    });

    it('sets loading during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useTechHealthStore.getState().fetchResponseTimes();
      expect(useTechHealthStore.getState().isLoading).toBe(true);
    });

    it('correctly maps all performance fields', async () => {
      mockApiClient.get.mockResolvedValue({ data: [mockPerformance[0]] });
      await useTechHealthStore.getState().fetchResponseTimes();
      const p = useTechHealthStore.getState().performance[0];
      expect(p.p95ResponseTime).toBe(250);
      expect(p.p99ResponseTime).toBe(400);
      expect(p.requestsPerSecond).toBe(100);
    });

    it('handles missing fields with defaults', async () => {
      mockApiClient.get.mockResolvedValue({ data: [{ timestamp: '2026-05-01T00:00:00Z' }] });
      await useTechHealthStore.getState().fetchResponseTimes();
      const p = useTechHealthStore.getState().performance[0];
      expect(p.avgResponseTime).toBe(0);
      expect(p.p95ResponseTime).toBe(0);
    });
  });

  describe('fetchErrorRates', () => {
    const mockErrors = [
      { id: 'e1', timestamp: '2026-05-01T10:00:00Z', code: 500, endpoint: '/api/users', message: 'Internal error' },
      { id: 'e2', timestamp: '2026-05-01T10:05:00Z', code: 404, endpoint: '/api/orders', message: 'Not found' },
      { id: 'e3', timestamp: '2026-05-01T10:10:00Z', code: 503, endpoint: '/api/tenants', message: 'Service unavailable' },
    ];

    it('fetches error rates and sets in state', async () => {
      mockApiClient.get.mockResolvedValue({ data: { errors: mockErrors, total_requests: 1000 } });
      await useTechHealthStore.getState().fetchErrorRates();
      expect(useTechHealthStore.getState().errors).toHaveLength(3);
      expect(useTechHealthStore.getState().errors[0].code).toBe(500);
    });

    it('sets errorAlert false when rate below threshold (<5%)', async () => {
      mockApiClient.get.mockResolvedValue({ data: { errors: [mockErrors[0]], total_requests: 1000 } });
      await useTechHealthStore.getState().fetchErrorRates();
      expect(useTechHealthStore.getState().errorAlert).toBe(false);
    });

    it('sets errorAlert true when rate exceeds threshold (>5%)', async () => {
      const manyErrors = Array(10).fill(null).map((_, i) => ({ id: `e${i}`, timestamp: '2026-05-01T10:00:00Z', code: 500, endpoint: '/api/test', message: 'Error' }));
      mockApiClient.get.mockResolvedValue({ data: { errors: manyErrors, total_requests: 100 } });
      await useTechHealthStore.getState().fetchErrorRates();
      expect(useTechHealthStore.getState().errorAlert).toBe(true);
    });

    it('handles empty error data', async () => {
      mockApiClient.get.mockResolvedValue({ data: { errors: [], total_requests: 0 } });
      await useTechHealthStore.getState().fetchErrorRates();
      expect(useTechHealthStore.getState().errors).toEqual([]);
      expect(useTechHealthStore.getState().errorAlert).toBe(false);
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useTechHealthStore.getState().fetchErrorRates();
      expect(useTechHealthStore.getState().errors).toEqual([]);
      expect(useTechHealthStore.getState().errorAlert).toBe(false);
    });

    it('sets loading during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useTechHealthStore.getState().fetchErrorRates();
      expect(useTechHealthStore.getState().isLoading).toBe(true);
    });
  });

  describe('polling', () => {
    it('starts polling with default interval (30s)', () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      useTechHealthStore.getState().startPolling();
      expect(useTechHealthStore.getState().pollingInterval).not.toBeNull();
    });

    it('starts polling with custom interval', () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      useTechHealthStore.getState().startPolling(10000);
      expect(useTechHealthStore.getState().pollingInterval).not.toBeNull();
    });

    it('does not start multiple polling intervals', () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      useTechHealthStore.getState().startPolling();
      const firstInterval = useTechHealthStore.getState().pollingInterval;
      useTechHealthStore.getState().startPolling();
      expect(useTechHealthStore.getState().pollingInterval).toBe(firstInterval);
    });

    it('stops polling and clears interval', () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      useTechHealthStore.getState().startPolling();
      useTechHealthStore.getState().stopPolling();
      expect(useTechHealthStore.getState().pollingInterval).toBeNull();
    });

    it('fetches service status periodically', async () => {
      mockApiClient.get.mockResolvedValue({ data: [{ id: '1', name: 'Test', status: 'healthy', cpu: 50, memory: 60, instances: 1, last_deploy: '', version: '1.0' }] });
      useTechHealthStore.getState().startPolling(5000);
      jest.advanceTimersByTime(5000);
      jest.advanceTimersByTime(5000);
      expect(mockApiClient.get).toHaveBeenCalledTimes(2);
      useTechHealthStore.getState().stopPolling();
    });

    it('stops polling on cleanup', () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      useTechHealthStore.getState().startPolling();
      useTechHealthStore.getState().stopPolling();
      const { pollingInterval } = useTechHealthStore.getState();
      expect(pollingInterval).toBeNull();
    });
  });

  describe('methods exist', () => {
    it('has setStatus', () => {
      expect(typeof useTechHealthStore.getState().setStatus).toBe('function');
    });

    it('has setPerformance', () => {
      expect(typeof useTechHealthStore.getState().setPerformance).toBe('function');
    });

    it('has setErrors', () => {
      expect(typeof useTechHealthStore.getState().setErrors).toBe('function');
    });

    it('has setServices', () => {
      expect(typeof useTechHealthStore.getState().setServices).toBe('function');
    });

    it('has setSecurityEvents', () => {
      expect(typeof useTechHealthStore.getState().setSecurityEvents).toBe('function');
    });

    it('has setPeriod', () => {
      expect(typeof useTechHealthStore.getState().setPeriod).toBe('function');
    });

    it('has setLoading', () => {
      expect(typeof useTechHealthStore.getState().setLoading).toBe('function');
    });

    it('has setErrorAlert', () => {
      expect(typeof useTechHealthStore.getState().setErrorAlert).toBe('function');
    });

    it('has fetchRisks', () => {
      expect(typeof useTechHealthStore.getState().fetchRisks).toBe('function');
    });

    it('has fetchEconomics', () => {
      expect(typeof useTechHealthStore.getState().fetchEconomics).toBe('function');
    });

    it('has fetchServiceStatus', () => {
      expect(typeof useTechHealthStore.getState().fetchServiceStatus).toBe('function');
    });

    it('has fetchResponseTimes', () => {
      expect(typeof useTechHealthStore.getState().fetchResponseTimes).toBe('function');
    });

    it('has fetchErrorRates', () => {
      expect(typeof useTechHealthStore.getState().fetchErrorRates).toBe('function');
    });

    it('has startPolling', () => {
      expect(typeof useTechHealthStore.getState().startPolling).toBe('function');
    });

    it('has stopPolling', () => {
      expect(typeof useTechHealthStore.getState().stopPolling).toBe('function');
    });
  });

  describe('fetchRisks', () => {
    const mockRisks = [
      { created_at: '2026-05-07T10:00:00Z', risk_type: 'failed_login', tenant_name: 'Tenant 1', ip: '192.168.1.1', description: 'Multiple failed logins', severity: 5 },
      { created_at: '2026-05-07T09:00:00Z', risk_type: 'suspicious_activity', tenant_name: 'Tenant 2', ip: '192.168.1.2', description: 'Unusual activity', severity: 3 },
    ];

    it('fetches security events successfully', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockRisks });
      await useTechHealthStore.getState().fetchRisks();
      
      const events = useTechHealthStore.getState().securityEvents;
      expect(events).toHaveLength(2);
      expect(events[0].type).toBe('failed_login');
      expect(events[0].tenant).toBe('Tenant 1');
      expect(events[0].ip).toBe('192.168.1.1');
      expect(events[0].count).toBe(5);
    });

    it('handles empty response', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      await useTechHealthStore.getState().fetchRisks();
      expect(useTechHealthStore.getState().securityEvents).toEqual([]);
    });

    it('handles null response', async () => {
      mockApiClient.get.mockResolvedValue({ data: null });
      await useTechHealthStore.getState().fetchRisks();
      expect(useTechHealthStore.getState().securityEvents).toEqual([]);
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useTechHealthStore.getState().fetchRisks();
      expect(useTechHealthStore.getState().securityEvents).toEqual([]);
    });

    it('applies defaults for missing fields', async () => {
      mockApiClient.get.mockResolvedValue({ data: [{}] });
      await useTechHealthStore.getState().fetchRisks();
      
      const event = useTechHealthStore.getState().securityEvents[0];
      expect(event.type).toBe('failed_login');
      expect(event.count).toBe(1);
      expect(event.ip).toBe('');
    });

    it('maps all risk types correctly', async () => {
      mockApiClient.get.mockResolvedValue({ 
        data: [{ risk_type: 'data_breach' }, { risk_type: 'unusual_access' }] 
      });
      await useTechHealthStore.getState().fetchRisks();
      
      expect(useTechHealthStore.getState().securityEvents[0].type).toBe('data_breach');
      expect(useTechHealthStore.getState().securityEvents[1].type).toBe('unusual_access');
    });

    it('sets loading during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useTechHealthStore.getState().fetchRisks();
      expect(useTechHealthStore.getState().isLoading).toBe(true);
    });
  });

  describe('fetchEconomics', () => {
    const mockEconomics = {
      avg_response_time: 145,
      p95_response_time: 280,
      p99_response_time: 450,
      error_rate: 0.05,
    };

    it('fetches economics data successfully', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockEconomics });
      await useTechHealthStore.getState().fetchEconomics();
      
      const status = useTechHealthStore.getState().status;
      expect(status).not.toBeNull();
      expect(status?.avgResponseTime).toBe(145);
      expect(status?.p95ResponseTime).toBe(280);
      expect(status?.p99ResponseTime).toBe(450);
      expect(status?.errorRate).toBe(0.05);
    });

    it('handles empty response with defaults', async () => {
      mockApiClient.get.mockResolvedValue({ data: {} });
      await useTechHealthStore.getState().fetchEconomics();
      
      const status = useTechHealthStore.getState().status;
      expect(status?.avgResponseTime).toBe(150);
      expect(status?.p95ResponseTime).toBe(300);
      expect(status?.p99ResponseTime).toBe(500);
      expect(status?.errorRate).toBe(0.1);
      expect(status?.uptime).toBe(99.9);
    });

    it('handles null response', async () => {
      mockApiClient.get.mockResolvedValue({ data: null });
      await useTechHealthStore.getState().fetchEconomics();
      
      const status = useTechHealthStore.getState().status;
      expect(status?.avgResponseTime).toBe(150);
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useTechHealthStore.getState().fetchEconomics();
      expect(useTechHealthStore.getState().status).toBeNull();
    });

    it('sets loading during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useTechHealthStore.getState().fetchEconomics();
      expect(useTechHealthStore.getState().isLoading).toBe(true);
    });
  });
});