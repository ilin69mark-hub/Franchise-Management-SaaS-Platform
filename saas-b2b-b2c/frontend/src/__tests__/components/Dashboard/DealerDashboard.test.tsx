import { useDealerDashboardStore } from '@/store/dealerDashboardStore';

describe('DealerDashboard Store', () => {
  beforeEach(() => {
    useDealerDashboardStore.setState({
      activeTab: 'summary',
      summary: {
        planPercent: 85,
        forecastPercent: 90,
        conversion: 12,
        margin: 28,
        activeClients: 45,
        pendingActions: 5,
      },
      dealerCenterName: 'Dealer Center Moscow',
      isLoading: false,
      lastUpdated: null,
      alerts: [],
    });
  });

  describe('initial state', () => {
    it('has default activeTab', () => {
      expect(useDealerDashboardStore.getState().activeTab).toBe('summary');
    });

    it('has summary object', () => {
      const summary = useDealerDashboardStore.getState().summary;
      expect(summary).toHaveProperty('planPercent');
      expect(summary).toHaveProperty('forecastPercent');
      expect(summary).toHaveProperty('conversion');
      expect(summary).toHaveProperty('margin');
    });

    it('has dealerCenterName', () => {
      expect(useDealerDashboardStore.getState().dealerCenterName).toBe('Dealer Center Moscow');
    });

    it('has isLoading false', () => {
      expect(useDealerDashboardStore.getState().isLoading).toBe(false);
    });
  });

  describe('setActiveTab', () => {
    it('changes active tab', () => {
      useDealerDashboardStore.getState().setActiveTab('products');
      expect(useDealerDashboardStore.getState().activeTab).toBe('products');
    });

    it('changes to funnel tab', () => {
      useDealerDashboardStore.getState().setActiveTab('funnel');
      expect(useDealerDashboardStore.getState().activeTab).toBe('funnel');
    });

    it('changes to communications tab', () => {
      useDealerDashboardStore.getState().setActiveTab('communications');
      expect(useDealerDashboardStore.getState().activeTab).toBe('communications');
    });
  });

  describe('setSummary', () => {
    it('updates summary data', () => {
      useDealerDashboardStore.getState().setSummary({
        planPercent: 100,
        forecastPercent: 95,
        conversion: 15,
        margin: 30,
        activeClients: 50,
        pendingActions: 2,
      });
      const summary = useDealerDashboardStore.getState().summary;
      expect(summary.planPercent).toBe(100);
      expect(summary.forecastPercent).toBe(95);
    });

    it('preserves other state when updating summary', () => {
      useDealerDashboardStore.getState().setSummary({
        planPercent: 80,
        forecastPercent: 85,
        conversion: 10,
        margin: 25,
        activeClients: 40,
        pendingActions: 3,
      });
      expect(useDealerDashboardStore.getState().dealerCenterName).toBe('Dealer Center Moscow');
      expect(useDealerDashboardStore.getState().activeTab).toBe('summary');
    });
  });

  describe('setLoading', () => {
    it('sets loading to true', () => {
      useDealerDashboardStore.getState().setLoading(true);
      expect(useDealerDashboardStore.getState().isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      useDealerDashboardStore.getState().setLoading(true);
      useDealerDashboardStore.getState().setLoading(false);
      expect(useDealerDashboardStore.getState().isLoading).toBe(false);
    });
  });

  describe('setLastUpdated', () => {
    it('sets last updated date', () => {
      const now = new Date();
      useDealerDashboardStore.getState().setLastUpdated(now);
      expect(useDealerDashboardStore.getState().lastUpdated).toEqual(now);
    });
  });

  describe('summary calculations', () => {
    it('calculates plan vs forecast difference', () => {
      const summary = useDealerDashboardStore.getState().summary;
      const diff = summary.forecastPercent - summary.planPercent;
      expect(diff).toBe(5);
    });

    it('calculates average metrics', () => {
      const summary = useDealerDashboardStore.getState().summary;
      const avgPerformance = (summary.planPercent + summary.forecastPercent) / 2;
      expect(avgPerformance).toBe(87.5);
    });

    it('identifies pending actions', () => {
      const summary = useDealerDashboardStore.getState().summary;
      expect(summary.pendingActions).toBeGreaterThan(0);
    });
  });

  describe('store methods exist', () => {
    it('has setActiveTab', () => {
      expect(typeof useDealerDashboardStore.getState().setActiveTab).toBe('function');
    });

    it('has setSummary', () => {
      expect(typeof useDealerDashboardStore.getState().setSummary).toBe('function');
    });

    it('has setLoading', () => {
      expect(typeof useDealerDashboardStore.getState().setLoading).toBe('function');
    });

    it('has setLastUpdated', () => {
      expect(typeof useDealerDashboardStore.getState().setLastUpdated).toBe('function');
    });

    it('has setLastUpdated method', () => {
      expect(typeof useDealerDashboardStore.getState().setLastUpdated).toBe('function');
    });
  });
});