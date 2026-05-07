import { useFranchiserAlertStore, FranchiserAlert, FranchiserAlertSettings } from '@/store/franchiserAlertStore';

describe('franchiserAlertStore', () => {
  beforeEach(() => {
    useFranchiserAlertStore.setState({
      alerts: [
        {
          id: '1',
          type: 'plan' as const,
          severity: 'critical' as const,
          title: 'Test Alert 1',
          description: 'Description 1',
          metric: 85,
          createdAt: '2026-01-01T00:00:00Z',
          status: 'new' as const,
        },
        {
          id: '2',
          type: 'dealer' as const,
          severity: 'warning' as const,
          title: 'Test Alert 2',
          description: 'Description 2',
          createdAt: '2026-01-02T00:00:00Z',
          status: 'acknowledged' as const,
        },
      ],
      settings: {
        enabled: true,
        channel: { inApp: true, email: false, push: false },
        thresholds: { criticalForecastPercent: 90, dealerChurnPercent: 5, managerKpiPercent: 70 },
      },
      isOpen: false,
      filter: 'all',
    });
  });

  describe('initial state', () => {
    it('has alerts from mock data', () => {
      expect(useFranchiserAlertStore.getState().alerts).toHaveLength(2);
    });

    it('has default settings', () => {
      const settings = useFranchiserAlertStore.getState().settings;
      expect(settings.enabled).toBe(true);
      expect(settings.channel.inApp).toBe(true);
    });

    it('has isOpen as false', () => {
      expect(useFranchiserAlertStore.getState().isOpen).toBe(false);
    });

    it('has filter as all', () => {
      expect(useFranchiserAlertStore.getState().filter).toBe('all');
    });
  });

  describe('setAlerts', () => {
    it('replaces all alerts', () => {
      const newAlerts: FranchiserAlert[] = [
        { id: '3', type: 'manager' as const, severity: 'warning' as const, title: 'New Alert', description: 'Desc', createdAt: '2026-01-03', status: 'new' as const },
      ];
      
      useFranchiserAlertStore.getState().setAlerts(newAlerts);
      expect(useFranchiserAlertStore.getState().alerts).toHaveLength(1);
      expect(useFranchiserAlertStore.getState().alerts[0].title).toBe('New Alert');
    });
  });

  describe('addAlert', () => {
    it('adds alert to beginning of list', () => {
      const newAlert: FranchiserAlert = {
        id: 'new',
        type: 'escalation' as const,
        severity: 'critical' as const,
        title: 'New Alert',
        description: 'New Desc',
        createdAt: '2026-01-10',
        status: 'new' as const,
      };
      
      useFranchiserAlertStore.getState().addAlert(newAlert);
      const alerts = useFranchiserAlertStore.getState().alerts;
      expect(alerts[0].id).toBe('new');
      expect(alerts).toHaveLength(3);
    });
  });

  describe('markAsRead', () => {
    it('changes alert status to acknowledged', () => {
      useFranchiserAlertStore.getState().markAsRead('1');
      const alert = useFranchiserAlertStore.getState().alerts.find(a => a.id === '1');
      expect(alert?.status).toBe('acknowledged');
    });

    it('does not affect other alerts', () => {
      useFranchiserAlertStore.getState().markAsRead('1');
      const alert2 = useFranchiserAlertStore.getState().alerts.find(a => a.id === '2');
      expect(alert2?.status).toBe('acknowledged');
    });
  });

  describe('markAllAsRead', () => {
    it('marks all alerts as acknowledged', () => {
      useFranchiserAlertStore.getState().markAllAsRead();
      const alerts = useFranchiserAlertStore.getState().alerts;
      alerts.forEach(a => expect(a.status).toBe('acknowledged'));
    });
  });

  describe('assignAlert', () => {
    it('assigns alert to manager and marks as acknowledged', () => {
      useFranchiserAlertStore.getState().assignAlert('1', 'manager-123');
      const alert = useFranchiserAlertStore.getState().alerts.find(a => a.id === '1');
      expect(alert?.assignedTo).toBe('manager-123');
      expect(alert?.status).toBe('acknowledged');
    });
  });

  describe('resolveAlert', () => {
    it('changes alert status to resolved with timestamp', () => {
      const beforeResolve = new Date().toISOString();
      useFranchiserAlertStore.getState().resolveAlert('1');
      const alert = useFranchiserAlertStore.getState().alerts.find(a => a.id === '1');
      expect(alert?.status).toBe('resolved');
      expect(alert?.resolvedAt).toBeDefined();
    });
  });

  describe('setOpen', () => {
    it('opens alert panel', () => {
      useFranchiserAlertStore.getState().setOpen(true);
      expect(useFranchiserAlertStore.getState().isOpen).toBe(true);
    });

    it('closes alert panel', () => {
      useFranchiserAlertStore.setState({ isOpen: true });
      useFranchiserAlertStore.getState().setOpen(false);
      expect(useFranchiserAlertStore.getState().isOpen).toBe(false);
    });
  });

  describe('setFilter', () => {
    it('sets filter to critical', () => {
      useFranchiserAlertStore.getState().setFilter('critical');
      expect(useFranchiserAlertStore.getState().filter).toBe('critical');
    });

    it('sets filter to specific type', () => {
      useFranchiserAlertStore.getState().setFilter('dealer');
      expect(useFranchiserAlertStore.getState().filter).toBe('dealer');
    });
  });

  describe('setSettings', () => {
    it('updates settings partially', () => {
      useFranchiserAlertStore.getState().setSettings({ enabled: false });
      expect(useFranchiserAlertStore.getState().settings.enabled).toBe(false);
      expect(useFranchiserAlertStore.getState().settings.channel.inApp).toBe(true);
    });

    it('updates channel settings', () => {
      useFranchiserAlertStore.getState().setSettings({
        channel: { inApp: false, email: true, push: true },
      });
      const settings = useFranchiserAlertStore.getState().settings;
      expect(settings.channel.inApp).toBe(false);
      expect(settings.channel.email).toBe(true);
      expect(settings.channel.push).toBe(true);
    });
  });

  describe('getUnreadCount', () => {
    it('returns count of new alerts', () => {
      const count = useFranchiserAlertStore.getState().getUnreadCount();
      expect(count).toBe(1);
    });

    it('returns 0 when all alerts acknowledged', () => {
      useFranchiserAlertStore.getState().markAllAsRead();
      const count = useFranchiserAlertStore.getState().getUnreadCount();
      expect(count).toBe(0);
    });
  });

  describe('getCriticalCount', () => {
    it('returns count of unresolved critical alerts', () => {
      const count = useFranchiserAlertStore.getState().getCriticalCount();
      expect(count).toBe(1);
    });

    it('excludes resolved alerts', () => {
      useFranchiserAlertStore.getState().resolveAlert('1');
      const count = useFranchiserAlertStore.getState().getCriticalCount();
      expect(count).toBe(0);
    });
  });

  describe('logout handler (clear state)', () => {
    it('clears alerts on logout simulation', () => {
      useFranchiserAlertStore.getState().setAlerts([]);
      const alerts = useFranchiserAlertStore.getState().alerts;
      expect(alerts).toHaveLength(0);
    });

    it('resets filter to all', () => {
      useFranchiserAlertStore.setState({ filter: 'critical' });
      useFranchiserAlertStore.getState().setFilter('all');
      expect(useFranchiserAlertStore.getState().filter).toBe('all');
    });
  });
});