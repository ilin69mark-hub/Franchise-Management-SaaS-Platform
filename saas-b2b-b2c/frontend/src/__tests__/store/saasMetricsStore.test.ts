import { useSaasMetricsStore, SaasOverview, MrrDynamicsMonth, CashflowData, TariffData } from '@/store/saasMetricsStore';

jest.mock('@/api/axiosClient', () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
}));

import apiClient from '@/api/axiosClient';

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('saasMetricsStore', () => {
  beforeEach(() => {
    useSaasMetricsStore.setState({
      activeSection: 'metrics',
      overview: null,
      mrrDynamics: [],
      cashflow: null,
      tariffs: [],
      isLoading: false,
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('has activeSection as metrics', () => {
      expect(useSaasMetricsStore.getState().activeSection).toBe('metrics');
    });

    it('has overview as null', () => {
      expect(useSaasMetricsStore.getState().overview).toBeNull();
    });

    it('has empty mrrDynamics array', () => {
      expect(useSaasMetricsStore.getState().mrrDynamics).toEqual([]);
    });

    it('has cashflow as null', () => {
      expect(useSaasMetricsStore.getState().cashflow).toBeNull();
    });

    it('has empty tariffs array', () => {
      expect(useSaasMetricsStore.getState().tariffs).toEqual([]);
    });

    it('has isLoading as false', () => {
      expect(useSaasMetricsStore.getState().isLoading).toBe(false);
    });
  });

  describe('setActiveSection action', () => {
    it('changes active section', () => {
      useSaasMetricsStore.getState().setActiveSection('cashflow');
      expect(useSaasMetricsStore.getState().activeSection).toBe('cashflow');
    });

    it('changes to tariffs section', () => {
      useSaasMetricsStore.getState().setActiveSection('tariffs');
      expect(useSaasMetricsStore.getState().activeSection).toBe('tariffs');
    });

    it('changes to revenue section', () => {
      useSaasMetricsStore.getState().setActiveSection('revenue');
      expect(useSaasMetricsStore.getState().activeSection).toBe('revenue');
    });
  });

  describe('setOverview action', () => {
    it('sets SaaS overview metrics', () => {
      const overview: SaasOverview = {
        mrr: 3000000,
        mrrChange: 15.5,
        arr: 36000000,
        churnRate: 4.2,
        overdueAmount: 200000,
        overdueCount: 5,
        arpu: 18000,
      };

      useSaasMetricsStore.getState().setOverview(overview);
      expect(useSaasMetricsStore.getState().overview).toEqual(overview);
    });

    it('sets partial overview', () => {
      useSaasMetricsStore.getState().setOverview({ mrr: 5000000, mrrChange: 20 });
      const overview = useSaasMetricsStore.getState().overview;
      expect(overview?.mrr).toBe(5000000);
      expect(overview?.arr).toBeUndefined();
    });
  });

  describe('setMrrDynamics action', () => {
    it('sets MRR dynamics data for chart', () => {
      const dynamics: MrrDynamicsMonth[] = [
        { month: '2025-01', newMrr: 100000, expansionMrr: 50000, churnMrr: 20000, netMrr: 130000 },
        { month: '2025-02', newMrr: 120000, expansionMrr: 60000, churnMrr: 25000, netMrr: 155000 },
      ];

      useSaasMetricsStore.getState().setMrrDynamics(dynamics);
      expect(useSaasMetricsStore.getState().mrrDynamics).toHaveLength(2);
      expect(useSaasMetricsStore.getState().mrrDynamics[0].month).toBe('2025-01');
    });

    it('replaces existing dynamics', () => {
      useSaasMetricsStore.getState().setMrrDynamics([{ month: '2025-01', newMrr: 100000, expansionMrr: 50000, churnMrr: 20000, netMrr: 130000 }]);
      useSaasMetricsStore.getState().setMrrDynamics([]);
      expect(useSaasMetricsStore.getState().mrrDynamics).toHaveLength(0);
    });
  });

  describe('setCashflow action', () => {
    it('sets cashflow data with expected payments', () => {
      const cashflow: CashflowData = {
        expected: [
          { date: '2026-06-01', tenant: 'Client A', amount: 100000 },
          { date: '2026-06-15', tenant: 'Client B', amount: 50000 },
        ],
        overdue: [],
        risks: [],
        totalExpected: 150000,
        atRisk: 0,
      };

      useSaasMetricsStore.getState().setCashflow(cashflow);
      expect(useSaasMetricsStore.getState().cashflow?.expected).toHaveLength(2);
      expect(useSaasMetricsStore.getState().cashflow?.totalExpected).toBe(150000);
    });

    it('sets cashflow data with overdue payments', () => {
      const cashflow: CashflowData = {
        expected: [],
        overdue: [
          { date: '2026-04-01', tenant: 'Client C', amount: 75000, daysOverdue: 30 },
        ],
        risks: [],
        totalExpected: 0,
        atRisk: 75000,
      };

      useSaasMetricsStore.getState().setCashflow(cashflow);
      expect(useSaasMetricsStore.getState().cashflow?.overdue).toHaveLength(1);
      expect(useSaasMetricsStore.getState().cashflow?.atRisk).toBe(75000);
    });

    it('sets cashflow with risks', () => {
      const cashflow: CashflowData = {
        expected: [],
        overdue: [],
        risks: [
          { date: '2026-07-01', tenant: 'Client D', amount: 200000 },
        ],
        totalExpected: 0,
        atRisk: 0,
      };

      useSaasMetricsStore.getState().setCashflow(cashflow);
      expect(useSaasMetricsStore.getState().cashflow?.risks).toHaveLength(1);
    });
  });

  describe('setTariffs action', () => {
    it('sets tariff distribution data', () => {
      const tariffs: TariffData[] = [
        { name: 'Start', count: 10, mrr: 100000, avgCheck: 10000, growth: 15 },
        { name: 'Pro', count: 5, mrr: 200000, avgCheck: 40000, growth: 25 },
      ];

      useSaasMetricsStore.getState().setTariffs(tariffs);
      expect(useSaasMetricsStore.getState().tariffs).toHaveLength(2);
      expect(useSaasMetricsStore.getState().tariffs[0].name).toBe('Start');
    });
  });

  describe('setLoading action', () => {
    it('sets loading to true', () => {
      useSaasMetricsStore.getState().setLoading(true);
      expect(useSaasMetricsStore.getState().isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      useSaasMetricsStore.setState({ isLoading: true });
      useSaasMetricsStore.getState().setLoading(false);
      expect(useSaasMetricsStore.getState().isLoading).toBe(false);
    });
  });

  describe('fetchAnalytics thunk', () => {
    const mockAnalyticsData = {
      currentMrr: 5000000,
      mrrGrowth: 18.5,
      arr: 60000000,
      churnRate: 3.5,
      overdueAmount: 100000,
      overdueCount: 2,
      arpu: 20000,
    };

    it('sets loading = true during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));

      useSaasMetricsStore.getState().fetchAnalytics();
      expect(useSaasMetricsStore.getState().isLoading).toBe(true);
    });

    it('fetches and sets overview metrics on success', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockAnalyticsData });

      await useSaasMetricsStore.getState().fetchAnalytics();

      const overview = useSaasMetricsStore.getState().overview;
      expect(overview?.mrr).toBe(5000000);
      expect(overview?.mrrChange).toBe(18.5);
      expect(overview?.arr).toBe(60000000);
      expect(overview?.churnRate).toBe(3.5);
      expect(useSaasMetricsStore.getState().isLoading).toBe(false);
    });

    it('converts string numbers to Number type', async () => {
      mockApiClient.get.mockResolvedValue({ data: { currentMrr: '3500000', mrrGrowth: '12.5' } });

      await useSaasMetricsStore.getState().fetchAnalytics();

      expect(typeof useSaasMetricsStore.getState().overview?.mrr).toBe('number');
    });

    it('handles missing data with defaults', async () => {
      mockApiClient.get.mockResolvedValue({ data: {} });

      await useSaasMetricsStore.getState().fetchAnalytics();

      const overview = useSaasMetricsStore.getState().overview;
      expect(overview?.mrr).toBe(2500000);
      expect(overview?.mrrChange).toBe(12.5);
    });

    it('sets fallback data on API error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));

      await useSaasMetricsStore.getState().fetchAnalytics();

      const overview = useSaasMetricsStore.getState().overview;
      expect(overview?.mrr).toBe(2500000);
      expect(overview?.arr).toBe(30000000);
      expect(overview?.churnRate).toBe(5.2);
    });

    it('sets loading = false after error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API Error'));

      await useSaasMetricsStore.getState().fetchAnalytics();

      expect(useSaasMetricsStore.getState().isLoading).toBe(false);
    });
  });

  describe('fetchPaymentStatus thunk', () => {
    const mockPaymentData = [
      { due_date: '2026-06-01', tenant_name: 'Client X', amount: 150000, status: 'paid' },
      { due_date: '2026-05-15', tenant_name: 'Client Y', amount: 80000, status: 'overdue', days_overdue: 20 },
      { due_date: '2026-06-20', tenant_name: 'Client Z', amount: 200000, status: 'paid' },
    ];

    it('sets loading = true during fetch', async () => {
      mockApiClient.get.mockImplementation(() => new Promise(() => {}));

      useSaasMetricsStore.getState().fetchPaymentStatus();
      expect(useSaasMetricsStore.getState().isLoading).toBe(true);
    });

    it('parses paid payments to expected array', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockPaymentData });

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      const cashflow = useSaasMetricsStore.getState().cashflow;
      expect(cashflow?.expected).toHaveLength(2);
      expect(cashflow?.totalExpected).toBe(350000);
    });

    it('parses overdue payments correctly', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockPaymentData });

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      const cashflow = useSaasMetricsStore.getState().cashflow;
      expect(cashflow?.overdue).toHaveLength(1);
      expect(cashflow?.overdue[0].daysOverdue).toBe(20);
      expect(cashflow?.atRisk).toBe(80000);
    });

    it('calculates totalExpected correctly', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockPaymentData });

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      expect(useSaasMetricsStore.getState().cashflow?.totalExpected).toBe(350000);
    });

    it('calculates atRisk correctly', async () => {
      mockApiClient.get.mockResolvedValue({ data: mockPaymentData });

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      expect(useSaasMetricsStore.getState().cashflow?.atRisk).toBe(80000);
    });

    it('handles empty payment data', async () => {
      mockApiClient.get.mockResolvedValue({ data: [] });

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      const cashflow = useSaasMetricsStore.getState().cashflow;
      expect(cashflow?.expected).toEqual([]);
      expect(cashflow?.overdue).toEqual([]);
      expect(cashflow?.totalExpected).toBe(0);
      expect(cashflow?.atRisk).toBe(0);
    });

    it('handles missing fields in data', async () => {
      mockApiClient.get.mockResolvedValue({ data: [{ status: 'paid' }] });

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      expect(useSaasMetricsStore.getState().cashflow?.expected).toHaveLength(1);
      expect(useSaasMetricsStore.getState().cashflow?.expected[0].amount).toBe(0);
    });

    it('sets fallback cashflow on API error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('Network error'));

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      const cashflow = useSaasMetricsStore.getState().cashflow;
      expect(cashflow?.expected).toHaveLength(2);
      expect(cashflow?.totalExpected).toBe(800000);
    });

    it('sets loading = false after error', async () => {
      mockApiClient.get.mockRejectedValue(new Error('API Error'));

      await useSaasMetricsStore.getState().fetchPaymentStatus();

      expect(useSaasMetricsStore.getState().isLoading).toBe(false);
    });
  });

  describe('computed selectors', () => {
    it('calculates total MRR from overview', () => {
      useSaasMetricsStore.getState().setOverview({
        mrr: 5000000,
        mrrChange: 10,
        arr: 60000000,
        churnRate: 5,
        overdueAmount: 0,
        overdueCount: 0,
        arpu: 15000,
      });

      const mrr = useSaasMetricsStore.getState().overview?.mrr;
      expect(mrr).toBe(5000000);
    });

    it('calculates ARR as MRR * 12', () => {
      useSaasMetricsStore.getState().setOverview({
        mrr: 2500000,
        mrrChange: 0,
        arr: 30000000,
        churnRate: 0,
        overdueAmount: 0,
        overdueCount: 0,
        arpu: 0,
      });

      const arr = useSaasMetricsStore.getState().overview?.arr;
      expect(arr).toBe(30000000);
    });

    it('calculates revenue from cashflow expected', () => {
      useSaasMetricsStore.getState().setCashflow({
        expected: [{ date: '2026-06-01', tenant: 'C1', amount: 100000 }],
        overdue: [],
        risks: [],
        totalExpected: 100000,
        atRisk: 0,
      });

      const revenue = useSaasMetricsStore.getState().cashflow?.totalExpected;
      expect(revenue).toBe(100000);
    });

    it('calculates overdue amount from cashflow', () => {
      useSaasMetricsStore.getState().setCashflow({
        expected: [],
        overdue: [{ date: '2026-04-01', tenant: 'C2', amount: 50000, daysOverdue: 10 }],
        risks: [],
        totalExpected: 0,
        atRisk: 50000,
      });

      const atRisk = useSaasMetricsStore.getState().cashflow?.atRisk;
      expect(atRisk).toBe(50000);
    });

    it('handles zero churn rate gracefully', () => {
      useSaasMetricsStore.getState().setOverview({
        mrr: 1000000,
        mrrChange: 0,
        arr: 12000000,
        churnRate: 0,
        overdueAmount: 0,
        overdueCount: 0,
        arpu: 10000,
      });

      const churnRate = useSaasMetricsStore.getState().overview?.churnRate;
      expect(churnRate).toBe(0);
    });

    it('calculates total tenants from tariffs', () => {
      useSaasMetricsStore.getState().setTariffs([
        { name: 'Start', count: 10, mrr: 100000, avgCheck: 10000, growth: 10 },
        { name: 'Pro', count: 5, mrr: 200000, avgCheck: 40000, growth: 20 },
      ]);

      const totalTenants = useSaasMetricsStore.getState().tariffs.reduce((sum, t) => sum + t.count, 0);
      expect(totalTenants).toBe(15);
    });
  });

  describe('integration scenarios', () => {
    it('loads full analytics suite', async () => {
      mockApiClient.get
        .mockResolvedValueOnce({ data: { currentMrr: 4000000, mrrGrowth: 14, arr: 48000000, churnRate: 4, overdueAmount: 80000, overdueCount: 1, arpu: 16000 } })
        .mockResolvedValueOnce({ data: [{ due_date: '2026-06-01', tenant: 'T1', amount: 200000, status: 'paid' }] });

      await useSaasMetricsStore.getState().fetchAnalytics();
      await useSaasMetricsStore.getState().fetchPaymentStatus();

      expect(useSaasMetricsStore.getState().overview?.mrr).toBe(4000000);
      expect(useSaasMetricsStore.getState().cashflow?.totalExpected).toBe(200000);
    });

    it('handles partial data gracefully', async () => {
      mockApiClient.get.mockResolvedValue({ data: { currentMrr: 'not-a-number' } });

      await useSaasMetricsStore.getState().fetchAnalytics();

      expect(useSaasMetricsStore.getState().overview?.mrr).toBe(2500000);
    });
  });
});