import { useActivityStore } from '@/store/activityStore';

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
    it('has overview as null', () => {
      expect(useActivityStore.getState().overview).toBeNull();
    });

    it('has empty dynamics array', () => {
      expect(useActivityStore.getState().dynamics).toEqual([]);
    });

    it('has empty byTenant array', () => {
      expect(useActivityStore.getState().byTenant).toEqual([]);
    });

    it('has empty featureAdoption array', () => {
      expect(useActivityStore.getState().featureAdoption).toEqual([]);
    });

    it('has empty ttv array', () => {
      expect(useActivityStore.getState().ttv).toEqual([]);
    });

    it('has default period of 30', () => {
      expect(useActivityStore.getState().period).toBe(30);
    });

    it('has empty selectedTenant', () => {
      expect(useActivityStore.getState().selectedTenant).toBe('');
    });

    it('has isLoading as false', () => {
      expect(useActivityStore.getState().isLoading).toBe(false);
    });
  });

  describe('setOverview', () => {
    it('sets activity overview', () => {
      const overview = {
        dau: 1500,
        dauChange: 5.2,
        wau: 8000,
        wauChange: 3.1,
        mau: 25000,
        mauChange: 2.5,
        stickiness: 6.0,
        avgSessionTime: 245,
        activeSessions: 42,
      };
      
      useActivityStore.getState().setOverview(overview);
      expect(useActivityStore.getState().overview).toEqual(overview);
    });
  });

  describe('setDynamics', () => {
    it('sets activity dynamics', () => {
      const dynamics = [
        { date: '2026-01-01', dau: 1400, wau: 7800, mau: 24000 },
        { date: '2026-01-02', dau: 1500, wau: 8000, mau: 25000 },
      ];
      
      useActivityStore.getState().setDynamics(dynamics);
      expect(useActivityStore.getState().dynamics).toHaveLength(2);
    });

    it('replaces existing dynamics', () => {
      useActivityStore.getState().setDynamics([{ date: '2026-01-01', dau: 1000, wau: 5000, mau: 20000 }]);
      useActivityStore.getState().setDynamics([]);
      expect(useActivityStore.getState().dynamics).toHaveLength(0);
    });
  });

  describe('setByTenant', () => {
    it('sets tenant activity list', () => {
      const byTenant = [
        { id: 't1', name: 'Tenant 1', dau: 100, wau: 500, mau: 2000, stickiness: 5.0, activeLicensePercent: 85 },
        { id: 't2', name: 'Tenant 2', dau: 150, wau: 600, mau: 2500, stickiness: 6.0, activeLicensePercent: 90 },
      ];
      
      useActivityStore.getState().setByTenant(byTenant);
      expect(useActivityStore.getState().byTenant).toHaveLength(2);
    });
  });

  describe('setFeatureAdoption', () => {
    it('sets feature adoption list', () => {
      const featureAdoption = [
        { id: 'f1', name: 'Feature 1', tenantPercent: 75, userPercent: 45, frequency: 'daily' as const, trend: 'up' as const },
        { id: 'f2', name: 'Feature 2', tenantPercent: 60, userPercent: 30, frequency: 'weekly' as const, trend: 'stable' as const },
      ];
      
      useActivityStore.getState().setFeatureAdoption(featureAdoption);
      expect(useActivityStore.getState().featureAdoption).toHaveLength(2);
    });
  });

  describe('setTtv', () => {
    it('sets time to value list', () => {
      const ttv = [
        { id: 't1', name: 'Client 1', createdAt: '2026-01-01', firstSaleAt: '2026-01-15', ttvDays: 14 },
        { id: 't2', name: 'Client 2', createdAt: '2026-01-10', firstSaleAt: '2026-01-30', ttvDays: 20 },
      ];
      
      useActivityStore.getState().setTtv(ttv);
      expect(useActivityStore.getState().ttv).toHaveLength(2);
    });
  });

  describe('setPeriod', () => {
    it('sets period to 7 days', () => {
      useActivityStore.getState().setPeriod(7);
      expect(useActivityStore.getState().period).toBe(7);
    });

    it('sets period to 90 days', () => {
      useActivityStore.getState().setPeriod(90);
      expect(useActivityStore.getState().period).toBe(90);
    });
  });

  describe('setSelectedTenant', () => {
    it('sets selected tenant ID', () => {
      useActivityStore.getState().setSelectedTenant('tenant-123');
      expect(useActivityStore.getState().selectedTenant).toBe('tenant-123');
    });

    it('clears selected tenant', () => {
      useActivityStore.setState({ selectedTenant: 'tenant-123' });
      useActivityStore.getState().setSelectedTenant('');
      expect(useActivityStore.getState().selectedTenant).toBe('');
    });
  });

  describe('setLoading', () => {
    it('sets loading to true', () => {
      useActivityStore.getState().setLoading(true);
      expect(useActivityStore.getState().isLoading).toBe(true);
    });

    it('sets loading to false', () => {
      useActivityStore.setState({ isLoading: true });
      useActivityStore.getState().setLoading(false);
      expect(useActivityStore.getState().isLoading).toBe(false);
    });
  });
});