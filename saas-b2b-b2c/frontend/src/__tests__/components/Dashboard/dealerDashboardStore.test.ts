// __tests__/components/Dashboard/dealerDashboardStore.test.ts
import { useDealerDashboardStore } from '@/store/dealerDashboardStore';

describe('dealerDashboardStore', () => {
  beforeEach(() => {
    useDealerDashboardStore.setState({
      activeTab: 'profit',
      summary: null,
      dealerCenterName: 'Дилерский центр',
      isLoading: false,
      lastUpdated: null,
    });
  });

  it('should have initial state', () => {
    const state = useDealerDashboardStore.getState();
    expect(state.activeTab).toBe('profit');
    expect(state.summary).toBeNull();
    expect(state.dealerCenterName).toBe('Дилерский центр');
    expect(state.isLoading).toBe(false);
  });

  it('should set active tab', () => {
    const { setActiveTab } = useDealerDashboardStore.getState();
    setActiveTab('funnel');
    expect(useDealerDashboardStore.getState().activeTab).toBe('funnel');
  });

  it('should set summary data', () => {
    const { setSummary } = useDealerDashboardStore.getState();
    setSummary({
      netProfit: 100000,
      grossRevenue: 500000,
      planCompletionPercent: 75,
      marginProfit: 80000,
      activeAlerts: 2,
    });
    const state = useDealerDashboardStore.getState();
    expect(state.summary?.netProfit).toBe(100000);
    expect(state.summary?.grossRevenue).toBe(500000);
    expect(state.summary?.planCompletionPercent).toBe(75);
    expect(state.summary?.marginProfit).toBe(80000);
    expect(state.summary?.activeAlerts).toBe(2);
  });

  it('should set loading state', () => {
    const { setLoading } = useDealerDashboardStore.getState();
    setLoading(true);
    expect(useDealerDashboardStore.getState().isLoading).toBe(true);
  });

  it('should set last updated date', () => {
    const now = new Date();
    const { setLastUpdated } = useDealerDashboardStore.getState();
    setLastUpdated(now);
    expect(useDealerDashboardStore.getState().lastUpdated).toEqual(now);
  });
});