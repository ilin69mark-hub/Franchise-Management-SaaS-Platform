// src/__tests__/components/Dashboard/territoryManagerStore.test.ts
import { useTerritoryManagerStore, TerritorySummary, DealerMetrics } from '@/store/territoryManagerStore';

describe('useTerritoryManagerStore', () => {
  it('initial state', () => {
    const store = useTerritoryManagerStore.getState();
    expect(store.activeTab).toBe('map');
    expect(store.summary).toBeNull();
    expect(store.dealers).toEqual([]);
    expect(store.isLoading).toBe(false);
    expect(store.summaryModalOpen).toBe(false);
  });

  it('setActiveTab', () => {
    const { setActiveTab } = useTerritoryManagerStore.getState();
    setActiveTab('funnel');
    expect(useTerritoryManagerStore.getState().activeTab).toBe('funnel');
  });

  it('setSummary', () => {
    const summary: TerritorySummary = {
      planCompletionPercent: 85,
      quarterForecastPercent: 92,
      redZoneDealersCount: 2,
      avgConversion: 4.1,
      activeAlerts: 3,
    };
    const { setSummary } = useTerritoryManagerStore.getState();
    setSummary(summary);
    expect(useTerritoryManagerStore.getState().summary).toEqual(summary);
  });

  it('setDealers', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'Test', salonCount: 2, planPercent: 90, forecastPercent: 100, conversion: 4.5, status: 'green' },
    ];
    const { setDealers } = useTerritoryManagerStore.getState();
    setDealers(dealers);
    expect(useTerritoryManagerStore.getState().dealers).toEqual(dealers);
  });

  it('setSummaryModalOpen', () => {
    const { setSummaryModalOpen } = useTerritoryManagerStore.getState();
    setSummaryModalOpen(true);
    expect(useTerritoryManagerStore.getState().summaryModalOpen).toBe(true);
  });

  it('calculates red zone count', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 60, forecastPercent: 70, conversion: 2, status: 'red' },
      { dealerId: '3', dealerName: 'C', salonCount: 1, planPercent: 50, forecastPercent: 60, conversion: 1.5, status: 'red' },
    ];
    const redCount = dealers.filter(d => d.status === 'red').length;
    expect(redCount).toBe(2);
  });

  it('calculates average conversion', () => {
    const dealers: DealerMetrics[] = [
      { dealerId: '1', dealerName: 'A', salonCount: 1, planPercent: 90, forecastPercent: 100, conversion: 4, status: 'green' },
      { dealerId: '2', dealerName: 'B', salonCount: 1, planPercent: 80, forecastPercent: 90, conversion: 3, status: 'yellow' },
    ];
    const avg = dealers.reduce((s, d) => s + d.conversion, 0) / dealers.length;
    expect(avg).toBe(3.5);
  });
});