import { useFranchiserStore, FranchiserSummary } from '@/store/franchiserStore';

describe('franchiserStore', () => {
  const initialState = useFranchiserStore.getState();

  it('has initial summary structure', () => {
    expect(initialState.summary).toBeDefined();
    expect(initialState.summary).toHaveProperty('planPercent');
    expect(initialState.summary).toHaveProperty('forecastPercent');
    expect(initialState.summary).toHaveProperty('activeDealers');
  });

  it('has initial dealers array', () => {
    expect(initialState.dealers).toBeDefined();
    expect(Array.isArray(initialState.dealers)).toBe(true);
  });

  it('has initial alerts array', () => {
    expect(initialState.alerts).toBeDefined();
    expect(Array.isArray(initialState.alerts)).toBe(true);
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

  it('has fetchSummary method', () => {
    expect(typeof useFranchiserStore.getState().fetchSummary).toBe('function');
  });

  it('has fetchNetwork method', () => {
    expect(typeof useFranchiserStore.getState().fetchNetwork).toBe('function');
  });

  it('has setLoading', () => {
    useFranchiserStore.getState().setLoading(true);
    expect(useFranchiserStore.getState().isLoading).toBe(true);
    useFranchiserStore.getState().setLoading(false);
    expect(useFranchiserStore.getState().isLoading).toBe(false);
  });
});