import { useSuperAdminStore } from '@/store/superAdminStore';

describe('superAdminStore', () => {
  it('should have default values', () => {
    const store = useSuperAdminStore.getState();
    expect(store.activeSection).toBe('metrics');
    expect(store.isLoading).toBe(false);
    expect(store.alertCount).toBe(0);
  });

  it('should set active section', () => {
    const { setActiveSection } = useSuperAdminStore.getState();
    setActiveSection('tenants');
    expect(useSuperAdminStore.getState().activeSection).toBe('tenants');
  });

  it('should set loading state', () => {
    const { setLoading } = useSuperAdminStore.getState();
    setLoading(true);
    expect(useSuperAdminStore.getState().isLoading).toBe(true);
  });

  it('should set stats', () => {
    const { setStats } = useSuperAdminStore.getState();
    setStats({
      mrr: 1000000,
      arr: 12000000,
      churnRate: 3.5,
      overduePayments: 2,
      arpu: 10000,
    });
    const store = useSuperAdminStore.getState();
    expect(store.stats.mrr).toBe(1000000);
    expect(store.stats.arr).toBe(12000000);
    expect(store.stats.churnRate).toBe(3.5);
    expect(store.stats.overduePayments).toBe(2);
    expect(store.stats.arpu).toBe(10000);
  });

  it('should set alert count', () => {
    const { setAlertCount } = useSuperAdminStore.getState();
    setAlertCount(5);
    expect(useSuperAdminStore.getState().alertCount).toBe(5);
  });

  it('should set last updated', () => {
    const now = new Date();
    const { setLastUpdated } = useSuperAdminStore.getState();
    setLastUpdated(now);
    expect(useSuperAdminStore.getState().lastUpdated).toEqual(now);
  });
});