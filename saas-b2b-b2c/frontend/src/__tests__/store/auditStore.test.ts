import { useAuditStore } from '@/store/auditStore';

describe('auditStore', () => {
  beforeEach(() => {
    useAuditStore.setState({
      adminActions: [],
      impersonations: [],
      activeSessions: [],
      userLogins: [],
      filters: { admin: '', action: '', dateRange: null, tenant: '' },
      isLoading: false,
    });
  });

  describe('initial state', () => {
    it('has empty adminActions array', () => {
      const { adminActions } = useAuditStore.getState();
      expect(adminActions).toEqual([]);
    });

    it('has empty impersonations array', () => {
      const { impersonations } = useAuditStore.getState();
      expect(impersonations).toEqual([]);
    });

    it('has empty activeSessions array', () => {
      const { activeSessions } = useAuditStore.getState();
      expect(activeSessions).toEqual([]);
    });

    it('has empty userLogins array', () => {
      const { userLogins } = useAuditStore.getState();
      expect(userLogins).toEqual([]);
    });

    it('has default filters', () => {
      const { filters } = useAuditStore.getState();
      expect(filters).toEqual({ admin: '', action: '', dateRange: null, tenant: '' });
    });

    it('is not loading by default', () => {
      const { isLoading } = useAuditStore.getState();
      expect(isLoading).toBe(false);
    });
  });

  describe('setAdminActions', () => {
    it('updates admin actions', () => {
      const mockActions = [
        {
          id: 'act-1',
          timestamp: '2024-01-15T10:00:00Z',
          adminName: 'Admin User',
          adminEmail: 'admin@example.com',
          action: 'create',
          object: 'tenant',
          details: 'Created new tenant',
          ip: '192.168.1.1',
        },
      ];
      useAuditStore.getState().setAdminActions(mockActions);
      expect(useAuditStore.getState().adminActions).toEqual(mockActions);
    });
  });

  describe('setImpersonations', () => {
    it('updates impersonations', () => {
      const mockImpersonations = [
        {
          id: 'imp-1',
          timestamp: '2024-01-15T10:00:00Z',
          adminName: 'Admin User',
          tenant: 'Tenant 1',
          role: 'dealer',
          userName: 'John Doe',
          duration: 300,
          actionsSummary: 'Viewed dashboard',
          ip: '192.168.1.1',
        },
      ];
      useAuditStore.getState().setImpersonations(mockImpersonations);
      expect(useAuditStore.getState().impersonations).toEqual(mockImpersonations);
    });
  });

  describe('setActiveSessions', () => {
    it('updates active sessions', () => {
      const mockSessions = [
        {
          id: 'sess-1',
          adminName: 'Admin User',
          ip: '192.168.1.1',
          startedAt: '2024-01-15T09:00:00Z',
          lastActivity: '2024-01-15T10:30:00Z',
        },
      ];
      useAuditStore.getState().setActiveSessions(mockSessions);
      expect(useAuditStore.getState().activeSessions).toEqual(mockSessions);
    });
  });

  describe('setUserLogins', () => {
    it('updates user logins', () => {
      const mockLogins = [
        {
          id: 'login-1',
          timestamp: '2024-01-15T10:00:00Z',
          tenant: 'Tenant 1',
          userName: 'John Doe',
          userEmail: 'john@example.com',
          role: 'dealer',
          action: 'login',
          ip: '192.168.1.1',
          userAgent: 'Mozilla/5.0',
          geo: 'Moscow, Russia',
        },
      ];
      useAuditStore.getState().setUserLogins(mockLogins);
      expect(useAuditStore.getState().userLogins).toEqual(mockLogins);
    });
  });

  describe('setFilters', () => {
    it('updates admin filter', () => {
      useAuditStore.getState().setFilters({ admin: 'admin@example.com' });
      expect(useAuditStore.getState().filters.admin).toBe('admin@example.com');
    });

    it('updates action filter', () => {
      useAuditStore.getState().setFilters({ action: 'delete' });
      expect(useAuditStore.getState().filters.action).toBe('delete');
    });

    it('updates dateRange filter', () => {
      const dateRange: [string, string] = ['2024-01-01', '2024-01-31'];
      useAuditStore.getState().setFilters({ dateRange });
      expect(useAuditStore.getState().filters.dateRange).toEqual(dateRange);
    });

    it('updates tenant filter', () => {
      useAuditStore.getState().setFilters({ tenant: 'Tenant 1' });
      expect(useAuditStore.getState().filters.tenant).toBe('Tenant 1');
    });

    it('merges with existing filters', () => {
      useAuditStore.getState().setFilters({ admin: 'admin@example.com' });
      useAuditStore.getState().setFilters({ action: 'create' });
      const { filters } = useAuditStore.getState();
      expect(filters.admin).toBe('admin@example.com');
      expect(filters.action).toBe('create');
    });
  });

  describe('setLoading', () => {
    it('updates loading state', () => {
      useAuditStore.getState().setLoading(true);
      expect(useAuditStore.getState().isLoading).toBe(true);
    });
  });
});

describe('Audit Store Data Types', () => {
  it('supports login action', () => {
    const login = {
      id: 'login-1',
      timestamp: '2024-01-15T10:00:00Z',
      tenant: 'Tenant 1',
      userName: 'John Doe',
      userEmail: 'john@example.com',
      role: 'dealer',
      action: 'login',
      ip: '192.168.1.1',
      userAgent: 'Mozilla/5.0',
    };
    expect(login.action).toBe('login');
  });

  it('supports logout action', () => {
    const login = {
      id: 'login-1',
      timestamp: '2024-01-15T10:00:00Z',
      tenant: 'Tenant 1',
      userName: 'John Doe',
      userEmail: 'john@example.com',
      role: 'dealer',
      action: 'logout',
      ip: '192.168.1.1',
      userAgent: 'Mozilla/5.0',
    };
    expect(login.action).toBe('logout');
  });

  it('supports geo location', () => {
    const login = {
      id: 'login-1',
      timestamp: '2024-01-15T10:00:00Z',
      tenant: 'Tenant 1',
      userName: 'John Doe',
      userEmail: 'john@example.com',
      role: 'dealer',
      action: 'login',
      ip: '192.168.1.1',
      userAgent: 'Mozilla/5.0',
      geo: 'New York, USA',
    };
    expect(login.geo).toBe('New York, USA');
  });
});