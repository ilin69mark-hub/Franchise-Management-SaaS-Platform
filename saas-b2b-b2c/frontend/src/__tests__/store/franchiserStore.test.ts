import { useFranchiserStore, FranchiserSummary, DealerMetrics, FranchiseAlert } from '@/store/franchiserStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('franchiserStore', () => {
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
      activeTab: 'network',
      isLoading: false,
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('has default summary with correct values', () => {
      const summary = useFranchiserStore.getState().summary;
      expect(summary.planPercent).toBe(78);
      expect(summary.forecastPercent).toBe(85);
      expect(summary.activeDealers).toBe(24);
      expect(summary.avgConversion).toBe(12);
      expect(summary.avgMargin).toBe(32);
    });

    it('has empty dealers array', () => {
      expect(useFranchiserStore.getState().dealers).toEqual([]);
    });

    it('has empty alerts array', () => {
      expect(useFranchiserStore.getState().alerts).toEqual([]);
    });

    it('has alertCount as 0', () => {
      expect(useFranchiserStore.getState().alertCount).toBe(0);
    });

    it('has activeTab as network', () => {
      expect(useFranchiserStore.getState().activeTab).toBe('network');
    });

    it('has isLoading as false', () => {
      expect(useFranchiserStore.getState().isLoading).toBe(false);
    });
  });

  describe('setSummary action', () => {
    it('sets summary data completely', () => {
      const newSummary: FranchiserSummary = {
        planPercent: 90,
        forecastPercent: 95,
        activeDealers: 30,
        avgConversion: 15,
        avgMargin: 35,
      };

      useFranchiserStore.getState().setSummary(newSummary);
      expect(useFranchiserStore.getState().summary).toEqual(newSummary);
    });

    it('replaces entire summary object', () => {
      useFranchiserStore.getState().setSummary({ planPercent: 100, forecastPercent: 100, activeDealers: 0, avgConversion: 0, avgMargin: 0 });
      const summary = useFranchiserStore.getState().summary;
      expect(summary.planPercent).toBe(100);
      expect(summary.forecastPercent).toBe(100);
    });
  });

  describe('setDealers action', () => {
    it('sets dealers list', () => {
      const dealers: DealerMetrics[] = [
        { dealerId: 'd1', dealerName: 'Dealer 1', salonCount: 5, planPercent: 90, forecastPercent: 100, conversion: 12, margin: 30, status: 'green' },
        { dealerId: 'd2', dealerName: 'Dealer 2', salonCount: 3, planPercent: 60, forecastPercent: 70, conversion: 8, margin: 25, status: 'yellow' },
      ];

      useFranchiserStore.getState().setDealers(dealers);
      expect(useFranchiserStore.getState().dealers).toHaveLength(2);
      expect(useFranchiserStore.getState().dealers[0].dealerName).toBe('Dealer 1');
    });

    it('replaces existing dealers', () => {
      useFranchiserStore.getState().setDealers([{ dealerId: 'd1', dealerName: 'Old', salonCount: 1, planPercent: 50, forecastPercent: 50, conversion: 5, margin: 20, status: 'red' }]);
      useFranchiserStore.getState().setDealers([]);
      expect(useFranchiserStore.getState().dealers).toHaveLength(0);
    });
  });

  describe('setAlerts action', () => {
    it('sets alerts and calculates alertCount for critical', () => {
      const alerts: FranchiseAlert[] = [
        { id: 'a1', type: 'critical', message: 'Critical alert', createdAt: '2026-01-01' },
        { id: 'a2', type: 'warning', message: 'Warning alert', createdAt: '2026-01-02' },
        { id: 'a3', type: 'critical', message: 'Another critical', createdAt: '2026-01-03' },
      ];

      useFranchiserStore.getState().setAlerts(alerts);
      expect(useFranchiserStore.getState().alerts).toHaveLength(3);
      expect(useFranchiserStore.getState().alertCount).toBe(2);
    });

    it('handles empty alerts', () => {
      useFranchiserStore.getState().setAlerts([]);
      expect(useFranchiserStore.getState().alerts).toEqual([]);
      expect(useFranchiserStore.getState().alertCount).toBe(0);
    });
  });

  describe('setAlertCount action', () => {
    it('sets alert count directly', () => {
      useFranchiserStore.getState().setAlertCount(5);
      expect(useFranchiserStore.getState().alertCount).toBe(5);
    });
  });

  describe('setActiveTab action', () => {
    it('changes active tab', () => {
      useFranchiserStore.getState().setActiveTab('health');
      expect(useFranchiserStore.getState().activeTab).toBe('health');
    });

    it('changes to team tab', () => {
      useFranchiserStore.getState().setActiveTab('team');
      expect(useFranchiserStore.getState().activeTab).toBe('team');
    });
  });

  describe('setLoading action', () => {
    it('sets loading to true', () => {
      useFranchiserStore.getState().setLoading(true);
      expect(useFranchiserStore.getState().isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      useFranchiserStore.setState({ isLoading: true });
      useFranchiserStore.getState().setLoading(false);
      expect(useFranchiserStore.getState().isLoading).toBe(false);
    });
  });

  describe('fetchSummary thunk', () => {
    const mockSummaryData = {
      planPercent: 85,
      forecastPercent: 92,
      activeDealers: 28,
      avgConversion: 14,
      avgMargin: 34,
    };

    it('sets loading = true during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));

      useFranchiserStore.getState().fetchSummary();
      expect(useFranchiserStore.getState().isLoading).toBe(true);
    });

    it('fetches and sets summary on success', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockSummaryData });

      await useFranchiserStore.getState().fetchSummary();

      expect(useFranchiserStore.getState().summary.planPercent).toBe(85);
      expect(useFranchiserStore.getState().summary.forecastPercent).toBe(92);
      expect(useFranchiserStore.getState().isLoading).toBe(false);
    });

    it('handles partial data with defaults', async () => {
      mockApiClient.get.mockResolvedValue({ data: { planPercent: 50 } });

      await useFranchiserStore.getState().fetchSummary();

      const summary = useFranchiserStore.getState().summary;
      expect(summary.planPercent).toBe(50);
      expect(summary.forecastPercent).toBe(0);
      expect(summary.activeDealers).toBe(0);
    });

    it('sets loading = false on error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));

      await useFranchiserStore.getState().fetchSummary();

      expect(useFranchiserStore.getState().isLoading).toBe(false);
    });

    it('preserves default summary on API error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API error'));

      await useFranchiserStore.getState().fetchSummary();

      expect(useFranchiserStore.getState().summary.planPercent).toBe(78);
    });
  });

  describe('fetchNetwork thunk', () => {
    const mockNetworkData = {
      network_data: [
        { id: 'd1', name: 'Dealer One', salon_count: 5, plan_percent: 90, conversion: 12, margin: 30, forecast: 'green', plan: 100000, fact: 95000 },
        { id: 'd2', name: 'Dealer Two', salon_count: 3, plan_percent: 60, conversion: 8, margin: 25, forecast: 'yellow', plan: 50000, fact: 30000 },
        { id: 'd3', name: 'Dealer Three', salon_count: 2, plan_percent: 40, conversion: 5, margin: 20, forecast: 'red', plan: 30000, fact: 12000 },
      ],
    };

    it('fetches and maps network data correctly', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockNetworkData });

      await useFranchiserStore.getState().fetchNetwork();

      const dealers = useFranchiserStore.getState().dealers;
      expect(dealers).toHaveLength(3);
      expect(dealers[0].dealerName).toBe('Dealer One');
      expect(dealers[0].status).toBe('green');
      expect(dealers[1].status).toBe('yellow');
      expect(dealers[2].status).toBe('red');
    });

    it('maps forecastPercent correctly for different plan percentages', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockNetworkData });

      await useFranchiserStore.getState().fetchNetwork();

      const dealers = useFranchiserStore.getState().dealers;
      expect(dealers[0].forecastPercent).toBe(100);
      expect(dealers[1].forecastPercent).toBe(70);
      expect(dealers[2].forecastPercent).toBe(30);
    });

    it('handles empty network data', async () => {
      mockApiClient.get.mockResolvedValue({ data: { network_data: [] } });

      await useFranchiserStore.getState().fetchNetwork();

      expect(useFranchiserStore.getState().dealers).toEqual([]);
    });

    it('handles missing network_data', async () => {
      mockApiClient.get.mockResolvedValue({ data: {} });

      await useFranchiserStore.getState().fetchNetwork();

      expect(useFranchiserStore.getState().dealers).toEqual([]);
    });

    it('maps all dealer fields correctly', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockNetworkData });

      await useFranchiserStore.getState().fetchNetwork();

      const dealer = useFranchiserStore.getState().dealers[0];
      expect(dealer.dealerId).toBe('d1');
      expect(dealer.salonCount).toBe(5);
      expect(dealer.planPercent).toBe(90);
      expect(dealer.conversion).toBe(12);
      expect(dealer.margin).toBe(30);
      expect(dealer.plan).toBe(100000);
      expect(dealer.fact).toBe(95000);
    });
  });

  describe('fetchHealth thunk', () => {
    it('returns health data on success', async () => {
      const healthData = { status: 'healthy', issues: [] };
      mockApiClient.get.mockResolvedValue({ data: healthData });

      const result = await useFranchiserStore.getState().fetchHealth();

      expect(result).toEqual(healthData);
    });

    it('returns null on error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));

      const result = await useFranchiserStore.getState().fetchHealth();

      expect(result).toBeNull();
    });
  });

  describe('fetchTeam thunk', () => {
    it('returns team data on success', async () => {
      const teamData = { managers: [{ id: 'm1', name: 'Manager 1' }] };
      mockApiClient.get.mockResolvedValue({ data: teamData });

      const result = await useFranchiserStore.getState().fetchTeam();

      expect(result).toEqual(teamData);
    });

    it('returns null on error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));

      const result = await useFranchiserStore.getState().fetchTeam();

      expect(result).toBeNull();
    });
  });

  describe('selectors (computed values)', () => {
    it('calculates total active dealers from summary', () => {
      const state = useFranchiserStore.getState();
      expect(state.summary.activeDealers).toBe(24);
    });

    it('calculates average conversion from summary', () => {
      const state = useFranchiserStore.getState();
      expect(state.summary.avgConversion).toBe(12);
    });

    it('calculates total revenue from dealers', () => {
      useFranchiserStore.getState().setDealers([
        { dealerId: 'd1', dealerName: 'D1', salonCount: 1, planPercent: 100, forecastPercent: 100, conversion: 10, margin: 30, status: 'green', plan: 100000, fact: 100000 },
        { dealerId: 'd2', dealerName: 'D2', salonCount: 1, planPercent: 80, forecastPercent: 80, conversion: 8, margin: 25, status: 'green', plan: 50000, fact: 40000 },
      ]);

      const totalRevenue = useFranchiserStore.getState().dealers.reduce((sum, d) => sum + (d.fact || 0), 0);
      expect(totalRevenue).toBe(140000);
    });

    it('counts green status dealers', () => {
      useFranchiserStore.getState().setDealers([
        { dealerId: 'd1', dealerName: 'D1', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 10, margin: 30, status: 'green' },
        { dealerId: 'd2', dealerName: 'D2', salonCount: 1, planPercent: 60, forecastPercent: 70, conversion: 8, margin: 25, status: 'yellow' },
        { dealerId: 'd3', dealerName: 'D3', salonCount: 1, planPercent: 40, forecastPercent: 30, conversion: 5, margin: 20, status: 'red' },
      ]);

      const greenDealers = useFranchiserStore.getState().dealers.filter(d => d.status === 'green');
      expect(greenDealers).toHaveLength(1);
    });
  });

  describe('integration scenarios', () => {
    it('loads full franchise data', async () => {
      mockApiClient.get
        .mockResolvedValueOnce({ data: { planPercent: 88, forecastPercent: 90, activeDealers: 25, avgConversion: 13, avgMargin: 33 } })
        .mockResolvedValueOnce({ data: { network_data: [{ id: 'd1', name: 'Test', salon_count: 1, plan_percent: 100, conversion: 10, margin: 30, forecast: 'green' }] } });

      await useFranchiserStore.getState().fetchSummary();
      await useFranchiserStore.getState().fetchNetwork();

      expect(useFranchiserStore.getState().summary.planPercent).toBe(88);
      expect(useFranchiserStore.getState().dealers).toHaveLength(1);
    });

    it('handles error gracefully and keeps previous state', async () => {
      useFranchiserStore.getState().setSummary({ planPercent: 80, forecastPercent: 85, activeDealers: 20, avgConversion: 10, avgMargin: 28 });

      mockApiClient.get.mockRejectedValue(new Error('API Error'));

      await useFranchiserStore.getState().fetchSummary();

      expect(useFranchiserStore.getState().summary.planPercent).toBe(80);
    });
  });
});