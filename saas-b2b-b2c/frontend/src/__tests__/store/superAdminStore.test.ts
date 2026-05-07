import { useSuperAdminStore, Tenant, AuditLogEntry } from '@/store/superAdminStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  patch: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('superAdminStore', () => {
  beforeEach(() => {
    useSuperAdminStore.setState({
      activeSection: 'metrics',
      stats: { mrr: 0, arr: 0, churnRate: 0, overduePayments: 0, arpu: 0 },
      tenants: [],
      auditLog: [],
      isLoading: false,
      alertCount: 0,
      lastUpdated: null,
      pagination: { page: 1, limit: 20, total: 0 },
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('has metrics section by default', () => {
      expect(useSuperAdminStore.getState().activeSection).toBe('metrics');
    });

    it('has zero stats by default', () => {
      const { stats } = useSuperAdminStore.getState();
      expect(stats.mrr).toBe(0);
      expect(stats.arr).toBe(0);
    });

    it('has empty tenants array', () => {
      expect(useSuperAdminStore.getState().tenants).toEqual([]);
    });

    it('has empty auditLog array', () => {
      expect(useSuperAdminStore.getState().auditLog).toEqual([]);
    });

    it('is not loading by default', () => {
      expect(useSuperAdminStore.getState().isLoading).toBe(false);
    });

    it('has default pagination', () => {
      const { pagination } = useSuperAdminStore.getState();
      expect(pagination.page).toBe(1);
      expect(pagination.limit).toBe(20);
      expect(pagination.total).toBe(0);
    });
  });

  describe('setActiveSection', () => {
    it('updates active section to tenants', () => {
      useSuperAdminStore.getState().setActiveSection('tenants');
      expect(useSuperAdminStore.getState().activeSection).toBe('tenants');
    });

    it('updates active section to audit', () => {
      useSuperAdminStore.getState().setActiveSection('audit');
      expect(useSuperAdminStore.getState().activeSection).toBe('audit');
    });
  });

  describe('setStats', () => {
    it('updates stats with full data', () => {
      const mockStats = { mrr: 2500000, arr: 30000000, churnRate: 5.2, overduePayments: 150000, arpu: 15000 };
      useSuperAdminStore.getState().setStats(mockStats);
      expect(useSuperAdminStore.getState().stats).toEqual(mockStats);
    });
  });

  describe('setTenants', () => {
    it('sets tenants list', () => {
      const tenants: Tenant[] = [
        { id: 't1', name: 'Tenant 1', slug: 'tenant1', status: 'active', plan: 'Pro', mrr: 10000, createdAt: '2026-01-01' },
        { id: 't2', name: 'Tenant 2', slug: 'tenant2', status: 'active', plan: 'Start', mrr: 5000, createdAt: '2026-02-01' },
      ];
      useSuperAdminStore.getState().setTenants(tenants);
      expect(useSuperAdminStore.getState().tenants).toHaveLength(2);
    });

    it('replaces existing tenants', () => {
      useSuperAdminStore.getState().setTenants([{ id: 't1', name: 'Old', slug: 'old', status: 'active', plan: 'Start', mrr: 1000, createdAt: '' }]);
      useSuperAdminStore.getState().setTenants([]);
      expect(useSuperAdminStore.getState().tenants).toHaveLength(0);
    });
  });

  describe('setAuditLog', () => {
    it('sets audit log entries', () => {
      const log: AuditLogEntry[] = [
        { id: 'a1', tenantId: 't1', tenantName: 'Tenant 1', action: 'login', userId: 'u1', userName: 'Admin', timestamp: '2026-05-01T10:00:00Z' },
      ];
      useSuperAdminStore.getState().setAuditLog(log);
      expect(useSuperAdminStore.getState().auditLog).toHaveLength(1);
    });
  });

  describe('setPagination', () => {
    it('updates pagination parameters', () => {
      useSuperAdminStore.getState().setPagination(2, 50, 150);
      const { pagination } = useSuperAdminStore.getState();
      expect(pagination.page).toBe(2);
      expect(pagination.limit).toBe(50);
      expect(pagination.total).toBe(150);
    });
  });

  describe('setLoading', () => {
    it('sets loading to true', () => {
      useSuperAdminStore.getState().setLoading(true);
      expect(useSuperAdminStore.getState().isLoading).toBe(true);
    });
  });

  describe('fetchStats', () => {
    it('sets loading = true during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useSuperAdminStore.getState().fetchStats();
      expect(useSuperAdminStore.getState().isLoading).toBe(true);
    });

    it('fetches and sets stats on success', async () => {
      mockApiClient.get.mockResolvedValue({ data: { mrr: 5000000, arr: 60000000, churnRate: 3.5, overduePayments: 80000, arpu: 20000 } });
      await useSuperAdminStore.getState().fetchStats();
      const { stats } = useSuperAdminStore.getState();
      expect(stats.mrr).toBe(5000000);
      expect(stats.arr).toBe(60000000);
      expect(useSuperAdminStore.getState().isLoading).toBe(false);
    });

    it('sets lastUpdated on success', async () => {
      mockApiClient.get.mockResolvedValue({ data: { mrr: 1000000 } });
      await useSuperAdminStore.getState().fetchStats();
      expect(useSuperAdminStore.getState().lastUpdated).not.toBeNull();
    });

    it('uses fallback data on API error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useSuperAdminStore.getState().fetchStats();
      expect(useSuperAdminStore.getState().stats.mrr).toBe(2500000);
    });

    it('sets loading = false after error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API Error'));
      await useSuperAdminStore.getState().fetchStats();
      expect(useSuperAdminStore.getState().isLoading).toBe(false);
    });
  });

  describe('fetchAllTenants', () => {
    const mockTenants = [
      { id: 't1', name: 'Client A', slug: 'client-a', status: 'active', plan: 'Pro', mrr: 15000, createdAt: '2026-01-01' },
      { id: 't2', name: 'Client B', slug: 'client-b', status: 'active', plan: 'Start', mrr: 5000, createdAt: '2026-02-01' },
    ];

    it('fetches tenants with default pagination', async () => {
      mockApiClient.get.mockResolvedValue({ data: { tenants: mockTenants, total: 2 } });
      await useSuperAdminStore.getState().fetchAllTenants();
      expect(useSuperAdminStore.getState().tenants).toHaveLength(2);
      expect(useSuperAdminStore.getState().pagination.total).toBe(2);
    });

    it('fetches tenants with custom page/limit', async () => {
      mockApiClient.get.mockResolvedValue({ data: { tenants: mockTenants, total: 100 } });
      await useSuperAdminStore.getState().fetchAllTenants(2, 50);
      expect(mockApiClient.get).toHaveBeenCalledWith('/admin/tenants', { params: { page: 2, limit: 50 } });
    });

    it('handles empty tenants list', async () => {
      mockApiClient.get.mockResolvedValue({ data: { tenants: [], total: 0 } });
      await useSuperAdminStore.getState().fetchAllTenants();
      expect(useSuperAdminStore.getState().tenants).toEqual([]);
    });

    it('sets loading = true during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useSuperAdminStore.getState().fetchAllTenants();
      expect(useSuperAdminStore.getState().isLoading).toBe(true);
    });

    it('updates pagination correctly', async () => {
      mockApiClient.get.mockResolvedValue({ data: { tenants: mockTenants, total: 500 } });
      await useSuperAdminStore.getState().fetchAllTenants(3, 100);
      const { pagination } = useSuperAdminStore.getState();
      expect(pagination.page).toBe(3);
      expect(pagination.limit).toBe(100);
      expect(pagination.total).toBe(500);
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useSuperAdminStore.getState().fetchAllTenants();
      expect(useSuperAdminStore.getState().tenants).toEqual([]);
    });
  });

  describe('suspendTenant', () => {
    const tenants: Tenant[] = [
      { id: 't1', name: 'Active Client', slug: 'active', status: 'active', plan: 'Pro', mrr: 10000, createdAt: '2026-01-01' },
      { id: 't2', name: 'Another Client', slug: 'another', status: 'active', plan: 'Start', mrr: 5000, createdAt: '2026-02-01' },
    ];

    it('suspends tenant and updates status in state', async () => {
      useSuperAdminStore.setState({ tenants });
      mockApiClient.patch.mockResolvedValue({});
      await useSuperAdminStore.getState().suspendTenant('t1');
      expect(useSuperAdminStore.getState().tenants[0].status).toBe('suspended');
      expect(useSuperAdminStore.getState().tenants[1].status).toBe('active');
    });

    it('throws error on API failure (403)', async () => {
      useSuperAdminStore.setState({ tenants });
      mockApiClient.patch.mockRejectedValue({ response: { status: 403 } });
      await expect(useSuperAdminStore.getState().suspendTenant('t1')).rejects.toEqual({ response: { status: 403 } });
    });

    it('does not modify state on API error', async () => {
      useSuperAdminStore.setState({ tenants: [...tenants] });
      mockApiClient.patch.mockRejectedValue(new Error('Forbidden'));
      try {
        await useSuperAdminStore.getState().suspendTenant('t1');
      } catch {}
      expect(useSuperAdminStore.getState().tenants[0].status).toBe('active');
    });
  });

  describe('deleteTenant', () => {
    const tenants: Tenant[] = [
      { id: 't1', name: 'To Delete', slug: 'delete', status: 'active', plan: 'Pro', mrr: 10000, createdAt: '2026-01-01' },
      { id: 't2', name: 'To Keep', slug: 'keep', status: 'active', plan: 'Start', mrr: 5000, createdAt: '2026-02-01' },
    ];

    it('removes tenant from state on success', async () => {
      useSuperAdminStore.setState({ tenants });
      mockApiClient.delete.mockResolvedValue({});
      await useSuperAdminStore.getState().deleteTenant('t1');
      expect(useSuperAdminStore.getState().tenants).toHaveLength(1);
      expect(useSuperAdminStore.getState().tenants[0].id).toBe('t2');
    });

    it('calls correct API endpoint', async () => {
      mockApiClient.delete.mockResolvedValue({});
      await useSuperAdminStore.getState().deleteTenant('t1');
      expect(mockApiClient.delete).toHaveBeenCalledWith('/admin/tenants/t1');
    });

    it('throws error on API failure', async () => {
      mockApiClient.delete.mockRejectedValue(new Error('Delete failed'));
      await expect(useSuperAdminStore.getState().deleteTenant('t1')).rejects.toEqual(new Error('Delete failed'));
    });

    it('does not modify state on API error', async () => {
      useSuperAdminStore.setState({ tenants: [...tenants] });
      mockApiClient.delete.mockRejectedValue(new Error('Error'));
      try {
        await useSuperAdminStore.getState().deleteTenant('t1');
      } catch {}
      expect(useSuperAdminStore.getState().tenants).toHaveLength(2);
    });
  });

  describe('fetchAuditLog', () => {
    const mockAuditLog: AuditLogEntry[] = [
      { id: 'a1', tenantId: 't1', tenantName: 'Client A', action: 'user_created', userId: 'u1', userName: 'Admin', timestamp: '2026-05-01T10:00:00Z' },
      { id: 'a2', tenantId: 't2', tenantName: 'Client B', action: 'plan_upgraded', userId: 'u2', userName: 'Manager', timestamp: '2026-05-02T14:30:00Z' },
    ];

    it('fetches audit log without filter', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockAuditLog });
      await useSuperAdminStore.getState().fetchAuditLog();
      expect(useSuperAdminStore.getState().auditLog).toHaveLength(2);
    });

    it('fetches audit log with tenant filter', async () => {
      mockApiClient.get.mockResolvedValue({ data: [mockAuditLog[0]] });
      await useSuperAdminStore.getState().fetchAuditLog('t1');
      expect(mockApiClient.get).toHaveBeenCalledWith('/admin/audit-log', { params: { tenant_id: 't1' } });
    });

    it('handles empty audit log', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });
      await useSuperAdminStore.getState().fetchAuditLog();
      expect(useSuperAdminStore.getState().auditLog).toEqual([]);
    });

    it('sets loading = true during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));
      useSuperAdminStore.getState().fetchAuditLog();
      expect(useSuperAdminStore.getState().isLoading).toBe(true);
    });

    it('handles API error gracefully', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));
      await useSuperAdminStore.getState().fetchAuditLog();
      expect(useSuperAdminStore.getState().auditLog).toEqual([]);
    });

    it('correctly parses audit log entry fields', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockAuditLog });
      await useSuperAdminStore.getState().fetchAuditLog();
      const entry = useSuperAdminStore.getState().auditLog[0];
      expect(entry.tenantId).toBe('t1');
      expect(entry.action).toBe('user_created');
      expect(entry.userName).toBe('Admin');
    });
  });

  describe('computed selectors', () => {
    it('calculates total MRR from tenants', () => {
      useSuperAdminStore.setState({
        tenants: [
          { id: 't1', name: 'T1', slug: 't1', status: 'active', plan: 'Pro', mrr: 15000, createdAt: '' },
          { id: 't2', name: 'T2', slug: 't2', status: 'active', plan: 'Start', mrr: 5000, createdAt: '' },
        ],
      });
      const totalMrr = useSuperAdminStore.getState().tenants.reduce((sum, t) => sum + t.mrr, 0);
      expect(totalMrr).toBe(20000);
    });

    it('counts active tenants', () => {
      useSuperAdminStore.setState({
        tenants: [
          { id: 't1', name: 'T1', slug: 't1', status: 'active', plan: 'Pro', mrr: 10000, createdAt: '' },
          { id: 't2', name: 'T2', slug: 't2', status: 'suspended', plan: 'Start', mrr: 5000, createdAt: '' },
        ],
      });
      const activeCount = useSuperAdminStore.getState().tenants.filter(t => t.status === 'active').length;
      expect(activeCount).toBe(1);
    });

    it('calculates average tenant MRR', () => {
      useSuperAdminStore.setState({
        tenants: [
          { id: 't1', name: 'T1', slug: 't1', status: 'active', plan: 'Pro', mrr: 10000, createdAt: '' },
          { id: 't2', name: 'T2', slug: 't2', status: 'active', plan: 'Start', mrr: 5000, createdAt: '' },
        ],
      });
      const avgMrr = useSuperAdminStore.getState().tenants.reduce((sum, t) => sum + t.mrr, 0) / 2;
      expect(avgMrr).toBe(7500);
    });
  });
});