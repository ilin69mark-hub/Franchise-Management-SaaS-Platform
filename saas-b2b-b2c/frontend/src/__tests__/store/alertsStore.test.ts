import { useAlertStore } from '@/store/alertsStore';

describe('alertsStore', () => {
  beforeEach(() => {
    useAlertStore.setState({
      alerts: [],
      unreadCount: 0,
      settings: null,
      isLoading: false,
    });
  });

  describe('initial state', () => {
    it('has empty alerts array', () => {
      expect(useAlertStore.getState().alerts).toEqual([]);
    });

    it('has unreadCount as 0', () => {
      expect(useAlertStore.getState().unreadCount).toBe(0);
    });

    it('has settings as null', () => {
      expect(useAlertStore.getState().settings).toBeNull();
    });

    it('has isLoading as false', () => {
      expect(useAlertStore.getState().isLoading).toBe(false);
    });
  });

  describe('setAlerts', () => {
    it('sets alerts and calculates unread count', () => {
      const alerts = [
        { id: '1', type: 'billing' as const, title: 'Alert 1', message: 'Msg 1', timestamp: '2026-01-01', read: false },
        { id: '2', type: 'technical' as const, title: 'Alert 2', message: 'Msg 2', timestamp: '2026-01-02', read: true },
        { id: '3', type: 'security' as const, title: 'Alert 3', message: 'Msg 3', timestamp: '2026-01-03', read: false },
      ];
      
      useAlertStore.getState().setAlerts(alerts);
      
      expect(useAlertStore.getState().alerts).toHaveLength(3);
      expect(useAlertStore.getState().unreadCount).toBe(2);
    });

    it('sets empty alerts and unreadCount to 0', () => {
      useAlertStore.getState().setAlerts([]);
      expect(useAlertStore.getState().alerts).toHaveLength(0);
      expect(useAlertStore.getState().unreadCount).toBe(0);
    });
  });

  describe('setSettings', () => {
    it('sets alert settings', () => {
      const settings = {
        channels: ['in-app', 'email'] as const,
        thresholds: {
          errorRate: 5,
          responseTime: 200,
          uptime: 99.9,
          churnRate: 10,
        },
      };
      
      useAlertStore.getState().setSettings(settings);
      expect(useAlertStore.getState().settings).toEqual(settings);
    });
  });

  describe('setLoading', () => {
    it('sets loading state to true', () => {
      useAlertStore.getState().setLoading(true);
      expect(useAlertStore.getState().isLoading).toBe(true);
    });

    it('sets loading state to false', () => {
      useAlertStore.setState({ isLoading: true });
      useAlertStore.getState().setLoading(false);
      expect(useAlertStore.getState().isLoading).toBe(false);
    });
  });

  describe('loadAlerts async', () => {
    it('sets loading true during fetch', async () => {
      global.fetch = jest.fn().mockImplementation(() => 
        Promise.resolve({
          json: () => Promise.resolve([
            { id: '1', type: 'billing' as const, title: 'Alert', message: 'Msg', timestamp: '2026-01-01', read: false },
          ]),
        })
      ) as jest.Mock;

      const loadPromise = useAlertStore.getState().loadAlerts();
      
      expect(useAlertStore.getState().isLoading).toBe(true);
      
      await loadPromise;
      
      expect(useAlertStore.getState().isLoading).toBe(false);
    });

    it('loads alerts successfully', async () => {
      const mockAlerts = [
        { id: '1', type: 'billing' as const, title: 'Alert 1', message: 'Msg 1', timestamp: '2026-01-01', read: false },
      ];

      global.fetch = jest.fn().mockImplementation(() => 
        Promise.resolve({
          json: () => Promise.resolve(mockAlerts),
        })
      ) as jest.Mock;

      await useAlertStore.getState().loadAlerts();

      expect(useAlertStore.getState().alerts).toEqual(mockAlerts);
      expect(useAlertStore.getState().unreadCount).toBe(1);
    });

    it('handles empty alerts response', async () => {
      global.fetch = jest.fn().mockImplementation(() => 
        Promise.resolve({
          json: () => Promise.resolve([]),
        })
      ) as jest.Mock;

      await useAlertStore.getState().loadAlerts();

      expect(useAlertStore.getState().alerts).toEqual([]);
      expect(useAlertStore.getState().unreadCount).toBe(0);
    });
  });

  describe('markAsRead async', () => {
    it('marks alert as read in state', async () => {
      useAlertStore.setState({
        alerts: [
          { id: '1', type: 'billing' as const, title: 'Alert', message: 'Msg', timestamp: '2026-01-01', read: false },
          { id: '2', type: 'technical' as const, title: 'Alert 2', message: 'Msg 2', timestamp: '2026-01-02', read: false },
        ],
        unreadCount: 2,
      });

      global.fetch = jest.fn().mockImplementation(() => 
        Promise.resolve({})
      ) as jest.Mock;

      await useAlertStore.getState().markAsRead('1');

      const alerts = useAlertStore.getState().alerts;
      expect(alerts[0].read).toBe(true);
      expect(alerts[1].read).toBe(false);
      expect(useAlertStore.getState().unreadCount).toBe(1);
    });

    it('does not modify state when alert not found', async () => {
      useAlertStore.setState({
        alerts: [
          { id: '1', type: 'billing' as const, title: 'Alert', message: 'Msg', timestamp: '2026-01-01', read: false },
        ],
        unreadCount: 1,
      });

      global.fetch = jest.fn().mockImplementation(() => 
        Promise.resolve({})
      ) as jest.Mock;

      await useAlertStore.getState().markAsRead('nonexistent');

      expect(useAlertStore.getState().alerts[0].read).toBe(false);
    });
  });

  describe('updateSettings async', () => {
    it('calls API and reloads alerts', async () => {
      const fetchMock = jest.fn().mockImplementation(() => 
        Promise.resolve({ json: () => Promise.resolve([]) })
      ) as jest.Mock;
      global.fetch = fetchMock;

      await useAlertStore.getState().updateSettings({
        channels: ['telegram'],
      });

      expect(fetchMock).toHaveBeenCalledWith(
        '/api/admin/alert-settings',
        expect.objectContaining({
          method: 'PUT',
          body: expect.any(String),
        })
      );
    });
  });
});