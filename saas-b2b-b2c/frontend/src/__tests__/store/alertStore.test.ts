import { useAlertStore } from '@/store/alertStore';

describe('useAlertStore', () => {
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
        inactiveDealerDays: 3,
        nonLiquidDays: 90,
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

  describe('setAlerts', () => {
    it('sets alerts array', () => {
      const { result } = renderHook(() => useAlertStore());
      const mockAlerts = [
        { id: '1', title: 'Alert 1', status: 'new' as const },
        { id: '2', title: 'Alert 2', status: 'read' as const },
      ];
      act(() => result.current.setAlerts(mockAlerts));
      expect(result.current.alerts).toEqual(mockAlerts);
    });

    it('replaces existing alerts', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({ alerts: [{ id: 'old', title: 'Old' }] as any });
      act(() => result.current.setAlerts([{ id: 'new', title: 'New' }] as any));
      expect(result.current.alerts).toHaveLength(1);
      expect(result.current.alerts[0].id).toBe('new');
    });
  });

  describe('addAlert', () => {
    it('adds alert to beginning of array', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({ alerts: [{ id: '1', title: 'First' }] as any });
      const newAlert = { id: '2', title: 'New', status: 'new' as const };
      act(() => result.current.addAlert(newAlert as any));
      expect(result.current.alerts).toHaveLength(2);
      expect(result.current.alerts[0].title).toBe('New');
    });

    it('handles empty alerts array', () => {
      const { result } = renderHook(() => useAlertStore());
      const alert = { id: '1', title: 'Alert', status: 'new' as const };
      act(() => result.current.addAlert(alert as any));
      expect(result.current.alerts).toHaveLength(1);
    });
  });

  describe('markAsRead', () => {
    it('marks single alert as read', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({
        alerts: [
          { id: '1', status: 'new' as const },
          { id: '2', status: 'new' as const },
        ] as any,
      });
      act(() => result.current.markAsRead('1'));
      expect(result.current.alerts[0].status).toBe('read');
      expect(result.current.alerts[1].status).toBe('new');
    });

    it('does nothing if alert not found', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({
        alerts: [{ id: '1', status: 'new' as const }] as any,
      });
      act(() => result.current.markAsRead('invalid'));
      expect(result.current.alerts[0].status).toBe('new');
    });
  });

  describe('markAllAsRead', () => {
    it('marks all alerts as read', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({
        alerts: [
          { id: '1', status: 'new' as const },
          { id: '2', status: 'new' as const },
          { id: '3', status: 'acknowledged' as const },
        ] as any,
      });
      act(() => result.current.markAllAsRead());
      expect(result.current.alerts.every(a => a.status === 'read')).toBe(true);
    });

    it('handles empty alerts', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.markAllAsRead());
      expect(result.current.alerts).toHaveLength(0);
    });
  });

  describe('setFilter', () => {
    it('sets filter to all', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setFilter('all'));
      expect(result.current.filter).toBe('all');
    });

    it('sets filter to critical', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setFilter('critical'));
      expect(result.current.filter).toBe('critical');
    });

    it('sets filter to category', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setFilter('plan'));
      expect(result.current.filter).toBe('plan');
    });
  });

  describe('setSettings', () => {
    it('updates single setting', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setSettings({ enabled: false }));
      expect(result.current.settings.enabled).toBe(false);
    });

    it('updates nested setting', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setSettings({ planThreshold: 50 }));
      expect(result.current.settings.planThreshold).toBe(50);
    });

    it('updates channels', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setSettings({
        channels: { inApp: false, emailInstant: true, emailDigest: false, push: true },
      }));
      expect(result.current.settings.channels.inApp).toBe(false);
      expect(result.current.settings.channels.emailInstant).toBe(true);
      expect(result.current.settings.channels.push).toBe(true);
    });

    it('preserves other settings', () => {
      const { result } = renderHook(() => useAlertStore());
      const original = { ...result.current.settings };
      act(() => result.current.setSettings({ enabled: false }));
      expect(result.current.settings.planThreshold).toBe(original.planThreshold);
      expect(result.current.settings.stuckDealDays).toBe(original.stuckDealDays);
    });
  });

  describe('getUnreadCount', () => {
    it('returns count of new alerts', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({
        alerts: [
          { id: '1', status: 'new' as const },
          { id: '2', status: 'new' as const },
          { id: '3', status: 'read' as const },
        ] as any,
      });
      expect(result.current.getUnreadCount()).toBe(2);
    });

    it('returns 0 when no new alerts', () => {
      const { result } = renderHook(() => useAlertStore());
      useAlertStore.setState({
        alerts: [
          { id: '1', status: 'read' as const },
          { id: '2', status: 'acknowledged' as const },
        ] as any,
      });
      expect(result.current.getUnreadCount()).toBe(0);
    });

    it('returns 0 for empty alerts', () => {
      const { result } = renderHook(() => useAlertStore());
      expect(result.current.getUnreadCount()).toBe(0);
    });
  });

  describe('initial state', () => {
    it('has correct default values', () => {
      const { result } = renderHook(() => useAlertStore());
      expect(result.current.alerts).toEqual([]);
      expect(result.current.filter).toBe('all');
      expect(result.current.isLoading).toBe(false);
      expect(result.current.settings.enabled).toBe(true);
      expect(result.current.settings.planThreshold).toBe(70);
      expect(result.current.settings.channels.inApp).toBe(true);
    });
  });

  describe('persistence', () => {
    it('persists settings', () => {
      const { result } = renderHook(() => useAlertStore());
      act(() => result.current.setSettings({ planThreshold: 80 }));
      const stored = localStorage.getItem('alert-storage');
      expect(stored).toBeTruthy();
    });
  });
});

import { renderHook, act } from '@testing-library/react';