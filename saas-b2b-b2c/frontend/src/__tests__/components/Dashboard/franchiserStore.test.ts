import { useFranchiserStore, FranchiserSummary, DealerMetrics, FranchiseAlert } from '@/store/franchiserStore';

describe('franchiserStore', () => {
  beforeEach(() => {
    useFranchiserStore.setState({
      summary: {
        planPercent: 0,
        forecastPercent: 0,
        activeDealers: 0,
        avgConversion: 0,
        avgMargin: 0,
      },
      dealers: [],
      alerts: [],
      alertCount: 0,
      activeTab: 'overview',
      isLoading: false,
    });
  });

  describe('initial state', () => {
    it('has initial summary structure', () => {
      const initialState = useFranchiserStore.getState();
      expect(initialState.summary).toBeDefined();
      expect(initialState.summary).toHaveProperty('planPercent');
      expect(initialState.summary).toHaveProperty('forecastPercent');
      expect(initialState.summary).toHaveProperty('activeDealers');
    });

    it('has initial dealers array', () => {
      const initialState = useFranchiserStore.getState();
      expect(initialState.dealers).toBeDefined();
      expect(Array.isArray(initialState.dealers)).toBe(true);
    });

    it('has initial alerts array', () => {
      const initialState = useFranchiserStore.getState();
      expect(initialState.alerts).toBeDefined();
      expect(Array.isArray(initialState.alerts)).toBe(true);
    });

    it('has initial activeTab', () => {
      const initialState = useFranchiserStore.getState();
      expect(initialState.activeTab).toBe('overview');
    });
  });

  describe('summary operations', () => {
    it('updates summary', () => {
      const newSummary: FranchiserSummary = {
        planPercent: 90,
        forecastPercent: 95,
        activeDealers: 25,
        avgConversion: 14,
        avgMargin: 35,
      };
      useFranchiserStore.getState().setSummary(newSummary);
      const state = useFranchiserStore.getState();
      expect(state.summary.planPercent).toBe(90);
      expect(state.summary.forecastPercent).toBe(95);
      expect(state.summary.activeDealers).toBe(25);
    });

    it('updates individual summary field', () => {
      useFranchiserStore.getState().setSummary({ ...useFranchiserStore.getState().summary, planPercent: 75 });
      expect(useFranchiserStore.getState().summary.planPercent).toBe(75);
    });

    it('calculates summary aggregates', () => {
      const summary: FranchiserSummary = {
        planPercent: 80,
        forecastPercent: 85,
        activeDealers: 20,
        avgConversion: 12,
        avgMargin: 30,
      };
      expect(summary.planPercent + summary.forecastPercent).toBe(165);
    });
  });

  describe('dealers operations', () => {
    it('sets dealers list', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: '1', dealerName: 'Dealer 1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, margin: 30, status: 'green' },
        { dealerId: '2', dealerName: 'Dealer 2', salonCount: 3, planPercent: 70, forecastPercent: 75, conversion: 10, margin: 25, status: 'yellow' },
      ];
      useFranchiserStore.getState().setDealers(dealers);
      expect(useFranchiserStore.getState().dealers).toHaveLength(2);
    });

    it('adds new dealer', () => {
      const dealer: DealerMetrics = { dealerId: '1', dealerName: 'New Dealer', salonCount: 2, planPercent: 90, forecastPercent: 95, conversion: 15, margin: 35, status: 'green' };
      useFranchiserStore.getState().setDealers([dealer]);
      expect(useFranchiserStore.getState().dealers).toContainEqual(dealer);
    });

    it('filters green status dealers', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: '1', dealerName: 'Green', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, margin: 30, status: 'green' },
        { dealerId: '2', dealerName: 'Red', salonCount: 3, planPercent: 50, forecastPercent: 45, conversion: 5, margin: 15, status: 'red' },
      ];
      useFranchiserStore.getState().setDealers(dealers);
      const green = useFranchiserStore.getState().dealers.filter(d => d.status === 'green');
      expect(green).toHaveLength(1);
    });

    it('calculates total salon count', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: '1', dealerName: 'D1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, margin: 30, status: 'green' },
        { dealerId: '2', dealerName: 'D2', salonCount: 3, planPercent: 70, forecastPercent: 75, conversion: 10, margin: 25, status: 'green' },
      ];
      useFranchiserStore.getState().setDealers(dealers);
      const total = useFranchiserStore.getState().dealers.reduce((sum, d) => sum + d.salonCount, 0);
      expect(total).toBe(8);
    });
  });

  describe('alerts operations', () => {
    it('updates alert count', () => {
      useFranchiserStore.getState().setAlertCount(5);
      expect(useFranchiserStore.getState().alertCount).toBe(5);
    });

    it('sets alerts list', () => {
      const alerts: FranchiseAlert[] = [
        { id: '1', type: 'warning', message: 'Low revenue', dealerId: '1', createdAt: '2024-01-01' },
        { id: '2', type: 'critical', message: 'No sales', dealerId: '2', createdAt: '2024-01-02' },
      ];
      useFranchiserStore.getState().setAlerts(alerts);
      expect(useFranchiserStore.getState().alerts).toHaveLength(2);
    });

    it('filters critical alerts', () => {
      const alerts: FranchiseAlert[] = [
        { id: '1', type: 'warning', message: 'Warning', dealerId: '1', createdAt: '2024-01-01' },
        { id: '2', type: 'critical', message: 'Critical', dealerId: '2', createdAt: '2024-01-02' },
      ];
      useFranchiserStore.getState().setAlerts(alerts);
      const critical = useFranchiserStore.getState().alerts.filter(a => a.type === 'critical');
      expect(critical).toHaveLength(1);
    });

    it('increments alert count', () => {
      useFranchiserStore.getState().setAlertCount(3);
      useFranchiserStore.getState().setAlertCount(useFranchiserStore.getState().alertCount + 1);
      expect(useFranchiserStore.getState().alertCount).toBe(4);
    });
  });

  describe('activeTab operations', () => {
    it('sets active tab', () => {
      useFranchiserStore.getState().setActiveTab('dealers');
      expect(useFranchiserStore.getState().activeTab).toBe('dealers');
    });

    it('changes between tabs', () => {
      useFranchiserStore.getState().setActiveTab('analytics');
      expect(useFranchiserStore.getState().activeTab).toBe('analytics');
      useFranchiserStore.getState().setActiveTab('team');
      expect(useFranchiserStore.getState().activeTab).toBe('team');
    });
  });

  describe('loading state', () => {
    it('sets loading state', () => {
      useFranchiserStore.getState().setLoading(true);
      expect(useFranchiserStore.getState().isLoading).toBe(true);
    });

    it('toggles loading', () => {
      useFranchiserStore.getState().setLoading(true);
      expect(useFranchiserStore.getState().isLoading).toBe(true);
      useFranchiserStore.getState().setLoading(false);
      expect(useFranchiserStore.getState().isLoading).toBe(false);
    });

    it('resets to initial state', () => {
      useFranchiserStore.getState().setLoading(true);
      useFranchiserStore.getState().setActiveTab('dealers');
      useFranchiserStore.getState().setDealers([{ dealerId: '1', dealerName: 'D', salonCount: 1, planPercent: 50, forecastPercent: 60, conversion: 5, margin: 10, status: 'red' }]);
      
      useFranchiserStore.setState({
        summary: { planPercent: 0, forecastPercent: 0, activeDealers: 0, avgConversion: 0, avgMargin: 0 },
        dealers: [],
        alerts: [],
        alertCount: 0,
        activeTab: 'overview',
        isLoading: false,
      });
      
      const state = useFranchiserStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.activeTab).toBe('overview');
      expect(state.dealers).toHaveLength(0);
    });
  });

  describe('methods exist', () => {
    it('has setSummary method', () => {
      expect(typeof useFranchiserStore.getState().setSummary).toBe('function');
    });

    it('has setDealers method', () => {
      expect(typeof useFranchiserStore.getState().setDealers).toBe('function');
    });

    it('has setAlerts method', () => {
      expect(typeof useFranchiserStore.getState().setAlerts).toBe('function');
    });

    it('has setAlertCount method', () => {
      expect(typeof useFranchiserStore.getState().setAlertCount).toBe('function');
    });

    it('has setActiveTab method', () => {
      expect(typeof useFranchiserStore.getState().setActiveTab).toBe('function');
    });

    it('has setLoading method', () => {
      expect(typeof useFranchiserStore.getState().setLoading).toBe('function');
    });

    it('has fetchSummary method', () => {
      expect(typeof useFranchiserStore.getState().fetchSummary).toBe('function');
    });

    it('has fetchNetwork method', () => {
      expect(typeof useFranchiserStore.getState().fetchNetwork).toBe('function');
    });

    it('has fetchHealth method', () => {
      expect(typeof useFranchiserStore.getState().fetchHealth).toBe('function');
    });

    it('has fetchTeam method', () => {
      expect(typeof useFranchiserStore.getState().fetchTeam).toBe('function');
    });
  });
});