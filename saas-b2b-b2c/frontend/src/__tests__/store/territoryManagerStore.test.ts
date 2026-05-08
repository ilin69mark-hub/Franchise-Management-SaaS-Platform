import { useTerritoryManagerStore, TerritoryManager, TerritorySummary, DealerMetrics } from '@/store/territoryManagerStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

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
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('has default activeTab', () => {
      expect(useTerritoryManagerStore.getState().activeTab).toBe('map');
    });

    it('has manager as null', () => {
      expect(useTerritoryManagerStore.getState().manager).toBeNull();
    });

    it('has summary as null', () => {
      expect(useTerritoryManagerStore.getState().summary).toBeNull();
    });

    it('has empty dealers array', () => {
      expect(useTerritoryManagerStore.getState().dealers).toEqual([]);
    });

    it('has isLoading as false', () => {
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });

    it('has summaryModalOpen as false', () => {
      expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(false);
    });
  });

  describe('setActiveTab', () => {
    it('sets active tab', () => {
      useTerritoryManagerStore.getState().setActiveTab('dealers');
      expect(useTerritoryManagerStore.getState().activeTab).toBe('dealers');
    });
  });

  describe('setManager', () => {
    it('sets territory manager', () => {
      const manager: TerritoryManager = {
        id: 'tm-1',
        name: 'Иван Петров',
        territoryName: 'Центральный округ',
        region: 'ЦФО',
      };
      
      useTerritoryManagerStore.getState().setManager(manager);
      expect(useTerritoryManagerStore.getState().manager).toEqual(manager);
    });
  });

  describe('setSummary', () => {
    it('sets territory summary', () => {
      const summary: TerritorySummary = {
        planCompletionPercent: 85,
        quarterForecastPercent: 92,
        redZoneDealersCount: 3,
        avgConversion: 4.5,
        activeAlerts: 7,
      };
      
      useTerritoryManagerStore.getState().setSummary(summary);
      expect(useTerritoryManagerStore.getState().summary).toEqual(summary);
    });
  });

  describe('setDealers', () => {
    it('sets dealers list', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: 'd1', dealerName: 'Dealer 1', salonCount: 5, planPercent: 90, forecastPercent: 100, conversion: 5, status: 'green' },
        { dealerId: 'd2', dealerName: 'Dealer 2', salonCount: 3, planPercent: 60, forecastPercent: 70, conversion: 3, status: 'yellow' },
      ];
      
      useTerritoryManagerStore.getState().setDealers(dealers);
      expect(useTerritoryManagerStore.getState().dealers).toHaveLength(2);
    });

    it('replaces existing dealers', () => {
      useTerritoryManagerStore.getState().setDealers([{ dealerId: 'd1', dealerName: 'Dealer 1', salonCount: 5, planPercent: 90, forecastPercent: 100, conversion: 5, status: 'green' }]);
      useTerritoryManagerStore.getState().setDealers([]);
      expect(useTerritoryManagerStore.getState().dealers).toHaveLength(0);
    });
  });

  describe('setLoading', () => {
    it('sets loading to true', () => {
      useTerritoryManagerStore.getState().setLoading(true);
      expect(useTerritoryManagerStore.getState().isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      useTerritoryManagerStore.setState({ isLoading: true });
      useTerritoryManagerStore.getState().setLoading(false);
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });
  });

  describe('setSummaryModalOpen', () => {
    it('opens summary modal', () => {
      useTerritoryManagerStore.getState().setSummaryModalOpen(true);
      expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(true);
    });

    it('closes summary modal', () => {
      useTerritoryManagerStore.setState({ summaryModalOpen: true });
      useTerritoryManagerStore.getState().setSummaryModalOpen(false);
      expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(false);
    });
  });

  describe('fetchSummary async', () => {
    it('fetches and sets summary successfully', async () => {
      const mockData = {
        planCompletionPercent: 82,
        quarterForecastPercent: 91,
        redZoneDealersCount: 2,
        avgConversion: 4.2,
        activeAlerts: 5,
      };

      mockApiClient.get.mockResolvedValue({ data: mockData });

      await useTerritoryManagerStore.getState().fetchSummary();

      expect(useTerritoryManagerStore.getState().summary).toEqual(mockData);
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });

    it('falls back to default summary on error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API Error'));

      await useTerritoryManagerStore.getState().fetchSummary();

      expect(useTerritoryManagerStore.getState().summary).toEqual({
        planCompletionPercent: 82,
        quarterForecastPercent: 91,
        redZoneDealersCount: 2,
        avgConversion: 4.2,
        activeAlerts: 5,
      });
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });

    it('sets isLoading during fetch', async () => {
      let resolvePromise: (value: unknown) => void;
      mockApiClient.get.mockImplementation(() => new Promise((resolve) => {
        resolvePromise = resolve;
      }));

      const fetchPromise = useTerritoryManagerStore.getState().fetchSummary();
      expect(useTerritoryManagerStore.getState().isLoading).toBe(true);

      resolvePromise!({ data: { planCompletionPercent: 82 } });
      await fetchPromise;
      expect(useTerritoryManagerStore.getState().isLoading).toBe(false);
    });
  });

  describe('fetchDealers async', () => {
    it('fetches and maps dealers correctly', async () => {
      const mockData = {
        dealers: [
          { id: 'd1', dealer_name: 'Dealer One', salon_count: 5, plan_percent: 90, conversion: 5 },
          { id: 'd2', dealer_name: 'Dealer Two', salon_count: 3, plan_percent: 60, conversion: 3 },
        ],
      };

      mockApiClient.get.mockResolvedValue({ data: mockData });

      await useTerritoryManagerStore.getState().fetchDealers();

      const dealers = useTerritoryManagerStore.getState().dealers;
      expect(dealers).toHaveLength(2);
      expect(dealers[0].dealerName).toBe('Dealer One');
      expect(dealers[0].status).toBe('green');
      expect(dealers[1].status).toBe('yellow');
    });

    it('handles empty dealers array', async () => {
      mockApiClient.get.mockResolvedValue({ data: { dealers: [] } });

      await useTerritoryManagerStore.getState().fetchDealers();

      expect(useTerritoryManagerStore.getState().dealers).toEqual([]);
    });

    it('maps plan percent to forecast correctly', async () => {
      const mockData = {
        dealers: [
          { id: 'd1', dealer_name: 'High Plan', salon_count: 5, plan_percent: 95, conversion: 5 },
          { id: 'd2', dealer_name: 'Medium Plan', salon_count: 3, plan_percent: 60, conversion: 3 },
          { id: 'd3', dealer_name: 'Low Plan', salon_count: 2, plan_percent: 30, conversion: 1 },
        ],
      };

      mockApiClient.get.mockResolvedValue({ data: mockData });

      await useTerritoryManagerStore.getState().fetchDealers();

      const dealers = useTerritoryManagerStore.getState().dealers;
      expect(dealers[0].forecastPercent).toBe(100);
      expect(dealers[1].forecastPercent).toBe(70);
      expect(dealers[2].forecastPercent).toBe(30);
    });
  });

  describe('fetchFunnel, fetchPlanFact, fetchBenchmarks', () => {
    it('fetchFunnel returns data on success', async () => {
      const mockData = { funnel: 'data' };
      mockApiClient.get.mockResolvedValue({ data: mockData });

      const result = await useTerritoryManagerStore.getState().fetchFunnel('week');
      expect(result).toEqual(mockData);
    });

    it('fetchFunnel returns null on error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Error'));

      const result = await useTerritoryManagerStore.getState().fetchFunnel();
      expect(result).toBeNull();
    });

    it('fetchPlanFact returns data', async () => {
      const mockData = { planFact: 'data' };
      mockApiClient.get.mockResolvedValue({ data: mockData });

      const result = await useTerritoryManagerStore.getState().fetchPlanFact('quarter');
      expect(result).toEqual(mockData);
    });

    it('fetchBenchmarks returns data', async () => {
      const mockData = { benchmarks: 'data' };
      mockApiClient.get.mockResolvedValue({ data: mockData });

      const result = await useTerritoryManagerStore.getState().fetchBenchmarks();
      expect(result).toEqual(mockData);
    });
  });
});