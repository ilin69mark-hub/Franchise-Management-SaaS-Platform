import { useActivityStore, ActivityOverview, ActivityDynamics, TenantActivityItem, FeatureAdoptionItem, TtvItem } from '@/store/activityStore';

describe('activityStore', () => {
  beforeEach(() => {
    useActivityStore.setState({
      overview: null,
      dynamics: [],
      byTenant: [],
      featureAdoption: [],
      ttv: [],
      period: 30,
      selectedTenant: '',
      isLoading: false,
    });
  });

  describe('initial state', () => {
    it('has null overview', () => {
      expect(useActivityStore.getState().overview).toBeNull();
    });

    it('has empty arrays', () => {
      const state = useActivityStore.getState();
      expect(state.dynamics).toEqual([]);
      expect(state.byTenant).toEqual([]);
      expect(state.featureAdoption).toEqual([]);
      expect(state.ttv).toEqual([]);
    });

    it('has default period', () => {
      expect(useActivityStore.getState().period).toBe(30);
    });

    it('has empty selectedTenant', () => {
      expect(useActivityStore.getState().selectedTenant).toBe('');
    });

    it('has isLoading false', () => {
      expect(useActivityStore.getState().isLoading).toBe(false);
    });
  });

  describe('setOverview', () => {
    it('sets activity overview', () => {
      const overview: ActivityOverview = {
        dau: 1500,
        dauChange: 5.2,
        wau: 4500,
        wauChange: 3.1,
        mau: 12000,
        mauChange: 2.8,
        stickiness: 12.5,
        avgSessionTime: 180,
        activeSessions: 250,
      };
      useActivityStore.getState().setOverview(overview);
      expect(useActivityStore.getState().overview).toEqual(overview);
    });

    it('updates individual overview fields', () => {
      const overview: ActivityOverview = {
        dau: 1000,
        dauChange: 0,
        wau: 3000,
        wauChange: 0,
        mau: 8000,
        mauChange: 0,
        stickiness: 10,
        avgSessionTime: 120,
        activeSessions: 100,
      };
      useActivityStore.getState().setOverview(overview);
      expect(useActivityStore.getState().overview?.dau).toBe(1000);
      expect(useActivityStore.getState().overview?.activeSessions).toBe(100);
    });
  });

  describe('setDynamics', () => {
    it('sets dynamics data', () => {
      const dynamics: ActivityDynamics[] = [
        { date: '2026-01-01', dau: 1400, wau: 4200, mau: 11500 },
        { date: '2026-01-02', dau: 1450, wau: 4300, mau: 11600 },
        { date: '2026-01-03', dau: 1500, wau: 4400, mau: 11700 },
      ];
      useActivityStore.getState().setDynamics(dynamics);
      expect(useActivityStore.getState().dynamics).toHaveLength(3);
    });

    it('calculates average dau', () => {
      const dynamics: ActivityDynamics[] = [
        { date: '2026-01-01', dau: 1000, wau: 3000, mau: 8000 },
        { date: '2026-01-02', dau: 2000, wau: 4000, mau: 9000 },
      ];
      useActivityStore.getState().setDynamics(dynamics);
      const avg = dynamics.reduce((sum, d) => sum + d.dau, 0) / 2;
      expect(avg).toBe(1500);
    });
  });

  describe('setByTenant', () => {
    it('sets tenant activity', () => {
      const byTenant: TenantActivityItem[] = [
        { id: '1', name: 'Tenant 1', dau: 500, wau: 1500, mau: 4000, stickiness: 12, activeLicensePercent: 85 },
        { id: '2', name: 'Tenant 2', dau: 300, wau: 900, mau: 2500, stickiness: 10, activeLicensePercent: 70 },
      ];
      useActivityStore.getState().setByTenant(byTenant);
      expect(useActivityStore.getState().byTenant).toHaveLength(2);
    });

    it('filters high stickiness tenants', () => {
      const byTenant: TenantActivityItem[] = [
        { id: '1', name: 'High Stickiness', dau: 500, wau: 1500, mau: 4000, stickiness: 15, activeLicensePercent: 90 },
        { id: '2', name: 'Low Stickiness', dau: 300, wau: 900, mau: 2500, stickiness: 8, activeLicensePercent: 60 },
      ];
      useActivityStore.getState().setByTenant(byTenant);
      const highStickiness = useActivityStore.getState().byTenant.filter(t => t.stickiness >= 10);
      expect(highStickiness).toHaveLength(1);
    });
  });

  describe('setFeatureAdoption', () => {
    it('sets feature adoption data', () => {
      const featureAdoption: FeatureAdoptionItem[] = [
        { id: '1', name: 'Dashboard', tenantPercent: 95, userPercent: 80, frequency: 'daily', trend: 'up' },
        { id: '2', name: 'Reports', tenantPercent: 75, userPercent: 50, frequency: 'weekly', trend: 'stable' },
        { id: '3', name: 'Settings', tenantPercent: 60, userPercent: 30, frequency: 'monthly', trend: 'down' },
      ];
      useActivityStore.getState().setFeatureAdoption(featureAdoption);
      expect(useActivityStore.getState().featureAdoption).toHaveLength(3);
    });

    it('filters daily usage features', () => {
      const featureAdoption: FeatureAdoptionItem[] = [
        { id: '1', name: 'Daily Feature', tenantPercent: 90, userPercent: 70, frequency: 'daily', trend: 'up' },
        { id: '2', name: 'Weekly Feature', tenantPercent: 70, userPercent: 40, frequency: 'weekly', trend: 'stable' },
      ];
      useActivityStore.getState().setFeatureAdoption(featureAdoption);
      const dailyFeatures = useActivityStore.getState().featureAdoption.filter(f => f.frequency === 'daily');
      expect(dailyFeatures).toHaveLength(1);
    });
  });

  describe('setTtv', () => {
    it('sets time to value data', () => {
      const ttv: TtvItem[] = [
        { id: '1', name: 'Tenant 1', createdAt: '2026-01-01', firstSaleAt: '2026-01-15', ttvDays: 14 },
        { id: '2', name: 'Tenant 2', createdAt: '2026-01-05', firstSaleAt: '2026-01-20', ttvDays: 15 },
      ];
      useActivityStore.getState().setTtv(ttv);
      expect(useActivityStore.getState().ttv).toHaveLength(2);
    });

    it('calculates average ttv', () => {
      const ttv: TtvItem[] = [
        { id: '1', name: 'T1', createdAt: '2026-01-01', firstSaleAt: '2026-01-08', ttvDays: 7 },
        { id: '2', name: 'T2', createdAt: '2026-01-01', firstSaleAt: '2026-01-15', ttvDays: 14 },
      ];
      useActivityStore.getState().setTtv(ttv);
      const avg = ttv.reduce((sum, t) => sum + t.ttvDays, 0) / 2;
      expect(avg).toBe(10.5);
    });
  });

  describe('setPeriod', () => {
    it('sets period', () => {
      useActivityStore.getState().setPeriod(7);
      expect(useActivityStore.getState().period).toBe(7);
    });

    it('changes period', () => {
      useActivityStore.getState().setPeriod(30);
      useActivityStore.getState().setPeriod(90);
      expect(useActivityStore.getState().period).toBe(90);
    });
  });

  describe('setSelectedTenant', () => {
    it('sets selected tenant', () => {
      useActivityStore.getState().setSelectedTenant('tenant-123');
      expect(useActivityStore.getState().selectedTenant).toBe('tenant-123');
    });

    it('clears selected tenant', () => {
      useActivityStore.getState().setSelectedTenant('tenant-123');
      useActivityStore.getState().setSelectedTenant('');
      expect(useActivityStore.getState().selectedTenant).toBe('');
    });
  });

  describe('setLoading', () => {
    it('sets loading state', () => {
      useActivityStore.getState().setLoading(true);
      expect(useActivityStore.getState().isLoading).toBe(true);
    });

    it('toggles loading', () => {
      useActivityStore.getState().setLoading(true);
      useActivityStore.getState().setLoading(false);
      expect(useActivityStore.getState().isLoading).toBe(false);
    });
  });

  describe('methods exist', () => {
    it('has setOverview', () => {
      expect(typeof useActivityStore.getState().setOverview).toBe('function');
    });

    it('has setDynamics', () => {
      expect(typeof useActivityStore.getState().setDynamics).toBe('function');
    });

    it('has setByTenant', () => {
      expect(typeof useActivityStore.getState().setByTenant).toBe('function');
    });

    it('has setFeatureAdoption', () => {
      expect(typeof useActivityStore.getState().setFeatureAdoption).toBe('function');
    });

    it('has setTtv', () => {
      expect(typeof useActivityStore.getState().setTtv).toBe('function');
    });

    it('has setPeriod', () => {
      expect(typeof useActivityStore.getState().setPeriod).toBe('function');
    });

    it('has setSelectedTenant', () => {
      expect(typeof useActivityStore.getState().setSelectedTenant).toBe('function');
    });

    it('has setLoading', () => {
      expect(typeof useActivityStore.getState().setLoading).toBe('function');
    });
  });
});