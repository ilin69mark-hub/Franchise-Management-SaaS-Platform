import { useAlertStore, Alert, AlertSettings } from '@/store/alertStore';

describe('alertStore', () => {
  beforeEach(() => {
    useAlertStore.setState({
      alerts: [],
      settings: {
        enabled: true,
        planThreshold: 70,
        conversionDropThreshold: 15,
        trafficDropThreshold: 30,
        stuckDealDays: 7,
        stuckDealAmount: 200000,
        inactiveDealerDays: 30,
        nonLiquidDays: 14,
        channels: {
          inApp: true,
          emailInstant: false,
          emailDigest: true,
          push: false,
        },
      },
      filter: 'all',
      isLoading: false,
    });
  });

  describe('initial state', () => {
    it('has initial alerts array', () => {
      expect(useAlertStore.getState().alerts).toEqual([]);
    });

    it('has default settings', () => {
      const settings = useAlertStore.getState().settings;
      expect(settings.enabled).toBe(true);
      expect(settings.planThreshold).toBe(70);
    });

    it('has default filter', () => {
      expect(useAlertStore.getState().filter).toBe('all');
    });

    it('has isLoading as false', () => {
      expect(useAlertStore.getState().isLoading).toBe(false);
    });
  });

  describe('setAlerts', () => {
    it('sets alerts list', () => {
      const alerts: Alert[] = [
        { id: '1', category: 'plan', priority: 'critical', title: 'Alert 1', description: 'Desc 1', createdAt: '2026-01-01', status: 'new' },
        { id: '2', category: 'funnel', priority: 'warning', title: 'Alert 2', description: 'Desc 2', createdAt: '2026-01-02', status: 'new' },
      ];
      useAlertStore.getState().setAlerts(alerts);
      expect(useAlertStore.getState().alerts).toHaveLength(2);
    });

    it('replaces existing alerts', () => {
      useAlertStore.getState().setAlerts([{ id: '1', category: 'plan', priority: 'critical', title: 'A', description: 'D', createdAt: '2026-01-01', status: 'new' }]);
      useAlertStore.getState().setAlerts([]);
      expect(useAlertStore.getState().alerts).toHaveLength(0);
    });
  });

  describe('addAlert', () => {
    it('adds new alert to list', () => {
      const alert: Alert = { id: '1', category: 'stock', priority: 'warning', title: 'New Alert', description: 'Desc', createdAt: '2026-01-01', status: 'new' };
      useAlertStore.getState().addAlert(alert);
      expect(useAlertStore.getState().alerts).toContainEqual(alert);
    });

    it('adds multiple alerts', () => {
      const alert1: Alert = { id: '1', category: 'plan', priority: 'critical', title: 'A1', description: 'D1', createdAt: '2026-01-01', status: 'new' };
      const alert2: Alert = { id: '2', category: 'plan', priority: 'warning', title: 'A2', description: 'D2', createdAt: '2026-01-02', status: 'new' };
      useAlertStore.getState().addAlert(alert1);
      useAlertStore.getState().addAlert(alert2);
      expect(useAlertStore.getState().alerts).toHaveLength(2);
    });
  });

  describe('markAsRead', () => {
    it('marks single alert as read', () => {
      const alert: Alert = { id: '1', category: 'plan', priority: 'critical', title: 'Alert', description: 'Desc', createdAt: '2026-01-01', status: 'new' };
      useAlertStore.getState().setAlerts([alert]);
      useAlertStore.getState().markAsRead('1');
      expect(useAlertStore.getState().alerts[0].status).toBe('read');
    });

    it('does not affect other alerts', () => {
      const alerts: Alert[] = [
        { id: '1', category: 'plan', priority: 'critical', title: 'A1', description: 'D1', createdAt: '2026-01-01', status: 'new' },
        { id: '2', category: 'plan', priority: 'warning', title: 'A2', description: 'D2', createdAt: '2026-01-02', status: 'new' },
      ];
      useAlertStore.getState().setAlerts(alerts);
      useAlertStore.getState().markAsRead('1');
      expect(useAlertStore.getState().alerts[1].status).toBe('new');
    });
  });

  describe('markAllAsRead', () => {
    it('marks all alerts as read', () => {
      const alerts: Alert[] = [
        { id: '1', category: 'plan', priority: 'critical', title: 'A1', description: 'D1', createdAt: '2026-01-01', status: 'new' },
        { id: '2', category: 'funnel', priority: 'warning', title: 'A2', description: 'D2', createdAt: '2026-01-02', status: 'new' },
      ];
      useAlertStore.getState().setAlerts(alerts);
      useAlertStore.getState().markAllAsRead();
      expect(useAlertStore.getState().alerts.every(a => a.status === 'read')).toBe(true);
    });
  });

  describe('setFilter', () => {
    it('sets filter to category', () => {
      useAlertStore.getState().setFilter('plan');
      expect(useAlertStore.getState().filter).toBe('plan');
    });

    it('sets filter to critical', () => {
      useAlertStore.getState().setFilter('critical');
      expect(useAlertStore.getState().filter).toBe('critical');
    });

    it('sets filter to all', () => {
      useAlertStore.getState().setFilter('plan');
      useAlertStore.getState().setFilter('all');
      expect(useAlertStore.getState().filter).toBe('all');
    });
  });

  describe('setSettings', () => {
    it('updates plan threshold', () => {
      useAlertStore.getState().setSettings({ planThreshold: 80 });
      expect(useAlertStore.getState().settings.planThreshold).toBe(80);
    });

    it('updates multiple settings', () => {
      useAlertStore.getState().setSettings({ planThreshold: 85, conversionDropThreshold: 20 });
      const settings = useAlertStore.getState().settings;
      expect(settings.planThreshold).toBe(85);
      expect(settings.conversionDropThreshold).toBe(20);
    });

    it('updates channels', () => {
      useAlertStore.getState().setSettings({ channels: { inApp: false, emailInstant: true, emailDigest: false, push: true } });
      const channels = useAlertStore.getState().settings.channels;
      expect(channels.inApp).toBe(false);
      expect(channels.emailInstant).toBe(true);
      expect(channels.push).toBe(true);
    });
  });

  describe('getUnreadCount', () => {
    it('returns zero for empty alerts', () => {
      expect(useAlertStore.getState().getUnreadCount()).toBe(0);
    });

    it('counts new status alerts', () => {
      const alerts: Alert[] = [
        { id: '1', category: 'plan', priority: 'critical', title: 'A', description: 'D', createdAt: '2026-01-01', status: 'new' },
        { id: '2', category: 'funnel', priority: 'warning', title: 'A', description: 'D', createdAt: '2026-01-02', status: 'read' },
        { id: '3', category: 'stock', priority: 'info', title: 'A', description: 'D', createdAt: '2026-01-03', status: 'new' },
      ];
      useAlertStore.getState().setAlerts(alerts);
      expect(useAlertStore.getState().getUnreadCount()).toBe(2);
    });

    it('does not count read or acknowledged', () => {
      const alerts: Alert[] = [
        { id: '1', category: 'plan', priority: 'critical', title: 'A', description: 'D', createdAt: '2026-01-01', status: 'acknowledged' },
        { id: '2', category: 'funnel', priority: 'warning', title: 'A', description: 'D', createdAt: '2026-01-02', status: 'read' },
      ];
      useAlertStore.getState().setAlerts(alerts);
      expect(useAlertStore.getState().getUnreadCount()).toBe(0);
    });
  });

  describe('methods exist', () => {
    it('has setAlerts', () => {
      expect(typeof useAlertStore.getState().setAlerts).toBe('function');
    });

    it('has addAlert', () => {
      expect(typeof useAlertStore.getState().addAlert).toBe('function');
    });

    it('has markAsRead', () => {
      expect(typeof useAlertStore.getState().markAsRead).toBe('function');
    });

    it('has markAllAsRead', () => {
      expect(typeof useAlertStore.getState().markAllAsRead).toBe('function');
    });

    it('has setFilter', () => {
      expect(typeof useAlertStore.getState().setFilter).toBe('function');
    });

    it('has setSettings', () => {
      expect(typeof useAlertStore.getState().setSettings).toBe('function');
    });

    it('has getUnreadCount', () => {
      expect(typeof useAlertStore.getState().getUnreadCount).toBe('function');
    });
  });
});