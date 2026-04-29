import { useFranchiserStore } from '@/store/franchiserStore';

describe('franchiserStore', () => {
  const initialState = useFranchiserStore.getState();

  it('has initial summary', () => {
    expect(initialState.summary).toBeDefined();
    expect(initialState.summary.planPercent).toBe(78);
    expect(initialState.summary.forecastPercent).toBe(85);
    expect(initialState.summary.activeDealers).toBe(24);
  });

  it('has initial dealers', () => {
    expect(initialState.dealers).toBeDefined();
    expect(initialState.dealers.length).toBeGreaterThan(0);
  });

  it('has initial alerts', () => {
    expect(initialState.alerts).toBeDefined();
    expect(initialState.alertCount).toBeGreaterThanOrEqual(1);
  });

  it('updates summary', () => {
    const newSummary: FranchiserSummary = {
      planPercent: 90,
      forecastPercent: 95,
      activeDealers: 25,
      avgConversion: 14,
      avgMargin: 35,
    };
    useFranchiserStore.getState().setSummary(newSummary);
    expect(useFranchiserStore.getState().summary.planPercent).toBe(90);
  });

  it('updates alert count', () => {
    useFranchiserStore.getState().setAlertCount(5);
    expect(useFranchiserStore.getState().alertCount).toBe(5);
  });

  it('fetches summary', async () => {
    const before = useFranchiserStore.getState().isLoading;
    const fetchPromise = useFranchiserStore.getState().fetchSummary();
    expect(useFranchiserStore.getState().isLoading).toBe(true);
    await fetchPromise;
    expect(useFranchiserStore.getState().isLoading).toBe(false);
  });

  it('has setLoading', () => {
    useFranchiserStore.getState().setLoading(true);
    expect(useFranchiserStore.getState().isLoading).toBe(true);
    useFranchiserStore.getState().setLoading(false);
    expect(useFranchiserStore.getState().isLoading).toBe(false);
  });
});