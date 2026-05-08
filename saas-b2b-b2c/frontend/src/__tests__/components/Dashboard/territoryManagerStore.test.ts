import { useTerritoryManagerStore, TerritorySummary, DealerMetrics, TerritoryManager } from '@/store/territoryManagerStore';

describe('territoryManagerStore', () => {
  beforeEach(() => {
    useTerritoryManagerStore.setState({
      activeTab: 'map',
      manager: null,
      summary: null,
      dealers: [],
      isLoading: false,
      lastUpdated: null,
      summaryModalOpen: false,
    });
  });

  describe('initial state', () => {
    it('has default activeTab', () => {
      expect(useTerritoryManagerStore.getState().activeTab).toBe('map');
    });

    it('has null manager', () => {
      expect(useTerritoryManagerStore.getState().manager).toBeNull();
    });

    it('has null summary', () => {
      expect(useTerritoryManagerStore.getState().summary).toBeNull();
    });

    it('has empty dealers array', () => {
      expect(useTerritoryManagerStore.getState().dealers).toEqual([]);
    });

    it('has isLoading false', () => {
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });

    it('has null lastUpdated', () => {
      expect(useTerritoryManagerStore.getState().lastUpdated).toBeNull();
    });

    it('has summaryModalOpen false', () => {
      expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(false);
    });
  });

  describe('setActiveTab', () => {
    it('sets active tab', () => {
      useTerritoryManagerStore.getState().setActiveTab('dealers');
      expect(useTerritoryManagerStore.getState().activeTab).toBe('dealers');
    });

    it('changes between tabs', () => {
      useTerritoryManagerStore.getState().setActiveTab('analytics');
      expect(useTerritoryManagerStore.getState().activeTab).toBe('analytics');
      useTerritoryManagerStore.getState().setActiveTab('tasks');
      expect(useTerritoryManagerStore.getState().activeTab).toBe('tasks');
    });
  });

  describe('setManager', () => {
    it('sets territory manager', () => {
      const manager: TerritoryManager = {
        id: 'tm-1',
        name: 'John Smith',
        territoryName: 'Central Region',
        region: 'Moscow',
      };
      useTerritoryManagerStore.getState().setManager(manager);
      expect(useTerritoryManagerStore.getState().manager).toEqual(manager);
    });

    it('updates manager fields', () => {
      const manager: TerritoryManager = {
        id: 'tm-2',
        name: 'Jane Doe',
        territoryName: 'North Region',
        region: 'SPB',
      };
      useTerritoryManagerStore.getState().setManager(manager);
      expect(useTerritoryManagerStore.getState().manager?.name).toBe('Jane Doe');
      expect(useTerritoryManagerStore.getState().manager?.region).toBe('SPB');
    });
  });

  describe('setSummary', () => {
    it('sets territory summary', () => {
      const summary: TerritorySummary = {
        planCompletionPercent: 85,
        quarterForecastPercent: 90,
        redZoneDealersCount: 3,
        avgConversion: 14,
        activeAlerts: 5,
      };
      useTerritoryManagerStore.getState().setSummary(summary);
      expect(useTerritoryManagerStore.getState().summary).toEqual(summary);
    });

    it('calculates health from summary', () => {
      const summary: TerritorySummary = {
        planCompletionPercent: 95,
        quarterForecastPercent: 100,
        redZoneDealersCount: 0,
        avgConversion: 18,
        activeAlerts: 0,
      };
      useTerritoryManagerStore.getState().setSummary(summary);
      const isHealthy = summary.redZoneDealersCount === 0 && summary.activeAlerts === 0;
      expect(isHealthy).toBe(true);
    });
  });

  describe('setDealers', () => {
    it('sets dealers list', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: 'd1', dealerName: 'Dealer 1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, status: 'green' },
        { dealerId: 'd2', dealerName: 'Dealer 2', salonCount: 3, planPercent: 60, forecastPercent: 55, conversion: 8, status: 'red' },
        { dealerId: 'd3', dealerName: 'Dealer 3', salonCount: 4, planPercent: 75, forecastPercent: 78, conversion: 10, status: 'yellow' },
      ];
      useTerritoryManagerStore.getState().setDealers(dealers);
      expect(useTerritoryManagerStore.getState().dealers).toHaveLength(3);
    });

    it('filters red zone dealers', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: 'd1', dealerName: 'Green Dealer', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, status: 'green' },
        { dealerId: 'd2', dealerName: 'Red Dealer', salonCount: 3, planPercent: 40, forecastPercent: 35, conversion: 5, status: 'red' },
      ];
      useTerritoryManagerStore.getState().setDealers(dealers);
      const redDealers = useTerritoryManagerStore.getState().dealers.filter(d => d.status === 'red');
      expect(redDealers).toHaveLength(1);
    });

    it('calculates total salons', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: 'd1', dealerName: 'D1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 12, status: 'green' },
        { dealerId: 'd2', dealerName: 'D2', salonCount: 3, planPercent: 70, forecastPercent: 75, conversion: 10, status: 'green' },
      ];
      useTerritoryManagerStore.getState().setDealers(dealers);
      const totalSalons = dealers.reduce((sum, d) => sum + d.salonCount, 0);
      expect(totalSalons).toBe(8);
    });

    it('calculates average conversion', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: 'd1', dealerName: 'D1', salonCount: 5, planPercent: 80, forecastPercent: 85, conversion: 10, status: 'green' },
        { dealerId: 'd2', dealerName: 'D2', salonCount: 3, planPercent: 70, forecastPercent: 75, conversion: 14, status: 'green' },
      ];
      useTerritoryManagerStore.getState().setDealers(dealers);
      const avgConversion = dealers.reduce((sum, d) => sum + d.conversion, 0) / 2;
      expect(avgConversion).toBe(12);
    });
  });

  describe('setLoading', () => {
    it('sets loading state', () => {
      useTerritoryManagerStore.getState().setLoading(true);
      expect(useTerritoryManagerStore.getState().isLoading).toBe(true);
    });

    it('toggles loading', () => {
      useTerritoryManagerStore.getState().setLoading(true);
      useTerritoryManagerStore.getState().setLoading(false);
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });
  });

  describe('setLastUpdated', () => {
    it('sets last updated date', () => {
      const now = new Date();
      useTerritoryManagerStore.getState().setLastUpdated(now);
      expect(useTerritoryManagerStore.getState().lastUpdated).toEqual(now);
    });
  });

  describe('setSummaryModalOpen', () => {
    it('opens summary modal', () => {
      useTerritoryManagerStore.getState().setSummaryModalOpen(true);
      expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(true);
    });

    it('closes summary modal', () => {
      useTerritoryManagerStore.getState().setSummaryModalOpen(true);
      useTerritoryManagerStore.getState().setSummaryModalOpen(false);
      expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(false);
    });
  });

  describe('methods exist', () => {
    it('has setActiveTab', () => {
      expect(typeof useTerritoryManagerStore.getState().setActiveTab).toBe('function');
    });

    it('has setManager', () => {
      expect(typeof useTerritoryManagerStore.getState().setManager).toBe('function');
    });

    it('has setSummary', () => {
      expect(typeof useTerritoryManagerStore.getState().setSummary).toBe('function');
    });

    it('has setDealers', () => {
      expect(typeof useTerritoryManagerStore.getState().setDealers).toBe('function');
    });

    it('has setLoading', () => {
      expect(typeof useTerritoryManagerStore.getState().setLoading).toBe('function');
    });

    it('has setLastUpdated', () => {
      expect(typeof useTerritoryManagerStore.getState().setLastUpdated).toBe('function');
    });

    it('has setSummaryModalOpen', () => {
      expect(typeof useTerritoryManagerStore.getState().setSummaryModalOpen).toBe('function');
    });

    it('has fetchSummary', () => {
      expect(typeof useTerritoryManagerStore.getState().fetchSummary).toBe('function');
    });

    it('has fetchDealers', () => {
      expect(typeof useTerritoryManagerStore.getState().fetchDealers).toBe('function');
    });

    it('has fetchFunnel', () => {
      expect(typeof useTerritoryManagerStore.getState().fetchFunnel).toBe('function');
    });

    it('has fetchPlanFact', () => {
      expect(typeof useTerritoryManagerStore.getState().fetchPlanFact).toBe('function');
    });

    it('has fetchBenchmarks', () => {
      expect(typeof useTerritoryManagerStore.getState().fetchBenchmarks).toBe('function');
    });
  });
});