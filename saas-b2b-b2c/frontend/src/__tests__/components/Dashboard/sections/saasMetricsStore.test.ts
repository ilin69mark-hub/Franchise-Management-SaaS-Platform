import { useSaasMetricsStore } from '@/store/saasMetricsStore';

describe('saasMetricsStore', () => {
  it('should have default values', () => {
    const store = useSaasMetricsStore.getState();
    expect(store.activeSection).toBe('metrics');
    expect(store.isLoading).toBe(false);
  });

  it('should set loading state', () => {
    const { setLoading } = useSaasMetricsStore.getState();
    setLoading(true);
    expect(useSaasMetricsStore.getState().isLoading).toBe(true);
  });

  it('should set overview data', () => {
    const { setOverview } = useSaasMetricsStore.getState();
    setOverview({
      mrr: 2500000,
      mrrChange: 12.5,
      arr: 30000000,
      churnRate: 5.2,
      overdueAmount: 150000,
      overdueCount: 3,
      arpu: 15000,
    });
    const store = useSaasMetricsStore.getState();
    expect(store.overview?.mrr).toBe(2500000);
    expect(store.overview?.mrrChange).toBe(12.5);
    expect(store.overview?.churnRate).toBe(5.2);
  });

  it('should set mrr dynamics', () => {
    const { setMrrDynamics } = useSaasMetricsStore.getState();
    const dynamics = [
      { month: '2026-01', newMrr: 100000, expansionMrr: 50000, churnMrr: 20000, netMrr: 2300000 },
      { month: '2026-02', newMrr: 120000, expansionMrr: 60000, churnMrr: 15000, netMrr: 2460000 },
    ];
    setMrrDynamics(dynamics);
    expect(useSaasMetricsStore.getState().mrrDynamics).toEqual(dynamics);
  });

  it('should set cashflow data', () => {
    const { setCashflow } = useSaasMetricsStore.getState();
    const cashflow = {
      expected: [
        { date: '2026-04-30', tenant: 'ООО Тест', amount: 50000 },
      ],
      overdue: [
        { date: '2026-04-01', tenant: 'ООО Должник', amount: 25000, daysOverdue: 15 },
      ],
      risks: [],
      totalExpected: 50000,
      atRisk: 25000,
    };
    setCashflow(cashflow);
    expect(useSaasMetricsStore.getState().cashflow?.expected.length).toBe(1);
  });

  it('should set tariff data', () => {
    const { setTariffs } = useSaasMetricsStore.getState();
    const tariffs = [
      { name: 'Start', count: 10, mrr: 100000, avgCheck: 10000, growth: 15 },
      { name: 'Pro', count: 5, mrr: 150000, avgCheck: 30000, growth: -5 },
    ];
    setTariffs(tariffs);
    expect(useSaasMetricsStore.getState().tariffs?.length).toBe(2);
  });
});