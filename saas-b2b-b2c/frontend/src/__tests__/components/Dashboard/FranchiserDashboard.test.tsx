import { useFranchiserStore, FranchiserSummary, DealerMetrics, FranchiseAlert } from '@/store/franchiserStore';

describe('FranchiserDashboard Store', () => {
  beforeEach(() => {
    useFranchiserStore.setState({
      summary: {
        planPercent: 78,
        forecastPercent: 85,
        activeDealers: 24,
        avgConversion: 12,
        avgMargin: 32,
      },
      dealers: [],
      alerts: [],
      alertCount: 0,
      activeTab: 'overview',
      isLoading: false,
    });
  });

  describe('initial state', () => {
    it('has default summary', () => {
      const summary = useFranchiserStore.getState().summary;
      expect(summary.planPercent).toBe(78);
      expect(summary.forecastPercent).toBe(85);
      expect(summary.activeDealers).toBe(24);
    });

    it('has empty dealers array', () => {
      expect(useFranchiserStore.getState().dealers).toEqual([]);
    });

    it('has empty alerts array', () => {
      expect(useFranchiserStore.getState().alerts).toEqual([]);
    });

    it('has alertCount 0', () => {
      expect(useFranchiserStore.getState().alertCount).toBe(0);
    });

    it('has activeTab overview', () => {
      expect(useFranchiserStore.getState().activeTab).toBe('overview');
    });

    it('has isLoading false', () => {
      expect(useFranchiserStore.getState().isLoading).toBe(false);
    });
  });

  describe('setSummary', () => {
    it('updates franchiser summary', () => {
      const newSummary: FranchiserSummary = {
        planPercent: 90,
        forecastPercent: 95,
        activeDealers: 30,
        avgConversion: 15,
        avgMargin: 35,
      };
      useFranchiserStore.getState().setSummary(newSummary);
      const summary = useFranchiserStore.getState().summary;
      expect(summary.planPercent).toBe(90);
      expect(summary.activeDealers).toBe(30);
    });

    it('preserves other state', () => {
      useFranchiserStore.getState().setSummary({ ...useFranchiserStore.getState().summary, planPercent: 80 });
      expect(useFranchiserStore.getState().activeTab).toBe('overview');
    });
  });

  describe('setAlertCount', () => {
    it('sets alert count', () => {
      useFranchiserStore.getState().setAlertCount(5);
      expect(useFranchiserStore.getState().alertCount).toBe(5);
    });

    it('increments alert count', () => {
      useFranchiserStore.getState().setAlertCount(3);
      useFranchiserStore.getState().setAlertCount(useFranchiserStore.getState().alertCount + 2);
      expect(useFranchiserStore.getState().alertCount).toBe(5);
    });
  });

  describe('setActiveTab', () => {
    it('changes tab to network', () => {
      useFranchiserStore.getState().setActiveTab('network');
      expect(useFranchiserStore.getState().activeTab).toBe('network');
    });

    it('changes tab to team', () => {
      useFranchiserStore.getState().setActiveTab('team');
      expect(useFranchiserStore.getState().activeTab).toBe('team');
    });

    it('changes tab to health', () => {
      useFranchiserStore.getState().setActiveTab('health');
      expect(useFranchiserStore.getState().activeTab).toBe('health');
    });

    it('changes tab to report', () => {
      useFranchiserStore.getState().setActiveTab('report');
      expect(useFranchiserStore.getState().activeTab).toBe('report');
    });
  });

  describe('dealers management', () => {
    it('sets dealers list', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: '1', dealerName: 'Dealer 1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, margin: 30, status: 'green' },
        { dealerId: '2', dealerName: 'Dealer 2', salonCount: 3, planPercent: 60, forecastPercent: 55, conversion: 8, margin: 25, status: 'red' },
      ];
      useFranchiserStore.getState().setDealers(dealers);
      expect(useFranchiserStore.getState().dealers).toHaveLength(2);
    });

    it('filters green dealers', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: '1', dealerName: 'Green', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, margin: 30, status: 'green' },
        { dealerId: '2', dealerName: 'Red', salonCount: 3, planPercent: 40, forecastPercent: 35, conversion: 5, margin: 15, status: 'red' },
      ];
      useFranchiserStore.getState().setDealers(dealers);
      const greenDealers = useFranchiserStore.getState().dealers.filter(d => d.status === 'green');
      expect(greenDealers).toHaveLength(1);
    });

    it('calculates total salons', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: '1', dealerName: 'D1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, margin: 30, status: 'green' },
        { dealerId: '2', dealerName: 'D2', salonCount: 3, planPercent: 70, forecastPercent: 75, conversion: 10, margin: 28, status: 'green' },
      ];
      useFranchiserStore.getState().setDealers(dealers);
      const total = dealers.reduce((sum, d) => sum + d.salonCount, 0);
      expect(total).toBe(8);
    });
  });

  describe('alerts management', () => {
    it('sets alerts', () => {
      const alerts: FranchiseAlert[] = [
        { id: '1', type: 'critical', message: 'Critical alert', dealerId: 'd1', createdAt: '2026-01-01' },
        { id: '2', type: 'warning', message: 'Warning alert', dealerId: 'd2', createdAt: '2026-01-02' },
      ];
      useFranchiserStore.getState().setAlerts(alerts);
      expect(useFranchiserStore.getState().alerts).toHaveLength(2);
    });

    it('filters critical alerts', () => {
      const alerts: FranchiseAlert[] = [
        { id: '1', type: 'critical', message: 'Critical', dealerId: 'd1', createdAt: '2026-01-01' },
        { id: '2', type: 'warning', message: 'Warning', dealerId: 'd2', createdAt: '2026-01-02' },
        { id: '3', type: 'critical', message: 'Critical 2', dealerId: 'd3', createdAt: '2026-01-03' },
      ];
      useFranchiserStore.getState().setAlerts(alerts);
      const critical = useFranchiserStore.getState().alerts.filter(a => a.type === 'critical');
      expect(critical).toHaveLength(2);
    });
  });

  describe('summary calculations', () => {
    it('calculates performance gap', () => {
      const summary = useFranchiserStore.getState().summary;
      const gap = summary.forecastPercent - summary.planPercent;
      expect(gap).toBe(7);
    });

    it('calculates average performance', () => {
      const summary = useFranchiserStore.getState().summary;
      const avg = (summary.planPercent + summary.forecastPercent) / 2;
      expect(avg).toBe(81.5);
    });

    it('calculates weighted conversion', () => {
      const summary = useFranchiserStore.getState().summary;
      const weighted = (summary.avgConversion * summary.activeDealers) / Math.max(summary.activeDealers, 1);
      expect(weighted).toBe(12);
    });
  });

  describe('store methods exist', () => {
    it('has setSummary', () => {
      expect(typeof useFranchiserStore.getState().setSummary).toBe('function');
    });

    it('has setDealers', () => {
      expect(typeof useFranchiserStore.getState().setDealers).toBe('function');
    });

    it('has setAlerts', () => {
      expect(typeof useFranchiserStore.getState().setAlerts).toBe('function');
    });

    it('has setAlertCount', () => {
      expect(typeof useFranchiserStore.getState().setAlertCount).toBe('function');
    });

    it('has setActiveTab', () => {
      expect(typeof useFranchiserStore.getState().setActiveTab).toBe('function');
    });

    it('has setLoading', () => {
      expect(typeof useFranchiserStore.getState().setLoading).toBe('function');
    });

    it('has fetchSummary', () => {
      expect(typeof useFranchiserStore.getState().fetchSummary).toBe('function');
    });

    it('has fetchNetwork', () => {
      expect(typeof useFranchiserStore.getState().fetchNetwork).toBe('function');
    });
  });
});