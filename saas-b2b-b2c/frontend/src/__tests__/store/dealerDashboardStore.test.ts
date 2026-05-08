import { useDealerDashboardStore } from '@/store/dealerDashboardStore';
import { renderHook, act } from '@testing-library/react';

describe('useDealerDashboardStore', () => {
  beforeEach(() => {
    useDealerDashboardStore.setState({
      activeTab: 'profit',
      summary: null,
      dealerCenterName: 'Дилерский центр',
      isLoading: false,
      lastUpdated: null,
    });
  });

  describe('setActiveTab', () => {
    it('sets active tab to profit', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      act(() => result.current.setActiveTab('profit'));
      expect(result.current.activeTab).toBe('profit');
    });

    it('sets active tab to revenue', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      act(() => result.current.setActiveTab('revenue'));
      expect(result.current.activeTab).toBe('revenue');
    });

    it('sets active tab to plan', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      act(() => result.current.setActiveTab('plan'));
      expect(result.current.activeTab).toBe('plan');
    });
  });

  describe('setSummary', () => {
    it('sets summary with all fields', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      const summary = {
        netProfit: 100000,
        grossRevenue: 500000,
        planCompletionPercent: 75,
        marginProfit: 25,
        activeAlerts: 3,
      };
      act(() => result.current.setSummary(summary));
      expect(result.current.summary).toEqual(summary);
    });

    it('updates existing summary', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      useDealerDashboardStore.setState({
        summary: { netProfit: 100, grossRevenue: 500, planCompletionPercent: 50, marginProfit: 20, activeAlerts: 1 },
      });
      act(() => result.current.setSummary({ netProfit: 200, grossRevenue: 600, planCompletionPercent: 60, marginProfit: 25, activeAlerts: 0 }));
      expect(result.current.summary?.netProfit).toBe(200);
      expect(result.current.summary?.activeAlerts).toBe(0);
    });
  });

  describe('setDealerCenterName', () => {
    it('sets dealer center name', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      act(() => result.current.setDealerCenterName('New Center'));
      expect(result.current.dealerCenterName).toBe('New Center');
    });
  });

  describe('setLoading', () => {
    it('sets loading to true', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      act(() => result.current.setLoading(true));
      expect(result.current.isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      useDealerDashboardStore.setState({ isLoading: true });
      act(() => result.current.setLoading(false));
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe('setLastUpdated', () => {
    it('sets last updated date', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      const date = new Date('2024-01-01');
      act(() => result.current.setLastUpdated(date));
      expect(result.current.lastUpdated).toEqual(date);
    });
  });

  describe('initial state', () => {
    it('has correct default values', () => {
      const { result } = renderHook(() => useDealerDashboardStore());
      expect(result.current.activeTab).toBe('profit');
      expect(result.current.summary).toBeNull();
      expect(result.current.dealerCenterName).toBe('Дилерский центр');
      expect(result.current.isLoading).toBe(false);
      expect(result.current.lastUpdated).toBeNull();
    });
  });
});